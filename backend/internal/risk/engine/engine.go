package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Taf0711/financial-risk-monitor/internal/marketdata/providers"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/calculator"
	"github.com/Taf0711/financial-risk-monitor/internal/services"
)

// RiskEngine orchestrates all risk calculations and monitoring
type RiskEngine struct {
	varCalculator       *calculator.VaRCalculator
	liquidityCalculator *calculator.LiquidityCalculator
	riskService         *services.RiskEngineService
}

// NewRiskEngine creates a new risk engine instance
func NewRiskEngine(portfolioValue float64) *RiskEngine {
	// Real market data provider using Alpaca
	alpacaProvider := providers.NewAlpacaProvider("", "")

	return &RiskEngine{
		varCalculator:       calculator.NewVaRCalculator(portfolioValue),
		liquidityCalculator: calculator.NewLiquidityCalculator(alpacaProvider),
		riskService:         services.NewRiskEngineService(),
	}
}

// CalculatePortfolioRisk performs comprehensive risk analysis
func (re *RiskEngine) CalculatePortfolioRisk(ctx context.Context, portfolioID uuid.UUID) (*PortfolioRiskReport, error) {
	// Calculate VaR
	varResult, err := re.riskService.CalculateVaR(services.VaRCalculationRequest{
		PortfolioID:     portfolioID,
		ConfidenceLevel: 95.0,
		TimeHorizon:     1,
		Method:          "monte_carlo",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to calculate VaR: %w", err)
	}

	// Calculate Liquidity Risk
	liquidityResult, err := re.riskService.CalculateLiquidityRisk(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate liquidity risk: %w", err)
	}

	// Check Position Limits
	positionLimits, err := re.riskService.CheckPositionLimits(portfolioID, 25.0) // 25% limit
	if err != nil {
		return nil, fmt.Errorf("failed to check position limits: %w", err)
	}

	// Compile comprehensive report
	report := &PortfolioRiskReport{
		PortfolioID:      portfolioID,
		CalculatedAt:     time.Now(),
		VaRMetrics:       varResult,
		LiquidityMetrics: liquidityResult,
		PositionLimits:   positionLimits,
		OverallRiskScore: calculateOverallRiskScore(varResult, liquidityResult, positionLimits),
	}

	return report, nil
}

// MonitorRealTimeRisk continuously monitors portfolio risk
func (re *RiskEngine) MonitorRealTimeRisk(ctx context.Context, portfolioID uuid.UUID, interval time.Duration) <-chan *RiskUpdate {
	updates := make(chan *RiskUpdate, 10)

	go func() {
		defer close(updates)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				report, err := re.CalculatePortfolioRisk(ctx, portfolioID)
				if err != nil {
					updates <- &RiskUpdate{
						PortfolioID: portfolioID,
						Error:       err,
						Timestamp:   time.Now(),
					}
					continue
				}

				updates <- &RiskUpdate{
					PortfolioID: portfolioID,
					Report:      report,
					Timestamp:   time.Now(),
				}
			}
		}
	}()

	return updates
}

// PortfolioRiskReport contains comprehensive risk analysis
type PortfolioRiskReport struct {
	PortfolioID      uuid.UUID                     `json:"portfolio_id"`
	CalculatedAt     time.Time                     `json:"calculated_at"`
	VaRMetrics       *services.VaRResult           `json:"var_metrics"`
	LiquidityMetrics *services.LiquidityResult     `json:"liquidity_metrics"`
	PositionLimits   *services.PositionLimitResult `json:"position_limits"`
	OverallRiskScore decimal.Decimal               `json:"overall_risk_score"`
}

// RiskUpdate represents a real-time risk update
type RiskUpdate struct {
	PortfolioID uuid.UUID            `json:"portfolio_id"`
	Report      *PortfolioRiskReport `json:"report,omitempty"`
	Error       error                `json:"error,omitempty"`
	Timestamp   time.Time            `json:"timestamp"`
}

// calculateOverallRiskScore computes a composite risk score
func calculateOverallRiskScore(var_ *services.VaRResult, liquidity *services.LiquidityResult, positions *services.PositionLimitResult) decimal.Decimal {
	score := decimal.Zero

	// VaR contribution (0-40 points)
	if var_ != nil {
		switch var_.Status {
		case "SAFE":
			score = score.Add(decimal.NewFromInt(10))
		case "WARNING":
			score = score.Add(decimal.NewFromInt(25))
		case "CRITICAL":
			score = score.Add(decimal.NewFromInt(40))
		}
	}

	// Liquidity contribution (0-30 points)
	if liquidity != nil {
		switch liquidity.RiskAssessment {
		case "LOW_RISK":
			score = score.Add(decimal.NewFromInt(5))
		case "MEDIUM_RISK":
			score = score.Add(decimal.NewFromInt(15))
		case "HIGH_RISK":
			score = score.Add(decimal.NewFromInt(30))
		}
	}

	// Position limits contribution (0-30 points)
	if positions != nil {
		violationCount := len(positions.Violations)
		if violationCount == 0 {
			score = score.Add(decimal.NewFromInt(5))
		} else if violationCount <= 2 {
			score = score.Add(decimal.NewFromInt(15))
		} else {
			score = score.Add(decimal.NewFromInt(30))
		}
	}

	return score
}
