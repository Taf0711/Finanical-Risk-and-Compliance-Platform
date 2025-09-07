package optimization

import (
	"fmt"
	"math"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/optimize"
)

// OptimizationObjective represents the optimization goal
type OptimizationObjective string

const (
	MaximizeSharpeRatio      OptimizationObjective = "MAX_SHARPE"
	MinimizeRisk             OptimizationObjective = "MIN_RISK"
	MaximizeReturn           OptimizationObjective = "MAX_RETURN"
	MinimizeTrackingError    OptimizationObjective = "MIN_TRACKING_ERROR"
	MaximizeInformationRatio OptimizationObjective = "MAX_INFO_RATIO"
	RiskParity               OptimizationObjective = "RISK_PARITY"
)

// Constraint represents a portfolio constraint
type Constraint struct {
	Type        string                `json:"type"`
	Assets      []string              `json:"assets,omitempty"`
	MinWeight   float64               `json:"min_weight,omitempty"`
	MaxWeight   float64               `json:"max_weight,omitempty"`
	TargetValue float64               `json:"target_value,omitempty"`
	Bounds      map[string][2]float64 `json:"bounds,omitempty"`
}

// OptimizationConfig contains optimization parameters
type OptimizationConfig struct {
	Objective         OptimizationObjective `json:"objective"`
	RiskFreeRate      float64               `json:"risk_free_rate"`
	TargetReturn      float64               `json:"target_return,omitempty"`
	TargetRisk        float64               `json:"target_risk,omitempty"`
	BenchmarkWeights  []float64             `json:"benchmark_weights,omitempty"`
	Constraints       []Constraint          `json:"constraints"`
	MaxIterations     int                   `json:"max_iterations"`
	Tolerance         float64               `json:"tolerance"`
	RiskAversionParam float64               `json:"risk_aversion_param"`
}

// OptimizationResult contains optimization results
type OptimizationResult struct {
	OptimalWeights       map[string]float64 `json:"optimal_weights"`
	ExpectedReturn       float64            `json:"expected_return"`
	ExpectedRisk         float64            `json:"expected_risk"`
	SharpeRatio          float64            `json:"sharpe_ratio"`
	SortinoRatio         float64            `json:"sortino_ratio"`
	MaxDrawdown          float64            `json:"max_drawdown"`
	InformationRatio     float64            `json:"information_ratio,omitempty"`
	TrackingError        float64            `json:"tracking_error,omitempty"`
	DiversificationRatio float64            `json:"diversification_ratio"`
	ConcentrationRisk    float64            `json:"concentration_risk"`
	EffectiveN           float64            `json:"effective_n"`
	Converged            bool               `json:"converged"`
	Iterations           int                `json:"iterations"`
}

// PortfolioOptimizer performs portfolio optimization
type PortfolioOptimizer struct {
	assets           []string
	returns          *mat.Dense
	expectedReturns  []float64
	covarianceMatrix *mat.SymDense
	config           OptimizationConfig
}

// NewPortfolioOptimizer creates a new portfolio optimizer
func NewPortfolioOptimizer(
	positions []models.Position,
	priceHistory map[string][]float64,
	config OptimizationConfig,
) (*PortfolioOptimizer, error) {

	// Extract asset symbols
	assets := make([]string, len(positions))
	for i, pos := range positions {
		assets[i] = pos.Symbol
	}

	// Calculate returns matrix
	returnsMatrix := calculateReturnsMatrix(positions, priceHistory)
	if returnsMatrix == nil {
		return nil, fmt.Errorf("insufficient price history for optimization")
	}

	// Calculate expected returns and covariance
	expectedReturns := calculateExpectedReturns(returnsMatrix)
	covMatrix := calculateCovarianceMatrix(returnsMatrix)

	return &PortfolioOptimizer{
		assets:           assets,
		returns:          returnsMatrix,
		expectedReturns:  expectedReturns,
		covarianceMatrix: covMatrix,
		config:           config,
	}, nil
}

