// backend/internal/handlers/monte_carlo.go
package handlers

import (
	"strconv"

	"github.com/Taf0711/financial-risk-monitor/internal/risk/calculator"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/simulation"
	"github.com/Taf0711/financial-risk-monitor/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type MonteCarloHandler struct {
	portfolioService *services.PortfolioService
}

func NewMonteCarloHandler(portfolioService *services.PortfolioService) *MonteCarloHandler {
	return &MonteCarloHandler{
		portfolioService: portfolioService,
	}
}

// RunSimulation executes Monte Carlo simulation for a portfolio
func (h *MonteCarloHandler) RunSimulation(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	portfolioID := c.Params("portfolio_id")

	// Parse simulation configuration
	var config simulation.MonteCarloConfig
	if err := c.BodyParser(&config); err != nil {
		// Use default configuration if parsing fails
		config = simulation.MonteCarloConfig{
			NumSimulations:    5000,
			TimeHorizonDays:   22,
			ConfidenceLevel:   0.95,
			NumWorkers:        4,
			MarketRegime:      "NORMAL",
			CorrelationMatrix: true,
		}
	}

	// Validate configuration
	if config.NumSimulations <= 0 || config.NumSimulations > 100000 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "NumSimulations must be between 1 and 100,000",
		})
	}

	if config.TimeHorizonDays <= 0 || config.TimeHorizonDays > 252 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "TimeHorizonDays must be between 1 and 252",
		})
	}

	// Get portfolio and positions
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	portfolio, err := h.portfolioService.GetPortfolio(portfolioUUID, userUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Portfolio not found",
		})
	}

	if len(portfolio.Positions) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Portfolio has no positions to simulate",
		})
	}

	// Initialize calculators
	portfolioValue := portfolio.TotalValue.InexactFloat64()
	varCalculator := calculator.NewVaRCalculator(portfolioValue)

	// Mock market data provider for now - in production, use real market data
	liquidityCalculator := calculator.NewLiquidityCalculator(&MockMarketDataProvider{})

	// Create Monte Carlo simulator
	simulator := simulation.NewMonteCarloSimulator(varCalculator, liquidityCalculator)

	// Run simulation
	result, err := simulator.RunSimulation(portfolio.Positions, config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to run Monte Carlo simulation: " + err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":    true,
		"portfolio":  portfolio,
		"simulation": result,
	})
}

// GetSimulationStatus returns the status of a running simulation
func (h *MonteCarloHandler) GetSimulationStatus(c *fiber.Ctx) error {
	simulationID := c.Params("simulation_id")

	// In a production system, you would track simulation status
	// For now, return a simple status
	return c.JSON(fiber.Map{
		"simulation_id": simulationID,
		"status":        "completed",
		"progress":      100,
	})
}

// RunQuickValidation performs a quick Monte Carlo validation
func (h *MonteCarloHandler) RunQuickValidation(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	portfolioID := c.Params("portfolio_id")

	// Get portfolio
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	portfolio, err := h.portfolioService.GetPortfolio(portfolioUUID, userUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Portfolio not found",
		})
	}

	// Quick validation configuration
	config := simulation.MonteCarloConfig{
		NumSimulations:    1000, // Fewer simulations for speed
		TimeHorizonDays:   22,   // Monthly horizon
		ConfidenceLevel:   0.95,
		NumWorkers:        2, // Fewer workers
		MarketRegime:      "NORMAL",
		CorrelationMatrix: false, // Disable for speed
	}

	// Initialize calculators
	portfolioValue := portfolio.TotalValue.InexactFloat64()
	varCalculator := calculator.NewVaRCalculator(portfolioValue)
	liquidityCalculator := calculator.NewLiquidityCalculator(&MockMarketDataProvider{})

	// Create simulator and run
	simulator := simulation.NewMonteCarloSimulator(varCalculator, liquidityCalculator)
	result, err := simulator.RunSimulation(portfolio.Positions, config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to run validation: " + err.Error(),
		})
	}

	// Return simplified results for quick validation
	return c.JSON(fiber.Map{
		"success": true,
		"validation": fiber.Map{
			"portfolio_value": result.PortfolioValue,
			"var_95":          result.VaREstimates["VaR_95"],
			"var_99":          result.VaREstimates["VaR_99"],
			"volatility":      result.Volatility * 100,
			"max_drawdown":    result.MaxDrawdown * 100,
			"liquidity_score": func() float64 {
				if result.LiquidityMetrics != nil {
					return result.LiquidityMetrics.AverageLiquidityScore
				}
				return 0.0
			}(),
			"simulation_time": result.Duration.String(),
			"performance": fiber.Map{
				"simulations_per_sec": result.PerformanceMetrics.SimulationsPerSecond,
				"memory_usage_mb":     float64(result.PerformanceMetrics.MemoryUsage) / (1024 * 1024),
			},
		},
	})
}

