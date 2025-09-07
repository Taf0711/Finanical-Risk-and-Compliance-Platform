package calculator

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// ProfessionalVaRCalculator implements institutional-grade VaR calculations
type ProfessionalVaRCalculator struct {
	portfolioValue   float64
	confidenceLevels []float64
	timeHorizon      int // in days
	mu               sync.RWMutex
}

// NewProfessionalVaRCalculator creates a new professional VaR calculator
func NewProfessionalVaRCalculator(portfolioValue float64, timeHorizon int) *ProfessionalVaRCalculator {
	return &ProfessionalVaRCalculator{
		portfolioValue:   portfolioValue,
		confidenceLevels: []float64{0.90, 0.95, 0.99}, // 90%, 95%, 99% confidence
		timeHorizon:      timeHorizon,
	}
}

// VaRResult contains comprehensive VaR metrics
type ProfessionalVaRResult struct {
	TimeHorizon        int                 `json:"time_horizon"`
	ConfidenceLevel    float64             `json:"confidence_level"`
	HistoricalVaR      map[float64]float64 `json:"historical_var"`
	ParametricVaR      map[float64]float64 `json:"parametric_var"`
	MonteCarloVaR      map[float64]float64 `json:"monte_carlo_var"`
	CornishFisherVaR   map[float64]float64 `json:"cornish_fisher_var"`
	ConditionalVaR     map[float64]float64 `json:"conditional_var"`
	ComponentVaR       map[string]float64  `json:"component_var"`
	MarginalVaR        map[string]float64  `json:"marginal_var"`
	IncrementalVaR     map[string]float64  `json:"incremental_var"`
	StressVaR          float64             `json:"stress_var"`
	Volatility         float64             `json:"volatility"`
	Skewness           float64             `json:"skewness"`
	Kurtosis           float64             `json:"kurtosis"`
	MaxDrawdown        float64             `json:"max_drawdown"`
	TailRisk           float64             `json:"tail_risk"`
	BacktestingMetrics *BacktestingResult  `json:"backtesting_metrics"`
}

// CalculateProfessionalVaR performs comprehensive VaR calculation
func (v *ProfessionalVaRCalculator) CalculateProfessionalVaR(
	positions []models.Position,
	priceHistory map[string][]float64,
) (*ProfessionalVaRResult, error) {

	// Calculate returns matrix
	returnsMatrix := v.calculateReturnsMatrix(positions, priceHistory)
	if len(returnsMatrix) == 0 {
		return nil, fmt.Errorf("insufficient data for VaR calculation")
	}

	// Calculate portfolio returns
	portfolioReturns := v.calculatePortfolioReturns(positions, returnsMatrix)

	result := &ProfessionalVaRResult{
		TimeHorizon:      v.timeHorizon,
		HistoricalVaR:    make(map[float64]float64),
		ParametricVaR:    make(map[float64]float64),
		MonteCarloVaR:    make(map[float64]float64),
		CornishFisherVaR: make(map[float64]float64),
		ConditionalVaR:   make(map[float64]float64),
		ComponentVaR:     make(map[string]float64),
		MarginalVaR:      make(map[string]float64),
		IncrementalVaR:   make(map[string]float64),
	}

	// Calculate statistical moments
	result.Volatility = v.calculateVolatility(portfolioReturns)
	result.Skewness = v.calculateSkewness(portfolioReturns)
	result.Kurtosis = v.calculateKurtosis(portfolioReturns)
	result.MaxDrawdown = v.calculateMaxDrawdown(portfolioReturns)

	// Calculate VaR using different methods
	for _, confidence := range v.confidenceLevels {
		// 1. Historical Simulation
		result.HistoricalVaR[confidence] = v.historicalVaR(portfolioReturns, confidence)

		// 2. Parametric VaR (Delta-Normal)
		result.ParametricVaR[confidence] = v.parametricVaR(portfolioReturns, confidence)

		// 3. Cornish-Fisher VaR (adjusts for skewness and kurtosis)
		result.CornishFisherVaR[confidence] = v.cornishFisherVaR(
			portfolioReturns, confidence, result.Skewness, result.Kurtosis,
		)

		// 4. Conditional VaR (Expected Shortfall)
		result.ConditionalVaR[confidence] = v.conditionalVaR(portfolioReturns, confidence)

		// 5. Monte Carlo VaR with proper correlation
		result.MonteCarloVaR[confidence] = v.monteCarloVaRWithCorrelation(
			positions, returnsMatrix, confidence, 10000,
		)
	}

	// Calculate risk decomposition
	v.calculateRiskDecomposition(result, positions, returnsMatrix)

	// Calculate Stress VaR
	result.StressVaR = v.calculateStressVaR(portfolioReturns)

	// Calculate Tail Risk metrics
	result.TailRisk = v.calculateTailRisk(portfolioReturns)

	// Perform backtesting
	result.BacktestingMetrics = v.performBacktesting(portfolioReturns, result.HistoricalVaR[0.99])

	return result, nil
}

