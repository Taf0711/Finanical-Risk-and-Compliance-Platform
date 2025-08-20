package calculator

import (
	"fmt"
	"math"
	"math/rand/v2"
	"sort"

	"github.com/shopspring/decimal"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

type VARCalculator struct {
	ConfidenceLevel float64
	TimeHorizon     int // in days
}

func NewVARCalculator(confidenceLevel float64, timeHorizon int) *VARCalculator {
	return &VARCalculator{
		ConfidenceLevel: confidenceLevel,
		TimeHorizon:     timeHorizon,
	}
}

// CalculateVAR using Historical Simulation method
func (v *VARCalculator) CalculateVAR(portfolio *models.Portfolio, historicalReturns []float64) (decimal.Decimal, error) {
	if len(historicalReturns) == 0 {
		return decimal.Zero, fmt.Errorf("no historical returns provided")
	}

	// Sort returns in ascending order
	sort.Float64s(historicalReturns)

	// Calculate the index for the VaR percentile
	index := int(math.Floor(float64(len(historicalReturns)) * (1 - v.ConfidenceLevel)))
	if index < 0 {
		index = 0
	}
	if index >= len(historicalReturns) {
		index = len(historicalReturns) - 1
	}

	// Get the VaR value (potential loss)
	varValue := historicalReturns[index]

	// Scale by time horizon (assuming returns are daily)
	scaledVAR := varValue * math.Sqrt(float64(v.TimeHorizon))

	// Apply to portfolio value
	portfolioValue := portfolio.TotalValue.InexactFloat64()
	potentialLoss := portfolioValue * math.Abs(scaledVAR)

	return decimal.NewFromFloat(potentialLoss), nil
}

// CalculateParametricVAR using variance-covariance method
func (v *VARCalculator) CalculateParametricVAR(portfolio *models.Portfolio, volatility float64) decimal.Decimal {
	// Z-score for confidence level (e.g., 1.645 for 95% confidence)
	zScore := getNormalInverse(v.ConfidenceLevel)

	// Calculate VaR
	portfolioValue := portfolio.TotalValue.InexactFloat64()
	sqrtTime := math.Sqrt(float64(v.TimeHorizon) / 252.0) // Assuming 252 trading days

	varValue := portfolioValue * volatility * zScore * sqrtTime

	return decimal.NewFromFloat(varValue)
}

// Helper function to get normal inverse (simplified)
func getNormalInverse(confidenceLevel float64) float64 {
	// Simplified z-scores for common confidence levels
	switch {
	case confidenceLevel >= 0.99:
		return 2.326
	case confidenceLevel >= 0.95:
		return 1.645
	case confidenceLevel >= 0.90:
		return 1.282
	default:
		return 1.645 // Default to 95%
	}
}

// GenerateMockReturns for testing purposes
func GenerateMockReturns(days int) []float64 {
	returns := make([]float64, days)

	// Generate random returns with normal distribution characteristics
	for i := 0; i < days; i++ {
		// Simulate daily returns between -5% and +5%
		returns[i] = (rand.Float64() - 0.5) * 0.1
	}

	return returns
}
