// backend/tests/monte_carlo_test.go
package tests

import (
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/marketdata"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/calculator"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/simulation"
	"github.com/shopspring/decimal"
)

// createTestMarketDataProvider creates a market data provider for testing
func createTestMarketDataProvider() calculator.MarketDataProvider {
	// Create market data service configuration for testing
	config := &marketdata.ServiceConfig{
		PrimaryProvider:   "test",
		FallbackProviders: []string{},
		CacheTTL:          5 * time.Minute,
		RateLimits: map[string]marketdata.RateLimitConfig{
			"test": {
				RequestsPerMinute: 1000,
				BurstLimit:        50,
			},
		},
	}

	// Create market data service
	marketDataService := marketdata.NewService(config, database.GetRedis())

	// Create and return adapter
	return &TestMarketDataProviderAdapter{
		marketDataService: marketDataService,
	}
}

// TestMarketDataProviderAdapter adapts the market data service for testing
type TestMarketDataProviderAdapter struct {
	marketDataService *marketdata.Service
}

func (m *TestMarketDataProviderAdapter) GetAverageDailyVolume(symbol string) float64 {
	// Use realistic test data
	volumes := map[string]float64{
		"AAPL":  50000000, // 50M shares
		"GOOGL": 25000000, // 25M shares
		"MSFT":  40000000, // 40M shares
		"BTC":   2000000,  // 2M BTC
		"ETH":   5000000,  // 5M ETH
		"GOLD":  100000,   // 100K oz
	}
	if vol, exists := volumes[symbol]; exists {
		return vol
	}
	return 1000000 // Default 1M
}

func (m *TestMarketDataProviderAdapter) GetBidAskSpread(symbol string) float64 {
	// Use realistic test spreads
	spreads := map[string]float64{
		"AAPL":  0.0001, // 0.01%
		"GOOGL": 0.0001,
		"MSFT":  0.0001,
		"BTC":   0.001, // 0.1%
		"ETH":   0.001,
		"GOLD":  0.002, // 0.2%
	}
	if spread, exists := spreads[symbol]; exists {
		return spread
	}
	return 0.001 // Default 0.1%
}

func (m *TestMarketDataProviderAdapter) GetMarketDepth(symbol string) *calculator.MarketDepth {
	// Generate realistic test market depth
	return &calculator.MarketDepth{
		BidLevels: []calculator.PriceLevel{
			{Price: 100.0, Quantity: 1000, Orders: 10},
			{Price: 99.5, Quantity: 2000, Orders: 15},
		},
		AskLevels: []calculator.PriceLevel{
			{Price: 100.1, Quantity: 1000, Orders: 10},
			{Price: 100.6, Quantity: 2000, Orders: 15},
		},
		Timestamp: time.Now(),
	}
}

func (m *TestMarketDataProviderAdapter) GetMarketCap(symbol string) float64 {
	// Use realistic test market caps
	marketCaps := map[string]float64{
		"AAPL":  3000000000000,  // $3T
		"GOOGL": 1500000000000,  // $1.5T
		"MSFT":  2800000000000,  // $2.8T
		"BTC":   800000000000,   // $800B
		"ETH":   400000000000,   // $400B
		"GOLD":  12000000000000, // $12T
	}
	if cap, exists := marketCaps[symbol]; exists {
		return cap
	}
	return 1000000000 // Default $1B
}

