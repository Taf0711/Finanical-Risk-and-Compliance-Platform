package services

import (
	"math"
	"sort"
	"time"

	"github.com/shopspring/decimal"
)

// AdvancedRiskService provides sophisticated risk analytics
type AdvancedRiskService struct {
	tradingService *TradingService
}

// RiskMetrics contains comprehensive risk assessment
type RiskMetrics struct {
	PortfolioID       string                        `json:"portfolio_id"`
	Timestamp         time.Time                     `json:"timestamp"`
	VaR               VaRMetrics                    `json:"var"`
	StressTest        StressTestResult              `json:"stress_test"`
	ConcentrationRisk ConcentrationMetrics          `json:"concentration_risk"`
	LiquidityRisk     LiquidityMetrics              `json:"liquidity_risk"`
	VolatilityMetrics VolatilityAnalysis            `json:"volatility_metrics"`
	CorrelationMatrix map[string]map[string]float64 `json:"correlation_matrix"`
	RiskScore         float64                       `json:"risk_score"`
	Recommendations   []RiskRecommendation          `json:"recommendations"`
}

type VaRMetrics struct {
	Value1Day       decimal.Decimal `json:"value_1day"`
	Value5Day       decimal.Decimal `json:"value_5day"`
	Value10Day      decimal.Decimal `json:"value_10day"`
	ConfidenceLevel float64         `json:"confidence_level"`
	Method          string          `json:"method"`
	BacktestResults BacktestResult  `json:"backtest_results"`
}

type StressTestResult struct {
	MarketCrash2008   decimal.Decimal  `json:"market_crash_2008"`
	COVID19Impact     decimal.Decimal  `json:"covid19_impact"`
	InterestRateShock decimal.Decimal  `json:"interest_rate_shock"`
	InflationScenario decimal.Decimal  `json:"inflation_scenario"`
	CustomScenarios   []CustomScenario `json:"custom_scenarios"`
}

type CustomScenario struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Impact      decimal.Decimal `json:"impact"`
	Probability float64         `json:"probability"`
}

type ConcentrationMetrics struct {
	HerfindahlIndex     float64             `json:"herfindahl_index"`
	TopPositions        []ConcentrationItem `json:"top_positions"`
	SectorConcentration map[string]float64  `json:"sector_concentration"`
	GeographicRisk      map[string]float64  `json:"geographic_risk"`
	MaxSinglePosition   float64             `json:"max_single_position"`
}

type ConcentrationItem struct {
	Symbol    string  `json:"symbol"`
	Weight    float64 `json:"weight"`
	RiskScore float64 `json:"risk_score"`
}

type LiquidityMetrics struct {
	OverallScore        float64                   `json:"overall_score"`
	DaysToLiquidate     int                       `json:"days_to_liquidate"`
	LiquidityByPosition []PositionLiquidityMetric `json:"liquidity_by_position"`
	MarketImpactCost    decimal.Decimal           `json:"market_impact_cost"`
}

type PositionLiquidityMetric struct {
	Symbol             string          `json:"symbol"`
	LiquidityScore     float64         `json:"liquidity_score"`
	AverageDailyVolume decimal.Decimal `json:"average_daily_volume"`
	BidAskSpread       decimal.Decimal `json:"bid_ask_spread"`
	DaysToLiquidate    int             `json:"days_to_liquidate"`
}

type VolatilityAnalysis struct {
	PortfolioVolatility  float64            `json:"portfolio_volatility"`
	ImpliedVolatility    float64            `json:"implied_volatility"`
	HistoricalVolatility float64            `json:"historical_volatility"`
	VolatilityByAsset    map[string]float64 `json:"volatility_by_asset"`
	VolatilityTrend      string             `json:"volatility_trend"`
}

type BacktestResult struct {
	Accuracy           float64   `json:"accuracy"`
	Exceptions         int       `json:"exceptions"`
	ExpectedExceptions int       `json:"expected_exceptions"`
	TestPeriod         string    `json:"test_period"`
	LastUpdated        time.Time `json:"last_updated"`
}