// Optimize performs portfolio optimization
func (po *PortfolioOptimizer) Optimize() (*OptimizationResult, error) {
	n := len(po.assets)
	if n == 0 {
		return nil, fmt.Errorf("no assets to optimize")
	}

	// Initial weights (equal weight)
	initialWeights := make([]float64, n)
	for i := range initialWeights {
		initialWeights[i] = 1.0 / float64(n)
	}

	var optimalWeights []float64
	var err error

	switch po.config.Objective {
	case MaximizeSharpeRatio:
		optimalWeights, err = po.optimizeSharpeRatio(initialWeights)
	case MinimizeRisk:
		optimalWeights, err = po.minimizeRisk(initialWeights)
	case MaximizeReturn:
		optimalWeights, err = po.maximizeReturn(initialWeights)
	case RiskParity:
		optimalWeights, err = po.optimizeRiskParity(initialWeights)
	case MinimizeTrackingError:
		optimalWeights, err = po.minimizeTrackingError(initialWeights)
	default:
		return nil, fmt.Errorf("unsupported optimization objective: %s", po.config.Objective)
	}

	if err != nil {
		return nil, err
	}

	// Build result
	result := po.buildOptimizationResult(optimalWeights)

	return result, nil
}

// optimizeSharpeRatio maximizes the Sharpe ratio
func (po *PortfolioOptimizer) optimizeSharpeRatio(initialWeights []float64) ([]float64, error) {
	// Objective function: negative Sharpe ratio (for minimization)
	objective := func(weights []float64) float64 {
		portfolioReturn := po.calculatePortfolioReturn(weights)
		portfolioRisk := po.calculatePortfolioRisk(weights)

		if portfolioRisk == 0 {
			return math.Inf(1)
		}

		sharpeRatio := (portfolioReturn - po.config.RiskFreeRate) / portfolioRisk
		return -sharpeRatio // Negative because we're minimizing
	}

	// Gradient function
	gradient := func(grad, weights []float64) {
		h := 1e-6
		for i := range weights {
			weightsPlus := make([]float64, len(weights))
			copy(weightsPlus, weights)
			weightsPlus[i] += h

			weightsMinus := make([]float64, len(weights))
			copy(weightsMinus, weights)
			weightsMinus[i] -= h

			grad[i] = (objective(weightsPlus) - objective(weightsMinus)) / (2 * h)
		}
	}

	// Set up optimization problem
	problem := optimize.Problem{
		Func: objective,
		Grad: gradient,
	}

	// Add constraints
	constraints := po.buildConstraints()

	// Run optimization
	result, err := optimize.Minimize(problem, initialWeights, &optimize.Settings{
		GradientThreshold: po.config.Tolerance,
		FuncEvaluations:   po.config.MaxIterations,
	}, &optimize.LBFGS{})

	if err != nil {
		return nil, err
	}

	// Apply constraints and normalize
	optimalWeights := po.applyConstraints(result.X, constraints)

	return optimalWeights, nil
}

// minimizeRisk minimizes portfolio risk
func (po *PortfolioOptimizer) minimizeRisk(initialWeights []float64) ([]float64, error) {
	// Objective function: portfolio variance
	objective := func(weights []float64) float64 {
		variance := po.calculatePortfolioVariance(weights)
		return variance
	}

	// For minimum variance portfolio, we can use analytical solution
	n := len(po.assets)
	ones := mat.NewVecDense(n, nil)
	for i := 0; i < n; i++ {
		ones.SetVec(i, 1.0)
	}

	// Solve: w = (Σ^-1 * 1) / (1' * Σ^-1 * 1)
	var invCov mat.Dense
	err := invCov.Inverse(po.covarianceMatrix)
	if err != nil {
		// Use numerical optimization if matrix is singular
		return po.numericalOptimization(objective, initialWeights)
	}

	invCovOnes := mat.NewVecDense(n, nil)
	invCovOnes.MulVec(&invCov, ones)

	denominator := mat.Dot(ones, invCovOnes)

	optimalWeights := make([]float64, n)
	for i := 0; i < n; i++ {
		optimalWeights[i] = invCovOnes.AtVec(i) / denominator
	}

	// Apply constraints
	constraints := po.buildConstraints()
	optimalWeights = po.applyConstraints(optimalWeights, constraints)

	return optimalWeights, nil
}

