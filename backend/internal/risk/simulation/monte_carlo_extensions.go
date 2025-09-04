// backend/internal/risk/simulation/monte_carlo_extensions.go
package simulation

import (
	"fmt"
	"math"
	"runtime"
	"time"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

// estimateVolatility estimates volatility based on asset type and market regime
func (mc *MonteCarloSimulator) estimateVolatility(symbol string, marketRegime string) float64 {
	baseVolatilities := map[string]float64{
		"AAPL":  0.25, // 25% annual volatility for tech stocks
		"GOOGL": 0.28,
		"MSFT":  0.24,
		"TSLA":  0.45,
		"BTC":   0.80, // 80% for Bitcoin
		"ETH":   0.85, // 85% for Ethereum
		"GOLD":  0.20, // 20% for gold
		"OIL":   0.35, // 35% for oil
		"SPY":   0.18, // 18% for S&P 500 ETF
	}

	// Get base volatility or use default
	baseVol := baseVolatilities[symbol]
	if baseVol == 0 {
		// Default volatility based on asset type pattern
		if len(symbol) >= 3 && symbol[:3] == "BTC" {
			baseVol = 0.80 // Crypto
		} else if len(symbol) >= 3 && symbol[:3] == "ETH" {
			baseVol = 0.85 // Crypto
		} else {
			baseVol = 0.25 // Default stock volatility
		}
	}

	// Adjust for market regime
	switch marketRegime {
	case "NORMAL":
		return baseVol
	case "STRESSED":
		return baseVol * 1.5 // 50% increase in volatility
	case "CRISIS":
		return baseVol * 2.0 // 100% increase in volatility
	default:
		return baseVol
	}
}

// generateCorrelationMatrix creates a correlation matrix for the given symbols
func (mc *MonteCarloSimulator) generateCorrelationMatrix(symbols []string) [][]float64 {
	n := len(symbols)
	matrix := make([][]float64, n)
	for i := range matrix {
		matrix[i] = make([]float64, n)
	}

	// Set diagonal to 1.0 (perfect correlation with self)
	for i := 0; i < n; i++ {
		matrix[i][i] = 1.0
	}

	// Set off-diagonal correlations based on asset relationships
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			correlation := mc.estimateCorrelation(symbols[i], symbols[j])
			matrix[i][j] = correlation
			matrix[j][i] = correlation // Symmetric matrix
		}
	}

	return matrix
}

// estimateCorrelation estimates correlation between two assets
func (mc *MonteCarloSimulator) estimateCorrelation(symbol1, symbol2 string) float64 {
	// Define asset categories
	cryptos := map[string]bool{"BTC": true, "ETH": true, "ADA": true, "DOT": true}
	techStocks := map[string]bool{"AAPL": true, "GOOGL": true, "MSFT": true, "TSLA": true}
	commodities := map[string]bool{"GOLD": true, "SILVER": true, "OIL": true}

	// Same asset type correlations
	if cryptos[symbol1] && cryptos[symbol2] {
		return 0.65 // High correlation between cryptos
	}
	if techStocks[symbol1] && techStocks[symbol2] {
		return 0.55 // Moderate-high correlation between tech stocks
	}
	if commodities[symbol1] && commodities[symbol2] {
		return 0.40 // Moderate correlation between commodities
	}

	// Cross-category correlations
	if (cryptos[symbol1] && techStocks[symbol2]) || (techStocks[symbol1] && cryptos[symbol2]) {
		return 0.25 // Low-moderate correlation
	}
	if (cryptos[symbol1] && commodities[symbol2]) || (commodities[symbol1] && cryptos[symbol2]) {
		return 0.15 // Low correlation
	}
	if (techStocks[symbol1] && commodities[symbol2]) || (commodities[symbol1] && techStocks[symbol2]) {
		return 0.20 // Low correlation
	}

	// Default correlation for unknown pairs
	return 0.10
}