type RiskRecommendation struct {
	Type        string `json:"type"`
	Priority    string `json:"priority"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Action      string `json:"action"`
	Impact      string `json:"impact"`
}

func NewAdvancedRiskService(tradingService *TradingService) *AdvancedRiskService {
	return &AdvancedRiskService{
		tradingService: tradingService,
	}
}

// CalculateAdvancedRisk performs comprehensive risk analysis
func (s *AdvancedRiskService) CalculateAdvancedRisk(portfolioID string) (*RiskMetrics, error) {
	// Get current positions
	positions, err := s.tradingService.GetPositions()
	if err != nil {
		return nil, err
	}

	if len(positions) == 0 {
		return &RiskMetrics{
			PortfolioID: portfolioID,
			Timestamp:   time.Now(),
			RiskScore:   0,
		}, nil
	}

	// Calculate various risk metrics
	varMetrics := s.calculateAdvancedVaR(positions)
	stressTest := s.performStressTest(positions)
	concentrationRisk := s.calculateConcentrationRisk(positions)
	liquidityRisk := s.calculateLiquidityRisk(positions)
	volatilityMetrics := s.calculateVolatilityMetrics(positions)
	correlationMatrix := s.calculateCorrelationMatrix(positions)

	// Calculate overall risk score
	riskScore := s.calculateOverallRiskScore(varMetrics, concentrationRisk, liquidityRisk, volatilityMetrics)

	// Generate recommendations
	recommendations := s.generateRiskRecommendations(varMetrics, concentrationRisk, liquidityRisk, volatilityMetrics)

	return &RiskMetrics{
		PortfolioID:       portfolioID,
		Timestamp:         time.Now(),
		VaR:               varMetrics,
		StressTest:        stressTest,
		ConcentrationRisk: concentrationRisk,
		LiquidityRisk:     liquidityRisk,
		VolatilityMetrics: volatilityMetrics,
		CorrelationMatrix: correlationMatrix,
		RiskScore:         riskScore,
		Recommendations:   recommendations,
	}, nil
}

func (s *AdvancedRiskService) calculateAdvancedVaR(positions []Position) VaRMetrics {
	// Calculate portfolio value
	totalValue := decimal.Zero
	for _, pos := range positions {
		value, _ := decimal.NewFromString(pos.MarketValue)
		totalValue = totalValue.Add(value)
	}

	// Simulate VaR calculations (in production, use historical data)
	volatility := 0.15 // 15% annual volatility
	confidenceLevel := 0.95

	// Convert to daily volatility
	dailyVol := volatility / math.Sqrt(252)

	// Z-score for 95% confidence
	zScore := 1.645

	// Calculate VaR for different time horizons
	var1Day := totalValue.Mul(decimal.NewFromFloat(dailyVol * zScore))
	var5Day := var1Day.Mul(decimal.NewFromFloat(math.Sqrt(5)))
	var10Day := var1Day.Mul(decimal.NewFromFloat(math.Sqrt(10)))

	return VaRMetrics{
		Value1Day:       var1Day,
		Value5Day:       var5Day,
		Value10Day:      var10Day,
		ConfidenceLevel: confidenceLevel,
		Method:          "Parametric",
		BacktestResults: BacktestResult{
			Accuracy:           0.94,
			Exceptions:         3,
			ExpectedExceptions: 5,
			TestPeriod:         "90 days",
			LastUpdated:        time.Now().AddDate(0, 0, -1),
		},
	}
}

func (s *AdvancedRiskService) performStressTest(positions []Position) StressTestResult {
	// Calculate total portfolio value
	totalValue := decimal.Zero
	for _, pos := range positions {
		value, _ := decimal.NewFromString(pos.MarketValue)
		totalValue = totalValue.Add(value)
	}

	// Historical stress scenarios
	crash2008Impact := totalValue.Mul(decimal.NewFromFloat(-0.37)) // -37% during 2008 crisis
	covid19Impact := totalValue.Mul(decimal.NewFromFloat(-0.34))   // -34% during COVID-19
	rateShockImpact := totalValue.Mul(decimal.NewFromFloat(-0.15)) // -15% interest rate shock
	inflationImpact := totalValue.Mul(decimal.NewFromFloat(-0.12)) // -12% inflation scenario

	customScenarios := []CustomScenario{
		{
			Name:        "Tech Bubble Burst",
			Description: "Technology sector correction similar to 2000",
			Impact:      totalValue.Mul(decimal.NewFromFloat(-0.45)),
			Probability: 0.15,
		},
		{
			Name:        "Geopolitical Crisis",
			Description: "Major geopolitical event causing market disruption",
			Impact:      totalValue.Mul(decimal.NewFromFloat(-0.25)),
			Probability: 0.20,
		},
	}

	return StressTestResult{
		MarketCrash2008:   crash2008Impact,
		COVID19Impact:     covid19Impact,
		InterestRateShock: rateShockImpact,
		InflationScenario: inflationImpact,
		CustomScenarios:   customScenarios,
	}
}

func (s *AdvancedRiskService) calculateConcentrationRisk(positions []Position) ConcentrationMetrics {
	if len(positions) == 0 {
		return ConcentrationMetrics{
			HerfindahlIndex:   0,
			TopPositions:      []ConcentrationItem{},
			MaxSinglePosition: 0,
		}
	}

	// Calculate total portfolio value
	totalValue := decimal.Zero
	for _, pos := range positions {
		value, _ := decimal.NewFromString(pos.MarketValue)
		totalValue = totalValue.Add(value)
	}

	// Calculate position weights
	var weights []float64
	var topPositions []ConcentrationItem
	maxWeight := 0.0

	for _, pos := range positions {
		value, _ := decimal.NewFromString(pos.MarketValue)
		weight, _ := value.Div(totalValue).Float64()
		weights = append(weights, weight)

		if weight > maxWeight {
			maxWeight = weight
		}

		// Risk score based on position size and volatility (simplified)
		riskScore := weight * 100 // Simple risk score
		if weight > 0.1 {         // More than 10% concentration
			riskScore *= 1.5
		}

		topPositions = append(topPositions, ConcentrationItem{
			Symbol:    pos.Symbol,
			Weight:    weight,
			RiskScore: riskScore,
		})
	}

	// Sort top positions by weight
	sort.Slice(topPositions, func(i, j int) bool {
		return topPositions[i].Weight > topPositions[j].Weight
	})

	// Take top 10
	if len(topPositions) > 10 {
		topPositions = topPositions[:10]
	}

	// Calculate Herfindahl-Hirschman Index
	hhi := 0.0
	for _, weight := range weights {
		hhi += weight * weight
	}

	// Mock sector concentration (in production, use real sector data)
	sectorConcentration := map[string]float64{
		"Technology": 0.35,
		"Healthcare": 0.20,
		"Finance":    0.15,
		"Consumer":   0.15,
		"Energy":     0.10,
		"Other":      0.05,
	}

	geographicRisk := map[string]float64{
		"US":     0.70,
		"Europe": 0.15,
		"Asia":   0.10,
		"Other":  0.05,
	}

	return ConcentrationMetrics{
		HerfindahlIndex:     hhi,
		TopPositions:        topPositions,
		SectorConcentration: sectorConcentration,
		GeographicRisk:      geographicRisk,
		MaxSinglePosition:   maxWeight,
	}
}

func (s *AdvancedRiskService) calculateLiquidityRisk(positions []Position) LiquidityMetrics {
	if len(positions) == 0 {
		return LiquidityMetrics{
			OverallScore:     100,
			DaysToLiquidate:  0,
			MarketImpactCost: decimal.Zero,
		}
	}

	var positionMetrics []PositionLiquidityMetric
	totalValue := decimal.Zero
	weightedLiquidityScore := 0.0
	maxDaysToLiquidate := 0
	totalMarketImpact := decimal.Zero

	for _, pos := range positions {
		value, _ := decimal.NewFromString(pos.MarketValue)
		totalValue = totalValue.Add(value)

		// Mock liquidity data (in production, use real market data)
		avgVolume := decimal.NewFromFloat(1000000) // $1M average daily volume
		bidAskSpread := decimal.NewFromFloat(0.01) // 1% spread

		// Calculate days to liquidate (position size / 10% of daily volume)
		positionSize, _ := value.Float64()
		avgVolumeFloat, _ := avgVolume.Float64()
		daysToLiq := int(math.Ceil(positionSize / (avgVolumeFloat * 0.1)))

		if daysToLiq > maxDaysToLiquidate {
			maxDaysToLiquidate = daysToLiq
		}

		// Liquidity score (100 = most liquid, 0 = illiquid)
		liquidityScore := 100.0
		if daysToLiq > 1 {
			liquidityScore = math.Max(20, 100-float64(daysToLiq)*10)
		}

		// Market impact cost (simplified)
		marketImpact := value.Mul(decimal.NewFromFloat(0.005)) // 0.5% impact
		totalMarketImpact = totalMarketImpact.Add(marketImpact)

		weight, _ := value.Div(totalValue).Float64()
		weightedLiquidityScore += liquidityScore * weight

		positionMetrics = append(positionMetrics, PositionLiquidityMetric{
			Symbol:             pos.Symbol,
			LiquidityScore:     liquidityScore,
			AverageDailyVolume: avgVolume,
			BidAskSpread:       bidAskSpread,
			DaysToLiquidate:    daysToLiq,
		})
	}

	return LiquidityMetrics{
		OverallScore:        weightedLiquidityScore,
		DaysToLiquidate:     maxDaysToLiquidate,
		LiquidityByPosition: positionMetrics,
		MarketImpactCost:    totalMarketImpact,
	}
}

func (s *AdvancedRiskService) calculateVolatilityMetrics(positions []Position) VolatilityAnalysis {
	if len(positions) == 0 {
		return VolatilityAnalysis{
			PortfolioVolatility: 0,
			VolatilityTrend:     "stable",
		}
	}

	// Mock volatility calculations (in production, use historical price data)
	volatilityByAsset := make(map[string]float64)
	totalValue := decimal.Zero
	weightedVolatility := 0.0

	for _, pos := range positions {
		value, _ := decimal.NewFromString(pos.MarketValue)
		totalValue = totalValue.Add(value)

		// Mock individual asset volatility
		assetVol := 0.15 + (float64(len(pos.Symbol)) * 0.01) // Simple mock based on symbol length
		if assetVol > 0.50 {
			assetVol = 0.50
		}

		volatilityByAsset[pos.Symbol] = assetVol

		weight, _ := value.Div(totalValue).Float64()
		weightedVolatility += assetVol * weight
	}

	// Mock implied vs historical volatility
	impliedVol := weightedVolatility * 1.1 // Implied usually higher
	historicalVol := weightedVolatility * 0.95

	trend := "stable"
	if impliedVol > historicalVol*1.2 {
		trend = "increasing"
	} else if impliedVol < historicalVol*0.8 {
		trend = "decreasing"
	}

	return VolatilityAnalysis{
		PortfolioVolatility:  weightedVolatility,
		ImpliedVolatility:    impliedVol,
		HistoricalVolatility: historicalVol,
		VolatilityByAsset:    volatilityByAsset,
		VolatilityTrend:      trend,
	}
}

func (s *AdvancedRiskService) calculateCorrelationMatrix(positions []Position) map[string]map[string]float64 {
	correlationMatrix := make(map[string]map[string]float64)

	// Mock correlation data (in production, calculate from historical returns)
	for i, pos1 := range positions {
		correlationMatrix[pos1.Symbol] = make(map[string]float64)

		for j, pos2 := range positions {
			if i == j {
				correlationMatrix[pos1.Symbol][pos2.Symbol] = 1.0
			} else {
				// Mock correlation based on symbol similarity (simplified)
				correlation := 0.3 + (float64((len(pos1.Symbol)+len(pos2.Symbol))%10) * 0.05)
				if correlation > 0.9 {
					correlation = 0.9
				}
				correlationMatrix[pos1.Symbol][pos2.Symbol] = correlation
			}
		}
	}

	return correlationMatrix
}

func (s *AdvancedRiskService) calculateOverallRiskScore(
	var_ VaRMetrics,
	concentration ConcentrationMetrics,
	liquidity LiquidityMetrics,
	volatility VolatilityAnalysis,
) float64 {
	// Weighted risk score calculation
	varScore := math.Min(100, volatility.PortfolioVolatility*200)        // 0-100 scale
	concentrationScore := concentration.HerfindahlIndex * 100            // 0-100 scale
	liquidityScore := 100 - liquidity.OverallScore                       // Invert liquidity score
	volatilityScore := math.Min(100, volatility.PortfolioVolatility*200) // 0-100 scale

	// Weighted average (can be customized)
	overallScore := (varScore*0.3 + concentrationScore*0.25 + liquidityScore*0.25 + volatilityScore*0.2)

	return math.Min(100, overallScore)
}

func (s *AdvancedRiskService) generateRiskRecommendations(
	var_ VaRMetrics,
	concentration ConcentrationMetrics,
	liquidity LiquidityMetrics,
	volatility VolatilityAnalysis,
) []RiskRecommendation {
	var recommendations []RiskRecommendation

	// High concentration risk
	if concentration.MaxSinglePosition > 0.2 {
		recommendations = append(recommendations, RiskRecommendation{
			Type:        "concentration",
			Priority:    "high",
			Title:       "High Position Concentration",
			Description: "One or more positions represent over 20% of the portfolio",
			Action:      "Consider reducing position sizes or diversifying holdings",
			Impact:      "Reduces portfolio volatility and single-stock risk",
		})
	}

	// Low liquidity
	if liquidity.OverallScore < 60 {
		recommendations = append(recommendations, RiskRecommendation{
			Type:        "liquidity",
			Priority:    "medium",
			Title:       "Low Portfolio Liquidity",
			Description: "Portfolio may be difficult to liquidate quickly",
			Action:      "Consider adding more liquid positions or reducing illiquid holdings",
			Impact:      "Improves ability to exit positions during market stress",
		})
	}

	// High volatility
	if volatility.PortfolioVolatility > 0.25 {
		recommendations = append(recommendations, RiskRecommendation{
			Type:        "volatility",
			Priority:    "medium",
			Title:       "High Portfolio Volatility",
			Description: "Portfolio volatility is above recommended levels",
			Action:      "Consider adding defensive positions or reducing high-beta stocks",
			Impact:      "Reduces portfolio volatility and drawdown risk",
		})
	}

	// Correlation risk
	if concentration.HerfindahlIndex > 0.25 {
		recommendations = append(recommendations, RiskRecommendation{
			Type:        "diversification",
			Priority:    "low",
			Title:       "Improve Diversification",
			Description: "Portfolio shows high concentration in similar assets",
			Action:      "Consider adding uncorrelated assets or different sectors",
			Impact:      "Improves risk-adjusted returns through diversification",
		})
	}

	return recommendations
}