// maximizeReturn maximizes expected return with risk constraint
func (po *PortfolioOptimizer) maximizeReturn(initialWeights []float64) ([]float64, error) {
	// Objective function: negative expected return (for minimization)
	objective := func(weights []float64) float64 {
		return -po.calculatePortfolioReturn(weights)
	}

	// Risk constraint
	riskConstraint := func(weights []float64) float64 {
		risk := po.calculatePortfolioRisk(weights)
		return po.config.TargetRisk - risk // Should be >= 0
	}

	return po.numericalOptimizationWithConstraint(objective, riskConstraint, initialWeights)
}

// optimizeRiskParity creates a risk parity portfolio
func (po *PortfolioOptimizer) optimizeRiskParity(initialWeights []float64) ([]float64, error) {
	n := len(po.assets)

	// Risk parity: each asset contributes equally to portfolio risk
	// Minimize: Σ(RC_i - 1/n)^2 where RC_i is risk contribution of asset i
	objective := func(weights []float64) float64 {
		// Calculate marginal risk contributions
		portfolioRisk := po.calculatePortfolioRisk(weights)
		if portfolioRisk == 0 {
			return math.Inf(1)
		}

		marginalContributions := po.calculateMarginalRiskContributions(weights)

		// Risk contribution = weight * marginal contribution
		targetContribution := 1.0 / float64(n)
		sumSquaredDiff := 0.0

		for i := range weights {
			riskContribution := weights[i] * marginalContributions[i] / portfolioRisk
			diff := riskContribution - targetContribution
			sumSquaredDiff += diff * diff
		}

		return sumSquaredDiff
	}

	return po.numericalOptimization(objective, initialWeights)
}

// minimizeTrackingError minimizes tracking error vs benchmark
func (po *PortfolioOptimizer) minimizeTrackingError(initialWeights []float64) ([]float64, error) {
	if len(po.config.BenchmarkWeights) != len(po.assets) {
		return nil, fmt.Errorf("benchmark weights dimension mismatch")
	}

	// Objective: minimize tracking error variance
	objective := func(weights []float64) float64 {
		trackingError := 0.0

		// Calculate active weights
		for i := range weights {
			activeWeight := weights[i] - po.config.BenchmarkWeights[i]
			for j := range weights {
				activeWeightJ := weights[j] - po.config.BenchmarkWeights[j]
				trackingError += activeWeight * activeWeightJ * po.covarianceMatrix.At(i, j)
			}
		}

		return trackingError
	}

	return po.numericalOptimization(objective, initialWeights)
}

// calculatePortfolioReturn calculates expected portfolio return
func (po *PortfolioOptimizer) calculatePortfolioReturn(weights []float64) float64 {
	portfolioReturn := 0.0
	for i, w := range weights {
		portfolioReturn += w * po.expectedReturns[i]
	}
	return portfolioReturn * 252 // Annualized
}

// calculatePortfolioRisk calculates portfolio standard deviation
func (po *PortfolioOptimizer) calculatePortfolioRisk(weights []float64) float64 {
	variance := po.calculatePortfolioVariance(weights)
	return math.Sqrt(variance) * math.Sqrt(252) // Annualized
}

// calculatePortfolioVariance calculates portfolio variance
func (po *PortfolioOptimizer) calculatePortfolioVariance(weights []float64) float64 {
	n := len(weights)
	variance := 0.0

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			variance += weights[i] * weights[j] * po.covarianceMatrix.At(i, j)
		}
	}

	return variance
}