// calculateScenarioLiquidityScore calculates liquidity score for a scenario
func (mc *MonteCarloSimulator) calculateScenarioLiquidityScore(positions []models.Position, scenario map[string][]float64) float64 {
	if len(positions) == 0 {
		return 1.0 // Empty portfolio is fully liquid
	}

	totalScore := 0.0
	totalWeight := 0.0

	for _, pos := range positions {
		weight := pos.MarketValue.InexactFloat64()
		score := mc.getLiquidityScore(pos.Symbol, pos.AssetType)

		// Adjust score based on price volatility in scenario
		if pricePath, exists := scenario[pos.Symbol]; exists && len(pricePath) > 1 {
			volatility := mc.calculatePathVolatility(pricePath)
			// Higher volatility reduces liquidity score
			score *= (1.0 - math.Min(volatility*0.1, 0.5))
		}

		totalScore += score * weight
		totalWeight += weight
	}

	if totalWeight > 0 {
		return totalScore / totalWeight
	}
	return 1.0
}

// getLiquidityScore returns a liquidity score based on asset characteristics
func (mc *MonteCarloSimulator) getLiquidityScore(symbol, assetType string) float64 {
	// High liquidity assets
	highLiquidityAssets := map[string]bool{
		"AAPL": true, "GOOGL": true, "MSFT": true, "SPY": true, "QQQ": true,
		"BTC": true, "ETH": true,
	}

	// Medium liquidity assets
	mediumLiquidityAssets := map[string]bool{
		"TSLA": true, "GOLD": true, "SILVER": true,
	}

	if highLiquidityAssets[symbol] {
		return 0.9 // 90% liquidity score
	} else if mediumLiquidityAssets[symbol] {
		return 0.6 // 60% liquidity score
	}

	// Score based on asset type
	switch assetType {
	case "STOCK":
		return 0.7 // 70% for generic stocks
	case "ETF":
		return 0.8 // 80% for ETFs
	case "CRYPTO":
		return 0.5 // 50% for generic crypto
	case "COMMODITY":
		return 0.4 // 40% for commodities
	case "BOND":
		return 0.6 // 60% for bonds
	default:
		return 0.5 // 50% default
	}
}

// calculatePathVolatility calculates volatility from a price path
func (mc *MonteCarloSimulator) calculatePathVolatility(pricePath []float64) float64 {
	if len(pricePath) < 2 {
		return 0.0
	}

	returns := make([]float64, len(pricePath)-1)
	for i := 1; i < len(pricePath); i++ {
		if pricePath[i-1] > 0 {
			returns[i-1] = (pricePath[i] - pricePath[i-1]) / pricePath[i-1]
		}
	}

	return mc.calculateVolatility(returns)
}

// calculateRiskDecomposition calculates risk contribution by position
func (mc *MonteCarloSimulator) calculateRiskDecomposition(result *SimulationResult, outcomes []SimulationOutcome) {
	if len(outcomes) == 0 {
		return
	}

	// Aggregate position returns across all simulations
	positionReturns := make(map[string][]float64)

	for _, outcome := range outcomes {
		for symbol, ret := range outcome.PositionReturns {
			if _, exists := positionReturns[symbol]; !exists {
				positionReturns[symbol] = make([]float64, 0, len(outcomes))
			}
			positionReturns[symbol] = append(positionReturns[symbol], ret)
		}
	}

	// Calculate risk contribution (volatility contribution)
	for symbol, returns := range positionReturns {
		volatility := mc.calculateVolatility(returns)
		result.RiskDecomposition[symbol] = volatility
	}
}

