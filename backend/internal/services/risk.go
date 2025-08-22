package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/calculator"
)

type RiskService struct {
	db            *gorm.DB
	varCalculator *calculator.VARCalculator
	liquidityCalc *calculator.LiquidityCalculator
}

func NewRiskService() *RiskService {
	return &RiskService{
		db:            database.GetDB(),
		varCalculator: calculator.NewVARCalculator(0.95, 1), // 95% confidence, 1 day horizon
		liquidityCalc: calculator.NewLiquidityCalculator(),
	}
}

// CalculatePortfolioVAR calculates Value at Risk for a portfolio
func (s *RiskService) CalculatePortfolioVAR(portfolioID uuid.UUID) (*models.RiskMetric, error) {
	// Get portfolio with positions
	var portfolio models.Portfolio
	if err := s.db.Preload("Positions").First(&portfolio, portfolioID).Error; err != nil {
		return nil, fmt.Errorf("portfolio not found: %w", err)
	}

	if len(portfolio.Positions) == 0 {
		return nil, fmt.Errorf("portfolio has no positions")
	}

	// Generate mock historical returns for demonstration
	// In production, this would come from actual market data
	historicalReturns := calculator.GenerateMockReturns(252) // 1 year of daily returns

	// Calculate VaR
	varValue, err := s.varCalculator.CalculateVAR(&portfolio, historicalReturns)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate VaR: %w", err)
	}

	// Determine status based on threshold (e.g., 5% of portfolio value)
	threshold := portfolio.TotalValue.Mul(decimal.NewFromFloat(0.05))
	status := "SAFE"
	if varValue.GreaterThan(threshold) {
		if varValue.GreaterThan(threshold.Mul(decimal.NewFromFloat(1.5))) {
			status = "CRITICAL"
		} else {
			status = "WARNING"
		}
	}

	// Create risk metric
	riskMetric := &models.RiskMetric{
		PortfolioID:     portfolioID,
		MetricType:      "VAR",
		Value:           varValue,
		Threshold:       threshold,
		Status:          status,
		TimeHorizon:     s.varCalculator.TimeHorizon,
		ConfidenceLevel: decimal.NewFromFloat(s.varCalculator.ConfidenceLevel),
		Details: models.JSON{
			"calculation_method": "historical_simulation",
			"position_count":     len(portfolio.Positions),
			"portfolio_value":    portfolio.TotalValue,
		},
	}

	// Save to database
	if err := s.db.Create(riskMetric).Error; err != nil {
		return nil, fmt.Errorf("failed to save risk metric: %w", err)
	}

	// Also save to risk history
	riskHistory := &models.RiskHistory{
		PortfolioID: portfolioID,
		MetricType:  "VAR",
		Value:       varValue,
		RecordedAt:  time.Now(),
	}

	if err := s.db.Create(riskHistory).Error; err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to save risk history: %v\n", err)
	}

	return riskMetric, nil
}

// CalculatePortfolioLiquidity calculates liquidity risk for a portfolio
func (s *RiskService) CalculatePortfolioLiquidity(portfolioID uuid.UUID) (*models.RiskMetric, error) {
	// Get portfolio with positions
	var portfolio models.Portfolio
	if err := s.db.Preload("Positions").First(&portfolio, portfolioID).Error; err != nil {
		return nil, fmt.Errorf("portfolio not found: %w", err)
	}

	if len(portfolio.Positions) == 0 {
		return nil, fmt.Errorf("portfolio has no positions")
	}

	// Calculate liquidity ratio
	liquidityRatio, breakdown := s.liquidityCalc.CalculateLiquidityRatio(portfolio.Positions)

	// Determine status
	status := s.liquidityCalc.AssessLiquidityRisk(liquidityRatio)
	threshold := decimal.NewFromFloat(0.3) // 30% liquidity threshold

	// Create risk metric
	riskMetric := &models.RiskMetric{
		PortfolioID: portfolioID,
		MetricType:  "LIQUIDITY_RATIO",
		Value:       liquidityRatio,
		Threshold:   threshold,
		Status:      status,
		Details: models.JSON{
			"breakdown":       breakdown,
			"total_positions": len(portfolio.Positions),
			"portfolio_value": portfolio.TotalValue,
		},
	}

	// Save to database
	if err := s.db.Create(riskMetric).Error; err != nil {
		return nil, fmt.Errorf("failed to save liquidity metric: %w", err)
	}

	// Also save to risk history
	riskHistory := &models.RiskHistory{
		PortfolioID: portfolioID,
		MetricType:  "LIQUIDITY_RATIO",
		Value:       liquidityRatio,
		RecordedAt:  time.Now(),
	}

	if err := s.db.Create(riskHistory).Error; err != nil {
		fmt.Printf("Warning: failed to save liquidity history: %v\n", err)
	}

	return riskMetric, nil
}

// GetPortfolioRiskMetrics returns all risk metrics for a portfolio
func (s *RiskService) GetPortfolioRiskMetrics(portfolioID uuid.UUID) ([]models.RiskMetric, error) {
	var metrics []models.RiskMetric
	err := s.db.Where("portfolio_id = ?", portfolioID).Order("calculated_at DESC").Find(&metrics).Error
	return metrics, err
}

// GetPortfolioRiskHistory returns historical risk data for a portfolio
func (s *RiskService) GetPortfolioRiskHistory(portfolioID uuid.UUID, metricType string, limit int) ([]models.RiskHistory, error) {
	var history []models.RiskHistory
	query := s.db.Where("portfolio_id = ?", portfolioID)

	if metricType != "" {
		query = query.Where("metric_type = ?", metricType)
	}

	err := query.Order("recorded_at DESC").Limit(limit).Find(&history).Error
	return history, err
}

// CalculateAllPortfolioRisks calculates risk metrics for all portfolios
func (s *RiskService) CalculateAllPortfolioRisks() error {
	var portfolios []models.Portfolio
	if err := s.db.Find(&portfolios).Error; err != nil {
		return fmt.Errorf("failed to fetch portfolios: %w", err)
	}

	for _, portfolio := range portfolios {
		// Calculate VaR
		if _, err := s.CalculatePortfolioVAR(portfolio.ID); err != nil {
			fmt.Printf("Warning: failed to calculate VaR for portfolio %s: %v\n", portfolio.ID, err)
		}

		// Calculate Liquidity
		if _, err := s.CalculatePortfolioLiquidity(portfolio.ID); err != nil {
			fmt.Printf("Warning: failed to calculate liquidity for portfolio %s: %v\n", portfolio.ID, err)
		}
	}

	return nil
}

// GetRiskMetricsByStatus returns risk metrics filtered by status
func (s *RiskService) GetRiskMetricsByStatus(status string) ([]models.RiskMetric, error) {
	var metrics []models.RiskMetric
	err := s.db.Where("status = ?", status).Preload("Portfolio").Find(&metrics).Error
	return metrics, err
}