// calculateMarginalRiskContributions calculates marginal risk contribution of each asset
func (po *PortfolioOptimizer) calculateMarginalRiskContributions(weights []float64) []float64 {
	n := len(weights)
	marginalContributions := make([]float64, n)

	portfolioStdDev := po.calculatePortfolioRisk(weights)
	if portfolioStdDev == 0 {
		return marginalContributions
	}

	// Marginal contribution = ∂σ/∂w_i = (Σw)_i / σ
	for i := 0; i < n; i++ {
		contribution := 0.0
		for j := 0; j < n; j++ {
			contribution += weights[j] * po.covarianceMatrix.At(i, j)
		}
		marginalContributions[i] = contribution / (portfolioStdDev / math.Sqrt(252))
	}

	return marginalContributions
}

// buildConstraints builds optimization constraints
func (po *PortfolioOptimizer) buildConstraints() []Constraint {
	constraints := make([]Constraint, 0)

	// Default constraint: weights sum to 1
	constraints = append(constraints, Constraint{
		Type:        "SUM_TO_ONE",
		TargetValue: 1.0,
	})

	// Default constraint: no short selling (weights >= 0)
	constraints = append(constraints, Constraint{
		Type:      "LONG_ONLY",
		MinWeight: 0.0,
		MaxWeight: 1.0,
	})

	// Add user-defined constraints
	constraints = append(constraints, po.config.Constraints...)

	return constraints
}

// applyConstraints applies constraints to weights
func (po *PortfolioOptimizer) applyConstraints(weights []float64, constraints []Constraint) []float64 {
	n := len(weights)
	constrainedWeights := make([]float64, n)
	copy(constrainedWeights, weights)

	// Apply individual constraints
	for _, constraint := range constraints {
		switch constraint.Type {
		case "LONG_ONLY":
			for i := range constrainedWeights {
				if constrainedWeights[i] < 0 {
					constrainedWeights[i] = 0
				}
			}
		case "MAX_WEIGHT":
			for i := range constrainedWeights {
				if constrainedWeights[i] > constraint.MaxWeight {
					constrainedWeights[i] = constraint.MaxWeight
				}
			}
		case "MIN_WEIGHT":
			for i := range constrainedWeights {
				if constrainedWeights[i] < constraint.MinWeight && constrainedWeights[i] > 0 {
					constrainedWeights[i] = constraint.MinWeight
				}
			}
		}
	}

	// Normalize to sum to 1
	sum := 0.0
	for _, w := range constrainedWeights {
		sum += w
	}

	if sum > 0 {
		for i := range constrainedWeights {
			constrainedWeights[i] /= sum
		}
	}

	return constrainedWeights
}

// numericalOptimization performs numerical optimization
func (po *PortfolioOptimizer) numericalOptimization(objective func([]float64) float64, initialWeights []float64) ([]float64, error) {
	problem := optimize.Problem{
		Func: objective,
	}

	result, err := optimize.Minimize(problem, initialWeights, &optimize.Settings{
		FuncEvaluations:   po.config.MaxIterations,
		GradientThreshold: po.config.Tolerance,
	}, &optimize.NelderMead{})

	if err != nil {
		return nil, err
	}

	constraints := po.buildConstraints()
	return po.applyConstraints(result.X, constraints), nil
}

// numericalOptimizationWithConstraint performs constrained optimization
func (po *PortfolioOptimizer) numericalOptimizationWithConstraint(
	objective func([]float64) float64,
	constraint func([]float64) float64,
	initialWeights []float64,
) ([]float64, error) {

	// Use penalty method for constraint
	penaltyObjective := func(weights []float64) float64 {
		obj := objective(weights)

		// Add penalty for constraint violation
		constraintValue := constraint(weights)
		if constraintValue < 0 {
			penalty := 1000.0 * constraintValue * constraintValue
			obj += penalty
		}

		return obj
	}

	return po.numericalOptimization(penaltyObjective, initialWeights)
}