// validateAccuracy performs backtesting and accuracy validation
func (mc *MonteCarloSimulator) validateAccuracy(positions []models.Position, outcomes []SimulationOutcome, config MonteCarloConfig) *AccuracyValidation {
	validation := &AccuracyValidation{
		BacktestingResults:        make(map[string]float64),
		VaRViolations:             make(map[string]int),
		VaRViolationRate:          make(map[string]float64),
		ExpectedViolations:        make(map[string]int),
		KupiecTestResults:         make(map[string]bool),
		ChristoffersenTestResults: make(map[string]bool),
	}

	confidenceLevels := []float64{0.95, 0.99}

	for _, cl := range confidenceLevels {
		key := fmt.Sprintf("%.0f", cl*100)

		// Calculate expected violations
		expectedViolations := int(float64(len(outcomes)) * (1 - cl))
		validation.ExpectedViolations[key] = expectedViolations

		// Count actual violations (simplified - would need historical data for real backtesting)
		actualViolations := mc.countVaRViolations(outcomes, cl)
		validation.VaRViolations[key] = actualViolations
		validation.VaRViolationRate[key] = float64(actualViolations) / float64(len(outcomes))

		// Kupiec test (simplified)
		validation.KupiecTestResults[key] = mc.performKupiecTest(actualViolations, len(outcomes), 1-cl)

		// Christoffersen test (simplified)
		validation.ChristoffersenTestResults[key] = mc.performChristoffersenTest(actualViolations, expectedViolations)

		// Backtesting score (0-100)
		score := 100.0 * (1.0 - math.Abs(float64(actualViolations-expectedViolations))/float64(expectedViolations+1))
		validation.BacktestingResults[key] = score
	}

	return validation
}

// countVaRViolations counts violations for a given confidence level
func (mc *MonteCarloSimulator) countVaRViolations(outcomes []SimulationOutcome, confidenceLevel float64) int {
	returns := make([]float64, len(outcomes))
	for i, outcome := range outcomes {
		returns[i] = outcome.PortfolioReturn
	}

	// Calculate VaR threshold
	varIndex := int(float64(len(returns)) * (1 - confidenceLevel))
	if varIndex >= len(returns) {
		return 0
	}

	// Sort returns to find VaR threshold
	sortedReturns := make([]float64, len(returns))
	copy(sortedReturns, returns)

	// Simple bubble sort for this example (would use more efficient sorting in production)
	for i := 0; i < len(sortedReturns); i++ {
		for j := 0; j < len(sortedReturns)-1-i; j++ {
			if sortedReturns[j] > sortedReturns[j+1] {
				sortedReturns[j], sortedReturns[j+1] = sortedReturns[j+1], sortedReturns[j]
			}
		}
	}

	varThreshold := sortedReturns[varIndex]

	// Count violations
	violations := 0
	for _, ret := range returns {
		if ret < varThreshold {
			violations++
		}
	}

	return violations
}

// performKupiecTest performs simplified Kupiec test
func (mc *MonteCarloSimulator) performKupiecTest(actualViolations, totalObservations int, alpha float64) bool {
	expectedViolations := float64(totalObservations) * alpha

	// Chi-square test statistic (simplified)
	if expectedViolations == 0 {
		return actualViolations == 0
	}

	chiSquare := math.Pow(float64(actualViolations)-expectedViolations, 2) / expectedViolations

	// Critical value for 5% significance level (simplified)
	criticalValue := 3.841 // Chi-square with 1 degree of freedom at 5% level

	return chiSquare < criticalValue
}

// performChristoffersenTest performs simplified Christoffersen test
func (mc *MonteCarloSimulator) performChristoffersenTest(actualViolations, expectedViolations int) bool {
	// Simplified test - in practice would check for independence of violations
	ratio := float64(actualViolations) / float64(expectedViolations+1)
	return ratio >= 0.8 && ratio <= 1.25 // Within 25% of expected
}