func runMonteCarloTests() {
	fmt.Println("🎲 Monte Carlo Risk Assessment Test")
	fmt.Println("===================================")

	// Create test portfolio
	portfolio := createTestPortfolio()

	// Initialize calculators with proper parameters
	varCalculator := calculator.NewVaRCalculator(portfolio.TotalValue.InexactFloat64(), 252) // 252 trading days lookback
	liquidityCalculator := calculator.NewLiquidityCalculator(createTestMarketDataProvider())

	// Create Monte Carlo simulator
	simulator := simulation.NewMonteCarloSimulator(varCalculator, liquidityCalculator)

	// Run different simulation configurations
	configs := []simulation.MonteCarloConfig{
		{
			NumSimulations:    1000,
			TimeHorizonDays:   22,
			ConfidenceLevel:   0.95,
			NumWorkers:        4,
			RandomSeed:        12345,
			MarketRegime:      "NORMAL",
			CorrelationMatrix: false,
		},
		{
			NumSimulations:    5000,
			TimeHorizonDays:   22,
			ConfidenceLevel:   0.99,
			NumWorkers:        4,
			RandomSeed:        12345,
			MarketRegime:      "STRESSED",
			CorrelationMatrix: true,
		},
		{
			NumSimulations:    10000,
			TimeHorizonDays:   66, // 3 months
			ConfidenceLevel:   0.99,
			NumWorkers:        8,
			RandomSeed:        54321,
			MarketRegime:      "CRISIS",
			CorrelationMatrix: true,
		},
	}

	// Run tests for each configuration
	for i, config := range configs {
		fmt.Printf("\n🔬 Test %d: %s Market Regime\n", i+1, config.MarketRegime)
		fmt.Printf("Simulations: %d, Horizon: %d days, Confidence: %.1f%%\n",
			config.NumSimulations, config.TimeHorizonDays, config.ConfidenceLevel*100)

		result, err := runSimulationTest(simulator, portfolio.Positions, config)
		if err != nil {
			log.Printf("❌ Test %d failed: %v", i+1, err)
			continue
		}

		// Display results
		displayResults(result, i+1)

		// Validate results
		validateResults(result, i+1)
	}

	// Run accuracy comparison test
	fmt.Println("\n🎯 Accuracy Comparison Test")
	fmt.Println("===========================")
	runAccuracyTest(simulator, portfolio.Positions)

	// Run performance benchmark
	fmt.Println("\n⚡ Performance Benchmark")
	fmt.Println("=======================")
	runPerformanceBenchmark(simulator, portfolio.Positions)

	fmt.Println("\n✅ All Monte Carlo tests completed!")
}

func createTestPortfolio() *models.Portfolio {
	rand.Seed(time.Now().UnixNano())

	positions := []models.Position{
		{
			Symbol:       "AAPL",
			Quantity:     decimal.NewFromFloat(100),
			AveragePrice: decimal.NewFromFloat(150.00),
			CurrentPrice: decimal.NewFromFloat(155.00),
			MarketValue:  decimal.NewFromFloat(15500),
			AssetType:    "STOCK",
			Liquidity:    "HIGH",
		},
		{
			Symbol:       "GOOGL",
			Quantity:     decimal.NewFromFloat(50),
			AveragePrice: decimal.NewFromFloat(2800.00),
			CurrentPrice: decimal.NewFromFloat(2850.00),
			MarketValue:  decimal.NewFromFloat(142500),
			AssetType:    "STOCK",
			Liquidity:    "HIGH",
		},
		{
			Symbol:       "MSFT",
			Quantity:     decimal.NewFromFloat(75),
			AveragePrice: decimal.NewFromFloat(350.00),
			CurrentPrice: decimal.NewFromFloat(360.00),
			MarketValue:  decimal.NewFromFloat(27000),
			AssetType:    "STOCK",
			Liquidity:    "HIGH",
		},
		{
			Symbol:       "BTC",
			Quantity:     decimal.NewFromFloat(2.5),
			AveragePrice: decimal.NewFromFloat(35000.00),
			CurrentPrice: decimal.NewFromFloat(45000.00),
			MarketValue:  decimal.NewFromFloat(112500),
			AssetType:    "CRYPTO",
			Liquidity:    "MEDIUM",
		},
		{
			Symbol:       "ETH",
			Quantity:     decimal.NewFromFloat(20),
			AveragePrice: decimal.NewFromFloat(2000.00),
			CurrentPrice: decimal.NewFromFloat(2500.00),
			MarketValue:  decimal.NewFromFloat(50000),
			AssetType:    "CRYPTO",
			Liquidity:    "MEDIUM",
		},
		{
			Symbol:       "GOLD",
			Quantity:     decimal.NewFromFloat(100),
			AveragePrice: decimal.NewFromFloat(1800.00),
			CurrentPrice: decimal.NewFromFloat(1950.00),
			MarketValue:  decimal.NewFromFloat(195000),
			AssetType:    "COMMODITY",
			Liquidity:    "LOW",
		},
	}

	totalValue := decimal.Zero
	for _, pos := range positions {
		totalValue = totalValue.Add(pos.MarketValue)
	}

	return &models.Portfolio{
		Name:       "Test Portfolio",
		TotalValue: totalValue,
		Currency:   "USD",
		Positions:  positions,
	}
}