// buildOptimizationResult builds the optimization result
func (po *PortfolioOptimizer) buildOptimizationResult(weights []float64) *OptimizationResult {
	result := &OptimizationResult{
		OptimalWeights: make(map[string]float64),
	}

	// Map weights to assets
	for i, asset := range po.assets {
		if weights[i] > 1e-6 { // Filter out very small weights
			result.OptimalWeights[asset] = weights[i]
		}
	}

	// Calculate portfolio metrics
	result.ExpectedReturn = po.calculatePortfolioReturn(weights)
	result.ExpectedRisk = po.calculatePortfolioRisk(weights)

	// Sharpe ratio
	if result.ExpectedRisk > 0 {
		result.SharpeRatio = (result.ExpectedReturn - po.config.RiskFreeRate) / result.ExpectedRisk
	}

	// Sortino ratio (using downside deviation)
	downsideRisk := po.calculateDownsideRisk(weights)
	if downsideRisk > 0 {
		result.SortinoRatio = (result.ExpectedReturn - po.config.RiskFreeRate) / downsideRisk
	}

	// Maximum drawdown
	result.MaxDrawdown = po.calculateMaxDrawdown(weights)

	// Diversification metrics
	result.DiversificationRatio = po.calculateDiversificationRatio(weights)
	result.ConcentrationRisk = po.calculateConcentrationRisk(weights)
	result.EffectiveN = po.calculateEffectiveN(weights)

	// Tracking error (if benchmark provided)
	if len(po.config.BenchmarkWeights) > 0 {
		result.TrackingError = po.calculateTrackingError(weights)
		if result.TrackingError > 0 {
			activeReturn := result.ExpectedReturn - po.calculatePortfolioReturn(po.config.BenchmarkWeights)
			result.InformationRatio = activeReturn / result.TrackingError
		}
	}

	result.Converged = true

	return result
}

// calculateDownsideRisk calculates downside deviation
func (po *PortfolioOptimizer) calculateDownsideRisk(weights []float64) float64 {
	// Calculate portfolio returns
	numPeriods := po.returns.RawMatrix().Rows
	portfolioReturns := make([]float64, numPeriods)

	for t := 0; t < numPeriods; t++ {
		portfolioReturn := 0.0
		for i, w := range weights {
			portfolioReturn += w * po.returns.At(t, i)
		}
		portfolioReturns[t] = portfolioReturn
	}

	// Calculate downside deviation (only negative returns)
	threshold := po.config.RiskFreeRate / 252 // Daily risk-free rate
	sumSquaredDownside := 0.0
	count := 0

	for _, ret := range portfolioReturns {
		if ret < threshold {
			diff := ret - threshold
			sumSquaredDownside += diff * diff
			count++
		}
	}

	if count > 0 {
		return math.Sqrt(sumSquaredDownside/float64(count)) * math.Sqrt(252)
	}

	return 0
}