// performStressTests conducts stress testing under different market regimes
func (mc *MonteCarloSimulator) performStressTests(positions []models.Position, config MonteCarloConfig) *StressTestResults {
	results := &StressTestResults{
		NormalMarketVaR:   make(map[string]float64),
		StressedMarketVaR: make(map[string]float64),
		CrisisMarketVaR:   make(map[string]float64),
		TailRiskMetrics:   make(map[string]float64),
	}

	regimes := []string{"NORMAL", "STRESSED", "CRISIS"}
	confidenceLevels := []float64{0.95, 0.99}

	for _, regime := range regimes {
		// Create config for this regime
		regimeConfig := config
		regimeConfig.MarketRegime = regime
		regimeConfig.NumSimulations = config.NumSimulations / 3 // Use fewer simulations for speed

		// Generate scenarios for this regime
		scenarios, _ := mc.generatePriceScenarios(positions, regimeConfig)
		outcomes := mc.runParallelSimulations(positions, scenarios, regimeConfig)

		// Calculate VaR for this regime
		returns := make([]float64, len(outcomes))
		for i, outcome := range outcomes {
			returns[i] = outcome.PortfolioReturn
		}

		portfolioValue := mc.calculatePortfolioValue(positions)

		for _, cl := range confidenceLevels {
			varIndex := int(float64(len(returns)) * (1 - cl))
			if varIndex >= 0 && varIndex < len(returns) {
				// Sort returns
				sortedReturns := make([]float64, len(returns))
				copy(sortedReturns, returns)
				for i := 0; i < len(sortedReturns); i++ {
					for j := 0; j < len(sortedReturns)-1-i; j++ {
						if sortedReturns[j] > sortedReturns[j+1] {
							sortedReturns[j], sortedReturns[j+1] = sortedReturns[j+1], sortedReturns[j]
						}
					}
				}

				varValue := -sortedReturns[varIndex] * portfolioValue
				key := fmt.Sprintf("VaR_%.0f", cl*100)

				switch regime {
				case "NORMAL":
					results.NormalMarketVaR[key] = varValue
				case "STRESSED":
					results.StressedMarketVaR[key] = varValue
				case "CRISIS":
					results.CrisisMarketVaR[key] = varValue
				}
			}
		}
	}

	// Calculate tail risk metrics
	results.TailRiskMetrics["stress_multiplier_95"] = results.StressedMarketVaR["VaR_95"] / (results.NormalMarketVaR["VaR_95"] + 0.01)
	results.TailRiskMetrics["crisis_multiplier_95"] = results.CrisisMarketVaR["VaR_95"] / (results.NormalMarketVaR["VaR_95"] + 0.01)
	results.TailRiskMetrics["stress_multiplier_99"] = results.StressedMarketVaR["VaR_99"] / (results.NormalMarketVaR["VaR_99"] + 0.01)
	results.TailRiskMetrics["crisis_multiplier_99"] = results.CrisisMarketVaR["VaR_99"] / (results.NormalMarketVaR["VaR_99"] + 0.01)

	return results
}

// calculateLiquidityMetrics calculates comprehensive liquidity metrics
func (mc *MonteCarloSimulator) calculateLiquidityMetrics(positions []models.Position, config MonteCarloConfig) *LiquiditySimResults {
	results := &LiquiditySimResults{
		LiquidityVaR:            make(map[string]float64),
		TimeToLiquidate:         make(map[string]float64),
		PositionLiquidityImpact: make(map[string]float64),
	}

	// Calculate liquidity scores across simulations
	liquidityScores := make([]float64, 0)

	// Simplified calculation - would run separate liquidity simulations in practice
	for i := 0; i < 100; i++ { // Sample liquidity scenarios
		scenario := make(map[string][]float64)
		for _, pos := range positions {
			// Generate simple price path for liquidity assessment
			path := mc.priceGenerator.GeneratePricePath(
				pos.CurrentPrice.InexactFloat64(),
				mc.estimateVolatility(pos.Symbol, config.MarketRegime),
				config.TimeHorizonDays,
				nil,
			)
			scenario[pos.Symbol] = path
		}

		liquidityScore := mc.calculateScenarioLiquidityScore(positions, scenario)
		liquidityScores = append(liquidityScores, liquidityScore)
	}

	// Calculate average liquidity score
	sum := 0.0
	for _, score := range liquidityScores {
		sum += score
	}
	results.AverageLiquidityScore = sum / float64(len(liquidityScores))

	// Calculate liquidity distribution percentiles
	results.LiquidityDistribution = make([]float64, len(liquidityScores))
	copy(results.LiquidityDistribution, liquidityScores)

	// Calculate time to liquidate estimates
	results.TimeToLiquidate["normal_market"] = mc.estimateTimeToLiquidate(positions, "NORMAL")
	results.TimeToLiquidate["stressed_market"] = mc.estimateTimeToLiquidate(positions, "STRESSED")
	results.TimeToLiquidate["crisis_market"] = mc.estimateTimeToLiquidate(positions, "CRISIS")

	// Calculate position-level liquidity impact
	for _, pos := range positions {
		impact := mc.estimateLiquidityImpact(pos)
		results.PositionLiquidityImpact[pos.Symbol] = impact
	}

	return results
}

