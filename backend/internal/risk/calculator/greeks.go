package calculator

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/stat/distuv"
)

// OptionType represents the type of option
type OptionType string

const (
	CallOption OptionType = "CALL"
	PutOption  OptionType = "PUT"
)

// OptionStyle represents the exercise style
type OptionStyle string

const (
	EuropeanStyle OptionStyle = "EUROPEAN"
	AmericanStyle OptionStyle = "AMERICAN"
)

// Option represents an option contract
type Option struct {
	Type            OptionType  `json:"type"`
	Style           OptionStyle `json:"style"`
	UnderlyingPrice float64     `json:"underlying_price"`
	StrikePrice     float64     `json:"strike_price"`
	TimeToExpiry    float64     `json:"time_to_expiry"` // in years
	RiskFreeRate    float64     `json:"risk_free_rate"`
	Volatility      float64     `json:"volatility"`
	DividendYield   float64     `json:"dividend_yield"`
}

// Greeks contains all option sensitivities
type Greeks struct {
	Delta  float64 `json:"delta"`  // Price sensitivity to underlying
	Gamma  float64 `json:"gamma"`  // Delta sensitivity to underlying
	Vega   float64 `json:"vega"`   // Price sensitivity to volatility
	Theta  float64 `json:"theta"`  // Price sensitivity to time
	Rho    float64 `json:"rho"`    // Price sensitivity to interest rate
	Lambda float64 `json:"lambda"` // Leverage (Omega)
	Vanna  float64 `json:"vanna"`  // Delta sensitivity to volatility
	Charm  float64 `json:"charm"`  // Delta sensitivity to time
	Vomma  float64 `json:"vomma"`  // Vega sensitivity to volatility
	Speed  float64 `json:"speed"`  // Gamma sensitivity to underlying
	Zomma  float64 `json:"zomma"`  // Gamma sensitivity to volatility
	Color  float64 `json:"color"`  // Gamma sensitivity to time
}

// GreeksCalculator calculates option Greeks
type GreeksCalculator struct {
	normal distuv.Normal
}

// NewGreeksCalculator creates a new Greeks calculator
func NewGreeksCalculator() *GreeksCalculator {
	return &GreeksCalculator{
		normal: distuv.Normal{Mu: 0, Sigma: 1},
	}
}

// CalculateGreeks calculates all Greeks for an option
func (g *GreeksCalculator) CalculateGreeks(opt Option) (*Greeks, error) {
	greeks := &Greeks{}

	// Calculate d1 and d2 for Black-Scholes
	d1, d2 := g.calculateD1D2(opt)

	// Standard normal PDF and CDF values
	nd1 := g.normal.CDF(d1)
	nd2 := g.normal.CDF(d2)
	npd1 := g.normal.Prob(d1)
	npd2 := g.normal.Prob(d2)

	// Calculate first-order Greeks
	greeks.Delta = g.calculateDelta(opt, d1, nd1)
	greeks.Gamma = g.calculateGamma(opt, d1, npd1)
	greeks.Vega = g.calculateVega(opt, d1, npd1)
	greeks.Theta = g.calculateTheta(opt, d1, d2, nd1, nd2, npd1, npd2)
	greeks.Rho = g.calculateRho(opt, d2, nd2)

	// Calculate leverage (Lambda/Omega)
	optionPrice := g.calculateOptionPrice(opt, d1, d2, nd1, nd2)
	if optionPrice != 0 {
		greeks.Lambda = greeks.Delta * opt.UnderlyingPrice / optionPrice
	}

	// Calculate second-order Greeks
	greeks.Vanna = g.calculateVanna(opt, d1, d2, npd1)
	greeks.Charm = g.calculateCharm(opt, d1, d2, npd1)
	greeks.Vomma = g.calculateVomma(opt, d1, npd1)

	// Calculate third-order Greeks
	greeks.Speed = g.calculateSpeed(opt, d1, npd1)
	greeks.Zomma = g.calculateZomma(opt, d1, npd1)
	greeks.Color = g.calculateColor(opt, d1, npd1)

	return greeks, nil
}

// calculateD1D2 calculates d1 and d2 for Black-Scholes formula
func (g *GreeksCalculator) calculateD1D2(opt Option) (float64, float64) {
	if opt.TimeToExpiry <= 0 || opt.Volatility <= 0 {
		return 0, 0
	}

	sqrtT := math.Sqrt(opt.TimeToExpiry)

	d1 := (math.Log(opt.UnderlyingPrice/opt.StrikePrice) +
		(opt.RiskFreeRate-opt.DividendYield+0.5*opt.Volatility*opt.Volatility)*opt.TimeToExpiry) /
		(opt.Volatility * sqrtT)

	d2 := d1 - opt.Volatility*sqrtT

	return d1, d2
}

