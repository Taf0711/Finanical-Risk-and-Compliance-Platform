package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/websocket/v2"
	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/Taf0711/financial-risk-monitor/internal/config"
	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/handlers"
	"github.com/Taf0711/financial-risk-monitor/internal/marketdata"
	"github.com/Taf0711/financial-risk-monitor/internal/marketdata/providers"
	"github.com/Taf0711/financial-risk-monitor/internal/middleware"
	"github.com/Taf0711/financial-risk-monitor/internal/services"
	wsHandler "github.com/Taf0711/financial-risk-monitor/internal/websocket"
)

func main() {
	// Load environment variables from .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// Initialize database connections
	if err := database.InitPostgres(&cfg.Database); err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}

	if err := database.InitRedis(&cfg.Redis); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: cfg.App.Name,
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:3000,http://localhost:3001",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Initialize services
	authService := services.NewAuthService(&cfg.JWT)
	portfolioService := services.NewPortfolioService()

	// Initialize trading service with Alpaca
	tradingService := services.NewTradingService(
		os.Getenv("ALPACA_API_KEY"),
		os.Getenv("ALPACA_SECRET_KEY"),
		true, // Use paper trading by default
	)

	// Initialize market data service
	marketDataConfig := &marketdata.ServiceConfig{
		PrimaryProvider:   "alpaca",
		FallbackProviders: []string{}, // We'll add more providers later
		CacheTTL:          5 * time.Minute,
		RateLimits: map[string]marketdata.RateLimitConfig{
			"alpaca": {
				RequestsPerMinute: 200, // Alpaca free tier: 200 requests/minute
				BurstLimit:        20,
			},
		},
	}
	marketDataService := marketdata.NewService(marketDataConfig, database.GetRedis())

	// Register Alpaca provider with paper trading credentials
	alpacaAPIKey := os.Getenv("ALPACA_API_KEY")
	alpacaSecretKey := os.Getenv("ALPACA_SECRET_KEY")

	if alpacaAPIKey == "" || alpacaSecretKey == "" {
		log.Println("Warning: ALPACA_API_KEY and ALPACA_SECRET_KEY not provided. Using fallback data.")
		alpacaAPIKey = "demo"
		alpacaSecretKey = "demo"
	}

	if len(alpacaAPIKey) >= 8 {
		log.Printf("Initializing Alpaca provider with API key: %s...", alpacaAPIKey[:8])
	} else {
		log.Printf("Initializing Alpaca provider with demo credentials")
	}
	alpacaProvider := providers.NewAlpacaProvider(alpacaAPIKey, alpacaSecretKey)
	marketDataService.RegisterProvider("alpaca", alpacaProvider)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	portfolioHandler := handlers.NewPortfolioHandler()
	transactionHandler := handlers.NewTransactionHandler()
	riskHandler := handlers.NewRiskHandler(&cfg.Risk)
	alertHandler := handlers.NewAlertHandler()
	complianceHandler := handlers.NewComplianceHandler()
	monteCarloHandler := handlers.NewMonteCarloHandler(portfolioService)
	marketDataHandler := handlers.NewMarketDataHandler(marketDataService)
	tradingHandler := handlers.NewTradingHandler(tradingService)
	adminHandler := handlers.NewAdminHandler(marketDataService)

	// Initialize WebSocket hub
	hub := wsHandler.NewHub()
	go hub.Run()

	// Initialize simple WebSocket hub for Fiber WebSocket connections
	simpleHub := wsHandler.NewSimpleHub()
	go simpleHub.Run()

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "Financial Risk Monitor API",
		})
	})

	// Serve dashboard at root
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "Financial Risk Monitor API",
			"version": "1.0.0",
			"status":  "operational",
		})
	})

	// Alpaca test endpoints (completely public)
	app.Get("/alpaca/quote/:symbol", func(c *fiber.Ctx) error {
		symbol := c.Params("symbol")
		quote, err := marketDataService.GetRealtimeQuote(symbol)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(quote)
	})

	app.Get("/alpaca/quotes", func(c *fiber.Ctx) error {
		symbolsParam := c.Query("symbols", "AAPL,GOOGL,MSFT")
		symbols := strings.Split(symbolsParam, ",")
		quotes, err := marketDataService.GetMultipleQuotes(symbols)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(fiber.Map{"quotes": quotes, "count": len(quotes)})
	})

	app.Get("/alpaca/status", func(c *fiber.Ctx) error {
		status := marketDataService.GetProviderStatus()
		return c.JSON(fiber.Map{"providers": status, "timestamp": time.Now().Unix()})
	})

	app.Get("/alpaca/historical/:symbol", func(c *fiber.Ctx) error {
		symbol := c.Params("symbol")
		period := c.Query("period", "1mo")
		data, err := marketDataService.GetHistoricalData(symbol, period)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(data)
	})

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected routes
	protected := api.Group("/protected", middleware.JWTMiddleware(authService))

	// Portfolio routes
	portfolios := protected.Group("/portfolios")
	portfolios.Get("/", portfolioHandler.GetPortfolios)
	portfolios.Get("/:id", portfolioHandler.GetPortfolio)
	portfolios.Post("/", portfolioHandler.CreatePortfolio)
	portfolios.Put("/:id", portfolioHandler.UpdatePortfolio)
	portfolios.Delete("/:id", portfolioHandler.DeletePortfolio)

	// Position routes
	portfolios.Get("/:id/positions", portfolioHandler.GetPositions)
	portfolios.Post("/:id/positions", portfolioHandler.AddPosition)
	portfolios.Put("/:id/positions/:positionId", portfolioHandler.UpdatePosition)
	portfolios.Delete("/:id/positions/:positionId", portfolioHandler.DeletePosition)

	// Transaction routes
	transactions := protected.Group("/transactions")
	transactions.Get("/", transactionHandler.GetTransactions)
	transactions.Get("/:id", transactionHandler.GetTransaction)
	transactions.Post("/", transactionHandler.CreateTransaction)
	transactions.Put("/:id", transactionHandler.UpdateTransaction)
	transactions.Put("/:id/status", transactionHandler.UpdateTransactionStatus)
	transactions.Delete("/:id", transactionHandler.DeleteTransaction)

	// Risk metrics routes
	risk := protected.Group("/risk")
	risk.Get("/portfolio/:id/metrics", riskHandler.GetRiskMetrics)
	risk.Get("/portfolio/:id/var", riskHandler.CalculateVAR)
	risk.Get("/portfolio/:id/liquidity", riskHandler.CalculateLiquidityRisk)
	risk.Get("/portfolio/:id/history", riskHandler.GetRiskHistory)

	// Monte Carlo simulation routes
	monteCarlo := protected.Group("/monte-carlo")
	monteCarlo.Post("/portfolio/:portfolio_id/simulation", monteCarloHandler.RunSimulation)
	monteCarlo.Get("/portfolio/:portfolio_id/validation", monteCarloHandler.RunQuickValidation)
	monteCarlo.Get("/simulation/:simulation_id/status", monteCarloHandler.GetSimulationStatus)
	monteCarlo.Get("/portfolio/:portfolio_id/history", monteCarloHandler.GetSimulationHistory)
	monteCarlo.Post("/portfolio/:portfolio_id/compare", monteCarloHandler.CompareSimulations)

	// Alert routes
	alerts := protected.Group("/alerts")
	alerts.Get("/", alertHandler.GetAlerts)
	alerts.Get("/active", alertHandler.GetActiveAlerts)
	alerts.Get("/:id", alertHandler.GetAlert)
	alerts.Put("/:id/acknowledge", alertHandler.AcknowledgeAlert)
	alerts.Put("/:id/resolve", alertHandler.ResolveAlert)
	alerts.Delete("/:id", alertHandler.DeleteAlert)

	// Compliance routes
	compliance := protected.Group("/compliance")
	compliance.Get("/portfolio/:id/check", complianceHandler.CheckCompliance)
	compliance.Get("/portfolio/:id/position-limits", complianceHandler.CheckPositionLimits)
	compliance.Post("/transaction/:id/aml-check", complianceHandler.CheckAML)

	// Market data routes (public for testing) - NOT under protected group
	app.Get("/api/v1/marketdata/historical/:symbol", marketDataHandler.GetHistoricalData)
	app.Get("/api/v1/marketdata/quote/:symbol", marketDataHandler.GetRealtimeQuote)
	app.Get("/api/v1/marketdata/quotes", marketDataHandler.GetMultipleQuotes)
	app.Get("/api/v1/marketdata/company/:symbol", marketDataHandler.GetCompanyInfo)
	app.Get("/api/v1/marketdata/status", marketDataHandler.GetProviderStatus)
	app.Get("/api/v1/marketdata/search", marketDataHandler.SearchSymbols)
	app.Get("/api/v1/marketdata/validate/:symbol", marketDataHandler.ValidateSymbol)
	app.Get("/api/v1/marketdata/periods", marketDataHandler.GetSupportedPeriods)

	// Trading routes (protected)
	trading := protected.Group("/trading")
	trading.Get("/account", tradingHandler.GetAccount)
	trading.Get("/status", tradingHandler.GetTradingStatus)

	// Order management
	trading.Post("/orders", tradingHandler.PlaceOrder)
	trading.Get("/orders", tradingHandler.GetOrders)
	trading.Get("/orders/:id", tradingHandler.GetOrder)
	trading.Delete("/orders/:id", tradingHandler.CancelOrder)
	trading.Post("/orders/validate", tradingHandler.ValidateOrder)

	// Position management
	trading.Get("/positions", tradingHandler.GetPositions)
	trading.Get("/positions/:symbol", tradingHandler.GetPosition)
	trading.Delete("/positions/:symbol", tradingHandler.ClosePosition)

	// Test endpoint for market data
	app.Get("/api/v1/marketdata/test/:symbol", func(c *fiber.Ctx) error {
		symbol := c.Params("symbol")
		quote, err := marketDataService.GetRealtimeQuote(symbol)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(quote)
	})

	// Admin endpoints
	admin := app.Group("/api/v1/admin")
	admin.Post("/update-transaction-prices", adminHandler.UpdateTransactionPrices)

	// WebSocket endpoint
	app.Use("/ws", func(c *fiber.Ctx) error {
		// IsWebSocketUpgrade returns true if the client
		// requested upgrade to the WebSocket protocol.
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		// Get user ID from query params
		userID := c.Query("user_id", "anonymous")
		clientID := uuid.New().String()

		log.Printf("WebSocket client connected: user_id=%s, client_id=%s", userID, clientID)

		// Register with simple hub
		simpleHub.RegisterConnection(c)
		defer simpleHub.UnregisterConnection(c)

		// Send welcome message
		welcome := map[string]interface{}{
			"type":      "welcome",
			"message":   "Connected to Financial Risk Monitor WebSocket",
			"user_id":   userID,
			"client_id": clientID,
			"timestamp": time.Now().Unix(),
		}

		if err := c.WriteJSON(welcome); err != nil {
			log.Println("WebSocket welcome error:", err)
			return
		}

		// Keep connection alive and handle incoming messages
		for {
			mt, msg, err := c.ReadMessage()
			if err != nil {
				log.Printf("WebSocket read error for client %s: %v", clientID, err)
				break
			}
			log.Printf("WebSocket received from %s: %s", clientID, msg)

			// Echo message back (optional)
			if err = c.WriteMessage(mt, msg); err != nil {
				log.Printf("WebSocket write error for client %s: %v", clientID, err)
				break
			}
		}

		log.Printf("WebSocket client disconnected: %s", clientID)
	}))

	// Start Alpaca real-time data integration in development
	if cfg.App.Env == "development" {
		// Start Alpaca real-time data integration
		go startAlpacaIntegration(marketDataService, hub, simpleHub)
	}

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-quit
		log.Println("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			log.Fatal("Server forced to shutdown:", err)
		}
	}()

	// Start server
	log.Printf("Server starting on port %s", cfg.App.Port)
	if err := app.Listen(":" + cfg.App.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

func startAlpacaIntegration(marketDataService *marketdata.Service, hub *wsHandler.Hub, simpleHub *wsHandler.SimpleHub) {
	log.Println("Starting Alpaca Markets real-time data integration...")

	// Define the symbols we want to track (expanded list with major stocks)
	symbols := []string{
		"AAPL", "GOOGL", "GOOG", "MSFT", "TSLA", "AMZN", "NVDA", "META",
		"JPM", "BAC", "GS", "WFC", "C", "MS",
		"NFLX", "DIS", "V", "MA", "PYPL", "ADBE",
		"CRM", "ORCL", "IBM", "INTC", "AMD", "QCOM",
		"JNJ", "PFE", "UNH", "ABBV", "MRK", "TMO",
		"KO", "PEP", "WMT", "HD", "MCD", "NKE",
		"XOM", "CVX", "COP", "SLB", "OXY",
	}

	// Fetch data immediately on startup
	log.Println("Fetching initial Alpaca market data...")
	quotes, err := marketDataService.GetMultipleQuotes(symbols)

	var priceUpdates map[string]interface{}
	var dataSource string

	if err != nil || len(quotes) == 0 {
		log.Printf("Error fetching Alpaca data or no data returned: %v. Using fallback data.", err)
		// Generate realistic fallback data
		priceUpdates = generateFallbackPriceData(symbols)
		dataSource = "Fallback Market Data"
	} else {
		// Convert to WebSocket format and broadcast
		priceUpdates = make(map[string]interface{})
		for symbol, quote := range quotes {
			priceUpdates[symbol] = map[string]interface{}{
				"price":          quote.Price,
				"change":         quote.Change,
				"change_percent": quote.ChangePercent,
				"volume":         quote.Volume,
				"timestamp":      quote.Timestamp.Unix(),
				"provider":       "Alpaca Markets",
				"is_market_open": quote.IsMarketOpen,
			}
		}
		dataSource = "Alpaca Markets"
	}

	// Broadcast to WebSocket clients
	message := map[string]interface{}{
		"type": "price_update",
		"data": priceUpdates,
	}

	// Send to simple hub for WebSocket connections
	if simpleHub != nil {
		simpleHub.BroadcastToAll(message)
	}

	log.Printf("Broadcasted initial %d symbols with %s data", len(priceUpdates), dataSource)

	ticker := time.NewTicker(10 * time.Second) // Update every 10 seconds for testing
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Fetch real market data from Alpaca
			quotes, err := marketDataService.GetMultipleQuotes(symbols)

			var priceUpdates map[string]interface{}
			var dataSource string

			if err != nil || len(quotes) == 0 {
				log.Printf("Error fetching Alpaca data: %v. Using fallback data.", err)
				// Generate realistic fallback data
				priceUpdates = generateFallbackPriceData(symbols)
				dataSource = "Fallback Market Data"
			} else {
				// Convert to WebSocket format and broadcast
				priceUpdates = make(map[string]interface{})
				for symbol, quote := range quotes {
					priceUpdates[symbol] = map[string]interface{}{
						"price":          quote.Price,
						"change":         quote.Change,
						"change_percent": quote.ChangePercent,
						"volume":         quote.Volume,
						"timestamp":      quote.Timestamp.Unix(),
						"provider":       "Alpaca Markets",
						"is_market_open": quote.IsMarketOpen,
					}
				}
				dataSource = "Alpaca Markets"
			}

			// Broadcast to WebSocket clients
			message := map[string]interface{}{
				"type": "price_update",
				"data": priceUpdates,
			}

			// Send to simple hub for WebSocket connections
			if simpleHub != nil {
				simpleHub.BroadcastToAll(message)
			}

			log.Printf("Updated %d symbols with %s data", len(priceUpdates), dataSource)
		}
	}
}