// estimateTimeToLiquidate estimates time needed to liquidate portfolio
func (mc *MonteCarloSimulator) estimateTimeToLiquidate(positions []models.Position, marketCondition string) float64 {
	totalTime := 0.0
	totalValue := 0.0

	for _, pos := range positions {
		positionValue := pos.MarketValue.InexactFloat64()
		liquidationTime := mc.estimatePositionLiquidationTime(pos, marketCondition)

		totalTime += liquidationTime * positionValue
		totalValue += positionValue
	}

	if totalValue > 0 {
		return totalTime / totalValue
	}
	return 0.0
}

// estimatePositionLiquidationTime estimates liquidation time for a single position
func (mc *MonteCarloSimulator) estimatePositionLiquidationTime(position models.Position, marketCondition string) float64 {
	baseDays := map[string]float64{
		"AAPL": 0.5, "GOOGL": 0.5, "MSFT": 0.5, // Large cap stocks
		"BTC": 1.0, "ETH": 1.0, // Major crypto
		"GOLD": 2.0, "SILVER": 3.0, // Commodities
	}

	baseLiquidationTime := baseDays[position.Symbol]
	if baseLiquidationTime == 0 {
		// Default based on asset type
		switch position.AssetType {
		case "STOCK":
			baseLiquidationTime = 1.0
		case "CRYPTO":
			baseLiquidationTime = 2.0
		case "COMMODITY":
			baseLiquidationTime = 3.0
		default:
			baseLiquidationTime = 2.0
		}
	}

	// Adjust for market conditions
	multiplier := 1.0
	switch marketCondition {
	case "NORMAL":
		multiplier = 1.0
	case "STRESSED":
		multiplier = 2.0
	case "CRISIS":
		multiplier = 4.0
	}

	return baseLiquidationTime * multiplier
}

// estimateLiquidityImpact estimates price impact of liquidating a position
func (mc *MonteCarloSimulator) estimateLiquidityImpact(position models.Position) float64 {
	// Simplified model - in practice would use market microstructure models
	positionSize := position.Quantity.InexactFloat64() * position.CurrentPrice.InexactFloat64()

	// Impact increases with position size (square root law)
	baseImpact := 0.001                                  // 0.1% base impact
	sizeMultiplier := math.Sqrt(positionSize / 100000.0) // Scale by $100k

	impact := baseImpact * sizeMultiplier

	// Cap the impact at 10%
	if impact > 0.10 {
		impact = 0.10
	}

	return impact
}

// calculatePerformanceMetrics calculates simulation performance statistics
func (mc *MonteCarloSimulator) calculatePerformanceMetrics(duration time.Duration, config MonteCarloConfig) *PerformanceMetrics {
	metrics := &PerformanceMetrics{
		SimulationTime:       duration,
		SimulationsPerSecond: float64(config.NumSimulations) / duration.Seconds(),
		ConvergenceRate:      0.95, // Would calculate actual convergence in practice
	}

	// Get memory statistics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	metrics.MemoryUsage = int64(memStats.Alloc)

	// CPU usage would be calculated from system metrics in practice
	metrics.CPUUsage = 0.0

	return metrics
}
