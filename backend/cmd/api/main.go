package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/Taf0711/financial-risk-monitor/internal/config"
	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/handlers"
	"github.com/Taf0711/financial-risk-monitor/internal/middleware"
	"github.com/Taf0711/financial-risk-monitor/internal/services"
	wsHandler "github.com/Taf0711/financial-risk-monitor/internal/websocket"
)

func main() {
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
		AllowOrigins:     "http://localhost:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
		AllowCredentials: true,
	}))

	// Initialize services
	authService := services.NewAuthService(&cfg.JWT)
	authHandler := handlers.NewAuthHandler(authService)
	portfolioHandler := handlers.NewPortfolioHandler()
	transactionHandler := handlers.NewTransactionHandler()
	riskHandler := handlers.NewRiskHandler(&cfg.Risk)
	alertHandler := handlers.NewAlertHandler()

	// Initialize WebSocket hub
	hub := wsHandler.NewHub()
	go hub.Run()

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "healthy",
			"service": "Financial Risk Monitor API",
		})
	})

	// API routes
	api := app.Group("/api/v1")

	// Auth routes (public)
	auth := api.Group("/auth")
	auth.Post("/register", authHandler.Register)
	auth.Post("/login", authHandler.Login)

	// Protected routes
	protected := api.Group("/", middleware.JWTMiddleware(authService))

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
	transactions.Delete("/:id", transactionHandler.DeleteTransaction)

	// Risk metrics routes
	risk := protected.Group("/risk")
	risk.Get("/portfolio/:id/metrics", riskHandler.GetRiskMetrics)
	risk.Get("/portfolio/:id/var", riskHandler.CalculateVAR)
	risk.Get("/portfolio/:id/liquidity", riskHandler.CalculateLiquidityRisk)
	risk.Get("/portfolio/:id/history", riskHandler.GetRiskHistory)

	// Alert routes
	alerts := protected.Group("/alerts")
	alerts.Get("/", alertHandler.GetAlerts)
	alerts.Get("/:id", alertHandler.GetAlert)
	alerts.Put("/:id/acknowledge", alertHandler.AcknowledgeAlert)
	alerts.Put("/:id/resolve", alertHandler.ResolveAlert)
	alerts.Delete("/:id", alertHandler.DeleteAlert)

	// WebSocket endpoint - simple placeholder for now
	app.Get("/ws", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "WebSocket endpoint - upgrade needed",
			"info":    "WebSocket functionality available but requires proper HTTP upgrade",
		})
	})

	// Start mock data generator in development
	if cfg.App.Env == "development" {
		go startMockDataGenerator(hub)
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

func startMockDataGenerator(hub *wsHandler.Hub) {
	// TODO: Implement mock data generator
	// generator := services.NewMockDataGenerator(hub)
	// generator.Start()
	log.Println("Mock data generator not yet implemented")
}
