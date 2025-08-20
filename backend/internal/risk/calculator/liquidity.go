package calculator

import (
	"github.com/shopspring/decimal"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

type LiquidityCalculator struct {
	HighLiquidityThreshold   float64 // e.g., 0.7 (70% of portfolio)
	MediumLiquidityThreshold float64 // e.g., 0.2 (20% of portfolio)
}

func NewLiquidityCalculator() *LiquidityCalculator {
	return &LiquidityCalculator{
		HighLiquidityThreshold:   0.7,
		MediumLiquidityThreshold: 0.2,
	}
}

// CalculateLiquidityRatio calculates the percentage of liquid assets in the portfolio
func (l *LiquidityCalculator) CalculateLiquidityRatio(positions []models.Position) (decimal.Decimal, map[string]decimal.Decimal) {
	totalValue := decimal.Zero
	liquidityBreakdown := map[string]decimal.Decimal{
		"HIGH":   decimal.Zero,
		"MEDIUM": decimal.Zero,
		"LOW":    decimal.Zero,
	}

	for _, position := range positions {
		totalValue = totalValue.Add(position.MarketValue)
		liquidityBreakdown[position.Liquidity] = liquidityBreakdown[position.Liquidity].Add(position.MarketValue)
	}

	if totalValue.IsZero() {
		return decimal.Zero, liquidityBreakdown
	}

	// Calculate liquid ratio (HIGH liquidity assets / total)
	liquidRatio := liquidityBreakdown["HIGH"].Div(totalValue)

	return liquidRatio, liquidityBreakdown
}

// AssessLiquidityRisk determines the liquidity risk level
func (l *LiquidityCalculator) AssessLiquidityRisk(liquidityRatio decimal.Decimal) string {
	ratio := liquidityRatio.InexactFloat64()

	switch {
	case ratio >= l.HighLiquidityThreshold:
		return "LOW_RISK"
	case ratio >= l.MediumLiquidityThreshold:
		return "MEDIUM_RISK"
	default:
		return "HIGH_RISK"
	}
}

// CalculateLiquidationTime estimates time to liquidate positions
func (l *LiquidityCalculator) CalculateLiquidationTime(positions []models.Position) map[string]int {
	liquidationTimes := map[string]int{
		"HIGH":   1,  // 1 day
		"MEDIUM": 5,  // 5 days
		"LOW":    30, // 30 days
	}

	timeBreakdown := make(map[string]int)

	for _, position := range positions {
		symbol := position.Symbol
		timeBreakdown[symbol] = liquidationTimes[position.Liquidity]
	}

	return timeBreakdown
}
