package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/Taf0711/financial-risk-monitor/internal/config"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
	"github.com/Taf0711/financial-risk-monitor/internal/services"
)

type RiskHandler struct {
	config      *config.RiskConfig
	riskService *services.RiskService
}

func NewRiskHandler(cfg *config.RiskConfig) *RiskHandler {
	return &RiskHandler{
		config:      cfg,
		riskService: services.NewRiskService(),
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

	// Calculate actual VaR
	riskMetric, err := h.riskService.CalculatePortfolioVAR(portfolioUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"portfolio_id":      portfolioID,
		"var_value":         riskMetric.Value,
		"var_threshold":     riskMetric.Threshold,
		"status":            riskMetric.Status,
		"confidence_level":  riskMetric.ConfidenceLevel,
		"time_horizon_days": riskMetric.TimeHorizon,
		"currency":          "USD",
		"calculated_at":     riskMetric.CalculatedAt,
		"details":           riskMetric.Details,
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

	// Calculate actual liquidity risk
	riskMetric, err := h.riskService.CalculatePortfolioLiquidity(portfolioUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Extract breakdown from details
	breakdown := riskMetric.Details["breakdown"].(map[string]interface{})
	highLiquidity := breakdown["HIGH"].(float64)
	mediumLiquidity := breakdown["MEDIUM"].(float64)
	lowLiquidity := breakdown["LOW"].(float64)

	totalValue := riskMetric.Details["portfolio_value"].(float64)
	liquidAssetsPct := (highLiquidity / totalValue) * 100
	illiquidAssetsPct := ((mediumLiquidity + lowLiquidity) / totalValue) * 100

	return c.JSON(fiber.Map{
		"portfolio_id":        portfolioID,
		"liquidity_ratio":     riskMetric.Value,
		"liquidity_score":     riskMetric.Status,
		"days_to_liquidate":   3.5, // This would need actual calculation
		"liquid_assets_pct":   liquidAssetsPct,
		"illiquid_assets_pct": illiquidAssetsPct,
		"calculated_at":       riskMetric.CalculatedAt,
		"breakdown":           breakdown,
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

	// Get actual risk metrics from database
	metrics, err := h.riskService.GetPortfolioRiskMetrics(portfolioUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve risk metrics",
		})
	}

	if len(metrics) == 0 {
		// Calculate metrics if none exist
		varMetric, err := h.riskService.CalculatePortfolioVAR(portfolioUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to calculate VaR: " + err.Error(),
			})
		}

		liquidityMetric, err := h.riskService.CalculatePortfolioLiquidity(portfolioUUID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to calculate liquidity: " + err.Error(),
			})
		}

		metrics = []models.RiskMetric{*varMetric, *liquidityMetric}
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

	// Get actual risk history from service
	history, err := h.riskService.GetPortfolioRiskHistory(portfolioUUID, metricType, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve risk history",
		})
	}

	return c.JSON(history)
}
