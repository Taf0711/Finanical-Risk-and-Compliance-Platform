// backend/internal/risk/simulation/monte_carlo.go
package simulation

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/calculator"
	"github.com/shopspring/decimal"
)

// MonteCarloConfig holds simulation configuration
type MonteCarloConfig struct {
	NumSimulations    int     `json:"num_simulations"`
	TimeHorizonDays   int     `json:"time_horizon_days"`
	ConfidenceLevel   float64 `json:"confidence_level"`
	NumWorkers        int     `json:"num_workers"`
	RandomSeed        int64   `json:"random_seed"`
	MarketRegime      string  `json:"market_regime"` // NORMAL, STRESSED, CRISIS
	CorrelationMatrix bool    `json:"correlation_matrix"`
}

// SimulationResult holds the results of Monte Carlo simulation
type SimulationResult struct {
	Config             MonteCarloConfig     `json:"config"`
	StartTime          time.Time            `json:"start_time"`
	EndTime            time.Time            `json:"end_time"`
	Duration           time.Duration        `json:"duration"`
	PortfolioValue     float64              `json:"portfolio_value"`
	SimulatedReturns   []float64            `json:"simulated_returns"`
	VaREstimates       map[string]float64   `json:"var_estimates"`
	ExpectedShortfall  map[string]float64   `json:"expected_shortfall"`
	MaxDrawdown        float64              `json:"max_drawdown"`
	Volatility         float64              `json:"volatility"`
	Skewness           float64              `json:"skewness"`
	Kurtosis           float64              `json:"kurtosis"`
	LiquidityMetrics   *LiquiditySimResults `json:"liquidity_metrics"`
	AccuracyValidation *AccuracyValidation  `json:"accuracy_validation"`
	PerformanceMetrics *PerformanceMetrics  `json:"performance_metrics"`
	RiskDecomposition  map[string]float64   `json:"risk_decomposition"`
	StresTesting       *StressTestResults   `json:"stress_testing"`
}

type LiquiditySimResults struct {
	AverageLiquidityScore   float64            `json:"average_liquidity_score"`
	LiquidityVaR            map[string]float64 `json:"liquidity_var"`
	TimeToLiquidate         map[string]float64 `json:"time_to_liquidate"`
	LiquidityDistribution   []float64          `json:"liquidity_distribution"`
	PositionLiquidityImpact map[string]float64 `json:"position_liquidity_impact"`
}

type AccuracyValidation struct {
	BacktestingResults        map[string]float64 `json:"backtesting_results"`
	VaRViolations             map[string]int     `json:"var_violations"`
	VaRViolationRate          map[string]float64 `json:"var_violation_rate"`
	ExpectedViolations        map[string]int     `json:"expected_violations"`
	KupiecTestResults         map[string]bool    `json:"kupiec_test_results"`
	ChristoffersenTestResults map[string]bool    `json:"christoffersen_test_results"`
}

type PerformanceMetrics struct {
	SimulationTime       time.Duration `json:"simulation_time"`
	MemoryUsage          int64         `json:"memory_usage"`
	CPUUsage             float64       `json:"cpu_usage"`
	SimulationsPerSecond float64       `json:"simulations_per_second"`
	ConvergenceRate      float64       `json:"convergence_rate"`
}

type StressTestResults struct {
	NormalMarketVaR   map[string]float64 `json:"normal_market_var"`
	StressedMarketVaR map[string]float64 `json:"stressed_market_var"`
	CrisisMarketVaR   map[string]float64 `json:"crisis_market_var"`
	TailRiskMetrics   map[string]float64 `json:"tail_risk_metrics"`
}

// MonteCarloSimulator handles Monte Carlo simulations for risk assessment
type MonteCarloSimulator struct {
	varCalculator       *calculator.VaRCalculator
	liquidityCalculator *calculator.LiquidityCalculator
	priceGenerator      *PriceGenerator
}

// NewMonteCarloSimulator creates a new Monte Carlo simulator
func NewMonteCarloSimulator(varCalc *calculator.VaRCalculator, liquidityCalc *calculator.LiquidityCalculator) *MonteCarloSimulator {
	return &MonteCarloSimulator{
		varCalculator:       varCalc,
		liquidityCalculator: liquidityCalc,
		priceGenerator:      NewPriceGenerator(),
	}
}

