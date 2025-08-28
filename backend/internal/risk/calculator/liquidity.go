// backend/internal/risk/calculator/liquidity.go
package calculator

import (
	"math"
	"time"

	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

// LiquidityCalculator handles liquidity risk calculations
type LiquidityCalculator struct {
	marketData MarketDataProvider
}

// MarketDataProvider interface for fetching market data
type MarketDataProvider interface {
	GetAverageDailyVolume(symbol string) float64
	GetBidAskSpread(symbol string) float64
	GetMarketDepth(symbol string) *MarketDepth
	GetMarketCap(symbol string) float64
}

// MarketDepth represents order book depth
type MarketDepth struct {
	BidLevels []PriceLevel `json:"bid_levels"`
	AskLevels []PriceLevel `json:"ask_levels"`
	Timestamp time.Time    `json:"timestamp"`
}

// PriceLevel represents a price level in the order book
type PriceLevel struct {
	Price    float64 `json:"price"`
	Quantity float64 `json:"quantity"`
	Orders   int     `json:"orders"`
}

// NewLiquidityCalculator creates a new liquidity calculator
func NewLiquidityCalculator(marketData MarketDataProvider) *LiquidityCalculator {
	return &LiquidityCalculator{
		marketData: marketData,
	}
}

// CalculateLiquidity performs comprehensive liquidity analysis
func (l *LiquidityCalculator) CalculateLiquidity(positions []models.Position, portfolioValue float64) (*LiquidityResult, error) {
	result := &LiquidityResult{
		Timestamp:      time.Now(),
		PortfolioValue: portfolioValue,
		Positions:      make([]PositionLiquidity, 0, len(positions)),
	}

	totalLiquidValue := 0.0
	totalIlliquidValue := 0.0
	weightedLiquidityScore := 0.0

	// Analyze each position
	for _, position := range positions {
		posLiquidity := l.analyzePositionLiquidity(position)
		result.Positions = append(result.Positions, posLiquidity)

		// Aggregate liquidity values
		switch posLiquidity.LiquidityClass {
		case "HIGHLY_LIQUID":
			totalLiquidValue += posLiquidity.MarketValue
		case "LIQUID":
			totalLiquidValue += posLiquidity.MarketValue * 0.75 // 75% counted as liquid
		case "SEMI_LIQUID":
			totalLiquidValue += posLiquidity.MarketValue * 0.25 // 25% counted as liquid
			totalIlliquidValue += posLiquidity.MarketValue * 0.75
		case "ILLIQUID":
			totalIlliquidValue += posLiquidity.MarketValue
		}

		// Calculate weighted liquidity score
		weightedLiquidityScore += posLiquidity.LiquidityScore * (posLiquidity.MarketValue / portfolioValue)
	}

	// Calculate portfolio-level metrics
	result.LiquidityRatio = totalLiquidValue / portfolioValue
	result.IlliquidityRatio = totalIlliquidValue / portfolioValue
	result.WeightedLiquidityScore = weightedLiquidityScore

	// Calculate time to liquidate under different market conditions
	result.NormalMarketDays = l.calculateLiquidationTime(positions, "NORMAL")
	result.StressedMarketDays = l.calculateLiquidationTime(positions, "STRESSED")
	result.CrisisMarketDays = l.calculateLiquidationTime(positions, "CRISIS")

	// Calculate liquidity-adjusted VaR
	result.LiquidityAdjustedVaR = l.calculateLiquidityAdjustedVaR(result)

	// Determine overall liquidity health
	result.LiquidityHealth = l.assessLiquidityHealth(result)

	// Generate alerts if needed
	result.Alerts = l.checkLiquidityAlerts(result)

	return result, nil
}

// analyzePositionLiquidity analyzes liquidity for a single position
func (l *LiquidityCalculator) analyzePositionLiquidity(position models.Position) PositionLiquidity {
	pl := PositionLiquidity{
		Symbol:      position.Symbol,
		AssetType:   position.AssetType,
		Quantity:    position.Quantity.InexactFloat64(),
		MarketValue: position.MarketValue.InexactFloat64(),
	}

	// Get market data
	avgDailyVolume := l.marketData.GetAverageDailyVolume(position.Symbol)
	bidAskSpread := l.marketData.GetBidAskSpread(position.Symbol)
	marketCap := l.marketData.GetMarketCap(position.Symbol)
	marketDepth := l.marketData.GetMarketDepth(position.Symbol)

	// Calculate liquidity metrics
	pl.DaysToLiquidate = l.calculateDaysToLiquidate(position.Quantity.InexactFloat64(), avgDailyVolume, 0.1) // 10% of daily volume
	pl.MarketImpact = l.calculateMarketImpact(position.Quantity.InexactFloat64(), avgDailyVolume, bidAskSpread)
	pl.BidAskSpread = bidAskSpread
	pl.SpreadCost = position.MarketValue.InexactFloat64() * bidAskSpread

	// Calculate liquidity score (0-100)
	pl.LiquidityScore = l.calculateLiquidityScore(
		avgDailyVolume,
		bidAskSpread,
		marketCap,
		position.MarketValue.InexactFloat64(),
		marketDepth,
	)

	// Classify liquidity based on multiple factors
	pl.LiquidityClass = l.classifyLiquidity(pl.LiquidityScore, pl.DaysToLiquidate, position.AssetType)

	// Calculate liquidation value under different scenarios
	pl.ImmediateLiquidationValue = l.calculateImmediateLiquidationValue(position, marketDepth)
	pl.OrdedlyLiquidationValue = position.MarketValue.InexactFloat64() * (1 - pl.MarketImpact)

	return pl
}

// calculateLiquidityScore computes a comprehensive liquidity score
func (l *LiquidityCalculator) calculateLiquidityScore(avgVolume, spread, marketCap, positionValue float64, depth *MarketDepth) float64 {
	score := 100.0

	// Volume score (0-25 points)
	volumeRatio := positionValue / avgVolume
	if volumeRatio < 0.01 {
		// Position is less than 1% of daily volume - very liquid
		score -= 0
	} else if volumeRatio < 0.1 {
		score -= 5
	} else if volumeRatio < 0.5 {
		score -= 15
	} else {
		score -= 25
	}

	// Spread score (0-25 points)
	if spread < 0.001 { // Less than 0.1%
		score -= 0
	} else if spread < 0.005 { // Less than 0.5%
		score -= 10
	} else if spread < 0.01 { // Less than 1%
		score -= 20
	} else {
		score -= 25
	}

	// Market cap score (0-25 points)
	if marketCap > 10e9 { // Large cap (>$10B)
		score -= 0
	} else if marketCap > 2e9 { // Mid cap ($2B-$10B)
		score -= 10
	} else if marketCap > 200e6 { // Small cap ($200M-$2B)
		score -= 20
	} else {
		score -= 25
	}

	// Market depth score (0-25 points)
	if depth != nil {
		depthScore := l.assessMarketDepth(depth, positionValue)
		score -= (25 - depthScore*25)
	} else {
		score -= 25 // No depth data available
	}

	return math.Max(0, score)
}

// assessMarketDepth evaluates order book depth
func (l *LiquidityCalculator) assessMarketDepth(depth *MarketDepth, positionValue float64) float64 {
	totalBidValue := 0.0
	totalAskValue := 0.0

	// Calculate total bid and ask values in top 5 levels
	levels := 5
	if len(depth.BidLevels) < levels {
		levels = len(depth.BidLevels)
	}

	for i := 0; i < levels && i < len(depth.BidLevels); i++ {
		totalBidValue += depth.BidLevels[i].Price * depth.BidLevels[i].Quantity
	}

	if len(depth.AskLevels) < levels {
		levels = len(depth.AskLevels)
	}

	for i := 0; i < levels && i < len(depth.AskLevels); i++ {
		totalAskValue += depth.AskLevels[i].Price * depth.AskLevels[i].Quantity
	}

	avgDepth := (totalBidValue + totalAskValue) / 2

	// Score based on how much of position can be absorbed by order book
	if avgDepth > positionValue*2 {
		return 1.0 // Excellent depth
	} else if avgDepth > positionValue {
		return 0.75 // Good depth
	} else if avgDepth > positionValue*0.5 {
		return 0.5 // Moderate depth
	} else if avgDepth > positionValue*0.25 {
		return 0.25 // Poor depth
	}

	return 0.0 // Very poor depth
}

// classifyLiquidity determines liquidity classification
func (l *LiquidityCalculator) classifyLiquidity(score, daysToLiquidate float64, assetType string) string {
	// Asset type specific adjustments
	switch assetType {
	case "GOVERNMENT_BOND", "CASH", "MONEY_MARKET":
		return "HIGHLY_LIQUID"
	case "CORPORATE_BOND":
		if score > 70 {
			return "LIQUID"
		}
		return "SEMI_LIQUID"
	case "CRYPTO":
		if score > 80 {
			return "LIQUID"
		} else if score > 50 {
			return "SEMI_LIQUID"
		}
		return "ILLIQUID"
	}

	// General classification based on score and liquidation time
	if score >= 85 && daysToLiquidate <= 1 {
		return "HIGHLY_LIQUID"
	} else if score >= 70 && daysToLiquidate <= 3 {
		return "LIQUID"
	} else if score >= 50 && daysToLiquidate <= 7 {
		return "SEMI_LIQUID"
	}

	return "ILLIQUID"
}

// calculateDaysToLiquidate estimates days needed to liquidate position
func (l *LiquidityCalculator) calculateDaysToLiquidate(quantity, avgDailyVolume, participationRate float64) float64 {
	if avgDailyVolume == 0 {
		return 999 // Effectively illiquid
	}

	dailyCapacity := avgDailyVolume * participationRate
	return quantity / dailyCapacity
}

// calculateMarketImpact estimates price impact of liquidating position
func (l *LiquidityCalculator) calculateMarketImpact(quantity, avgDailyVolume, spread float64) float64 {
	if avgDailyVolume == 0 {
		return 0.5 // 50% impact for illiquid assets
	}

	// Simple square-root market impact model
	volumeRatio := quantity / avgDailyVolume
	impactFromSize := 0.1 * math.Sqrt(volumeRatio) // 10% * sqrt(volume ratio)

	// Add spread cost
	totalImpact := impactFromSize + spread/2

	return math.Min(totalImpact, 0.5) // Cap at 50%
}

// calculateImmediateLiquidationValue calculates value if position sold immediately
func (l *LiquidityCalculator) calculateImmediateLiquidationValue(position models.Position, depth *MarketDepth) float64 {
	if depth == nil || len(depth.BidLevels) == 0 {
		// No depth data, estimate with haircut
		return position.MarketValue.InexactFloat64() * 0.95
	}

	remainingQty := position.Quantity.InexactFloat64()
	liquidationValue := 0.0

	// Walk through bid levels
	for _, level := range depth.BidLevels {
		if remainingQty <= 0 {
			break
		}

		fillQty := math.Min(remainingQty, level.Quantity)
		liquidationValue += fillQty * level.Price
		remainingQty -= fillQty
	}

	// If we couldn't fill entire order from book, apply heavy discount
	if remainingQty > 0 {
		lastPrice := depth.BidLevels[len(depth.BidLevels)-1].Price
		liquidationValue += remainingQty * lastPrice * 0.9 // 10% additional discount
	}

	return liquidationValue
}

// calculateLiquidationTime estimates time to liquidate portfolio
func (l *LiquidityCalculator) calculateLiquidationTime(positions []models.Position, marketCondition string) float64 {
	maxDays := 0.0

	participationRate := 0.1 // Normal: 10% of volume
	switch marketCondition {
	case "STRESSED":
		participationRate = 0.05 // Stressed: 5% of volume
	case "CRISIS":
		participationRate = 0.02 // Crisis: 2% of volume
	}

	for _, position := range positions {
		avgVolume := l.marketData.GetAverageDailyVolume(position.Symbol)
		days := l.calculateDaysToLiquidate(position.Quantity.InexactFloat64(), avgVolume, participationRate)
		if days > maxDays {
			maxDays = days
		}
	}

	return maxDays
}

// calculateLiquidityAdjustedVaR adjusts VaR for liquidity risk
func (l *LiquidityCalculator) calculateLiquidityAdjustedVaR(result *LiquidityResult) float64 {
	// Simple adjustment: multiply by liquidity factor
	liquidityFactor := 1.0

	if result.LiquidityRatio < 0.3 {
		liquidityFactor = 1.5 // 50% increase for low liquidity
	} else if result.LiquidityRatio < 0.5 {
		liquidityFactor = 1.3 // 30% increase
	} else if result.LiquidityRatio < 0.7 {
		liquidityFactor = 1.15 // 15% increase
	}

	// This would typically use the VaR from VaR calculator
	// For now, returning a placeholder
	return result.PortfolioValue * 0.05 * liquidityFactor
}

// assessLiquidityHealth determines overall liquidity health status
func (l *LiquidityCalculator) assessLiquidityHealth(result *LiquidityResult) string {
	if result.LiquidityRatio >= 0.7 && result.NormalMarketDays <= 3 {
		return "HEALTHY"
	} else if result.LiquidityRatio >= 0.5 && result.NormalMarketDays <= 7 {
		return "ADEQUATE"
	} else if result.LiquidityRatio >= 0.3 && result.NormalMarketDays <= 14 {
		return "CONCERNING"
	}

	return "CRITICAL"
}

// checkLiquidityAlerts generates alerts for liquidity issues
func (l *LiquidityCalculator) checkLiquidityAlerts(result *LiquidityResult) []LiquidityAlert {
	alerts := []LiquidityAlert{}

	// Check liquidity ratio
	if result.LiquidityRatio < 0.3 {
		alerts = append(alerts, LiquidityAlert{
			Type:      "LOW_LIQUIDITY_RATIO",
			Severity:  "CRITICAL",
			Message:   "Portfolio liquidity ratio below 30%",
			Value:     result.LiquidityRatio,
			Threshold: 0.3,
		})
	} else if result.LiquidityRatio < 0.5 {
		alerts = append(alerts, LiquidityAlert{
			Type:      "LOW_LIQUIDITY_RATIO",
			Severity:  "WARNING",
			Message:   "Portfolio liquidity ratio below 50%",
			Value:     result.LiquidityRatio,
			Threshold: 0.5,
		})
	}

	// Check liquidation time
	if result.NormalMarketDays > 10 {
		alerts = append(alerts, LiquidityAlert{
			Type:      "EXTENDED_LIQUIDATION_TIME",
			Severity:  "WARNING",
			Message:   "Portfolio liquidation would take more than 10 days",
			Value:     result.NormalMarketDays,
			Threshold: 10,
		})
	}

	// Check for concentrated illiquid positions
	for _, pos := range result.Positions {
		if pos.LiquidityClass == "ILLIQUID" && pos.MarketValue/result.PortfolioValue > 0.1 {
			alerts = append(alerts, LiquidityAlert{
				Type:      "CONCENTRATED_ILLIQUID_POSITION",
				Severity:  "WARNING",
				Message:   "Large illiquid position: " + pos.Symbol,
				Value:     pos.MarketValue / result.PortfolioValue,
				Threshold: 0.1,
			})
		}
	}

	return alerts
}

// Result structures

// LiquidityResult contains comprehensive liquidity analysis
type LiquidityResult struct {
	Timestamp              time.Time           `json:"timestamp"`
	PortfolioValue         float64             `json:"portfolio_value"`
	LiquidityRatio         float64             `json:"liquidity_ratio"`
	IlliquidityRatio       float64             `json:"illiquidity_ratio"`
	WeightedLiquidityScore float64             `json:"weighted_liquidity_score"`
	NormalMarketDays       float64             `json:"normal_market_days"`
	StressedMarketDays     float64             `json:"stressed_market_days"`
	CrisisMarketDays       float64             `json:"crisis_market_days"`
	LiquidityAdjustedVaR   float64             `json:"liquidity_adjusted_var"`
	LiquidityHealth        string              `json:"liquidity_health"`
	Positions              []PositionLiquidity `json:"positions"`
	Alerts                 []LiquidityAlert    `json:"alerts"`
}

// PositionLiquidity contains liquidity metrics for a single position
type PositionLiquidity struct {
	Symbol                    string  `json:"symbol"`
	AssetType                 string  `json:"asset_type"`
	Quantity                  float64 `json:"quantity"`
	MarketValue               float64 `json:"market_value"`
	LiquidityScore            float64 `json:"liquidity_score"`
	LiquidityClass            string  `json:"liquidity_class"`
	DaysToLiquidate           float64 `json:"days_to_liquidate"`
	MarketImpact              float64 `json:"market_impact"`
	BidAskSpread              float64 `json:"bid_ask_spread"`
	SpreadCost                float64 `json:"spread_cost"`
	ImmediateLiquidationValue float64 `json:"immediate_liquidation_value"`
	OrdedlyLiquidationValue   float64 `json:"orderly_liquidation_value"`
}

// LiquidityAlert represents a liquidity-related alert
type LiquidityAlert struct {
	Type      string  `json:"type"`
	Severity  string  `json:"severity"`
	Message   string  `json:"message"`
	Value     float64 `json:"value"`
	Threshold float64 `json:"threshold"`
}
