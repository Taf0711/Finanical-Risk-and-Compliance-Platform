package calculator

import (
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// VaRCalculator handles Value at Risk calculations
type VaRCalculator struct {
	portfolioValue   float64
	confidenceLevels []float64
}

// NewVaRCalculator creates a new VaR calculator instance
func NewVaRCalculator(portfolioValue float64) *VaRCalculator {
	return &VaRCalculator{
		portfolioValue:   portfolioValue,
		confidenceLevels: []float64{0.95, 0.99}, // 95% and 99% confidence levels
	}
}

// CalculateVaR calculates Value at Risk using multiple methods
func (v *VaRCalculator) CalculateVaR(positions []models.Position, priceHistory map[string][]float64, timeHorizon int) (*VaRResult, error) {
	result := &VaRResult{
		TimeHorizon: timeHorizon,
	}

	// Calculate portfolio returns from price history
	portfolioReturns := v.calculatePortfolioReturns(positions, priceHistory)

	// Method 1: Historical Simulation
	historicalVaR := v.historicalVaR(portfolioReturns)
	result.HistoricalVaR95 = historicalVaR[0.95]
	result.HistoricalVaR99 = historicalVaR[0.99]

	// Method 2: Parametric VaR (assumes normal distribution)
	parametricVaR := v.parametricVaR(portfolioReturns)
	result.ParametricVaR95 = parametricVaR[0.95]
	result.ParametricVaR99 = parametricVaR[0.99]

	// Method 3: Monte Carlo Simulation
	monteCarloVaR := v.monteCarloVaR(positions, priceHistory, 10000) // 10,000 simulations
	result.MonteCarloVaR95 = monteCarloVaR[0.95]
	result.MonteCarloVaR99 = monteCarloVaR[0.99]

	// Use the average of all methods for final VaR
	result.VaR95 = (result.HistoricalVaR95 + result.ParametricVaR95 + result.MonteCarloVaR95) / 3
	result.VaR99 = (result.HistoricalVaR99 + result.ParametricVaR99 + result.MonteCarloVaR99) / 3

	// Calculate additional risk metrics
	result.ExpectedShortfall95 = v.calculateExpectedShortfall(portfolioReturns, 0.95)
	result.ExpectedShortfall99 = v.calculateExpectedShortfall(portfolioReturns, 0.99)
	result.MaxDrawdown = v.calculateMaxDrawdown(portfolioReturns)

	return result, nil
}

// calculatePortfolioReturns calculates historical returns for the portfolio
func (v *VaRCalculator) calculatePortfolioReturns(positions []models.Position, priceHistory map[string][]float64) []float64 {
	if len(priceHistory) == 0 {
		return []float64{}
	}

	// Find minimum history length
	minLength := math.MaxInt32
	for _, prices := range priceHistory {
		if len(prices) < minLength {
			minLength = len(prices)
		}
	}

	if minLength < 2 {
		return []float64{}
	}

	portfolioReturns := make([]float64, minLength-1)

	for i := 1; i < minLength; i++ {
		portfolioValueYesterday := 0.0
		portfolioValueToday := 0.0

		for _, position := range positions {
			prices, exists := priceHistory[position.Symbol]
			if !exists || len(prices) < minLength {
				continue
			}

			portfolioValueYesterday += position.Quantity.InexactFloat64() * prices[i-1]
			portfolioValueToday += position.Quantity.InexactFloat64() * prices[i]
		}

		if portfolioValueYesterday > 0 {
			portfolioReturns[i-1] = (portfolioValueToday - portfolioValueYesterday) / portfolioValueYesterday
		}
	}

	return portfolioReturns
}

// historicalVaR calculates VaR using historical simulation
func (v *VaRCalculator) historicalVaR(returns []float64) map[float64]float64 {
	if len(returns) == 0 {
		return map[float64]float64{0.95: 0, 0.99: 0}
	}

	// Sort returns in ascending order
	sortedReturns := make([]float64, len(returns))
	copy(sortedReturns, returns)
	sort.Float64s(sortedReturns)

	result := make(map[float64]float64)

	for _, confidence := range v.confidenceLevels {
		percentileIndex := int((1 - confidence) * float64(len(sortedReturns)))
		if percentileIndex >= len(sortedReturns) {
			percentileIndex = len(sortedReturns) - 1
		}

		// VaR is the loss at the percentile (negative return)
		varReturn := sortedReturns[percentileIndex]
		result[confidence] = -varReturn * v.portfolioValue
	}

	return result
}

// parametricVaR calculates VaR assuming normal distribution
func (v *VaRCalculator) parametricVaR(returns []float64) map[float64]float64 {
	if len(returns) == 0 {
		return map[float64]float64{0.95: 0, 0.99: 0}
	}

	mean := v.calculateMean(returns)
	stdDev := v.calculateStdDev(returns, mean)

	result := make(map[float64]float64)

	// Z-scores for confidence levels
	zScores := map[float64]float64{
		0.95: 1.645,
		0.99: 2.326,
	}

	for confidence, z := range zScores {
		varReturn := mean - z*stdDev
		result[confidence] = -varReturn * v.portfolioValue
	}

	return result
}

// monteCarloVaR calculates VaR using Monte Carlo simulation
func (v *VaRCalculator) monteCarloVaR(positions []models.Position, priceHistory map[string][]float64, numSimulations int) map[float64]float64 {
	if len(positions) == 0 || len(priceHistory) == 0 {
		return map[float64]float64{0.95: 0, 0.99: 0}
	}

	// Calculate returns for each asset
	assetReturns := make(map[string][]float64)
	assetStats := make(map[string]struct{ mean, stdDev float64 })

	for symbol, prices := range priceHistory {
		returns := v.calculateReturns(prices)
		if len(returns) > 0 {
			assetReturns[symbol] = returns
			mean := v.calculateMean(returns)
			stdDev := v.calculateStdDev(returns, mean)
			assetStats[symbol] = struct{ mean, stdDev float64 }{mean, stdDev}
		}
	}

	// Run Monte Carlo simulations
	simulatedPortfolioReturns := make([]float64, numSimulations)

	for i := 0; i < numSimulations; i++ {
		portfolioReturn := 0.0
		totalValue := 0.0

		for _, position := range positions {
			stats, exists := assetStats[position.Symbol]
			if !exists {
				continue
			}

			// Generate random return based on historical mean and std dev
			randomReturn := v.generateRandomReturn(stats.mean, stats.stdDev)
			positionValue := position.Quantity.InexactFloat64() * position.CurrentPrice.InexactFloat64()
			portfolioReturn += randomReturn * positionValue
			totalValue += positionValue
		}

		if totalValue > 0 {
			simulatedPortfolioReturns[i] = portfolioReturn / totalValue
		}
	}

	// Calculate VaR from simulated returns
	return v.historicalVaR(simulatedPortfolioReturns)
}

// calculateExpectedShortfall calculates the expected loss beyond VaR
func (v *VaRCalculator) calculateExpectedShortfall(returns []float64, confidence float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	sortedReturns := make([]float64, len(returns))
	copy(sortedReturns, returns)
	sort.Float64s(sortedReturns)

	percentileIndex := int((1 - confidence) * float64(len(sortedReturns)))
	if percentileIndex >= len(sortedReturns) {
		percentileIndex = len(sortedReturns) - 1
	}

	// Calculate average of returns worse than VaR
	sum := 0.0
	count := 0

	for i := 0; i <= percentileIndex; i++ {
		sum += sortedReturns[i]
		count++
	}

	if count > 0 {
		return -sum / float64(count) * v.portfolioValue
	}

	return 0
}

// calculateMaxDrawdown calculates the maximum peak-to-trough decline
func (v *VaRCalculator) calculateMaxDrawdown(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	cumulative := 1.0
	peak := 1.0
	maxDrawdown := 0.0

	for _, ret := range returns {
		cumulative *= (1 + ret)

		if cumulative > peak {
			peak = cumulative
		}

		drawdown := (peak - cumulative) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown * v.portfolioValue
}

// Helper functions
func (v *VaRCalculator) calculateReturns(prices []float64) []float64 {
	if len(prices) < 2 {
		return []float64{}
	}

	returns := make([]float64, len(prices)-1)
	for i := 1; i < len(prices); i++ {
		if prices[i-1] > 0 {
			returns[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
		}
	}

	return returns
}

func (v *VaRCalculator) calculateMean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

func (v *VaRCalculator) calculateStdDev(data []float64, mean float64) float64 {
	if len(data) < 2 {
		return 0
	}

	variance := 0.0
	for _, v := range data {
		variance += math.Pow(v-mean, 2)
	}
	variance /= float64(len(data) - 1)

	return math.Sqrt(variance)
}

func (v *VaRCalculator) generateRandomReturn(mean, stdDev float64) float64 {
	// Box-Muller transform for normal distribution
	u1 := math.Max(1e-10, rand.Float64())
	u2 := rand.Float64()

	z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
	return mean + z*stdDev
}

// VaRResult contains the calculated VaR metrics
type VaRResult struct {
	TimeHorizon         int     `json:"time_horizon"`
	VaR95               float64 `json:"var_95"`
	VaR99               float64 `json:"var_99"`
	HistoricalVaR95     float64 `json:"historical_var_95"`
	HistoricalVaR99     float64 `json:"historical_var_99"`
	ParametricVaR95     float64 `json:"parametric_var_95"`
	ParametricVaR99     float64 `json:"parametric_var_99"`
	MonteCarloVaR95     float64 `json:"monte_carlo_var_95"`
	MonteCarloVaR99     float64 `json:"monte_carlo_var_99"`
	ExpectedShortfall95 float64 `json:"expected_shortfall_95"`
	ExpectedShortfall99 float64 `json:"expected_shortfall_99"`
	MaxDrawdown         float64 `json:"max_drawdown"`
}