// RunSimulation executes the Monte Carlo simulation
func (mc *MonteCarloSimulator) RunSimulation(positions []models.Position, config MonteCarloConfig) (*SimulationResult, error) {
	startTime := time.Now()

	// Set random seed for reproducibility
	if config.RandomSeed > 0 {
		// Create a new random source with the seed
		rng := rand.New(rand.NewSource(config.RandomSeed))
		// Store the seeded generator for use in simulations
		mc.priceGenerator.rng = rng
	}

	// Initialize result structure
	result := &SimulationResult{
		Config:            config,
		StartTime:         startTime,
		VaREstimates:      make(map[string]float64),
		ExpectedShortfall: make(map[string]float64),
		RiskDecomposition: make(map[string]float64),
	}

	// Calculate initial portfolio value
	portfolioValue := mc.calculatePortfolioValue(positions)
	result.PortfolioValue = portfolioValue

	// Generate price scenarios
	priceScenarios, err := mc.generatePriceScenarios(positions, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate price scenarios: %w", err)
	}

	// Run simulations in parallel
	simulationResults := mc.runParallelSimulations(positions, priceScenarios, config)

	// Process simulation results
	mc.processSimulationResults(result, simulationResults, config)

	// Perform accuracy validation
	result.AccuracyValidation = mc.validateAccuracy(positions, simulationResults, config)

	// Run stress tests
	result.StresTesting = mc.performStressTests(positions, config)

	// Calculate liquidity metrics
	result.LiquidityMetrics = mc.calculateLiquidityMetrics(positions, config)

	// Calculate performance metrics
	result.EndTime = time.Now()
	result.Duration = result.EndTime.Sub(result.StartTime)
	result.PerformanceMetrics = mc.calculatePerformanceMetrics(result.Duration, config)

	return result, nil
}

// generatePriceScenarios creates price scenarios for simulation
func (mc *MonteCarloSimulator) generatePriceScenarios(positions []models.Position, config MonteCarloConfig) ([]map[string][]float64, error) {
	scenarios := make([]map[string][]float64, config.NumSimulations)

	// Extract symbols and current prices
	symbols := make([]string, len(positions))
	currentPrices := make(map[string]float64)
	volatilities := make(map[string]float64)

	for i, pos := range positions {
		symbols[i] = pos.Symbol
		currentPrices[pos.Symbol] = pos.CurrentPrice.InexactFloat64()
		volatilities[pos.Symbol] = mc.estimateVolatility(pos.Symbol, config.MarketRegime)
	}

	// Generate correlation matrix if enabled
	var correlationMatrix [][]float64
	if config.CorrelationMatrix {
		correlationMatrix = mc.generateCorrelationMatrix(symbols)
	}

	// Generate price paths for each simulation
	for i := 0; i < config.NumSimulations; i++ {
		scenario := make(map[string][]float64)

		for _, symbol := range symbols {
			pricePath := mc.priceGenerator.GeneratePricePath(
				currentPrices[symbol],
				volatilities[symbol],
				config.TimeHorizonDays,
				correlationMatrix,
			)
			scenario[symbol] = pricePath
		}

		scenarios[i] = scenario
	}

	return scenarios, nil
}

// runParallelSimulations executes simulations in parallel using worker goroutines
func (mc *MonteCarloSimulator) runParallelSimulations(positions []models.Position, scenarios []map[string][]float64, config MonteCarloConfig) []SimulationOutcome {
	numWorkers := config.NumWorkers
	if numWorkers <= 0 {
		numWorkers = 4 // Default to 4 workers
	}

	jobs := make(chan int, config.NumSimulations)
	results := make(chan SimulationOutcome, config.NumSimulations)

	// Start workers
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go mc.simulationWorker(&wg, jobs, results, positions, scenarios, config)
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for i := 0; i < config.NumSimulations; i++ {
			jobs <- i
		}
	}()

	// Wait for workers to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	simulationResults := make([]SimulationOutcome, 0, config.NumSimulations)
	for outcome := range results {
		simulationResults = append(simulationResults, outcome)
	}

	return simulationResults
}

// simulationWorker processes individual simulations
func (mc *MonteCarloSimulator) simulationWorker(wg *sync.WaitGroup, jobs <-chan int, results chan<- SimulationOutcome, positions []models.Position, scenarios []map[string][]float64, config MonteCarloConfig) {
	defer wg.Done()

	for simIndex := range jobs {
		scenario := scenarios[simIndex]
		outcome := mc.runSingleSimulation(positions, scenario, config)
		results <- outcome
	}
}