// generateFallbackPriceData creates realistic market data when API calls fail
func generateFallbackPriceData(symbols []string) map[string]interface{} {
	priceUpdates := make(map[string]interface{})

	// Base prices for different symbols (realistic market prices)
	basePrices := map[string]float64{
		"AAPL": 175.00, "GOOGL": 2800.00, "GOOG": 2790.00, "MSFT": 350.00, "TSLA": 800.00,
		"AMZN": 3300.00, "NVDA": 450.00, "META": 320.00, "JPM": 140.00, "BAC": 35.00,
		"GS": 350.00, "WFC": 42.00, "C": 50.00, "MS": 95.00, "NFLX": 400.00, "DIS": 95.00,
		"V": 240.00, "MA": 380.00, "PYPL": 65.00, "ADBE": 550.00, "CRM": 220.00,
		"ORCL": 110.00, "IBM": 140.00, "INTC": 35.00, "AMD": 120.00, "QCOM": 140.00,
		"JNJ": 160.00, "PFE": 35.00, "UNH": 520.00, "ABBV": 150.00, "MRK": 105.00,
		"TMO": 520.00, "KO": 60.00, "PEP": 170.00, "WMT": 155.00, "HD": 330.00,
		"MCD": 280.00, "NKE": 105.00, "XOM": 110.00, "CVX": 155.00, "COP": 125.00,
		"SLB": 55.00, "OXY": 65.00,
	}

	// Volume estimates for different symbols
	volumes := map[string]int64{
		"AAPL": 75000000, "GOOGL": 30000000, "GOOG": 28000000, "MSFT": 45000000, "TSLA": 85000000,
		"AMZN": 35000000, "NVDA": 55000000, "META": 40000000, "JPM": 15000000, "BAC": 45000000,
		"GS": 3000000, "WFC": 25000000, "C": 20000000, "MS": 8000000, "NFLX": 12000000, "DIS": 18000000,
		"V": 8000000, "MA": 6000000, "PYPL": 15000000, "ADBE": 4000000, "CRM": 5000000,
		"ORCL": 25000000, "IBM": 8000000, "INTC": 35000000, "AMD": 40000000, "QCOM": 12000000,
		"JNJ": 12000000, "PFE": 30000000, "UNH": 8000000, "ABBV": 10000000, "MRK": 15000000,
		"TMO": 3000000, "KO": 18000000, "PEP": 8000000, "WMT": 15000000, "HD": 10000000,
		"MCD": 6000000, "NKE": 12000000, "XOM": 25000000, "CVX": 18000000, "COP": 12000000,
		"SLB": 8000000, "OXY": 10000000,
	}

	for _, symbol := range symbols {
		basePrice := basePrices[symbol]
		if basePrice == 0 {
			basePrice = 100.00 // Default price
		}

		volume := volumes[symbol]
		if volume == 0 {
			volume = 1000000 // Default volume
		}

		// Generate realistic price movement (-3% to +3%)
		changePercent := (rand.Float64() - 0.5) * 6.0
		currentPrice := basePrice * (1 + changePercent/100)
		changeAmount := currentPrice - basePrice

		priceUpdates[symbol] = map[string]interface{}{
			"price":          currentPrice,
			"change":         changeAmount,
			"change_percent": changePercent,
			"volume":         volume + int64(rand.Intn(100000)), // Add some volume variation
			"timestamp":      time.Now().Unix(),
			"provider":       "Fallback Market Data",
			"is_market_open": true,
		}
	}

	return priceUpdates
}
