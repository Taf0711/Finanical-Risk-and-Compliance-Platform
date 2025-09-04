// Enhanced VaR Calculator with improved statistical modeling
package calculator

import (
	"math"
	"sort"
)

// EnhancedVaRCalculator with dynamic parameter estimation
type EnhancedVaRCalculator struct {
	portfolioValue float64
	lookbackDays   int
}

// NewEnhancedVaRCalculator creates an improved VaR calculator
func NewEnhancedVaRCalculator(portfolioValue float64, lookbackDays int) *EnhancedVaRCalculator {
	if lookbackDays == 0 {
		lookbackDays = 252 // Default to 1 year
	}
	return &EnhancedVaRCalculator{
		portfolioValue: portfolioValue,
		lookbackDays:   lookbackDays,
	}
}

// EstimateGARCHVolatility implements GARCH(1,1) volatility estimation
func (e *EnhancedVaRCalculator) EstimateGARCHVolatility(returns []float64) float64 {
	if len(returns) < 10 {
		return e.calculateSimpleVolatility(returns)
	}

	// GARCH(1,1) parameters (simplified estimation)
	alpha0 := 0.00001 // Long-term variance
	alpha1 := 0.05    // ARCH parameter
	beta1 := 0.9      // GARCH parameter

	// Initialize with sample variance
	variance := e.calculateVariance(returns)

	// Apply GARCH recursion for last 30 observations
	start := len(returns) - 30
	if start < 0 {
		start = 0
	}

	for i := start; i < len(returns); i++ {
		variance = alpha0 + alpha1*returns[i]*returns[i] + beta1*variance
	}

	return math.Sqrt(variance * 252) // Annualized volatility
}

// EstimateCorrelationMatrix using exponential weighting
func (e *EnhancedVaRCalculator) EstimateCorrelationMatrix(assetReturns map[string][]float64) [][]float64 {
	symbols := make([]string, 0, len(assetReturns))
	for symbol := range assetReturns {
		symbols = append(symbols, symbol)
	}

	n := len(symbols)
	correlations := make([][]float64, n)
	for i := range correlations {
		correlations[i] = make([]float64, n)
	}

	// Exponential weighting parameter
	lambda := 0.94 // RiskMetrics standard

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i == j {
				correlations[i][j] = 1.0
			} else {
				correlations[i][j] = e.calculateExponentialCorrelation(
					assetReturns[symbols[i]],
					assetReturns[symbols[j]],
					lambda,
				)
			}
		}
	}

	return correlations
}

// CalculateEWMAVolatility uses Exponentially Weighted Moving Average
func (e *EnhancedVaRCalculator) CalculateEWMAVolatility(returns []float64, lambda float64) float64 {
	if len(returns) < 2 {
		return 0.0
	}

	if lambda == 0 {
		lambda = 0.94 // RiskMetrics standard
	}

	// Initialize with first squared return
	variance := returns[0] * returns[0]

	// Apply EWMA recursion
	for i := 1; i < len(returns); i++ {
		variance = lambda*variance + (1-lambda)*returns[i]*returns[i]
	}

	return math.Sqrt(variance * 252) // Annualized
}

// CalculateFilteredHistoricalVaR removes outliers before VaR calculation
func (e *EnhancedVaRCalculator) CalculateFilteredHistoricalVaR(returns []float64, confidence float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Remove extreme outliers (beyond 3 standard deviations)
	mean := e.calculateMean(returns)
	stdDev := e.calculateStdDev(returns, mean)

	filteredReturns := make([]float64, 0, len(returns))
	for _, ret := range returns {
		if math.Abs(ret-mean) <= 3*stdDev {
			filteredReturns = append(filteredReturns, ret)
		}
	}

	// Sort filtered returns
	sort.Float64s(filteredReturns)

	// Calculate VaR
	percentileIndex := int((1 - confidence) * float64(len(filteredReturns)))
	if percentileIndex >= len(filteredReturns) {
		percentileIndex = len(filteredReturns) - 1
	}

	return -filteredReturns[percentileIndex] * e.portfolioValue
}

// Helper functions
func (e *EnhancedVaRCalculator) calculateExponentialCorrelation(returns1, returns2 []float64, lambda float64) float64 {
	if len(returns1) != len(returns2) || len(returns1) < 2 {
		return 0.0
	}

	// Calculate exponentially weighted covariance and variances
	var cov, var1, var2 float64
	weight := 1.0
	totalWeight := 0.0

	for i := len(returns1) - 1; i >= 0; i-- {
		cov += weight * returns1[i] * returns2[i]
		var1 += weight * returns1[i] * returns1[i]
		var2 += weight * returns2[i] * returns2[i]
		totalWeight += weight
		weight *= lambda

		if weight < 0.001 { // Stop when weight becomes negligible
			break
		}
	}

	cov /= totalWeight
	var1 /= totalWeight
	var2 /= totalWeight

	if var1 <= 0 || var2 <= 0 {
		return 0.0
	}

	correlation := cov / math.Sqrt(var1*var2)

	// Bound correlation between -1 and 1
	if correlation > 1.0 {
		correlation = 1.0
	} else if correlation < -1.0 {
		correlation = -1.0
	}

	return correlation
}

func (e *EnhancedVaRCalculator) calculateSimpleVolatility(returns []float64) float64 {
	if len(returns) < 2 {
		return 0.0
	}

	mean := e.calculateMean(returns)
	variance := 0.0

	for _, ret := range returns {
		deviation := ret - mean
		variance += deviation * deviation
	}

	variance /= float64(len(returns) - 1)
	return math.Sqrt(variance * 252) // Annualized
}

func (e *EnhancedVaRCalculator) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, val := range values {
		sum += val
	}
	return sum / float64(len(values))
}

func (e *EnhancedVaRCalculator) calculateStdDev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	variance := e.calculateVariance(values)
	return math.Sqrt(variance)
}

func (e *EnhancedVaRCalculator) calculateVariance(values []float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	mean := e.calculateMean(values)
	variance := 0.0

	for _, val := range values {
		deviation := val - mean
		variance += deviation * deviation
	}

	return variance / float64(len(values)-1)
}