// calculateDelta calculates the Delta of the option
func (g *GreeksCalculator) calculateDelta(opt Option, d1 float64, nd1 float64) float64 {
	discountFactor := math.Exp(-opt.DividendYield * opt.TimeToExpiry)

	if opt.Type == CallOption {
		return discountFactor * nd1
	}
	// Put option
	return discountFactor * (nd1 - 1)
}

// calculateGamma calculates the Gamma of the option
func (g *GreeksCalculator) calculateGamma(opt Option, d1 float64, npd1 float64) float64 {
	if opt.TimeToExpiry <= 0 || opt.Volatility <= 0 {
		return 0
	}

	discountFactor := math.Exp(-opt.DividendYield * opt.TimeToExpiry)
	denominator := opt.UnderlyingPrice * opt.Volatility * math.Sqrt(opt.TimeToExpiry)

	if denominator == 0 {
		return 0
	}

	return discountFactor * npd1 / denominator
}

// calculateVega calculates the Vega of the option
func (g *GreeksCalculator) calculateVega(opt Option, d1 float64, npd1 float64) float64 {
	if opt.TimeToExpiry <= 0 {
		return 0
	}

	discountFactor := math.Exp(-opt.DividendYield * opt.TimeToExpiry)
	sqrtT := math.Sqrt(opt.TimeToExpiry)

	// Vega is typically expressed per 1% change in volatility
	return opt.UnderlyingPrice * discountFactor * npd1 * sqrtT / 100
}

// calculateTheta calculates the Theta of the option
func (g *GreeksCalculator) calculateTheta(opt Option, d1, d2 float64, nd1, nd2, npd1, npd2 float64) float64 {
	if opt.TimeToExpiry <= 0 {
		return 0
	}

	sqrtT := math.Sqrt(opt.TimeToExpiry)
	discountS := opt.UnderlyingPrice * math.Exp(-opt.DividendYield*opt.TimeToExpiry)
	discountK := opt.StrikePrice * math.Exp(-opt.RiskFreeRate*opt.TimeToExpiry)

	// First term (common to both call and put)
	term1 := -(discountS * npd1 * opt.Volatility) / (2 * sqrtT)

	if opt.Type == CallOption {
		// Call option theta
		term2 := opt.DividendYield * discountS * nd1
		term3 := -opt.RiskFreeRate * discountK * nd2
		// Theta is typically expressed per day (divide by 365)
		return (term1 + term2 + term3) / 365
	}

	// Put option theta
	term2 := -opt.DividendYield * discountS * (1 - nd1)
	term3 := opt.RiskFreeRate * discountK * (1 - nd2)
	// Theta is typically expressed per day (divide by 365)
	return (term1 + term2 + term3) / 365
}

// calculateRho calculates the Rho of the option
func (g *GreeksCalculator) calculateRho(opt Option, d2 float64, nd2 float64) float64 {
	if opt.TimeToExpiry <= 0 {
		return 0
	}

	discountK := opt.StrikePrice * math.Exp(-opt.RiskFreeRate*opt.TimeToExpiry)

	if opt.Type == CallOption {
		// Rho is typically expressed per 1% change in interest rate
		return opt.TimeToExpiry * discountK * nd2 / 100
	}

	// Put option rho
	return -opt.TimeToExpiry * discountK * (1 - nd2) / 100
}

// calculateVanna calculates Vanna (∂Delta/∂σ = ∂Vega/∂S)
func (g *GreeksCalculator) calculateVanna(opt Option, d1, d2 float64, npd1 float64) float64 {
	if opt.TimeToExpiry <= 0 || opt.Volatility <= 0 {
		return 0
	}

	discountFactor := math.Exp(-opt.DividendYield * opt.TimeToExpiry)

	// Vanna = -e^(-qT) * n(d1) * d2 / σ
	return -discountFactor * npd1 * d2 / opt.Volatility
}

// calculateCharm calculates Charm (∂Delta/∂T)
func (g *GreeksCalculator) calculateCharm(opt Option, d1, d2 float64, npd1 float64) float64 {
	if opt.TimeToExpiry <= 0 || opt.Volatility <= 0 {
		return 0
	}

	sqrtT := math.Sqrt(opt.TimeToExpiry)
	discountFactor := math.Exp(-opt.DividendYield * opt.TimeToExpiry)

	term1 := opt.DividendYield * discountFactor * g.normal.CDF(d1)

	numerator := 2*(opt.RiskFreeRate-opt.DividendYield)*opt.TimeToExpiry - d2*opt.Volatility*sqrtT
	denominator := 2 * opt.TimeToExpiry * opt.Volatility * sqrtT

	if denominator == 0 {
		return 0
	}

	term2 := discountFactor * npd1 * numerator / denominator

	if opt.Type == CallOption {
		return -term1 + term2
	}

	// Put option charm
	return term1 + term2
}