// calculateMaxDrawdown calculates maximum drawdown
func (po *PortfolioOptimizer) calculateMaxDrawdown(weights []float64) float64 {
	// Calculate cumulative portfolio returns
	numPeriods := po.returns.RawMatrix().Rows
	cumulative := 1.0
	peak := 1.0
	maxDrawdown := 0.0

	for t := 0; t < numPeriods; t++ {
		portfolioReturn := 0.0
		for i, w := range weights {
			portfolioReturn += w * po.returns.At(t, i)
		}

		cumulative *= (1 + portfolioReturn)

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

// calculateDiversificationRatio calculates diversification ratio
func (po *PortfolioOptimizer) calculateDiversificationRatio(weights []float64) float64 {
	// DR = (Σw_i*σ_i) / σ_p
	weightedAvgVolatility := 0.0

	for i, w := range weights {
		assetVolatility := math.Sqrt(po.covarianceMatrix.At(i, i)) * math.Sqrt(252)
		weightedAvgVolatility += w * assetVolatility
	}

	portfolioVolatility := po.calculatePortfolioRisk(weights)

	if portfolioVolatility > 0 {
		return weightedAvgVolatility / portfolioVolatility
	}

	return 1.0
}

// calculateConcentrationRisk calculates Herfindahl-Hirschman Index
func (po *PortfolioOptimizer) calculateConcentrationRisk(weights []float64) float64 {
	hhi := 0.0
	for _, w := range weights {
		hhi += w * w
	}
	return hhi
}

// calculateEffectiveN calculates effective number of assets
func (po *PortfolioOptimizer) calculateEffectiveN(weights []float64) float64 {
	hhi := po.calculateConcentrationRisk(weights)
	if hhi > 0 {
		return 1.0 / hhi
	}
	return 1.0
}

// calculateTrackingError calculates tracking error vs benchmark
func (po *PortfolioOptimizer) calculateTrackingError(weights []float64) float64 {
	if len(po.config.BenchmarkWeights) != len(weights) {
		return 0
	}

	trackingVariance := 0.0

	for i := range weights {
		activeWeight := weights[i] - po.config.BenchmarkWeights[i]
		for j := range weights {
			activeWeightJ := weights[j] - po.config.BenchmarkWeights[j]
			trackingVariance += activeWeight * activeWeightJ * po.covarianceMatrix.At(i, j)
		}
	}

	return math.Sqrt(trackingVariance) * math.Sqrt(252)
}

// Helper functions

func calculateReturnsMatrix(positions []models.Position, priceHistory map[string][]float64) *mat.Dense {
	if len(positions) == 0 || len(priceHistory) == 0 {
		return nil
	}

	// Find minimum history length
	minLength := math.MaxInt32
	for _, pos := range positions {
		if prices, exists := priceHistory[pos.Symbol]; exists {
			if len(prices) < minLength {
				minLength = len(prices)
			}
		}
	}

	if minLength < 2 {
		return nil
	}

	// Build returns matrix
	numAssets := len(positions)
	numReturns := minLength - 1

	returnsData := make([]float64, numReturns*numAssets)

	for j, pos := range positions {
		if prices, exists := priceHistory[pos.Symbol]; exists && len(prices) >= minLength {
			for i := 1; i < minLength; i++ {
				if prices[i-1] != 0 {
					ret := (prices[i] - prices[i-1]) / prices[i-1]
					returnsData[i-1+j*numReturns] = ret
				}
			}
		}
	}

	return mat.NewDense(numReturns, numAssets, returnsData)
}

func calculateExpectedReturns(returns *mat.Dense) []float64 {
	rows, cols := returns.Dims()
	expectedReturns := make([]float64, cols)

	for j := 0; j < cols; j++ {
		sum := 0.0
		for i := 0; i < rows; i++ {
			sum += returns.At(i, j)
		}
		expectedReturns[j] = sum / float64(rows)
	}

	return expectedReturns
}

func calculateCovarianceMatrix(returns *mat.Dense) *mat.SymDense {
	rows, cols := returns.Dims()

	// Center the returns
	centered := mat.NewDense(rows, cols, nil)
	expectedReturns := calculateExpectedReturns(returns)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			centered.Set(i, j, returns.At(i, j)-expectedReturns[j])
		}
	}

	// Calculate covariance matrix
	cov := mat.NewSymDense(cols, nil)

	for i := 0; i < cols; i++ {
		for j := i; j < cols; j++ {
			covariance := 0.0
			for k := 0; k < rows; k++ {
				covariance += centered.At(k, i) * centered.At(k, j)
			}
			covariance /= float64(rows - 1)
			cov.SetSym(i, j, covariance)
		}
	}

	return cov
}