// historicalVaR calculates VaR using historical simulation
func (v *ProfessionalVaRCalculator) historicalVaR(returns []float64, confidence float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Sort returns in ascending order
	sortedReturns := make([]float64, len(returns))
	copy(sortedReturns, returns)
	sort.Float64s(sortedReturns)

	// Calculate the percentile
	percentileIndex := int(math.Floor((1 - confidence) * float64(len(sortedReturns))))
	if percentileIndex >= len(sortedReturns) {
		percentileIndex = len(sortedReturns) - 1
	}

	// Scale by time horizon using square root rule
	timeScaling := math.Sqrt(float64(v.timeHorizon))

	// VaR is the loss at the percentile (negative return)
	varReturn := -sortedReturns[percentileIndex] * timeScaling

	return varReturn * v.portfolioValue
}

// parametricVaR calculates VaR assuming normal distribution
func (v *ProfessionalVaRCalculator) parametricVaR(returns []float64, confidence float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Calculate mean and standard deviation
	mean := stat.Mean(returns, nil)
	stdDev := stat.StdDev(returns, nil)

	// Create normal distribution
	normal := distuv.Normal{
		Mu:    mean,
		Sigma: stdDev,
	}

	// Get the quantile for the confidence level
	quantile := normal.Quantile(1 - confidence)

	// Scale by time horizon
	timeScaling := math.Sqrt(float64(v.timeHorizon))
	scaledVaR := -quantile * timeScaling

	return scaledVaR * v.portfolioValue
}

// cornishFisherVaR adjusts for non-normal distributions
func (v *ProfessionalVaRCalculator) cornishFisherVaR(
	returns []float64, confidence float64, skewness float64, kurtosis float64,
) float64 {
	if len(returns) == 0 {
		return 0
	}

	mean := stat.Mean(returns, nil)
	stdDev := stat.StdDev(returns, nil)

	// Standard normal quantile
	normal := distuv.Normal{Mu: 0, Sigma: 1}
	z := normal.Quantile(1 - confidence)

	// Cornish-Fisher expansion
	// Adjusts the normal quantile for skewness and excess kurtosis
	excessKurtosis := kurtosis - 3

	cfQuantile := z +
		(z*z-1)*skewness/6 +
		(z*z*z-3*z)*excessKurtosis/24 -
		(2*z*z*z-5*z)*skewness*skewness/36

	// Calculate adjusted VaR
	timeScaling := math.Sqrt(float64(v.timeHorizon))
	adjustedVaR := mean - cfQuantile*stdDev*timeScaling

	return -adjustedVaR * v.portfolioValue
}

// conditionalVaR calculates Expected Shortfall (CVaR)
func (v *ProfessionalVaRCalculator) conditionalVaR(returns []float64, confidence float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Sort returns
	sortedReturns := make([]float64, len(returns))
	copy(sortedReturns, returns)
	sort.Float64s(sortedReturns)

	// Find VaR threshold index
	varIndex := int(math.Floor((1 - confidence) * float64(len(sortedReturns))))

	// Calculate average of all returns worse than VaR
	sum := 0.0
	count := 0

	for i := 0; i <= varIndex; i++ {
		sum += sortedReturns[i]
		count++
	}

	if count == 0 {
		return 0
	}

	// Scale by time horizon
	timeScaling := math.Sqrt(float64(v.timeHorizon))
	expectedShortfall := -sum / float64(count) * timeScaling

	return expectedShortfall * v.portfolioValue
}

