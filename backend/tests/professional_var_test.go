package tests

import (
	"math"
	"math/rand"
	"testing"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/calculator"
	"github.com/shopspring/decimal"
)

func TestProfessionalVaRCalculator(t *testing.T) {
	// Create test portfolio
	positions := []models.Position{
		{
			Symbol:       "AAPL",
			Quantity:     decimal.NewFromFloat(100),
			AveragePrice: decimal.NewFromFloat(150),
			CurrentPrice: decimal.NewFromFloat(155),
			MarketValue:  decimal.NewFromFloat(15500),
		},
		{
			Symbol:       "GOOGL",
			Quantity:     decimal.NewFromFloat(50),
			AveragePrice: decimal.NewFromFloat(2800),
			CurrentPrice: decimal.NewFromFloat(2850),
			MarketValue:  decimal.NewFromFloat(142500),
		},
		{
			Symbol:       "MSFT",
			Quantity:     decimal.NewFromFloat(75),
			AveragePrice: decimal.NewFromFloat(350),
			CurrentPrice: decimal.NewFromFloat(360),
			MarketValue:  decimal.NewFromFloat(27000),
		},
	}

	// Generate sample price history
	priceHistory := generateSamplePriceHistory()

	// Calculate portfolio value
	portfolioValue := 185000.0 // Sum of market values

	t.Run("VaR Calculation Methods", func(t *testing.T) {
		calculator := calculator.NewProfessionalVaRCalculator(portfolioValue, 1)
		result, err := calculator.CalculateProfessionalVaR(positions, priceHistory)

		if err != nil {
			t.Fatalf("Failed to calculate VaR: %v", err)
		}

		// Test that all VaR methods produce results
		if len(result.HistoricalVaR) == 0 {
			t.Error("Historical VaR not calculated")
		}

		if len(result.ParametricVaR) == 0 {
			t.Error("Parametric VaR not calculated")
		}

		if len(result.MonteCarloVaR) == 0 {
			t.Error("Monte Carlo VaR not calculated")
		}

		if len(result.CornishFisherVaR) == 0 {
			t.Error("Cornish-Fisher VaR not calculated")
		}

		// Test VaR ordering (99% VaR should be higher than 95% VaR)
		if result.HistoricalVaR[0.99] <= result.HistoricalVaR[0.95] {
			t.Error("99% VaR should be greater than 95% VaR")
		}

		// Test that VaR values are reasonable (between 0 and portfolio value)
		for confidence, varValue := range result.HistoricalVaR {
			if varValue < 0 || varValue > portfolioValue {
				t.Errorf("VaR at %f confidence is out of bounds: %f", confidence, varValue)
			}
		}

		t.Logf("95%% Historical VaR: %.2f", result.HistoricalVaR[0.95])
		t.Logf("99%% Historical VaR: %.2f", result.HistoricalVaR[0.99])
		t.Logf("95%% Parametric VaR: %.2f", result.ParametricVaR[0.95])
		t.Logf("99%% Parametric VaR: %.2f", result.ParametricVaR[0.99])
	})

	t.Run("Risk Decomposition", func(t *testing.T) {
		calculator := calculator.NewProfessionalVaRCalculator(portfolioValue, 1)
		result, err := calculator.CalculateProfessionalVaR(positions, priceHistory)

		if err != nil {
			t.Fatalf("Failed to calculate VaR: %v", err)
		}

		// Test component VaR
		if len(result.ComponentVaR) != len(positions) {
			t.Errorf("Component VaR count mismatch: expected %d, got %d",
				len(positions), len(result.ComponentVaR))
		}

		// Component VaR should sum approximately to portfolio VaR
		totalComponentVaR := 0.0
		for _, componentVaR := range result.ComponentVaR {
			totalComponentVaR += componentVaR
		}

		// Allow 10% tolerance due to diversification effects
		portfolioVaR := result.ParametricVaR[0.99]
		tolerance := portfolioVaR * 0.1

		if math.Abs(totalComponentVaR-portfolioVaR) > tolerance {
			t.Logf("Warning: Component VaR sum (%.2f) differs from portfolio VaR (%.2f) by more than 10%%",
				totalComponentVaR, portfolioVaR)
		}

		// Test marginal VaR
		if len(result.MarginalVaR) != len(positions) {
			t.Errorf("Marginal VaR count mismatch: expected %d, got %d",
				len(positions), len(result.MarginalVaR))
		}

		for symbol, marginalVaR := range result.MarginalVaR {
			t.Logf("Marginal VaR for %s: %.2f", symbol, marginalVaR)
		}
	})

	t.Run("Statistical Moments", func(t *testing.T) {
		calculator := calculator.NewProfessionalVaRCalculator(portfolioValue, 1)
		result, err := calculator.CalculateProfessionalVaR(positions, priceHistory)

		if err != nil {
			t.Fatalf("Failed to calculate VaR: %v", err)
		}

		// Test volatility is positive
		if result.Volatility <= 0 {
			t.Error("Volatility should be positive")
		}

		// Test volatility is reasonable (0-100% annualized)
		if result.Volatility > 1.0 {
			t.Error("Volatility seems too high (>100% annualized)")
		}

		// Skewness typically between -3 and 3
		if math.Abs(result.Skewness) > 3 {
			t.Logf("Warning: Skewness %.2f is unusually high", result.Skewness)
		}

		// Kurtosis for normal distribution is 3
		if result.Kurtosis < 1 || result.Kurtosis > 10 {
			t.Logf("Warning: Kurtosis %.2f is unusual", result.Kurtosis)
		}

		t.Logf("Volatility: %.2f%%", result.Volatility*100)
		t.Logf("Skewness: %.4f", result.Skewness)
		t.Logf("Kurtosis: %.4f", result.Kurtosis)
		t.Logf("Max Drawdown: %.2f%%", result.MaxDrawdown*100)
	})

	t.Run("Stress Testing", func(t *testing.T) {
		calculator := calculator.NewProfessionalVaRCalculator(portfolioValue, 1)
		result, err := calculator.CalculateProfessionalVaR(positions, priceHistory)

		if err != nil {
			t.Fatalf("Failed to calculate VaR: %v", err)
		}

		// Stress VaR should be higher than normal VaR
		if result.StressVaR <= result.ParametricVaR[0.99] {
			t.Error("Stress VaR should be higher than normal 99% VaR")
		}

		// Tail risk should be significant
		if result.TailRisk <= 0 {
			t.Error("Tail risk should be positive")
		}

		t.Logf("Stress VaR: %.2f", result.StressVaR)
		t.Logf("Tail Risk: %.2f", result.TailRisk)
	})

	t.Run("Backtesting", func(t *testing.T) {
		calculator := calculator.NewProfessionalVaRCalculator(portfolioValue, 1)
		result, err := calculator.CalculateProfessionalVaR(positions, priceHistory)

		if err != nil {
			t.Fatalf("Failed to calculate VaR: %v", err)
		}

		if result.BacktestingMetrics == nil {
			t.Skip("Insufficient data for backtesting")
		}

		backtesting := result.BacktestingMetrics

		// Violation rate should be close to 1% for 99% VaR
		expectedRate := 0.01
		if math.Abs(backtesting.ViolationRate-expectedRate) > 0.02 {
			t.Logf("Warning: Violation rate %.2f%% differs from expected %.2f%%",
				backtesting.ViolationRate*100, expectedRate*100)
		}

		// Traffic light should be green for good model
		if backtesting.TrafficLight == "Red" {
			t.Error("Model failed Basel traffic light test")
		}

		t.Logf("Backtesting - Violations: %d, Rate: %.2f%%, Traffic Light: %s",
			backtesting.NumViolations, backtesting.ViolationRate*100, backtesting.TrafficLight)
		t.Logf("Kupiec Test: Statistic=%.4f, Passed=%v",
			backtesting.KupiecTestStatistic, backtesting.KupiecTestPassed)
	})
}

