package handlers

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/Taf0711/financial-risk-monitor/internal/config"
	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/marketdata"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/calculator"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/optimization"
	"github.com/shopspring/decimal"
)

// ProfessionalRiskHandler handles professional risk calculations
type ProfessionalRiskHandler struct {
	config            *config.RiskConfig
	marketDataService *marketdata.Service
}

// NewProfessionalRiskHandler creates a new professional risk handler
func NewProfessionalRiskHandler(cfg *config.RiskConfig, marketDataService *marketdata.Service) *ProfessionalRiskHandler {
	return &ProfessionalRiskHandler{
		config:            cfg,
		marketDataService: marketDataService,
	}
}

// CalculateProfessionalVaR calculates comprehensive Value at Risk
func (h *ProfessionalRiskHandler) CalculateProfessionalVaR(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	// Get time horizon from query params (default 1 day)
	timeHorizon := 1
	if th := c.Query("time_horizon"); th != "" {
		if parsed, err := strconv.Atoi(th); err == nil && parsed > 0 && parsed <= 252 {
			timeHorizon = parsed
		}
	}

	// Get portfolio and positions
	var portfolio models.Portfolio
	if err := database.GetDB().Preload("Positions").First(&portfolio, portfolioUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Portfolio not found",
		})
	}

	if len(portfolio.Positions) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Portfolio has no positions",
		})
	}

	// Get price history for all positions
	priceHistory := make(map[string][]float64)
	for _, position := range portfolio.Positions {
		// Get historical data from market data service
		historicalData, err := h.marketDataService.GetHistoricalData(position.Symbol, "1y")
		if err != nil {
			continue // Skip if we can't get data for this symbol
		}

		prices := make([]float64, len(historicalData.DataPoints))
		for i, point := range historicalData.DataPoints {
			prices[i] = point.Close
		}
		priceHistory[position.Symbol] = prices
	}

	if len(priceHistory) == 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to retrieve price history for portfolio positions",
		})
	}

	// Create professional VaR calculator
	portfolioValue := portfolio.TotalValue.InexactFloat64()
	varCalculator := calculator.NewProfessionalVaRCalculator(portfolioValue, timeHorizon)

	// Calculate comprehensive VaR
	result, err := varCalculator.CalculateProfessionalVaR(portfolio.Positions, priceHistory)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate VaR: " + err.Error(),
		})
	}

	// Store the most conservative VaR in database
	mostConservativeVaR := result.HistoricalVaR[0.99]
	if result.MonteCarloVaR[0.99] > mostConservativeVaR {
		mostConservativeVaR = result.MonteCarloVaR[0.99]
	}
	if result.CornishFisherVaR[0.99] > mostConservativeVaR {
		mostConservativeVaR = result.CornishFisherVaR[0.99]
	}

	// Determine status based on VaR percentage
	varPercentage := (mostConservativeVaR / portfolioValue) * 100
	status := "SAFE"
	if varPercentage > 15 {
		status = "CRITICAL"
	} else if varPercentage > 10 {
		status = "WARNING"
	}

	// Store in database
	riskMetric := models.RiskMetric{
		PortfolioID:     portfolioUUID,
		MetricType:      "PROFESSIONAL_VAR",
		Value:           decimal.NewFromFloat(mostConservativeVaR),
		Threshold:       decimal.NewFromFloat(portfolioValue * 0.15), // 15% threshold
		Status:          status,
		TimeHorizon:     timeHorizon,
		ConfidenceLevel: decimal.NewFromFloat(0.99),
		Details: models.JSON{
			"historical_var_99":     result.HistoricalVaR[0.99],
			"parametric_var_99":     result.ParametricVaR[0.99],
			"monte_carlo_var_99":    result.MonteCarloVaR[0.99],
			"cornish_fisher_var_99": result.CornishFisherVaR[0.99],
			"conditional_var_99":    result.ConditionalVaR[0.99],
			"stress_var":            result.StressVaR,
			"volatility":            result.Volatility,
			"skewness":              result.Skewness,
			"kurtosis":              result.Kurtosis,
			"max_drawdown":          result.MaxDrawdown,
			"tail_risk":             result.TailRisk,
		},
	}

	database.GetDB().Create(&riskMetric)

	// Generate alert if VaR exceeds threshold
	if status == "CRITICAL" || status == "WARNING" {
		alert := models.Alert{
			PortfolioID: portfolioUUID,
			AlertType:   "RISK_BREACH",
			Severity:    status,
			Title:       "VaR Limit Exceeded",
			Description: fmt.Sprintf("Portfolio VaR (%.2f%%) exceeds acceptable levels", varPercentage),
			Source:      "PROFESSIONAL_VAR_CALCULATOR",
			Status:      "ACTIVE",
			TriggeredBy: models.JSON{
				"var_value":      mostConservativeVaR,
				"var_percentage": varPercentage,
				"time_horizon":   timeHorizon,
			},
		}
		database.GetDB().Create(&alert)
	}

	return c.JSON(result)
}