func runSimulationTest(simulator *simulation.MonteCarloSimulator, positions []models.Position, config simulation.MonteCarloConfig) (*simulation.SimulationResult, error) {
	startTime := time.Now()
	fmt.Printf("⏳ Running simulation with %d iterations...\n", config.NumSimulations)

	result, err := simulator.RunSimulation(positions, config)
	if err != nil {
		return nil, err
	}

	fmt.Printf("✅ Simulation completed in %v\n", time.Since(startTime))
	return result, nil
}

func displayResults(result *simulation.SimulationResult, testNumber int) {
	fmt.Printf("\n📊 Test %d Results:\n", testNumber)
	fmt.Printf("Portfolio Value: $%.2f\n", result.PortfolioValue)
	fmt.Printf("Simulation Duration: %v\n", result.Duration)

	fmt.Println("\n💰 Value at Risk (VaR):")
	for level, var_val := range result.VaREstimates {
		fmt.Printf("  %s: $%.2f (%.2f%%)\n", level, var_val, var_val/result.PortfolioValue*100)
	}

	fmt.Println("\n📉 Expected Shortfall:")
	for level, es := range result.ExpectedShortfall {
		fmt.Printf("  %s: $%.2f (%.2f%%)\n", level, es, es/result.PortfolioValue*100)
	}

	fmt.Printf("\n📈 Risk Metrics:\n")
	fmt.Printf("  Volatility: %.2f%%\n", result.Volatility*100)
	fmt.Printf("  Max Drawdown: %.2f%%\n", result.MaxDrawdown*100)
	fmt.Printf("  Skewness: %.4f\n", result.Skewness)
	fmt.Printf("  Kurtosis: %.4f\n", result.Kurtosis)

	if result.LiquidityMetrics != nil {
		fmt.Printf("\n💧 Liquidity Metrics:\n")
		fmt.Printf("  Average Liquidity Score: %.2f\n", result.LiquidityMetrics.AverageLiquidityScore)
		for condition, days := range result.LiquidityMetrics.TimeToLiquidate {
			fmt.Printf("  Time to Liquidate (%s): %.1f days\n", condition, days)
		}
	}

	if result.StresTesting != nil {
		fmt.Printf("\n⚠️  Stress Testing:\n")
		for metric, value := range result.StresTesting.TailRiskMetrics {
			fmt.Printf("  %s: %.2fx\n", metric, value)
		}
	}

	if result.PerformanceMetrics != nil {
		fmt.Printf("\n⚡ Performance:\n")
		fmt.Printf("  Simulations/sec: %.0f\n", result.PerformanceMetrics.SimulationsPerSecond)
		fmt.Printf("  Memory Usage: %.2f MB\n", float64(result.PerformanceMetrics.MemoryUsage)/(1024*1024))
	}
}