// monteCarloVaRWithCorrelation performs correlated Monte Carlo simulation
func (v *ProfessionalVaRCalculator) monteCarloVaRWithCorrelation(
	positions []models.Position,
	returnsMatrix map[string][]float64,
	confidence float64,
	numSimulations int,
) float64 {

	// Calculate correlation matrix
	correlationMatrix := v.calculateCorrelationMatrix(positions, returnsMatrix)

	// Perform Cholesky decomposition for correlation
	var cholesky mat.Cholesky
	ok := cholesky.Factorize(correlationMatrix)
	if !ok {
		// Fall back to independent simulation if correlation matrix is not positive definite
		return v.monteCarloVaRIndependent(positions, returnsMatrix, confidence, numSimulations)
	}

	// Generate correlated random returns
	simulatedPortfolioReturns := make([]float64, numSimulations)

	for sim := 0; sim < numSimulations; sim++ {
		// Generate independent standard normal random variables
		n := len(positions)
		independent := make([]float64, n)
		for i := 0; i < n; i++ {
			independent[i] = rand.NormFloat64()
		}

		// Apply Cholesky transformation to get correlated returns
		correlatedVec := mat.NewVecDense(n, nil)
		independentVec := mat.NewVecDense(n, independent)
		L := &mat.TriDense{}
		cholesky.LTo(L)
		correlatedVec.MulVec(L, independentVec)

		// Calculate portfolio return for this simulation
		portfolioReturn := 0.0
		totalWeight := 0.0

		for i, pos := range positions {
			symbol := pos.Symbol
			if returns, exists := returnsMatrix[symbol]; exists && len(returns) > 0 {
				mean := stat.Mean(returns, nil)
				stdDev := stat.StdDev(returns, nil)

				// Scale the correlated random number
				assetReturn := mean + stdDev*correlatedVec.AtVec(i)

				weight := pos.MarketValue.InexactFloat64() / v.portfolioValue
				portfolioReturn += assetReturn * weight
				totalWeight += weight
			}
		}

		if totalWeight > 0 {
			simulatedPortfolioReturns[sim] = portfolioReturn
		}
	}

	// Calculate VaR from simulated returns
	sort.Float64s(simulatedPortfolioReturns)
	varIndex := int(math.Floor((1 - confidence) * float64(numSimulations)))

	timeScaling := math.Sqrt(float64(v.timeHorizon))
	monteCarloVaR := -simulatedPortfolioReturns[varIndex] * timeScaling

	return monteCarloVaR * v.portfolioValue
}

// calculateRiskDecomposition calculates component, marginal, and incremental VaR
func (v *ProfessionalVaRCalculator) calculateRiskDecomposition(
	result *ProfessionalVaRResult,
	positions []models.Position,
	returnsMatrix map[string][]float64,
) {

	// Calculate covariance matrix
	covMatrix := v.calculateCovarianceMatrix(positions, returnsMatrix)

	// Portfolio weights
	weights := make([]float64, len(positions))
	for i, pos := range positions {
		weights[i] = pos.MarketValue.InexactFloat64() / v.portfolioValue
	}

	// Portfolio variance
	portfolioVariance := 0.0
	for i := range positions {
		for j := range positions {
			portfolioVariance += weights[i] * weights[j] * covMatrix.At(i, j)
		}
	}
	portfolioStdDev := math.Sqrt(portfolioVariance)

	// Calculate Marginal VaR for each position
	// Marginal VaR = ∂VaR/∂wi
	z99 := 2.326 // 99% confidence z-score
	timeScaling := math.Sqrt(float64(v.timeHorizon))

	for i, pos := range positions {
		// Marginal contribution to portfolio risk
		marginalContribution := 0.0
		for j := range positions {
			marginalContribution += weights[j] * covMatrix.At(i, j)
		}
		marginalContribution = marginalContribution / portfolioStdDev

		// Marginal VaR
		marginalVaR := z99 * marginalContribution * timeScaling * v.portfolioValue
		result.MarginalVaR[pos.Symbol] = marginalVaR

		// Component VaR = Marginal VaR × Position Weight
		componentVaR := marginalVaR * weights[i]
		result.ComponentVaR[pos.Symbol] = componentVaR
	}

	// Calculate Incremental VaR (simplified - would need full recalculation in practice)
	for _, pos := range positions {
		// Approximate incremental VaR as position weight times marginal VaR
		incrementalVaR := result.MarginalVaR[pos.Symbol] * 0.01 // 1% position change
		result.IncrementalVaR[pos.Symbol] = incrementalVaR
	}
}

// calculateStressVaR calculates VaR under stressed market conditions
func (v *ProfessionalVaRCalculator) calculateStressVaR(returns []float64) float64 {
	if len(returns) < 20 {
		return 0
	}

	// Find the worst 10% period in the returns
	windowSize := len(returns) / 10
	if windowSize < 10 {
		windowSize = 10
	}

	worstWindowStdDev := 0.0

	for i := 0; i <= len(returns)-windowSize; i++ {
		window := returns[i : i+windowSize]
		stdDev := stat.StdDev(window, nil)
		if stdDev > worstWindowStdDev {
			worstWindowStdDev = stdDev
		}
	}

	// Stress VaR uses the worst period volatility
	z99 := 2.326 // 99% confidence
	timeScaling := math.Sqrt(float64(v.timeHorizon))
	stressVaR := z99 * worstWindowStdDev * timeScaling

	return stressVaR * v.portfolioValue
}