// CalculateGreeks calculates option Greeks for derivatives
func (h *ProfessionalRiskHandler) CalculateGreeks(c *fiber.Ctx) error {
	// Parse option parameters from request
	var req struct {
		Type            string  `json:"type"` // CALL or PUT
		UnderlyingPrice float64 `json:"underlying_price"`
		StrikePrice     float64 `json:"strike_price"`
		TimeToExpiry    float64 `json:"time_to_expiry"` // in years
		RiskFreeRate    float64 `json:"risk_free_rate"`
		Volatility      float64 `json:"volatility"`
		DividendYield   float64 `json:"dividend_yield"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Create option
	option := calculator.Option{
		Type:            calculator.OptionType(req.Type),
		Style:           calculator.EuropeanStyle,
		UnderlyingPrice: req.UnderlyingPrice,
		StrikePrice:     req.StrikePrice,
		TimeToExpiry:    req.TimeToExpiry,
		RiskFreeRate:    req.RiskFreeRate,
		Volatility:      req.Volatility,
		DividendYield:   req.DividendYield,
	}

	// Calculate Greeks
	greeksCalc := calculator.NewGreeksCalculator()
	greeks, err := greeksCalc.CalculateGreeks(option)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate Greeks: " + err.Error(),
		})
	}

	// Also run scenario analysis
	scenarios, err := greeksCalc.RunScenarioAnalysis(option)
	if err != nil {
		scenarios = nil // Don't fail if scenario analysis fails
	}

	return c.JSON(fiber.Map{
		"option":    option,
		"greeks":    greeks,
		"scenarios": scenarios,
	})
}

// OptimizePortfolio performs portfolio optimization
func (h *ProfessionalRiskHandler) OptimizePortfolio(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	// Parse optimization config
	var config optimization.OptimizationConfig
	if err := c.BodyParser(&config); err != nil {
		// Use default config
		config = optimization.OptimizationConfig{
			Objective:         optimization.MaximizeSharpeRatio,
			RiskFreeRate:      0.05,
			MaxIterations:     1000,
			Tolerance:         1e-6,
			RiskAversionParam: 2.0,
		}
	}

	// Get portfolio and positions
	var portfolio models.Portfolio
	if err := database.GetDB().Preload("Positions").First(&portfolio, portfolioUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Portfolio not found",
		})
	}

	if len(portfolio.Positions) < 2 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Portfolio needs at least 2 positions for optimization",
		})
	}

	// Get price history
	priceHistory := make(map[string][]float64)
	for _, position := range portfolio.Positions {
		historicalData, err := h.marketDataService.GetHistoricalData(position.Symbol, "1y")
		if err != nil {
			continue
		}

		prices := make([]float64, len(historicalData.DataPoints))
		for i, point := range historicalData.DataPoints {
			prices[i] = point.Close
		}
		priceHistory[position.Symbol] = prices
	}

	// Create optimizer
	optimizer, err := optimization.NewPortfolioOptimizer(portfolio.Positions, priceHistory, config)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to initialize optimizer: " + err.Error(),
		})
	}

	// Run optimization
	result, err := optimizer.Optimize()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Optimization failed: " + err.Error(),
		})
	}

	// Compare with current allocation
	currentWeights := make(map[string]float64)
	totalValue := portfolio.TotalValue.InexactFloat64()
	for _, position := range portfolio.Positions {
		currentWeights[position.Symbol] = position.MarketValue.InexactFloat64() / totalValue
	}

	return c.JSON(fiber.Map{
		"portfolio_id":       portfolioID,
		"current_weights":    currentWeights,
		"optimal_weights":    result.OptimalWeights,
		"optimization":       result,
		"rebalancing_needed": calculateRebalancingNeeded(currentWeights, result.OptimalWeights),
	})
}

// GetRiskDecomposition provides detailed risk decomposition
func (h *ProfessionalRiskHandler) GetRiskDecomposition(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	// Get portfolio
	var portfolio models.Portfolio
	if err := database.GetDB().Preload("Positions").First(&portfolio, portfolioUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Portfolio not found",
		})
	}

	// Get the latest professional VaR calculation
	var riskMetric models.RiskMetric
	if err := database.GetDB().
		Where("portfolio_id = ? AND metric_type = ?", portfolioUUID, "PROFESSIONAL_VAR").
		Order("calculated_at DESC").
		First(&riskMetric).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No risk metrics found. Please calculate VaR first.",
		})
	}

	// Extract decomposition from stored details
	details := make(map[string]interface{})
	if riskMetric.Details != nil {
		// Parse the JSON details
		// Note: In production, you'd properly unmarshal this
		details = map[string]interface{}(riskMetric.Details)
	}

	return c.JSON(fiber.Map{
		"portfolio_id":       portfolioID,
		"calculated_at":      riskMetric.CalculatedAt,
		"risk_decomposition": details,
		"total_var":          riskMetric.Value,
		"status":             riskMetric.Status,
	})
}

// GetStressTestResults performs stress testing
func (h *ProfessionalRiskHandler) GetStressTestResults(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	// Get portfolio
	var portfolio models.Portfolio
	if err := database.GetDB().Preload("Positions").First(&portfolio, portfolioUUID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Portfolio not found",
		})
	}

	// Define stress scenarios
	scenarios := []struct {
		Name        string
		Description string
		Shocks      map[string]float64
	}{
		{
			Name:        "Market Crash",
			Description: "2008-style financial crisis",
			Shocks: map[string]float64{
				"equity":        -0.40,
				"volatility":    2.5,
				"credit_spread": 3.0,
			},
		},
		{
			Name:        "Interest Rate Shock",
			Description: "Rapid rate increase",
			Shocks: map[string]float64{
				"rates":      0.03,
				"equity":     -0.15,
				"volatility": 1.5,
			},
		},
		{
			Name:        "Liquidity Crisis",
			Description: "Market liquidity dries up",
			Shocks: map[string]float64{
				"liquidity": -0.80,
				"bid_ask":   5.0,
				"equity":    -0.25,
			},
		},
		{
			Name:        "Tech Bubble Burst",
			Description: "Technology sector collapse",
			Shocks: map[string]float64{
				"tech_equity":  -0.60,
				"other_equity": -0.20,
				"volatility":   2.0,
			},
		},
	}

	// Calculate impact of each scenario
	results := make([]fiber.Map, 0)
	currentValue := portfolio.TotalValue.InexactFloat64()

	for _, scenario := range scenarios {
		// Simplified stress test calculation
		// In production, you'd use proper factor models
		impact := 0.0

		// Apply equity shock
		if equityShock, exists := scenario.Shocks["equity"]; exists {
			impact = currentValue * equityShock
		}

		// Add volatility impact (simplified)
		if volMultiplier, exists := scenario.Shocks["volatility"]; exists {
			additionalLoss := currentValue * 0.05 * (volMultiplier - 1)
			impact += additionalLoss
		}

		results = append(results, fiber.Map{
			"scenario":       scenario.Name,
			"description":    scenario.Description,
			"shocks":         scenario.Shocks,
			"impact":         impact,
			"impact_percent": (impact / currentValue) * 100,
			"new_value":      currentValue + impact,
		})
	}

	return c.JSON(fiber.Map{
		"portfolio_id":   portfolioID,
		"current_value":  currentValue,
		"stress_tests":   results,
		"worst_scenario": findWorstScenario(results),
		"timestamp":      time.Now(),
	})
}

// Helper functions

func calculateRebalancingNeeded(current, optimal map[string]float64) map[string]float64 {
	rebalancing := make(map[string]float64)

	for symbol, optimalWeight := range optimal {
		currentWeight := current[symbol]
		rebalancing[symbol] = optimalWeight - currentWeight
	}

	return rebalancing
}

func findWorstScenario(results []fiber.Map) fiber.Map {
	if len(results) == 0 {
		return fiber.Map{}
	}

	worst := results[0]
	worstImpact := worst["impact"].(float64)

	for _, result := range results[1:] {
		if impact := result["impact"].(float64); impact < worstImpact {
			worst = result
			worstImpact = impact
		}
	}

	return worst
}