func validateResults(result *simulation.SimulationResult, testNumber int) {
	fmt.Printf("\n🔍 Test %d Validation:\n", testNumber)

	// Validate VaR results
	var95 := result.VaREstimates["VaR_95"]
	var99 := result.VaREstimates["VaR_99"]

	if var99 <= var95 {
		fmt.Printf("❌ VaR ordering invalid: VaR99 (%.2f) should be > VaR95 (%.2f)\n", var99, var95)
	} else {
		fmt.Printf("✅ VaR ordering valid: VaR99 > VaR95\n")
	}

	// Validate reasonable ranges
	varRatio95 := var95 / result.PortfolioValue
	varRatio99 := var99 / result.PortfolioValue

	if varRatio95 < 0.01 || varRatio95 > 0.30 {
		fmt.Printf("⚠️  VaR95 seems unusual: %.2f%% of portfolio\n", varRatio95*100)
	} else {
		fmt.Printf("✅ VaR95 in reasonable range: %.2f%% of portfolio\n", varRatio95*100)
	}

	if varRatio99 < 0.02 || varRatio99 > 0.50 {
		fmt.Printf("⚠️  VaR99 seems unusual: %.2f%% of portfolio\n", varRatio99*100)
	} else {
		fmt.Printf("✅ VaR99 in reasonable range: %.2f%% of portfolio\n", varRatio99*100)
	}

	// Validate accuracy metrics
	if result.AccuracyValidation != nil {
		fmt.Printf("✅ Backtesting completed\n")
		for level, score := range result.AccuracyValidation.BacktestingResults {
			if score >= 80 {
				fmt.Printf("✅ %s backtesting score: %.1f%% (Good)\n", level, score)
			} else {
				fmt.Printf("⚠️  %s backtesting score: %.1f%% (Needs improvement)\n", level, score)
			}
		}
	}

	// Validate simulation convergence
	if len(result.SimulatedReturns) >= result.Config.NumSimulations {
		fmt.Printf("✅ All %d simulations completed\n", result.Config.NumSimulations)
	} else {
		fmt.Printf("⚠️  Only %d of %d simulations completed\n", len(result.SimulatedReturns), result.Config.NumSimulations)
	}
}

func runAccuracyTest(simulator *simulation.MonteCarloSimulator, positions []models.Position) {
	// Test different numbers of simulations to check convergence
	simulationCounts := []int{100, 500, 1000, 5000, 10000}

	fmt.Println("Testing VaR convergence with different simulation counts:")

	baseConfig := simulation.MonteCarloConfig{
		TimeHorizonDays:   22,
		ConfidenceLevel:   0.95,
		NumWorkers:        4,
		RandomSeed:        12345,
		MarketRegime:      "NORMAL",
		CorrelationMatrix: true,
	}

	var previousVaR float64

	for i, count := range simulationCounts {
		config := baseConfig
		config.NumSimulations = count

		result, err := simulator.RunSimulation(positions, config)
		if err != nil {
			fmt.Printf("❌ Error with %d simulations: %v\n", count, err)
			continue
		}

		currentVaR := result.VaREstimates["VaR_95"]

		if i > 0 {
			convergence := math.Abs(currentVaR-previousVaR) / previousVaR * 100
			fmt.Printf("%d simulations: VaR95 = $%.2f, Convergence = %.2f%%\n", count, currentVaR, convergence)

			if convergence < 2.0 {
				fmt.Printf("✅ Good convergence achieved at %d simulations\n", count)
			}
		} else {
			fmt.Printf("%d simulations: VaR95 = $%.2f\n", count, currentVaR)
		}

		previousVaR = currentVaR
	}
}

func runPerformanceBenchmark(simulator *simulation.MonteCarloSimulator, positions []models.Position) {
	config := simulation.MonteCarloConfig{
		NumSimulations:    10000,
		TimeHorizonDays:   22,
		ConfidenceLevel:   0.95,
		NumWorkers:        1, // Start with single worker
		RandomSeed:        12345,
		MarketRegime:      "NORMAL",
		CorrelationMatrix: true,
	}

	fmt.Println("Testing performance with different worker counts:")

	workerCounts := []int{1, 2, 4, 8}

	for _, workers := range workerCounts {
		config.NumWorkers = workers

		startTime := time.Now()
		result, err := simulator.RunSimulation(positions, config)
		duration := time.Since(startTime)

		if err != nil {
			fmt.Printf("❌ Error with %d workers: %v\n", workers, err)
			continue
		}

		simsPerSec := float64(config.NumSimulations) / duration.Seconds()
		fmt.Printf("%d workers: %.0f sims/sec, Duration: %v\n", workers, simsPerSec, duration)

		if result.PerformanceMetrics != nil {
			fmt.Printf("  Memory: %.2f MB\n", float64(result.PerformanceMetrics.MemoryUsage)/(1024*1024))
		}
	}
}
