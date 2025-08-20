package rules

import (
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

type PositionLimitChecker struct {
	MaxPositionPercent float64 // Maximum percentage for a single position
}

func NewPositionLimitChecker(maxPercent float64) *PositionLimitChecker {
	return &PositionLimitChecker{
		MaxPositionPercent: maxPercent,
	}
}

// CheckPositionLimits verifies if any position exceeds the limit
func (p *PositionLimitChecker) CheckPositionLimits(positions []models.Position) ([]PositionViolation, error) {
	violations := []PositionViolation{}

	totalValue := decimal.Zero
	for _, pos := range positions {
		totalValue = totalValue.Add(pos.MarketValue)
	}

	if totalValue.IsZero() {
		return violations, nil
	}

	maxAllowed := decimal.NewFromFloat(p.MaxPositionPercent / 100)

	for _, position := range positions {
		weight := position.MarketValue.Div(totalValue)

		if weight.GreaterThan(maxAllowed) {
			violations = append(violations, PositionViolation{
				Symbol:         position.Symbol,
				CurrentPercent: weight.Mul(decimal.NewFromInt(100)).InexactFloat64(),
				MaxPercent:     p.MaxPositionPercent,
				ExcessPercent:  weight.Sub(maxAllowed).Mul(decimal.NewFromInt(100)).InexactFloat64(),
				MarketValue:    position.MarketValue,
			})
		}
	}

	return violations, nil
}

type PositionViolation struct {
	Symbol         string
	CurrentPercent float64
	MaxPercent     float64
	ExcessPercent  float64
	MarketValue    decimal.Decimal
}

func (v PositionViolation) String() string {
	return fmt.Sprintf("Position %s exceeds limit: %.2f%% (max: %.2f%%, excess: %.2f%%)",
		v.Symbol, v.CurrentPercent, v.MaxPercent, v.ExcessPercent)
}