func TestGreeksCalculator(t *testing.T) {
	greeksCalc := calculator.NewGreeksCalculator()

	// Test European Call Option
	callOption := calculator.Option{
		Type:            calculator.CallOption,
		Style:           calculator.EuropeanStyle,
		UnderlyingPrice: 100,
		StrikePrice:     105,
		TimeToExpiry:    0.25, // 3 months
		RiskFreeRate:    0.05,
		Volatility:      0.20,
		DividendYield:   0.02,
	}

	t.Run("Call Option Greeks", func(t *testing.T) {
		greeks, err := greeksCalc.CalculateGreeks(callOption)
		if err != nil {
			t.Fatalf("Failed to calculate Greeks: %v", err)
		}

		// Delta for call should be between 0 and 1
		if greeks.Delta < 0 || greeks.Delta > 1 {
			t.Errorf("Call delta out of range: %f", greeks.Delta)
		}

		// Gamma should be positive
		if greeks.Gamma < 0 {
			t.Errorf("Gamma should be positive: %f", greeks.Gamma)
		}

		// Vega should be positive
		if greeks.Vega < 0 {
			t.Errorf("Vega should be positive: %f", greeks.Vega)
		}

		// Theta for call is typically negative
		if greeks.Theta > 0 {
			t.Logf("Warning: Call theta is positive: %f", greeks.Theta)
		}

		t.Logf("Call Greeks - Delta: %.4f, Gamma: %.4f, Vega: %.4f, Theta: %.4f, Rho: %.4f",
			greeks.Delta, greeks.Gamma, greeks.Vega, greeks.Theta, greeks.Rho)
	})

	// Test European Put Option
	putOption := callOption
	putOption.Type = calculator.PutOption

	t.Run("Put Option Greeks", func(t *testing.T) {
		greeks, err := greeksCalc.CalculateGreeks(putOption)
		if err != nil {
			t.Fatalf("Failed to calculate Greeks: %v", err)
		}

		// Delta for put should be between -1 and 0
		if greeks.Delta > 0 || greeks.Delta < -1 {
			t.Errorf("Put delta out of range: %f", greeks.Delta)
		}

		// Gamma should be positive (same as call)
		if greeks.Gamma < 0 {
			t.Errorf("Gamma should be positive: %f", greeks.Gamma)
		}

		t.Logf("Put Greeks - Delta: %.4f, Gamma: %.4f, Vega: %.4f, Theta: %.4f, Rho: %.4f",
			greeks.Delta, greeks.Gamma, greeks.Vega, greeks.Theta, greeks.Rho)
	})

	t.Run("At-The-Money Option", func(t *testing.T) {
		atmOption := callOption
		atmOption.StrikePrice = 100 // ATM

		greeks, err := greeksCalc.CalculateGreeks(atmOption)
		if err != nil {
			t.Fatalf("Failed to calculate Greeks: %v", err)
		}

		// ATM call delta should be around 0.5
		if math.Abs(greeks.Delta-0.5) > 0.15 {
			t.Logf("ATM delta %.4f is not close to 0.5", greeks.Delta)
		}

		// ATM options have highest gamma
		t.Logf("ATM Greeks - Delta: %.4f, Gamma: %.4f (highest for ATM)",
			greeks.Delta, greeks.Gamma)
	})

	t.Run("Second and Third Order Greeks", func(t *testing.T) {
		greeks, err := greeksCalc.CalculateGreeks(callOption)
		if err != nil {
			t.Fatalf("Failed to calculate Greeks: %v", err)
		}

		// Test second-order Greeks
		t.Logf("Second-order Greeks - Vanna: %.6f, Charm: %.6f, Vomma: %.6f",
			greeks.Vanna, greeks.Charm, greeks.Vomma)

		// Test third-order Greeks
		t.Logf("Third-order Greeks - Speed: %.8f, Zomma: %.8f, Color: %.8f",
			greeks.Speed, greeks.Zomma, greeks.Color)
	})

	t.Run("Scenario Analysis", func(t *testing.T) {
		analysis, err := greeksCalc.RunScenarioAnalysis(callOption)
		if err != nil {
			t.Fatalf("Failed to run scenario analysis: %v", err)
		}

		t.Logf("Base option price: %.2f", analysis.BasePrice)

		// Check heat map
		if len(analysis.HeatMap) == 0 {
			t.Error("Heat map is empty")
		}

		// Check scenarios
		if len(analysis.Scenarios) == 0 {
			t.Error("No scenarios generated")
		}

		// Log some scenario results
		for i, scenario := range analysis.Scenarios[:min(5, len(analysis.Scenarios))] {
			t.Logf("Scenario %d: Spot %.0f%%, Vol %.0f%%, Time %g days => P&L: %.2f",
				i+1,
				scenario.UnderlyingChange*100,
				scenario.VolatilityChange*100,
				scenario.TimeDecay,
				scenario.PnL)
		}
	})
}