// calculateVomma calculates Vomma/Volga (∂Vega/∂σ)
func (g *GreeksCalculator) calculateVomma(opt Option, d1 float64, npd1 float64) float64 {
	if opt.TimeToExpiry <= 0 || opt.Volatility <= 0 {
		return 0
	}

	sqrtT := math.Sqrt(opt.TimeToExpiry)
	d2 := d1 - opt.Volatility*sqrtT
	discountFactor := math.Exp(-opt.DividendYield * opt.TimeToExpiry)

	// Vomma = Vega * d1 * d2 / σ
	vega := opt.UnderlyingPrice * discountFactor * npd1 * sqrtT

	return vega * d1 * d2 / opt.Volatility
}

// calculateSpeed calculates Speed (∂Gamma/∂S)
func (g *GreeksCalculator) calculateSpeed(opt Option, d1 float64, npd1 float64) float64 {
	if opt.TimeToExpiry <= 0 || opt.Volatility <= 0 || opt.UnderlyingPrice <= 0 {
		return 0
	}

	sqrtT := math.Sqrt(opt.TimeToExpiry)
	discountFactor := math.Exp(-opt.DividendYield * opt.TimeToExpiry)

	gamma := discountFactor * npd1 / (opt.UnderlyingPrice * opt.Volatility * sqrtT)

	// Speed = -Gamma/S * (1 + d1/(σ*√T))
	return -gamma / opt.UnderlyingPrice * (1 + d1/(opt.Volatility*sqrtT))
}

// calculateZomma calculates Zomma (∂Gamma/∂σ)
func (g *GreeksCalculator) calculateZomma(opt Option, d1 float64, npd1 float64) float64 {
	if opt.TimeToExpiry <= 0 || opt.Volatility <= 0 {
		return 0
	}

	sqrtT := math.Sqrt(opt.TimeToExpiry)
	d2 := d1 - opt.Volatility*sqrtT
	discountFactor := math.Exp(-opt.DividendYield * opt.TimeToExpiry)

	gamma := discountFactor * npd1 / (opt.UnderlyingPrice * opt.Volatility * sqrtT)

	// Zomma = Gamma * (d1*d2 - 1) / σ
	return gamma * (d1*d2 - 1) / opt.Volatility
}

// calculateColor calculates Color (∂Gamma/∂T)
func (g *GreeksCalculator) calculateColor(opt Option, d1 float64, npd1 float64) float64 {
	if opt.TimeToExpiry <= 0 || opt.Volatility <= 0 {
		return 0
	}

	sqrtT := math.Sqrt(opt.TimeToExpiry)
	d2 := d1 - opt.Volatility*sqrtT
	discountFactor := math.Exp(-opt.DividendYield * opt.TimeToExpiry)

	numerator1 := 2*opt.DividendYield*opt.TimeToExpiry - d2*opt.Volatility*sqrtT
	denominator := 2 * opt.TimeToExpiry * opt.Volatility * sqrtT

	if denominator == 0 {
		return 0
	}

	term1 := discountFactor * npd1 / (2 * opt.UnderlyingPrice * opt.TimeToExpiry * opt.Volatility * sqrtT)
	term2 := 2*opt.DividendYield*opt.TimeToExpiry + 1 + numerator1*d1/denominator

	return -term1 * term2
}

// calculateOptionPrice calculates the Black-Scholes option price
func (g *GreeksCalculator) calculateOptionPrice(opt Option, d1, d2 float64, nd1, nd2 float64) float64 {
	if opt.TimeToExpiry <= 0 {
		// Option has expired
		if opt.Type == CallOption {
			return math.Max(opt.UnderlyingPrice-opt.StrikePrice, 0)
		}
		return math.Max(opt.StrikePrice-opt.UnderlyingPrice, 0)
	}

	discountS := opt.UnderlyingPrice * math.Exp(-opt.DividendYield*opt.TimeToExpiry)
	discountK := opt.StrikePrice * math.Exp(-opt.RiskFreeRate*opt.TimeToExpiry)

	if opt.Type == CallOption {
		return discountS*nd1 - discountK*nd2
	}

	// Put option price
	return discountK*(1-nd2) - discountS*(1-nd1)
}

// PortfolioGreeks calculates portfolio-level Greeks
type PortfolioGreeks struct {
	TotalDelta     float64            `json:"total_delta"`
	TotalGamma     float64            `json:"total_gamma"`
	TotalVega      float64            `json:"total_vega"`
	TotalTheta     float64            `json:"total_theta"`
	TotalRho       float64            `json:"total_rho"`
	PositionGreeks map[string]*Greeks `json:"position_greeks"`
}

