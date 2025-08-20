package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

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
	if _, err := uuid.Parse(portfolioID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	// Placeholder VaR calculation
	return c.JSON(fiber.Map{
		"portfolio_id":      portfolioID,
		"var_1_day_95":      -15000.50,
		"var_10_day_95":     -47500.25,
		"confidence_level":  0.95,
		"time_horizon_days": 1,
		"currency":          "USD",
		"calculated_at":     "2025-08-20T15:55:21Z",
	})
}

// CalculateLiquidityRisk calculates liquidity risk for a portfolio
func (h *RiskHandler) CalculateLiquidityRisk(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	if _, err := uuid.Parse(portfolioID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	// Placeholder liquidity calculation
	return c.JSON(fiber.Map{
		"portfolio_id":        portfolioID,
		"liquidity_ratio":     0.75,
		"liquidity_score":     "GOOD",
		"days_to_liquidate":   3.5,
		"liquid_assets_pct":   65.2,
		"illiquid_assets_pct": 34.8,
		"calculated_at":       "2025-08-20T15:55:21Z",
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
	if err := database.GetDB().Where("portfolio_id = ?", portfolioUUID).Find(&metrics).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve risk metrics",
		})
	}

	if len(metrics) == 0 {
		// Return placeholder data if no metrics exist
		return c.JSON([]fiber.Map{
			{
				"portfolio_id":     portfolioID,
				"metric_type":      "VAR",
				"value":            -15000.50,
				"threshold":        -20000.00,
				"status":           "OK",
				"calculated_at":    "2025-08-20T15:55:21Z",
				"time_horizon":     1,
				"confidence_level": 0.95,
			},
			{
				"portfolio_id":  portfolioID,
				"metric_type":   "LIQUIDITY",
				"value":         0.75,
				"threshold":     0.30,
				"status":        "OK",
				"calculated_at": "2025-08-20T15:55:21Z",
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
	metricType := c.Query("metric_type", "VAR")
	limitStr := c.Query("limit", "30")
	limit, _ := strconv.Atoi(limitStr)

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
		// Return placeholder historical data if none exists
		return c.JSON([]fiber.Map{
			{
				"portfolio_id": portfolioID,
				"metric_type":  metricType,
				"value":        -15000.50,
				"recorded_at":  "2025-08-20T15:55:21Z",
			},
			{
				"portfolio_id": portfolioID,
				"metric_type":  metricType,
				"value":        -14800.25,
				"recorded_at":  "2025-08-19T15:55:21Z",
			},
			{
				"portfolio_id": portfolioID,
				"metric_type":  metricType,
				"value":        -15200.75,
				"recorded_at":  "2025-08-18T15:55:21Z",
			},
		})
	}

	return c.JSON(history)
}