// SimulationOutcome represents the result of a single simulation
type SimulationOutcome struct {
	SimulationIndex  int                `json:"simulation_index"`
	PortfolioReturn  float64            `json:"portfolio_return"`
	MaxDrawdown      float64            `json:"max_drawdown"`
	Volatility       float64            `json:"volatility"`
	LiquidityScore   float64            `json:"liquidity_score"`
	PositionReturns  map[string]float64 `json:"position_returns"`
	LiquidityMetrics map[string]float64 `json:"liquidity_metrics"`
}

// runSingleSimulation executes a single Monte Carlo simulation
func (mc *MonteCarloSimulator) runSingleSimulation(positions []models.Position, scenario map[string][]float64, config MonteCarloConfig) SimulationOutcome {
	outcome := SimulationOutcome{
		PositionReturns:  make(map[string]float64),
		LiquidityMetrics: make(map[string]float64),
	}

	// Calculate portfolio returns for this scenario
	portfolioReturns := mc.calculateScenarioReturns(positions, scenario)

	if len(portfolioReturns) > 0 {
		outcome.PortfolioReturn = portfolioReturns[len(portfolioReturns)-1] // Final return
		outcome.MaxDrawdown = mc.calculateMaxDrawdown(portfolioReturns)
		outcome.Volatility = mc.calculateVolatility(portfolioReturns)
	}

	// Calculate liquidity score for this scenario
	outcome.LiquidityScore = mc.calculateScenarioLiquidityScore(positions, scenario)

	// Calculate individual position returns
	for _, pos := range positions {
		if pricePath, exists := scenario[pos.Symbol]; exists && len(pricePath) > 0 {
			initialPrice := pos.CurrentPrice.InexactFloat64()
			finalPrice := pricePath[len(pricePath)-1]
			positionReturn := (finalPrice - initialPrice) / initialPrice
			outcome.PositionReturns[pos.Symbol] = positionReturn
		}
	}

	return outcome
}

// processSimulationResults aggregates and analyzes simulation outcomes
func (mc *MonteCarloSimulator) processSimulationResults(result *SimulationResult, outcomes []SimulationOutcome, config MonteCarloConfig) {
	if len(outcomes) == 0 {
		return
	}

	// Extract returns and sort for VaR calculation
	returns := make([]float64, len(outcomes))
	for i, outcome := range outcomes {
		returns[i] = outcome.PortfolioReturn
	}
	sort.Float64s(returns)

	result.SimulatedReturns = returns

	// Calculate VaR at different confidence levels
	confidenceLevels := []float64{0.90, 0.95, 0.99}
	for _, cl := range confidenceLevels {
		varIndex := int(float64(len(returns)) * (1 - cl))
		if varIndex >= 0 && varIndex < len(returns) {
			key := fmt.Sprintf("VaR_%.0f", cl*100)
			result.VaREstimates[key] = -returns[varIndex] * result.PortfolioValue // Convert to loss
		}
	}

	// Calculate Expected Shortfall (CVaR)
	for _, cl := range confidenceLevels {
		varIndex := int(float64(len(returns)) * (1 - cl))
		if varIndex > 0 {
			tailReturns := returns[:varIndex]
			expectedShortfall := 0.0
			for _, ret := range tailReturns {
				expectedShortfall += ret
			}
			expectedShortfall = -expectedShortfall / float64(len(tailReturns)) * result.PortfolioValue
			key := fmt.Sprintf("ES_%.0f", cl*100)
			result.ExpectedShortfall[key] = expectedShortfall
		}
	}

	// Calculate portfolio-level risk metrics
	result.Volatility = mc.calculateVolatility(returns)
	result.Skewness = mc.calculateSkewness(returns)
	result.Kurtosis = mc.calculateKurtosis(returns)

	// Calculate max drawdown across all simulations
	maxDrawdowns := make([]float64, len(outcomes))
	for i, outcome := range outcomes {
		maxDrawdowns[i] = outcome.MaxDrawdown
	}
	sort.Float64s(maxDrawdowns)
	result.MaxDrawdown = maxDrawdowns[len(maxDrawdowns)-1] // Worst drawdown

	// Risk decomposition by position
	mc.calculateRiskDecomposition(result, outcomes)
}

