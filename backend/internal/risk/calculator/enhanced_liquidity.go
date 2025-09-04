// Enhanced Liquidity Risk Calculator with advanced modeling
package calculator

import (
	"fmt"
	"math"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

// EnhancedLiquidityCalculator with sophisticated liquidity modeling
type EnhancedLiquidityCalculator struct {
	marketDataProvider MarketDataProvider
}

// NewEnhancedLiquidityCalculator creates an enhanced liquidity calculator
func NewEnhancedLiquidityCalculator(provider MarketDataProvider) *EnhancedLiquidityCalculator {
	return &EnhancedLiquidityCalculator{
		marketDataProvider: provider,
	}
}

// LiquidityRiskMetrics comprehensive liquidity risk assessment
type LiquidityRiskMetrics struct {
	LiquidityScore    float64            `json:"liquidity_score"`
	TimeToLiquidate   map[string]float64 `json:"time_to_liquidate"`
	MarketImpactCost  float64            `json:"market_impact_cost"`
	BidAskCost        float64            `json:"bid_ask_cost"`
	LiquidityVaR      map[string]float64 `json:"liquidity_var"`
	ConcentrationRisk float64            `json:"concentration_risk"`
	LiquidityBuffer   float64            `json:"liquidity_buffer"`
	StressTestResults map[string]float64 `json:"stress_test_results"`
}

// CalculateAdvancedLiquidityRisk performs comprehensive liquidity assessment
func (lc *EnhancedLiquidityCalculator) CalculateAdvancedLiquidityRisk(positions []models.Position) (*LiquidityRiskMetrics, error) {
	metrics := &LiquidityRiskMetrics{
		TimeToLiquidate:   make(map[string]float64),
		LiquidityVaR:      make(map[string]float64),
		StressTestResults: make(map[string]float64),
	}

	// Calculate position-level metrics
	totalValue := 0.0
	weightedLiquidityScore := 0.0
	totalMarketImpact := 0.0
	totalBidAskCost := 0.0

	for _, position := range positions {
		positionValue := position.MarketValue.InexactFloat64()
		totalValue += positionValue

		// Position liquidity score
		positionLiquidity := lc.calculatePositionLiquidityScore(position)
		weightedLiquidityScore += positionLiquidity * positionValue

		// Market impact cost using square root law
		marketImpact := lc.calculateMarketImpactCost(position)
		totalMarketImpact += marketImpact

		// Bid-ask spread cost
		bidAskCost := lc.calculateBidAskCost(position)
		totalBidAskCost += bidAskCost

		// Time to liquidate under different scenarios
		metrics.TimeToLiquidate[position.Symbol+"_normal"] = lc.calculateTimeToLiquidateNormal(position)
		metrics.TimeToLiquidate[position.Symbol+"_stressed"] = lc.calculateTimeToLiquidateStressed(position)
		metrics.TimeToLiquidate[position.Symbol+"_crisis"] = lc.calculateTimeToLiquidateCrisis(position)
	}

	// Portfolio-level metrics
	if totalValue > 0 {
		metrics.LiquidityScore = weightedLiquidityScore / totalValue
	}

	metrics.MarketImpactCost = totalMarketImpact
	metrics.BidAskCost = totalBidAskCost

	// Calculate concentration risk
	metrics.ConcentrationRisk = lc.calculateConcentrationRisk(positions)

	// Calculate liquidity VaR
	metrics.LiquidityVaR = lc.calculateLiquidityVaR(positions)

	// Calculate required liquidity buffer
	metrics.LiquidityBuffer = lc.calculateLiquidityBuffer(positions)

	// Stress test results
	metrics.StressTestResults = lc.performLiquidityStressTests(positions)

	return metrics, nil
}

// calculatePositionLiquidityScore uses multiple factors for liquidity assessment
func (lc *EnhancedLiquidityCalculator) calculatePositionLiquidityScore(position models.Position) float64 {
	// Base score from asset characteristics
	baseScore := lc.getBaseScore(position.Symbol, position.AssetType)

	// Adjust for position size relative to average daily volume
	avgVolume := lc.marketDataProvider.GetAverageDailyVolume(position.Symbol)
	positionSize := position.Quantity.InexactFloat64()

	// Volume impact factor (0.5 to 1.0)
	volumeRatio := positionSize / (avgVolume * 0.1) // 10% of daily volume threshold
	volumeFactor := 1.0 / (1.0 + volumeRatio)
	if volumeFactor < 0.5 {
		volumeFactor = 0.5
	}

	// Market depth factor
	depth := lc.marketDataProvider.GetMarketDepth(position.Symbol)
	depthFactor := lc.calculateDepthFactor(depth, positionSize)

	// Bid-ask spread factor (tighter spreads = higher liquidity)
	spread := lc.marketDataProvider.GetBidAskSpread(position.Symbol)
	spreadFactor := math.Max(0.5, 1.0-spread*100) // Convert spread to penalty

	// Market cap factor for stocks
	marketCap := lc.marketDataProvider.GetMarketCap(position.Symbol)
	marketCapFactor := lc.calculateMarketCapFactor(marketCap, position.AssetType)

	// Combined score
	combinedScore := baseScore * volumeFactor * depthFactor * spreadFactor * marketCapFactor

	// Ensure score is between 0 and 1
	if combinedScore > 1.0 {
		combinedScore = 1.0
	} else if combinedScore < 0.0 {
		combinedScore = 0.0
	}

	return combinedScore
}

// calculateMarketImpactCost estimates price impact of liquidation
func (lc *EnhancedLiquidityCalculator) calculateMarketImpactCost(position models.Position) float64 {
	positionValue := position.MarketValue.InexactFloat64()
	avgVolume := lc.marketDataProvider.GetAverageDailyVolume(position.Symbol)
	currentPrice := position.CurrentPrice.InexactFloat64()

	// Daily dollar volume
	dailyDollarVolume := avgVolume * currentPrice

	// Participation rate (what fraction of daily volume we represent)
	participationRate := positionValue / dailyDollarVolume

	// Market impact using square root law: Impact ∝ √(participation_rate)
	// Typical impact for 10% participation is 0.5-1.5% depending on asset
	baseImpact := lc.getBaseMarketImpact(position.AssetType)
	impact := baseImpact * math.Sqrt(participationRate/0.1) // Scale relative to 10% participation

	// Apply market impact to position value
	return impact * positionValue
}

// calculateBidAskCost estimates cost of immediate execution
func (lc *EnhancedLiquidityCalculator) calculateBidAskCost(position models.Position) float64 {
	spread := lc.marketDataProvider.GetBidAskSpread(position.Symbol)
	positionValue := position.MarketValue.InexactFloat64()

	// Assume we pay half the spread on average
	return (spread / 2.0) * positionValue
}

// calculateTimeToLiquidateNormal estimates liquidation time under normal conditions
func (lc *EnhancedLiquidityCalculator) calculateTimeToLiquidateNormal(position models.Position) float64 {
	avgVolume := lc.marketDataProvider.GetAverageDailyVolume(position.Symbol)
	positionSize := position.Quantity.InexactFloat64()

	// Conservative assumption: can trade 20% of daily volume without major impact
	maxDailyTradingVolume := avgVolume * 0.2

	if maxDailyTradingVolume <= 0 {
		return 30.0 // Default to 30 days if no volume data
	}

	daysToLiquidate := positionSize / maxDailyTradingVolume

	// Apply asset-specific factors
	assetFactor := lc.getAssetLiquidationFactor(position.AssetType)

	return daysToLiquidate * assetFactor
}

// calculateTimeToLiquidateStressed estimates liquidation time under stress
func (lc *EnhancedLiquidityCalculator) calculateTimeToLiquidateStressed(position models.Position) float64 {
	normalTime := lc.calculateTimeToLiquidateNormal(position)

	// Stressed conditions: 50% reduction in available liquidity
	stressMultiplier := 2.0

	// Additional stress based on asset type
	switch position.AssetType {
	case "CRYPTO":
		stressMultiplier *= 1.5 // Crypto more volatile in stress
	case "COMMODITY":
		stressMultiplier *= 1.3 // Commodities less liquid in stress
	case "STOCK":
		if lc.marketDataProvider.GetMarketCap(position.Symbol) < 10000000000 { // < $10B
			stressMultiplier *= 1.4 // Small caps more affected
		}
	}

	return normalTime * stressMultiplier
}

// calculateTimeToLiquidateCrisis estimates liquidation time during crisis
func (lc *EnhancedLiquidityCalculator) calculateTimeToLiquidateCrisis(position models.Position) float64 {
	stressedTime := lc.calculateTimeToLiquidateStressed(position)

	// Crisis conditions: severe liquidity reduction
	crisisMultiplier := 2.5

	// Crisis affects different assets differently
	switch position.AssetType {
	case "CRYPTO":
		crisisMultiplier *= 2.0 // Crypto markets can freeze
	case "COMMODITY":
		crisisMultiplier *= 1.5 // Physical commodities affected
	case "STOCK":
		marketCap := lc.marketDataProvider.GetMarketCap(position.Symbol)
		if marketCap < 1000000000 { // < $1B
			crisisMultiplier *= 3.0 // Small caps severely affected
		} else if marketCap < 10000000000 { // < $10B
			crisisMultiplier *= 1.8
		}
	}

	return stressedTime * crisisMultiplier
}

// calculateConcentrationRisk assesses portfolio concentration risk
func (lc *EnhancedLiquidityCalculator) calculateConcentrationRisk(positions []models.Position) float64 {
	if len(positions) == 0 {
		return 0.0
	}

	totalValue := 0.0
	for _, pos := range positions {
		totalValue += pos.MarketValue.InexactFloat64()
	}

	// Calculate Herfindahl-Hirschman Index (HHI) for concentration
	hhi := 0.0
	for _, pos := range positions {
		weight := pos.MarketValue.InexactFloat64() / totalValue
		hhi += weight * weight
	}

	// Normalize HHI to 0-1 scale (1 = maximum concentration)
	// For n equal positions, HHI = 1/n, so we scale accordingly
	maxHHI := 1.0                           // Single position case
	minHHI := 1.0 / float64(len(positions)) // Equal weights case

	if maxHHI == minHHI {
		return 0.0
	}

	concentrationRisk := (hhi - minHHI) / (maxHHI - minHHI)
	return concentrationRisk
}

// calculateLiquidityVaR calculates liquidity-adjusted VaR
func (lc *EnhancedLiquidityCalculator) calculateLiquidityVaR(positions []models.Position) map[string]float64 {
	liquidityVaR := make(map[string]float64)

	// Calculate liquidity cost at different confidence levels
	confidenceLevels := []float64{0.95, 0.99}

	for _, confidence := range confidenceLevels {
		totalLiquidityCost := 0.0

		for _, position := range positions {
			// Base market impact cost
			marketImpact := lc.calculateMarketImpactCost(position)

			// Stress multiplier based on confidence level
			stressMultiplier := 1.0
			if confidence >= 0.99 {
				stressMultiplier = 2.5 // 99% confidence assumes crisis conditions
			} else if confidence >= 0.95 {
				stressMultiplier = 1.5 // 95% confidence assumes stressed conditions
			}

			totalLiquidityCost += marketImpact * stressMultiplier
		}

		key := fmt.Sprintf("LVaR_%.0f", confidence*100)
		liquidityVaR[key] = totalLiquidityCost
	}

	return liquidityVaR
}

// calculateLiquidityBuffer calculates required liquidity buffer
func (lc *EnhancedLiquidityCalculator) calculateLiquidityBuffer(positions []models.Position) float64 {
	totalValue := 0.0
	highLiquidityValue := 0.0

	for _, position := range positions {
		positionValue := position.MarketValue.InexactFloat64()
		totalValue += positionValue

		liquidityScore := lc.calculatePositionLiquidityScore(position)
		if liquidityScore > 0.8 { // Consider high liquidity
			highLiquidityValue += positionValue
		}
	}

	if totalValue == 0 {
		return 0.0
	}

	// Buffer should be higher for less liquid portfolios
	liquidityRatio := highLiquidityValue / totalValue

	// Recommended buffer: 5-20% based on portfolio liquidity
	bufferRate := 0.05 + (1.0-liquidityRatio)*0.15

	return bufferRate * totalValue
}

// performLiquidityStressTests runs various stress scenarios
func (lc *EnhancedLiquidityCalculator) performLiquidityStressTests(positions []models.Position) map[string]float64 {
	results := make(map[string]float64)

	// Scenario 1: Market volume drops by 50%
	volumeStressImpact := 0.0
	for _, position := range positions {
		normalImpact := lc.calculateMarketImpactCost(position)
		volumeStressImpact += normalImpact * 2.0 // Double impact with half volume
	}
	results["volume_stress_50pct"] = volumeStressImpact

	// Scenario 2: Bid-ask spreads widen by 300%
	spreadStressImpact := 0.0
	for _, position := range positions {
		normalBidAskCost := lc.calculateBidAskCost(position)
		spreadStressImpact += normalBidAskCost * 4.0 // 4x wider spreads
	}
	results["spread_stress_300pct"] = spreadStressImpact

	// Scenario 3: Large position liquidation (fire sale)
	fireSaleImpact := 0.0
	for _, position := range positions {
		normalImpact := lc.calculateMarketImpactCost(position)
		// Fire sale impact is non-linear
		fireSaleImpact += normalImpact * 3.0
	}
	results["fire_sale_scenario"] = fireSaleImpact

	// Scenario 4: Combined stress (market crisis)
	combinedStress := (volumeStressImpact + spreadStressImpact + fireSaleImpact) * 0.6
	results["combined_crisis"] = combinedStress

	return results
}

// Helper functions
func (lc *EnhancedLiquidityCalculator) getBaseScore(symbol, assetType string) float64 {
	// Enhanced base scores with more granular classification
	highLiquidityAssets := map[string]float64{
		"AAPL": 0.95, "GOOGL": 0.90, "MSFT": 0.95, "AMZN": 0.90,
		"SPY": 0.98, "QQQ": 0.95, "IWM": 0.85,
		"BTC": 0.80, "ETH": 0.75,
	}

	if score, exists := highLiquidityAssets[symbol]; exists {
		return score
	}

	// Default scores by asset type
	switch assetType {
	case "ETF":
		return 0.80
	case "STOCK":
		return 0.70
	case "CRYPTO":
		return 0.50
	case "COMMODITY":
		return 0.40
	case "BOND":
		return 0.60
	default:
		return 0.50
	}
}

func (lc *EnhancedLiquidityCalculator) calculateDepthFactor(depth *MarketDepth, positionSize float64) float64 {
	if depth == nil {
		return 0.8 // Default factor when no depth data
	}

	// Calculate total depth on both sides
	totalBidQuantity := 0.0
	totalAskQuantity := 0.0

	for _, level := range depth.BidLevels {
		totalBidQuantity += level.Quantity
	}
	for _, level := range depth.AskLevels {
		totalAskQuantity += level.Quantity
	}

	avgDepth := (totalBidQuantity + totalAskQuantity) / 2.0

	// Factor based on position size relative to market depth
	if avgDepth <= 0 {
		return 0.5
	}

	depthRatio := positionSize / avgDepth
	factor := 1.0 / (1.0 + depthRatio*2.0) // Penalty for large positions relative to depth

	if factor < 0.3 {
		factor = 0.3
	}

	return factor
}

func (lc *EnhancedLiquidityCalculator) calculateMarketCapFactor(marketCap float64, assetType string) float64 {
	if assetType != "STOCK" {
		return 1.0 // No market cap adjustment for non-stocks
	}

	// Market cap tiers (in USD)
	if marketCap >= 200000000000 { // >= $200B (mega cap)
		return 1.0
	} else if marketCap >= 10000000000 { // >= $10B (large cap)
		return 0.9
	} else if marketCap >= 2000000000 { // >= $2B (mid cap)
		return 0.7
	} else if marketCap >= 300000000 { // >= $300M (small cap)
		return 0.5
	} else { // micro cap
		return 0.3
	}
}

func (lc *EnhancedLiquidityCalculator) getBaseMarketImpact(assetType string) float64 {
	// Base market impact for 10% participation rate
	switch assetType {
	case "STOCK":
		return 0.008 // 0.8% for stocks
	case "ETF":
		return 0.005 // 0.5% for ETFs
	case "CRYPTO":
		return 0.015 // 1.5% for crypto
	case "COMMODITY":
		return 0.012 // 1.2% for commodities
	case "BOND":
		return 0.003 // 0.3% for bonds
	default:
		return 0.010 // 1.0% default
	}
}

func (lc *EnhancedLiquidityCalculator) getAssetLiquidationFactor(assetType string) float64 {
	// Multiplier for liquidation time based on asset characteristics
	switch assetType {
	case "ETF":
		return 0.8 // ETFs typically more liquid
	case "STOCK":
		return 1.0 // Base case
	case "CRYPTO":
		return 1.2 // Crypto can be more volatile
	case "COMMODITY":
		return 1.5 // Commodities typically less liquid
	case "BOND":
		return 1.3 // Bonds can have liquidity issues
	default:
		return 1.0
	}
}