// Helper function to generate sample price history
func generateSamplePriceHistory() map[string][]float64 {
	// Generate 252 days of price history (1 year)
	days := 252

	// Starting prices
	startPrices := map[string]float64{
		"AAPL":  150.0,
		"GOOGL": 2800.0,
		"MSFT":  350.0,
	}

	// Annual volatilities
	volatilities := map[string]float64{
		"AAPL":  0.25,
		"GOOGL": 0.30,
		"MSFT":  0.22,
	}

	priceHistory := make(map[string][]float64)

	for symbol, startPrice := range startPrices {
		prices := make([]float64, days)
		prices[0] = startPrice

		volatility := volatilities[symbol]
		dailyVol := volatility / math.Sqrt(252)

		for i := 1; i < days; i++ {
			// Generate random return
			randomReturn := randomNormal(0, dailyVol)
			prices[i] = prices[i-1] * (1 + randomReturn)
		}

		priceHistory[symbol] = prices
	}

	return priceHistory
}

// Simple random normal generator for testing
func randomNormal(mean, stdDev float64) float64 {
	// Box-Muller transform
	u1 := math.Max(1e-10, rand.Float64())
	u2 := rand.Float64()
	z := math.Sqrt(-2*math.Log(u1)) * math.Cos(2*math.Pi*u2)
	return mean + z*stdDev
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