// Helper functions for statistical calculations
func (mc *MonteCarloSimulator) calculatePortfolioValue(positions []models.Position) float64 {
	totalValue := decimal.Zero
	for _, pos := range positions {
		totalValue = totalValue.Add(pos.MarketValue)
	}
	return totalValue.InexactFloat64()
}

func (mc *MonteCarloSimulator) calculateScenarioReturns(positions []models.Position, scenario map[string][]float64) []float64 {
	// Find minimum path length
	minLength := math.MaxInt32
	for _, path := range scenario {
		if len(path) < minLength {
			minLength = len(path)
		}
	}

	if minLength <= 1 {
		return []float64{}
	}

	portfolioReturns := make([]float64, minLength-1)

	for t := 1; t < minLength; t++ {
		portfolioValueT := 0.0
		portfolioValueT1 := 0.0

		for _, pos := range positions {
			if path, exists := scenario[pos.Symbol]; exists && len(path) > t {
				quantity := pos.Quantity.InexactFloat64()
				portfolioValueT += quantity * path[t]
				portfolioValueT1 += quantity * path[t-1]
			}
		}

		if portfolioValueT1 > 0 {
			portfolioReturns[t-1] = (portfolioValueT - portfolioValueT1) / portfolioValueT1
		}
	}

	return portfolioReturns
}

func (mc *MonteCarloSimulator) calculateMaxDrawdown(returns []float64) float64 {
	if len(returns) == 0 {
		return 0.0
	}

	cumulativeValue := 1.0
	peak := 1.0
	maxDrawdown := 0.0

	for _, ret := range returns {
		cumulativeValue *= (1 + ret)
		if cumulativeValue > peak {
			peak = cumulativeValue
		}
		drawdown := (peak - cumulativeValue) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	return maxDrawdown
}

func (mc *MonteCarloSimulator) calculateVolatility(returns []float64) float64 {
	if len(returns) < 2 {
		return 0.0
	}

	// Calculate mean
	sum := 0.0
	for _, ret := range returns {
		sum += ret
	}
	mean := sum / float64(len(returns))

	// Calculate variance
	sumSquaredDeviations := 0.0
	for _, ret := range returns {
		deviation := ret - mean
		sumSquaredDeviations += deviation * deviation
	}
	variance := sumSquaredDeviations / float64(len(returns)-1)

	return math.Sqrt(variance) * math.Sqrt(252) // Annualized volatility
}

func (mc *MonteCarloSimulator) calculateSkewness(returns []float64) float64 {
	if len(returns) < 3 {
		return 0.0
	}

	n := float64(len(returns))
	mean := mc.calculateMean(returns)
	variance := mc.calculateVariance(returns, mean)
	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		return 0.0
	}

	sumCubedDeviations := 0.0
	for _, ret := range returns {
		deviation := (ret - mean) / stdDev
		sumCubedDeviations += deviation * deviation * deviation
	}

	return (n / ((n - 1) * (n - 2))) * sumCubedDeviations
}

func (mc *MonteCarloSimulator) calculateKurtosis(returns []float64) float64 {
	if len(returns) < 4 {
		return 0.0
	}

	n := float64(len(returns))
	mean := mc.calculateMean(returns)
	variance := mc.calculateVariance(returns, mean)
	stdDev := math.Sqrt(variance)

	if stdDev == 0 {
		return 0.0
	}

	sumQuartedDeviations := 0.0
	for _, ret := range returns {
		deviation := (ret - mean) / stdDev
		quartedDeviation := deviation * deviation * deviation * deviation
		sumQuartedDeviations += quartedDeviation
	}

	kurtosis := (n * (n + 1) / ((n - 1) * (n - 2) * (n - 3))) * sumQuartedDeviations
	excess := (3 * (n - 1) * (n - 1)) / ((n - 2) * (n - 3))

	return kurtosis - excess // Excess kurtosis
}

func (mc *MonteCarloSimulator) calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, val := range values {
		sum += val
	}
	return sum / float64(len(values))
}

func (mc *MonteCarloSimulator) calculateVariance(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0.0
	}

	sumSquaredDeviations := 0.0
	for _, val := range values {
		deviation := val - mean
		sumSquaredDeviations += deviation * deviation
	}
	return sumSquaredDeviations / float64(len(values)-1)
}