// calculateTailRisk measures extreme tail risk
func (v *ProfessionalVaRCalculator) calculateTailRisk(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Sort returns
	sortedReturns := make([]float64, len(returns))
	copy(sortedReturns, returns)
	sort.Float64s(sortedReturns)

	// Calculate the average of the worst 1% of returns
	tailSize := int(math.Max(1, float64(len(returns))*0.01))

	tailSum := 0.0
	for i := 0; i < tailSize; i++ {
		tailSum += sortedReturns[i]
	}

	tailRisk := -tailSum / float64(tailSize)
	timeScaling := math.Sqrt(float64(v.timeHorizon))

	return tailRisk * timeScaling * v.portfolioValue
}

// BacktestingResult contains VaR backtesting metrics
type BacktestingResult struct {
	NumViolations       int     `json:"num_violations"`
	ViolationRate       float64 `json:"violation_rate"`
	ExpectedViolations  float64 `json:"expected_violations"`
	KupiecTestStatistic float64 `json:"kupiec_test_statistic"`
	KupiecTestPassed    bool    `json:"kupiec_test_passed"`
	TrafficLight        string  `json:"traffic_light"` // Green, Yellow, Red
}

// performBacktesting validates VaR model accuracy
func (v *ProfessionalVaRCalculator) performBacktesting(returns []float64, var99 float64) *BacktestingResult {
	if len(returns) < 250 { // Need at least 250 observations
		return nil
	}

	// Count VaR violations
	violations := 0
	varThreshold := -var99 / v.portfolioValue // Convert to return

	for _, ret := range returns {
		if ret < varThreshold {
			violations++
		}
	}

	n := float64(len(returns))
	p := 0.01 // 1% for 99% VaR
	expectedViolations := n * p
	violationRate := float64(violations) / n

	// Kupiec POF Test (Proportion of Failures)
	// Test statistic: -2*ln(likelihood ratio)
	var kupiecStat float64
	if violations > 0 {
		kupiecStat = -2 * math.Log(
			math.Pow(p, float64(violations))*math.Pow(1-p, n-float64(violations))/
				math.Pow(violationRate, float64(violations))*math.Pow(1-violationRate, n-float64(violations)),
		)
	}

	// Critical value at 95% confidence (chi-square with 1 df)
	criticalValue := 3.841
	kupiecPassed := kupiecStat < criticalValue

	// Basel Traffic Light
	var trafficLight string
	if violations <= 4 {
		trafficLight = "Green"
	} else if violations <= 9 {
		trafficLight = "Yellow"
	} else {
		trafficLight = "Red"
	}

	return &BacktestingResult{
		NumViolations:       violations,
		ViolationRate:       violationRate,
		ExpectedViolations:  expectedViolations,
		KupiecTestStatistic: kupiecStat,
		KupiecTestPassed:    kupiecPassed,
		TrafficLight:        trafficLight,
	}
}

// Helper functions for calculations

func (v *ProfessionalVaRCalculator) calculateReturnsMatrix(
	positions []models.Position,
	priceHistory map[string][]float64,
) map[string][]float64 {

	returnsMatrix := make(map[string][]float64)

	for _, pos := range positions {
		if prices, exists := priceHistory[pos.Symbol]; exists && len(prices) > 1 {
			returns := make([]float64, len(prices)-1)
			for i := 1; i < len(prices); i++ {
				if prices[i-1] != 0 {
					returns[i-1] = (prices[i] - prices[i-1]) / prices[i-1]
				}
			}
			returnsMatrix[pos.Symbol] = returns
		}
	}

	return returnsMatrix
}

func (v *ProfessionalVaRCalculator) calculatePortfolioReturns(
	positions []models.Position,
	returnsMatrix map[string][]float64,
) []float64 {

	// Find minimum return series length
	minLength := math.MaxInt32
	for _, returns := range returnsMatrix {
		if len(returns) < minLength {
			minLength = len(returns)
		}
	}

	if minLength == 0 || minLength == math.MaxInt32 {
		return []float64{}
	}

	// Calculate weighted portfolio returns
	portfolioReturns := make([]float64, minLength)

	for t := 0; t < minLength; t++ {
		portfolioReturn := 0.0
		totalWeight := 0.0

		for _, pos := range positions {
			if returns, exists := returnsMatrix[pos.Symbol]; exists && len(returns) > t {
				weight := pos.MarketValue.InexactFloat64() / v.portfolioValue
				portfolioReturn += returns[t] * weight
				totalWeight += weight
			}
		}

		if totalWeight > 0 {
			portfolioReturns[t] = portfolioReturn
		}
	}

	return portfolioReturns
}

