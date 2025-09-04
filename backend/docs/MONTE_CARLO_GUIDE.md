# Monte Carlo Risk Assessment Guide

## Overview

This Monte Carlo simulation framework provides comprehensive testing and validation of your financial risk metrics and calculators. It helps you assess the accuracy, efficiency, and reliability of your risk management system.

## Features

### 🎲 Monte Carlo Simulation Engine
- **Multiple Price Models**: Geometric Brownian Motion, Jump Diffusion, Regime Switching
- **Correlated Asset Simulation**: Cholesky decomposition for realistic asset correlations
- **Market Regime Modeling**: Normal, Stressed, and Crisis market conditions
- **Parallel Processing**: Multi-threaded simulations for performance

### 📊 Risk Metrics Validation
- **Value at Risk (VaR)**: 95% and 99% confidence levels
- **Expected Shortfall (CVaR)**: Conditional Value at Risk
- **Risk Decomposition**: Position-level risk contributions
- **Liquidity Risk**: Time to liquidate under different market conditions

### 🔍 Accuracy Testing
- **Backtesting Framework**: Statistical validation of VaR models
- **Kupiec Test**: Frequency of VaR violations
- **Christoffersen Test**: Independence of VaR violations
- **Convergence Analysis**: Optimal number of simulations

### ⚡ Performance Benchmarking
- **Execution Speed**: Simulations per second
- **Memory Usage**: Resource consumption monitoring
- **Scalability Testing**: Multi-worker performance
- **Efficiency Metrics**: Cost-benefit analysis

## Getting Started

### 1. Build the Monte Carlo Tool

```bash
cd backend
go build -o monte-carlo-tool ./cmd/monte_carlo
```

### 2. Run Basic Simulation

```bash
./monte-carlo-tool
```

### 3. Interpret Results

The tool will output:
- Portfolio composition and total value
- VaR estimates at different confidence levels
- Expected Shortfall calculations
- Risk decomposition by asset
- Liquidity metrics and stress test results
- Performance benchmarks

## Configuration Options

### Simulation Parameters

```go
MonteCarloConfig{
    NumSimulations:    10000,        // Number of Monte Carlo paths
    TimeHorizonDays:   22,           // Risk horizon (trading days)
    ConfidenceLevel:   0.95,         // VaR confidence level
    NumWorkers:        4,            // Parallel workers
    RandomSeed:        12345,        // For reproducible results
    MarketRegime:      "NORMAL",     // NORMAL, STRESSED, CRISIS
    CorrelationMatrix: true,         // Enable asset correlations
}
```

### Market Regimes

- **NORMAL**: Standard market volatility
- **STRESSED**: 1.5x volatility increase
- **CRISIS**: 2x volatility increase

## Understanding the Results

### VaR Validation Criteria

✅ **Good VaR Model**:
- VaR99 > VaR95 (ordering property)
- VaR95 between 1-30% of portfolio value
- Backtesting score > 80%
- Convergence < 2% with sufficient simulations

### Liquidity Assessment

- **High Liquidity** (Score 0.8-1.0): Major stocks, ETFs
- **Medium Liquidity** (Score 0.5-0.8): Mid-cap stocks, major crypto
- **Low Liquidity** (Score 0.0-0.5): Small caps, commodities

### Performance Benchmarks

- **Fast**: >1000 simulations/second
- **Medium**: 100-1000 simulations/second  
- **Slow**: <100 simulations/second

## Best Practices

### 1. Choosing Simulation Count
- **Development**: 1,000-5,000 simulations
- **Testing**: 10,000-50,000 simulations
- **Production**: 100,000+ simulations

### 2. Time Horizons
- **Daily VaR**: 1 day
- **Monthly VaR**: 22 days
- **Quarterly VaR**: 66 days

### 3. Validation Frequency
- Run Monte Carlo validation monthly
- Backtest VaR models quarterly
- Stress test during market volatility

## Integration with Your Risk Engine

### Adding New Assets

Update the volatility estimates in `monte_carlo_extensions.go`:

```go
func (mc *MonteCarloSimulator) estimateVolatility(symbol string, marketRegime string) float64 {
    baseVolatilities := map[string]float64{
        "YOUR_ASSET": 0.30, // 30% annual volatility
        // ... existing assets
    }
    // ... rest of function
}
```

### Custom Market Data

Implement the `MarketDataProvider` interface:

```go
type YourMarketDataProvider struct {
    // Your market data source
}

func (m *YourMarketDataProvider) GetAverageDailyVolume(symbol string) float64 {
    // Return actual volume data
}
// ... implement other methods
```

### API Integration

Add Monte Carlo endpoint to your risk service:

```go
func (h *RiskHandler) RunMonteCarloSimulation(c *fiber.Ctx) error {
    // Parse request
    var config simulation.MonteCarloConfig
    if err := c.BodyParser(&config); err != nil {
        return c.Status(400).JSON(fiber.Map{"error": "Invalid config"})
    }
    
    // Get portfolio positions
    positions := getPortfolioPositions(c.Locals("user_id").(string))
    
    // Run simulation
    simulator := simulation.NewMonteCarloSimulator(varCalc, liquidityCalc)
    result, err := simulator.RunSimulation(positions, config)
    
    return c.JSON(result)
}
```

## Troubleshooting

### Common Issues

1. **High Memory Usage**
   - Reduce `NumSimulations`
   - Increase `NumWorkers` for better memory distribution
   - Use smaller `TimeHorizonDays`

2. **Slow Performance**
   - Increase `NumWorkers` (up to CPU cores)
   - Disable `CorrelationMatrix` for faster execution
   - Use smaller portfolios for testing

3. **Unrealistic VaR Values**
   - Check asset volatility estimates
   - Verify portfolio composition
   - Review market regime settings

4. **Poor Convergence**
   - Increase `NumSimulations`
   - Check for coding errors in price generation
   - Verify random seed consistency

### Debug Mode

Enable detailed logging by setting environment variable:

```bash
export MONTE_CARLO_DEBUG=true
./monte-carlo-tool
```

## Advanced Features

### Custom Price Models

Extend the `PriceGenerator` for specialized assets:

```go
func (pg *PriceGenerator) GenerateCustomPath(params CustomParams) []float64 {
    // Your custom price model
}
```

### Risk Factor Models

Add factor-based risk models:

```go
func (mc *MonteCarloSimulator) GenerateFactorReturns(factors []Factor) map[string][]float64 {
    // Multi-factor model implementation
}
```

### Real-time Monitoring

Set up automated validation:

```bash
# Run daily VaR validation
0 9 * * * cd /path/to/backend && ./monte-carlo-tool >> logs/monte_carlo.log 2>&1
```

## Performance Optimization

### Hardware Recommendations
- **CPU**: Multi-core processor (8+ cores recommended)
- **RAM**: 8GB+ for large portfolios
- **Storage**: SSD for faster I/O

### Software Optimizations
- Use Go build optimizations: `go build -ldflags="-s -w"`
- Profile memory usage: `go tool pprof`
- Benchmark different algorithms

## Conclusion

This Monte Carlo framework provides robust validation of your risk metrics, helping ensure your financial risk monitoring system is accurate, efficient, and reliable. Regular use of these tools will help maintain confidence in your risk assessments and regulatory compliance.

For questions or support, refer to the codebase documentation or create an issue in the repository.