// CalculatePortfolioGreeks calculates Greeks for a portfolio of options
func (g *GreeksCalculator) CalculatePortfolioGreeks(options map[string]Option, quantities map[string]float64) (*PortfolioGreeks, error) {
	portfolio := &PortfolioGreeks{
		PositionGreeks: make(map[string]*Greeks),
	}

	for symbol, opt := range options {
		greeks, err := g.CalculateGreeks(opt)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate Greeks for %s: %w", symbol, err)
		}

		quantity := quantities[symbol]
		if quantity == 0 {
			quantity = 1
		}

		// Store individual position Greeks
		portfolio.PositionGreeks[symbol] = greeks

		// Aggregate portfolio Greeks
		portfolio.TotalDelta += greeks.Delta * quantity
		portfolio.TotalGamma += greeks.Gamma * quantity
		portfolio.TotalVega += greeks.Vega * quantity
		portfolio.TotalTheta += greeks.Theta * quantity
		portfolio.TotalRho += greeks.Rho * quantity
	}

	return portfolio, nil
}

// ScenarioAnalysis performs scenario analysis on option Greeks
type ScenarioAnalysis struct {
	BasePrice float64                       `json:"base_price"`
	Scenarios []ScenarioResult              `json:"scenarios"`
	HeatMap   map[string]map[string]float64 `json:"heat_map"`
}

// ScenarioResult represents the result of a single scenario
type ScenarioResult struct {
	UnderlyingChange float64 `json:"underlying_change"`
	VolatilityChange float64 `json:"volatility_change"`
	TimeDecay        float64 `json:"time_decay"`
	NewPrice         float64 `json:"new_price"`
	PnL              float64 `json:"pnl"`
	NewGreeks        *Greeks `json:"new_greeks"`
}

// RunScenarioAnalysis performs scenario analysis on an option
func (g *GreeksCalculator) RunScenarioAnalysis(opt Option) (*ScenarioAnalysis, error) {
	d1, d2 := g.calculateD1D2(opt)
	nd1 := g.normal.CDF(d1)
	nd2 := g.normal.CDF(d2)
	basePrice := g.calculateOptionPrice(opt, d1, d2, nd1, nd2)

	analysis := &ScenarioAnalysis{
		BasePrice: basePrice,
		Scenarios: make([]ScenarioResult, 0),
		HeatMap:   make(map[string]map[string]float64),
	}

	// Define scenario ranges
	underlyingChanges := []float64{-0.20, -0.10, -0.05, 0, 0.05, 0.10, 0.20}
	volatilityChanges := []float64{-0.10, -0.05, 0, 0.05, 0.10}

	// Generate heat map for price changes
	for _, spotChange := range underlyingChanges {
		spotKey := fmt.Sprintf("%.0f%%", spotChange*100)
		analysis.HeatMap[spotKey] = make(map[string]float64)

		for _, volChange := range volatilityChanges {
			volKey := fmt.Sprintf("%.0f%%", volChange*100)

			// Create modified option
			modOpt := opt
			modOpt.UnderlyingPrice = opt.UnderlyingPrice * (1 + spotChange)
			modOpt.Volatility = opt.Volatility * (1 + volChange)

			// Calculate new price
			d1, d2 := g.calculateD1D2(modOpt)
			nd1 := g.normal.CDF(d1)
			nd2 := g.normal.CDF(d2)
			newPrice := g.calculateOptionPrice(modOpt, d1, d2, nd1, nd2)

			analysis.HeatMap[spotKey][volKey] = newPrice - basePrice
		}
	}

	// Generate detailed scenarios
	for _, spotChange := range []float64{-0.10, -0.05, 0, 0.05, 0.10} {
		for _, volChange := range []float64{-0.05, 0, 0.05} {
			for _, days := range []float64{0, 1, 7} {
				modOpt := opt
				modOpt.UnderlyingPrice = opt.UnderlyingPrice * (1 + spotChange)
				modOpt.Volatility = opt.Volatility * (1 + volChange)
				modOpt.TimeToExpiry = math.Max(0, opt.TimeToExpiry-days/365)

				newGreeks, _ := g.CalculateGreeks(modOpt)
				d1, d2 := g.calculateD1D2(modOpt)
				nd1 := g.normal.CDF(d1)
				nd2 := g.normal.CDF(d2)
				newPrice := g.calculateOptionPrice(modOpt, d1, d2, nd1, nd2)

				scenario := ScenarioResult{
					UnderlyingChange: spotChange,
					VolatilityChange: volChange,
					TimeDecay:        days,
					NewPrice:         newPrice,
					PnL:              newPrice - basePrice,
					NewGreeks:        newGreeks,
				}

				analysis.Scenarios = append(analysis.Scenarios, scenario)
			}
		}
	}

	return analysis, nil
}