func (v *ProfessionalVaRCalculator) calculateCorrelationMatrix(
	positions []models.Position,
	returnsMatrix map[string][]float64,
) *mat.SymDense {

	n := len(positions)
	correlation := mat.NewSymDense(n, nil)

	for i, pos1 := range positions {
		returns1, exists1 := returnsMatrix[pos1.Symbol]
		if !exists1 {
			continue
		}

		for j, pos2 := range positions {
			if j < i {
				continue // Symmetric matrix
			}

			returns2, exists2 := returnsMatrix[pos2.Symbol]
			if !exists2 {
				if i == j {
					correlation.SetSym(i, j, 1.0) // Diagonal
				}
				continue
			}

			// Calculate correlation
			corr := stat.Correlation(returns1, returns2, nil)
			correlation.SetSym(i, j, corr)
		}
	}

	return correlation
}

func (v *ProfessionalVaRCalculator) calculateCovarianceMatrix(
	positions []models.Position,
	returnsMatrix map[string][]float64,
) *mat.SymDense {

	n := len(positions)
	covariance := mat.NewSymDense(n, nil)

	for i, pos1 := range positions {
		returns1, exists1 := returnsMatrix[pos1.Symbol]
		if !exists1 {
			continue
		}

		for j, pos2 := range positions {
			if j < i {
				continue // Symmetric matrix
			}

			returns2, exists2 := returnsMatrix[pos2.Symbol]
			if !exists2 {
				continue
			}

			// Calculate covariance
			cov := stat.Covariance(returns1, returns2, nil)
			covariance.SetSym(i, j, cov)
		}
	}

	return covariance
}

func (v *ProfessionalVaRCalculator) calculateVolatility(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// Annualized volatility
	dailyVol := stat.StdDev(returns, nil)
	return dailyVol * math.Sqrt(252) // Assuming 252 trading days
}

func (v *ProfessionalVaRCalculator) calculateSkewness(returns []float64) float64 {
	if len(returns) < 3 {
		return 0
	}

	mean := stat.Mean(returns, nil)
	stdDev := stat.StdDev(returns, nil)

	if stdDev == 0 {
		return 0
	}

	n := float64(len(returns))
	sum := 0.0

	for _, r := range returns {
		z := (r - mean) / stdDev
		sum += z * z * z
	}

	return (n / ((n - 1) * (n - 2))) * sum
}

func (v *ProfessionalVaRCalculator) calculateKurtosis(returns []float64) float64 {
	if len(returns) < 4 {
		return 3.0 // Normal distribution kurtosis
	}

	mean := stat.Mean(returns, nil)
	stdDev := stat.StdDev(returns, nil)

	if stdDev == 0 {
		return 3.0
	}

	n := float64(len(returns))
	sum := 0.0

	for _, r := range returns {
		z := (r - mean) / stdDev
		sum += z * z * z * z
	}

	// Excess kurtosis (subtract 3 for normal distribution reference)
	kurtosis := (n*(n+1))/((n-1)*(n-2)*(n-3))*sum -
		(3*(n-1)*(n-1))/((n-2)*(n-3))

	return kurtosis + 3 // Return full kurtosis (not excess)
}

func (v *ProfessionalVaRCalculator) calculateMaxDrawdown(returns []float64) float64 {
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

	return maxDrawdown
}

// monteCarloVaRIndependent is a fallback for when correlation matrix is not positive definite
func (v *ProfessionalVaRCalculator) monteCarloVaRIndependent(
	positions []models.Position,
	returnsMatrix map[string][]float64,
	confidence float64,
	numSimulations int,
) float64 {

	simulatedReturns := make([]float64, numSimulations)

	for sim := 0; sim < numSimulations; sim++ {
		portfolioReturn := 0.0

		for _, pos := range positions {
			if returns, exists := returnsMatrix[pos.Symbol]; exists && len(returns) > 0 {
				// Bootstrap from historical returns
				randomIndex := rand.Intn(len(returns))
				assetReturn := returns[randomIndex]

				weight := pos.MarketValue.InexactFloat64() / v.portfolioValue
				portfolioReturn += assetReturn * weight
			}
		}

		simulatedReturns[sim] = portfolioReturn
	}

	// Calculate VaR from simulated returns
	sort.Float64s(simulatedReturns)
	varIndex := int(math.Floor((1 - confidence) * float64(numSimulations)))

	timeScaling := math.Sqrt(float64(v.timeHorizon))
	return -simulatedReturns[varIndex] * timeScaling * v.portfolioValue
}
