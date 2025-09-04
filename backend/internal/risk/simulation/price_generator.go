// backend/internal/risk/simulation/price_generator.go
package simulation

import (
	"math"
	"math/rand"
)

// PriceGenerator generates simulated price paths using various models
type PriceGenerator struct {
	rng *rand.Rand
}

// NewPriceGenerator creates a new price path generator
func NewPriceGenerator() *PriceGenerator {
	return &PriceGenerator{
		rng: rand.New(rand.NewSource(rand.Int63())),
	}
}

// GeneratePricePath creates a price path using Geometric Brownian Motion
func (pg *PriceGenerator) GeneratePricePath(initialPrice, volatility float64, days int, correlationMatrix [][]float64) []float64 {
	path := make([]float64, days+1)
	path[0] = initialPrice

	dt := 1.0 / 252.0 // Daily time step (assuming 252 trading days per year)
	drift := 0.05     // Assumed annual drift rate (5%)

	for i := 1; i <= days; i++ {
		// Generate random shock
		shock := pg.rng.NormFloat64()

		// Apply geometric Brownian motion
		logReturn := (drift-0.5*volatility*volatility)*dt + volatility*math.Sqrt(dt)*shock
		path[i] = path[i-1] * math.Exp(logReturn)
	}

	return path
}

// GenerateCorrelatedPaths generates correlated price paths for multiple assets
func (pg *PriceGenerator) GenerateCorrelatedPaths(initialPrices map[string]float64, volatilities map[string]float64, correlationMatrix [][]float64, days int) map[string][]float64 {
	symbols := make([]string, 0, len(initialPrices))
	for symbol := range initialPrices {
		symbols = append(symbols, symbol)
	}

	paths := make(map[string][]float64)
	for _, symbol := range symbols {
		paths[symbol] = make([]float64, days+1)
		paths[symbol][0] = initialPrices[symbol]
	}

	dt := 1.0 / 252.0
	drift := 0.05

	// Cholesky decomposition of correlation matrix for correlated random numbers
	cholesky := pg.choleskyDecomposition(correlationMatrix)

	for day := 1; day <= days; day++ {
		// Generate independent random shocks
		independentShocks := make([]float64, len(symbols))
		for i := range independentShocks {
			independentShocks[i] = pg.rng.NormFloat64()
		}

		// Apply correlation via Cholesky decomposition
		correlatedShocks := pg.multiplyMatrix(cholesky, independentShocks)

		// Apply shocks to each asset
		for i, symbol := range symbols {
			vol := volatilities[symbol]
			logReturn := (drift-0.5*vol*vol)*dt + vol*math.Sqrt(dt)*correlatedShocks[i]
			paths[symbol][day] = paths[symbol][day-1] * math.Exp(logReturn)
		}
	}

	return paths
}

// choleskyDecomposition performs Cholesky decomposition of a correlation matrix
func (pg *PriceGenerator) choleskyDecomposition(matrix [][]float64) [][]float64 {
	n := len(matrix)
	result := make([][]float64, n)
	for i := range result {
		result[i] = make([]float64, n)
	}

	for i := 0; i < n; i++ {
		for j := 0; j <= i; j++ {
			if i == j {
				sum := 0.0
				for k := 0; k < j; k++ {
					sum += result[j][k] * result[j][k]
				}
				result[i][j] = math.Sqrt(matrix[i][i] - sum)
			} else {
				sum := 0.0
				for k := 0; k < j; k++ {
					sum += result[i][k] * result[j][k]
				}
				result[i][j] = (matrix[i][j] - sum) / result[j][j]
			}
		}
	}

	return result
}

// multiplyMatrix multiplies matrix by vector
func (pg *PriceGenerator) multiplyMatrix(matrix [][]float64, vector []float64) []float64 {
	result := make([]float64, len(vector))
	for i := 0; i < len(matrix); i++ {
		for j := 0; j < len(vector); j++ {
			result[i] += matrix[i][j] * vector[j]
		}
	}
	return result
}

// GenerateJumpDiffusionPath generates price path with jumps (Merton jump-diffusion model)
func (pg *PriceGenerator) GenerateJumpDiffusionPath(initialPrice, volatility, jumpIntensity, jumpMean, jumpStd float64, days int) []float64 {
	path := make([]float64, days+1)
	path[0] = initialPrice

	dt := 1.0 / 252.0
	drift := 0.05

	for i := 1; i <= days; i++ {
		// Diffusion component
		shock := pg.rng.NormFloat64()
		diffusion := (drift-0.5*volatility*volatility)*dt + volatility*math.Sqrt(dt)*shock

		// Jump component
		jumpComponent := 0.0
		if pg.rng.Float64() < jumpIntensity*dt {
			jumpSize := pg.rng.NormFloat64()*jumpStd + jumpMean
			jumpComponent = jumpSize
		}

		// Apply price evolution
		logReturn := diffusion + jumpComponent
		path[i] = path[i-1] * math.Exp(logReturn)
	}

	return path
}

// GenerateRegimeSwitchingPath generates path with regime switching volatility
func (pg *PriceGenerator) GenerateRegimeSwitchingPath(initialPrice float64, regimeParams map[string]RegimeParams, days int) []float64 {
	path := make([]float64, days+1)
	path[0] = initialPrice

	// Start in normal regime
	currentRegime := "normal"
	dt := 1.0 / 252.0

	for i := 1; i <= days; i++ {
		params := regimeParams[currentRegime]

		// Check for regime switch
		if pg.rng.Float64() < params.TransitionProb*dt {
			// Switch regime
			if currentRegime == "normal" {
				currentRegime = "stressed"
			} else {
				currentRegime = "normal"
			}
			params = regimeParams[currentRegime]
		}

		// Generate price movement with current regime parameters
		shock := pg.rng.NormFloat64()
		logReturn := (params.Drift-0.5*params.Volatility*params.Volatility)*dt + params.Volatility*math.Sqrt(dt)*shock
		path[i] = path[i-1] * math.Exp(logReturn)
	}

	return path
}

// RegimeParams holds parameters for each market regime
type RegimeParams struct {
	Volatility     float64 `json:"volatility"`
	Drift          float64 `json:"drift"`
	TransitionProb float64 `json:"transition_prob"`
}
