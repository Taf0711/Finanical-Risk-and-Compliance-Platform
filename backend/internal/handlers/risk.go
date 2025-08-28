package handlers

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Taf0711/financial-risk-monitor/internal/config"
	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

type RiskHandler struct {
	config *config.RiskConfig
}

func NewRiskHandler(cfg *config.RiskConfig) *RiskHandler {
	return &RiskHandler{
		config: cfg,
	}
}

// CalculateVAR calculates Value at Risk for a portfolio
func (h *RiskHandler) CalculateVAR(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
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

	// Simple VaR calculation - 5% of portfolio value at 95% confidence
	varPercentage := 0.05 // 5% VaR
	varValue := portfolio.TotalValue.Mul(decimal.NewFromFloat(varPercentage))
	threshold := portfolio.TotalValue.Mul(decimal.NewFromFloat(0.08)) // 8% threshold

	status := "SAFE"
	if varValue.GreaterThan(threshold) {
		status = "CRITICAL"
	} else if varValue.GreaterThan(threshold.Mul(decimal.NewFromFloat(0.75))) {
		status = "WARNING"
	}

	// Store the metric in database
	riskMetric := models.RiskMetric{
		PortfolioID:     portfolioUUID,
		MetricType:      "VAR",
		Value:           varValue,
		Threshold:       threshold,
		Status:          status,
		TimeHorizon:     h.config.VARTimeHorizon,
		ConfidenceLevel: decimal.NewFromFloat(h.config.VARConfidenceLevel),
		Details: models.JSON{
			"method":          "simplified",
			"portfolio_value": portfolio.TotalValue.InexactFloat64(),
			"position_count":  len(portfolio.Positions),
		},
	}

	database.GetDB().Create(&riskMetric)

	return c.JSON(fiber.Map{
		"portfolio_id":     portfolioID,
		"var_value":        varValue,
		"var_percentage":   varPercentage * 100,
		"confidence_level": h.config.VARConfidenceLevel,
		"time_horizon":     h.config.VARTimeHorizon,
		"method":           "simplified",
		"portfolio_value":  portfolio.TotalValue,
		"status":           status,
		"threshold":        threshold,
		"calculated_at":    time.Now(),
	})
}

// CalculateLiquidityRisk calculates liquidity risk for a portfolio
func (h *RiskHandler) CalculateLiquidityRisk(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
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

	// Calculate liquidity breakdown
	totalValue := decimal.Zero
	highLiquid := decimal.Zero
	mediumLiquid := decimal.Zero
	lowLiquid := decimal.Zero

	for _, position := range portfolio.Positions {
		totalValue = totalValue.Add(position.MarketValue)
		switch position.Liquidity {
		case "HIGH":
			highLiquid = highLiquid.Add(position.MarketValue)
		case "MEDIUM":
			mediumLiquid = mediumLiquid.Add(position.MarketValue)
		case "LOW":
			lowLiquid = lowLiquid.Add(position.MarketValue)
		default:
			highLiquid = highLiquid.Add(position.MarketValue) // Default to high
		}
	}

	liquidityRatio := decimal.Zero
	if !totalValue.IsZero() {
		liquidityRatio = highLiquid.Div(totalValue)
	}

	// Determine risk assessment
	riskAssessment := "LOW_RISK"
	daysToLiquidate := 1.0
	status := "SAFE"
	ratio := liquidityRatio.InexactFloat64()

	if ratio < 0.3 {
		riskAssessment = "HIGH_RISK"
		daysToLiquidate = 10.0
		status = "CRITICAL"
	} else if ratio < 0.7 {
		riskAssessment = "MEDIUM_RISK"
		daysToLiquidate = 3.5
		status = "WARNING"
	}

	// Store the metric in database
	threshold := decimal.NewFromFloat(0.3) // 30% threshold
	riskMetric := models.RiskMetric{
		PortfolioID: portfolioUUID,
		MetricType:  "LIQUIDITY_RATIO",
		Value:       liquidityRatio,
		Threshold:   threshold,
		Status:      status,
		Details: models.JSON{
			"breakdown": fiber.Map{
				"HIGH":   highLiquid.InexactFloat64(),
				"MEDIUM": mediumLiquid.InexactFloat64(),
				"LOW":    lowLiquid.InexactFloat64(),
			},
			"portfolio_value": totalValue.InexactFloat64(),
			"position_count":  len(portfolio.Positions),
		},
	}

	database.GetDB().Create(&riskMetric)

	return c.JSON(fiber.Map{
		"portfolio_id":      portfolioID,
		"liquidity_ratio":   liquidityRatio,
		"liquidity_score":   riskAssessment,
		"days_to_liquidate": daysToLiquidate,
		"risk_assessment":   riskAssessment,
		"status":            status,
		"calculated_at":     time.Now(),
		"breakdown": fiber.Map{
			"HIGH":   highLiquid,
			"MEDIUM": mediumLiquid,
			"LOW":    lowLiquid,
		},
	})
}

// GetRiskMetrics returns risk metrics for a portfolio
func (h *RiskHandler) GetRiskMetrics(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	var metrics []models.RiskMetric
	if err := database.GetDB().Preload("Portfolio").Preload("Portfolio.User").
		Where("portfolio_id = ?", portfolioUUID).
		Order("calculated_at DESC").
		Find(&metrics).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve risk metrics",
		})
	}

	if len(metrics) == 0 {
		// Return instructions to calculate fresh metrics if none exist
		return c.JSON([]fiber.Map{
			{
				"portfolio_id":       portfolioID,
				"metric_type":        "NONE",
				"message":            "No historical metrics found - calculate VaR and liquidity separately",
				"var_endpoint":       "/api/v1/risk/portfolio/" + portfolioID + "/var",
				"liquidity_endpoint": "/api/v1/risk/portfolio/" + portfolioID + "/liquidity",
			},
		})
	}

	return c.JSON(metrics)
}

// GetRiskHistory returns historical risk data for a portfolio
func (h *RiskHandler) GetRiskHistory(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	// Get optional query parameters
	metricType := c.Query("metric_type", "")
	limitStr := c.Query("limit", "30")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 1000 {
		limit = 30
	}

	var history []models.RiskHistory
	query := database.GetDB().Where("portfolio_id = ?", portfolioUUID)

	if metricType != "" {
		query = query.Where("metric_type = ?", metricType)
	}

	if err := query.Order("recorded_at DESC").Limit(limit).Find(&history).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve risk history",
		})
	}

	if len(history) == 0 {
		// Return sample history if none exists
		return c.JSON([]fiber.Map{
			{
				"portfolio_id": portfolioID,
				"metric_type":  "VAR",
				"message":      "No historical data available",
				"suggestion":   "Calculate some risk metrics first to build history",
			},
		})
	}

	return c.JSON(history)
}