// GetSimulationHistory returns historical simulation results
func (h *MonteCarloHandler) GetSimulationHistory(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	portfolioID := c.Params("portfolio_id")

	// Parse query parameters
	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// In a production system, you would store simulation results in database
	// For now, return mock historical data
	history := []fiber.Map{
		{
			"id":         "sim_001",
			"timestamp":  "2025-08-28T10:00:00Z",
			"config":     fiber.Map{"simulations": 5000, "regime": "NORMAL"},
			"var_95":     12500.00,
			"var_99":     18200.00,
			"volatility": 22.5,
			"duration":   "1.2s",
		},
		{
			"id":         "sim_002",
			"timestamp":  "2025-08-27T15:30:00Z",
			"config":     fiber.Map{"simulations": 10000, "regime": "STRESSED"},
			"var_95":     18750.00,
			"var_99":     26800.00,
			"volatility": 31.2,
			"duration":   "2.8s",
		},
	}

	// Limit results
	if len(history) > limit {
		history = history[:limit]
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"portfolio_id": portfolioID,
		"user_id":      userID,
		"history":      history,
		"total":        len(history),
	})
}

// CompareSimulations compares different simulation scenarios
func (h *MonteCarloHandler) CompareSimulations(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)
	portfolioID := c.Params("portfolio_id")

	// Parse comparison request
	var request struct {
		Scenarios []simulation.MonteCarloConfig `json:"scenarios"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if len(request.Scenarios) == 0 || len(request.Scenarios) > 5 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Must provide between 1 and 5 scenarios to compare",
		})
	}

	// Get portfolio
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	portfolio, err := h.portfolioService.GetPortfolio(portfolioUUID, userUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Portfolio not found",
		})
	}

	// Initialize calculators
	portfolioValue := portfolio.TotalValue.InexactFloat64()
	varCalculator := calculator.NewVaRCalculator(portfolioValue)
	liquidityCalculator := calculator.NewLiquidityCalculator(&MockMarketDataProvider{})
	simulator := simulation.NewMonteCarloSimulator(varCalculator, liquidityCalculator)

	// Run simulations for each scenario
	results := make([]fiber.Map, 0, len(request.Scenarios))

	for i, config := range request.Scenarios {
		// Validate each configuration
		if config.NumSimulations <= 0 || config.NumSimulations > 50000 {
			config.NumSimulations = 5000 // Default for comparison
		}

		result, err := simulator.RunSimulation(portfolio.Positions, config)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to run scenario " + strconv.Itoa(i+1) + ": " + err.Error(),
			})
		}

		// Summarize results for comparison
		scenarioResult := fiber.Map{
			"scenario":              i + 1,
			"config":                config,
			"var_95":                result.VaREstimates["VaR_95"],
			"var_99":                result.VaREstimates["VaR_99"],
			"expected_shortfall_95": result.ExpectedShortfall["ES_95"],
			"expected_shortfall_99": result.ExpectedShortfall["ES_99"],
			"volatility":            result.Volatility * 100,
			"max_drawdown":          result.MaxDrawdown * 100,
			"skewness":              result.Skewness,
			"kurtosis":              result.Kurtosis,
			"simulation_time":       result.Duration.String(),
		}

		if result.LiquidityMetrics != nil {
			scenarioResult["liquidity_score"] = result.LiquidityMetrics.AverageLiquidityScore
		}

		results = append(results, scenarioResult)
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"portfolio_id": portfolioID,
		"comparison":   results,
		"summary": fiber.Map{
			"scenarios_compared": len(results),
			"portfolio_value":    portfolioValue,
		},
	})
}

// MockMarketDataProvider provides mock market data for the handler
type MockMarketDataProvider struct{}

func (m *MockMarketDataProvider) GetAverageDailyVolume(symbol string) float64 {
	volumes := map[string]float64{
		"AAPL": 50000000, "GOOGL": 25000000, "MSFT": 40000000,
		"BTC": 2000000, "ETH": 5000000, "GOLD": 100000,
	}
	if vol, exists := volumes[symbol]; exists {
		return vol
	}
	return 1000000
}

func (m *MockMarketDataProvider) GetBidAskSpread(symbol string) float64 {
	spreads := map[string]float64{
		"AAPL": 0.0001, "GOOGL": 0.0001, "MSFT": 0.0001,
		"BTC": 0.001, "ETH": 0.001, "GOLD": 0.002,
	}
	if spread, exists := spreads[symbol]; exists {
		return spread
	}
	return 0.001
}

func (m *MockMarketDataProvider) GetMarketDepth(symbol string) *calculator.MarketDepth {
	return &calculator.MarketDepth{
		BidLevels: []calculator.PriceLevel{{Price: 100.0, Quantity: 1000, Orders: 10}},
		AskLevels: []calculator.PriceLevel{{Price: 100.1, Quantity: 1000, Orders: 10}},
	}
}

func (m *MockMarketDataProvider) GetMarketCap(symbol string) float64 {
	marketCaps := map[string]float64{
		"AAPL": 3000000000000, "GOOGL": 1500000000000, "MSFT": 2800000000000,
		"BTC": 800000000000, "ETH": 400000000000, "GOLD": 12000000000000,
	}
	if cap, exists := marketCaps[symbol]; exists {
		return cap
	}
	return 1000000000
}
