// backend/internal/services/risk_engine.go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/marketdata"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/calculator"
)

type RiskEngineService struct {
	db            *gorm.DB
	varCalculator *calculator.VaRCalculator
	liquidityCalc *calculator.LiquidityCalculator
	alertService  *AlertService
}

// MarketDataProviderAdapter adapts the market data service to the calculator interface
type MarketDataProviderAdapter struct {
	marketDataService *marketdata.Service
}

func NewMarketDataProviderAdapter(service *marketdata.Service) *MarketDataProviderAdapter {
	return &MarketDataProviderAdapter{
		marketDataService: service,
	}
}

func (m *MarketDataProviderAdapter) GetAverageDailyVolume(symbol string) float64 {
	// Try to get recent historical data to calculate average volume
	historicalData, err := m.marketDataService.GetHistoricalData(symbol, "1mo")
	if err != nil || len(historicalData.DataPoints) == 0 {
		// Fallback to realistic estimates based on market cap and asset type
		return m.estimateVolumeFromSymbol(symbol)
	}

	// Calculate average volume from historical data
	totalVolume := int64(0)
	validDays := 0
	for _, point := range historicalData.DataPoints {
		if point.Volume > 0 {
			totalVolume += point.Volume
			validDays++
		}
	}

	if validDays > 0 {
		return float64(totalVolume) / float64(validDays)
	}

	return m.estimateVolumeFromSymbol(symbol)
}

func (m *MarketDataProviderAdapter) GetBidAskSpread(symbol string) float64 {
	// Try to get real-time quote to calculate actual spread
	_, err := m.marketDataService.GetRealtimeQuote(symbol)
	if err != nil {
		// Fallback to estimated spreads based on asset characteristics
		return m.estimateSpreadFromSymbol(symbol)
	}

	// For now, return estimated spread since quote might not have bid/ask
	// In a real implementation, you'd get this from order book data
	return m.estimateSpreadFromSymbol(symbol)
}

func (m *MarketDataProviderAdapter) GetMarketCap(symbol string) float64 {
	// Try to get company info which might include market cap
	companyInfo, err := m.marketDataService.GetCompanyInfo(symbol)
	if err != nil {
		return m.estimateMarketCapFromSymbol(symbol)
	}

	// For now, return estimated market cap since company info might not include it
	// In a real implementation, you'd get this from fundamental data
	_ = companyInfo
	return m.estimateMarketCapFromSymbol(symbol)
}

func (m *MarketDataProviderAdapter) GetMarketDepth(symbol string) *calculator.MarketDepth {
	// In a real implementation, this would fetch actual order book data
	// For now, generate realistic order book based on current price
	quote, err := m.marketDataService.GetRealtimeQuote(symbol)
	if err != nil {
		return m.generateRealisticOrderBook(symbol, 100.0) // Default price
	}

	return m.generateRealisticOrderBook(symbol, quote.Price)
}

// Helper methods for fallback estimates
func (m *MarketDataProviderAdapter) estimateVolumeFromSymbol(symbol string) float64 {
	// Use market cap tiers to estimate volume
	marketCap := m.estimateMarketCapFromSymbol(symbol)

	// Volume typically correlates with market cap and volatility
	switch {
	case marketCap >= 1000000000000: // $1T+ mega cap
		return 50000000.0
	case marketCap >= 200000000000: // $200B+ large cap
		return 25000000.0
	case marketCap >= 10000000000: // $10B+ mid cap
		return 5000000.0
	case marketCap >= 2000000000: // $2B+ small cap
		return 1000000.0
	default:
		return 500000.0 // Micro cap
	}
}

func (m *MarketDataProviderAdapter) estimateSpreadFromSymbol(symbol string) float64 {
	// Estimate spreads based on asset characteristics and market cap
	marketCap := m.estimateMarketCapFromSymbol(symbol)

	// Larger market cap typically means tighter spreads
	switch {
	case marketCap >= 1000000000000: // $1T+ mega cap
		return 0.00005 // 0.005%
	case marketCap >= 200000000000: // $200B+ large cap
		return 0.0001 // 0.01%
	case marketCap >= 10000000000: // $10B+ mid cap
		return 0.0003 // 0.03%
	case marketCap >= 2000000000: // $2B+ small cap
		return 0.001 // 0.1%
	default:
		return 0.005 // 0.5% for micro cap
	}
}

func (m *MarketDataProviderAdapter) estimateMarketCapFromSymbol(symbol string) float64 {
	// Realistic market cap estimates based on known symbols
	knownMarketCaps := map[string]float64{
		"AAPL":   3200000000000,  // $3.2T
		"GOOGL":  1800000000000,  // $1.8T
		"MSFT":   3000000000000,  // $3.0T
		"AMZN":   1500000000000,  // $1.5T
		"TSLA":   800000000000,   // $800B
		"JPM":    450000000000,   // $450B
		"BAC":    250000000000,   // $250B
		"GS":     110000000000,   // $110B
		"MS":     150000000000,   // $150B
		"WFC":    180000000000,   // $180B
		"BTC":    900000000000,   // $900B (crypto market cap)
		"ETH":    450000000000,   // $450B
		"GOLD":   13000000000000, // $13T (gold market)
		"SILVER": 1500000000000,  // $1.5T
		"OIL":    5000000000000,  // $5T (oil market)
	}

	if cap, exists := knownMarketCaps[symbol]; exists {
		return cap
	}

	// Default to mid-cap for unknown symbols
	return 5000000000 // $5B
}

func (m *MarketDataProviderAdapter) generateRealisticOrderBook(symbol string, currentPrice float64) *calculator.MarketDepth {
	// Generate realistic order book levels around current price
	spread := m.estimateSpreadFromSymbol(symbol)
	halfSpread := currentPrice * spread / 2

	bidLevels := []calculator.PriceLevel{}
	askLevels := []calculator.PriceLevel{}

	// Generate 5 levels on each side
	for i := 0; i < 5; i++ {
		levelSpread := halfSpread * (1 + float64(i)*0.5)

		// Bid levels (below current price)
		bidPrice := currentPrice - levelSpread
		bidQuantity := 5000.0 / (1 + float64(i)*0.3) // Decreasing quantity
		bidLevels = append(bidLevels, calculator.PriceLevel{
			Price:    bidPrice,
			Quantity: bidQuantity,
			Orders:   25 - i*3, // Decreasing order count
		})

		// Ask levels (above current price)
		askPrice := currentPrice + levelSpread
		askQuantity := 4500.0 / (1 + float64(i)*0.3) // Decreasing quantity
		askLevels = append(askLevels, calculator.PriceLevel{
			Price:    askPrice,
			Quantity: askQuantity,
			Orders:   22 - i*3, // Decreasing order count
		})
	}

	return &calculator.MarketDepth{
		BidLevels: bidLevels,
		AskLevels: askLevels,
		Timestamp: time.Now(),
	}
}

func NewRiskEngineService() *RiskEngineService {
	// Create market data service configuration
	config := &marketdata.ServiceConfig{
		PrimaryProvider:   "alpaca",
		FallbackProviders: []string{},
		CacheTTL:          5 * time.Minute,
		RateLimits: map[string]marketdata.RateLimitConfig{
			"alpaca": {
				RequestsPerMinute: 200,
				BurstLimit:        10,
			},
		},
	}

	// Create market data service and adapter
	marketDataService := marketdata.NewService(config, database.GetRedis())
	marketDataAdapter := NewMarketDataProviderAdapter(marketDataService)

	return &RiskEngineService{
		db:            database.GetDB(),
		varCalculator: calculator.NewVaRCalculator(1000000.0, 252), // Portfolio value and lookback days
		liquidityCalc: calculator.NewLiquidityCalculator(marketDataAdapter),
		alertService:  NewAlertService(),
	}
}

// TradeRiskAnalysis represents the risk assessment for a trade
type TradeRiskAnalysis struct {
	TradeID  uuid.UUID       `json:"trade_id"`
	Symbol   string          `json:"symbol"`
	Side     string          `json:"side"`
	Quantity decimal.Decimal `json:"quantity"`
	Price    decimal.Decimal `json:"price"`

	// Risk Metrics
	PositionRisk        decimal.Decimal `json:"position_risk"`
	PortfolioImpact     decimal.Decimal `json:"portfolio_impact"`
	ConcentrationImpact decimal.Decimal `json:"concentration_impact"`
	LiquidityImpact     decimal.Decimal `json:"liquidity_impact"`

	// Risk Checks
	Violations     []RiskViolation `json:"violations"`
	RiskScore      decimal.Decimal `json:"risk_score"`
	Approved       bool            `json:"approved"`
	RequiresReview bool            `json:"requires_review"`

	// Recommendations
	SuggestedStopLoss   decimal.Decimal `json:"suggested_stop_loss,omitempty"`
	SuggestedSize       decimal.Decimal `json:"suggested_size,omitempty"`
	HedgeRecommendation string          `json:"hedge_recommendation,omitempty"`
}

// RiskViolation represents a specific risk limit breach
type RiskViolation struct {
	Type         string          `json:"type"`
	Severity     string          `json:"severity"`
	Description  string          `json:"description"`
	CurrentValue decimal.Decimal `json:"current_value"`
	Limit        decimal.Decimal `json:"limit"`
	Impact       decimal.Decimal `json:"impact"`
}

// EvaluateTransaction performs pre-trade risk assessment
func (res *RiskEngineService) EvaluateTransaction(tx *models.Transaction) (*TradeRiskAnalysis, error) {
	// Get portfolio and positions
	var portfolio models.Portfolio
	if err := res.db.Preload("Positions").First(&portfolio, tx.PortfolioID).Error; err != nil {
		return nil, fmt.Errorf("portfolio not found: %w", err)
	}

	// Get or create risk thresholds
	thresholds, err := res.getOrCreateThresholds(tx.PortfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thresholds: %w", err)
	}

	analysis := &TradeRiskAnalysis{
		TradeID:    tx.ID,
		Symbol:     tx.Symbol,
		Side:       tx.TransactionType,
		Quantity:   tx.Quantity,
		Price:      tx.Price,
		Violations: []RiskViolation{},
	}

	// 1. Check Position Size Limit
	if violation := res.checkPositionSizeLimit(tx, &portfolio, thresholds); violation != nil {
		analysis.Violations = append(analysis.Violations, *violation)
	}

	// 2. Calculate VaR Impact
	varImpact, err := res.calculateVaRImpact(tx, &portfolio, thresholds)
	if err == nil {
		analysis.PortfolioImpact = varImpact.PortfolioImpact
		if varImpact.Violation != nil {
			analysis.Violations = append(analysis.Violations, *varImpact.Violation)
		}
	}

	// 3. Check Concentration Risk
	concentrationImpact := res.checkConcentrationRisk(tx, &portfolio, thresholds)
	analysis.ConcentrationImpact = concentrationImpact.Impact
	if concentrationImpact.Violation != nil {
		analysis.Violations = append(analysis.Violations, *concentrationImpact.Violation)
	}

	// 4. Check Liquidity Impact
	liquidityImpact := res.checkLiquidityImpact(tx, &portfolio, thresholds)
	analysis.LiquidityImpact = liquidityImpact.Impact
	if liquidityImpact.Violation != nil {
		analysis.Violations = append(analysis.Violations, *liquidityImpact.Violation)
	}

	// 5. Check Stop Loss Requirements
	if thresholds.RequireStopLoss && tx.StopLoss.IsZero() {
		analysis.Violations = append(analysis.Violations, RiskViolation{
			Type:        "STOP_LOSS_REQUIRED",
			Severity:    "WARNING",
			Description: "Stop loss is required but not set",
		})
		analysis.SuggestedStopLoss = res.calculateSuggestedStopLoss(tx)
	}

	// 6. Calculate Risk Score
	analysis.RiskScore = res.calculateRiskScore(analysis)

	// 7. Determine Approval Status
	analysis.Approved, analysis.RequiresReview = res.determineApprovalStatus(analysis)

	// 8. Generate Recommendations
	if analysis.RiskScore.GreaterThan(decimal.NewFromInt(70)) || len(analysis.Violations) > 0 {
		res.generateRecommendations(analysis, tx)
	}

	// 9. Update transaction with risk analysis
	res.updateTransactionRiskStatus(tx, analysis)

	// 10. Create alerts for critical violations
	if !analysis.Approved && len(analysis.Violations) > 0 {
		res.createRiskAlerts(tx, analysis)
	}

	return analysis, nil
}

// Helper methods

func (res *RiskEngineService) getOrCreateThresholds(portfolioID uuid.UUID) (*models.RiskThresholds, error) {
	var thresholds models.RiskThresholds
	err := res.db.Where("portfolio_id = ?", portfolioID).First(&thresholds).Error

	if err == gorm.ErrRecordNotFound {
		// Create default thresholds
		thresholds = *models.GetDefaultThresholds(portfolioID)
		if err := res.db.Create(&thresholds).Error; err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return &thresholds, nil
}

func (res *RiskEngineService) checkPositionSizeLimit(tx *models.Transaction, portfolio *models.Portfolio, thresholds *models.RiskThresholds) *RiskViolation {
	tradeValue := tx.Quantity.Mul(tx.Price)

	if portfolio.TotalValue.IsZero() {
		return nil
	}

	positionPercent := tradeValue.Div(portfolio.TotalValue)

	if positionPercent.GreaterThan(thresholds.MaxPositionSize) {
		impact := positionPercent.Sub(thresholds.MaxPositionSize).Div(thresholds.MaxPositionSize)
		return &RiskViolation{
			Type:         "POSITION_SIZE",
			Severity:     "VIOLATION",
			Description:  fmt.Sprintf("Position size %.2f%% exceeds maximum", positionPercent.Mul(decimal.NewFromInt(100)).InexactFloat64()),
			CurrentValue: positionPercent,
			Limit:        thresholds.MaxPositionSize,
			Impact:       impact,
		}
	}

	return nil
}

type VaRImpactResult struct {
	PortfolioImpact decimal.Decimal
	Violation       *RiskViolation
}

func (res *RiskEngineService) calculateVaRImpact(tx *models.Transaction, portfolio *models.Portfolio, thresholds *models.RiskThresholds) (*VaRImpactResult, error) {
	// Calculate current VaR using the calculator
	priceHistory := make(map[string][]float64) // Mock price history - would need real data
	// Use enhanced VaR calculation with filtered historical method
	returns := []float64{} // Mock returns - would need real price data
	for _, prices := range priceHistory {
		if len(prices) > 1 {
			for i := 1; i < len(prices); i++ {
				ret := (prices[i] - prices[i-1]) / prices[i-1]
				returns = append(returns, ret)
			}
		}
	}

	var currentVaR95 float64
	if len(returns) > 0 {
		currentVaR95 = res.varCalculator.CalculateFilteredHistoricalVaR(returns, 0.95)
	} else {
		currentVaR95 = portfolio.TotalValue.InexactFloat64() * 0.05 // 5% fallback
	}

	currentVaRResult := struct{ VaR95 float64 }{VaR95: currentVaR95}

	// Simulate trade impact (simplified)
	// In production, this would recalculate VaR with the new position
	estimatedImpact := decimal.NewFromFloat(0.02) // 2% estimated impact
	currentVaR := decimal.NewFromFloat(currentVaRResult.VaR95)
	newVaR := currentVaR.Mul(decimal.NewFromFloat(1).Add(estimatedImpact))

	result := &VaRImpactResult{
		PortfolioImpact: estimatedImpact,
	}

	if newVaR.GreaterThan(thresholds.MaxVaR95) {
		result.Violation = &RiskViolation{
			Type:         "VAR_LIMIT",
			Severity:     "CRITICAL",
			Description:  "Trade would increase VaR beyond limit",
			CurrentValue: newVaR,
			Limit:        thresholds.MaxVaR95,
			Impact:       newVaR.Sub(thresholds.MaxVaR95).Div(thresholds.MaxVaR95),
		}
	}

	return result, nil
}

type ConcentrationResult struct {
	Impact    decimal.Decimal
	Violation *RiskViolation
}

func (res *RiskEngineService) checkConcentrationRisk(tx *models.Transaction, portfolio *models.Portfolio, thresholds *models.RiskThresholds) *ConcentrationResult {
	// Calculate Herfindahl index
	totalValue := portfolio.TotalValue
	if totalValue.IsZero() {
		return &ConcentrationResult{Impact: decimal.Zero}
	}

	hhi := decimal.Zero
	for _, position := range portfolio.Positions {
		weight := position.MarketValue.Div(totalValue)
		hhi = hhi.Add(weight.Mul(weight))
	}

	// Add new position impact
	newPositionValue := tx.Quantity.Mul(tx.Price)
	newTotalValue := totalValue.Add(newPositionValue)
	newWeight := newPositionValue.Div(newTotalValue)
	newHHI := hhi.Add(newWeight.Mul(newWeight))

	result := &ConcentrationResult{
		Impact: newHHI.Sub(hhi),
	}

	if newHHI.GreaterThan(thresholds.MaxConcentration) {
		result.Violation = &RiskViolation{
			Type:         "CONCENTRATION_LIMIT",
			Severity:     "WARNING",
			Description:  "Portfolio concentration exceeds limit",
			CurrentValue: newHHI,
			Limit:        thresholds.MaxConcentration,
			Impact:       newHHI.Sub(thresholds.MaxConcentration).Div(thresholds.MaxConcentration),
		}
	}

	return result
}

// LiquidityResult contains liquidity analysis
type LiquidityResult struct {
	PortfolioID        uuid.UUID                  `json:"portfolio_id"`
	LiquidityRatio     decimal.Decimal            `json:"liquidity_ratio"`
	LiquidityScore     string                     `json:"liquidity_score"`
	DaysToLiquidate    decimal.Decimal            `json:"days_to_liquidate"`
	LiquidityBreakdown map[string]decimal.Decimal `json:"liquidity_breakdown"`
	RiskAssessment     string                     `json:"risk_assessment"`
	CalculatedAt       time.Time                  `json:"calculated_at"`
	Impact             decimal.Decimal            `json:"impact"`
	Violation          *RiskViolation             `json:"violation"`
}

// PositionLimitResult contains position limit analysis
type PositionLimitResult struct {
	PortfolioID     uuid.UUID           `json:"portfolio_id"`
	MaxLimit        decimal.Decimal     `json:"max_limit"`
	Violations      []PositionViolation `json:"violations"`
	ComplianceScore decimal.Decimal     `json:"compliance_score"`
	Status          string              `json:"status"`
	TotalPositions  int                 `json:"total_positions"`
	CalculatedAt    time.Time           `json:"calculated_at"`
}

type PositionViolation struct {
	Symbol         string          `json:"symbol"`
	CurrentPercent decimal.Decimal `json:"current_percent"`
	MaxPercent     decimal.Decimal `json:"max_percent"`
	ExcessPercent  decimal.Decimal `json:"excess_percent"`
	MarketValue    decimal.Decimal `json:"market_value"`
	Severity       string          `json:"severity"`
}

func (res *RiskEngineService) checkLiquidityImpact(tx *models.Transaction, portfolio *models.Portfolio, thresholds *models.RiskThresholds) *LiquidityResult {
	// Get current liquidity using the calculator
	// Use enhanced liquidity calculation
	liquidityMetrics, err := res.liquidityCalc.CalculateAdvancedLiquidityRisk(portfolio.Positions)
	if err != nil {
		// Return simplified result if calculation fails
		return &LiquidityResult{
			Impact: decimal.NewFromFloat(0.05),
		}
	}

	// Convert to expected format
	liquidityResult := struct {
		LiquidityRatio float64
	}{
		LiquidityRatio: liquidityMetrics.LiquidityScore,
	}
	if err != nil {
		// Return simplified result if calculation fails
		return &LiquidityResult{
			Impact: decimal.NewFromFloat(0.05),
		}
	}

	liquidityRatio := decimal.NewFromFloat(liquidityResult.LiquidityRatio)

	// Estimate impact (simplified)
	// In production, this would properly calculate the new liquidity ratio
	estimatedImpact := decimal.NewFromFloat(0.05) // 5% impact
	newLiquidityRatio := liquidityRatio.Sub(estimatedImpact)

	result := &LiquidityResult{
		Impact: estimatedImpact,
	}

	if newLiquidityRatio.LessThan(thresholds.MinLiquidityRatio) {
		result.Violation = &RiskViolation{
			Type:         "LIQUIDITY_RATIO",
			Severity:     "WARNING",
			Description:  "Trade reduces liquidity below minimum",
			CurrentValue: newLiquidityRatio,
			Limit:        thresholds.MinLiquidityRatio,
			Impact:       thresholds.MinLiquidityRatio.Sub(newLiquidityRatio).Div(thresholds.MinLiquidityRatio),
		}
	}

	return result
}

func (res *RiskEngineService) calculateSuggestedStopLoss(tx *models.Transaction) decimal.Decimal {
	// Simple 2% stop loss suggestion
	stopLossPercent := decimal.NewFromFloat(0.02)

	if tx.TransactionType == "BUY" {
		return tx.Price.Mul(decimal.NewFromFloat(1).Sub(stopLossPercent))
	}

	return tx.Price.Mul(decimal.NewFromFloat(1).Add(stopLossPercent))
}

func (res *RiskEngineService) calculateRiskScore(analysis *TradeRiskAnalysis) decimal.Decimal {
	score := decimal.Zero

	for _, violation := range analysis.Violations {
		switch violation.Severity {
		case "CRITICAL":
			score = score.Add(decimal.NewFromInt(30))
		case "VIOLATION":
			score = score.Add(decimal.NewFromInt(20))
		case "WARNING":
			score = score.Add(decimal.NewFromInt(10))
		}
	}

	// Add impact scores
	score = score.Add(analysis.PortfolioImpact.Mul(decimal.NewFromInt(20)))
	score = score.Add(analysis.ConcentrationImpact.Mul(decimal.NewFromInt(100)).Mul(decimal.NewFromInt(15)))
	score = score.Add(analysis.LiquidityImpact.Mul(decimal.NewFromInt(15)))

	// Cap at 100
	if score.GreaterThan(decimal.NewFromInt(100)) {
		return decimal.NewFromInt(100)
	}

	return score
}

func (res *RiskEngineService) determineApprovalStatus(analysis *TradeRiskAnalysis) (approved, requiresReview bool) {
	criticalCount := 0
	for _, v := range analysis.Violations {
		if v.Severity == "CRITICAL" {
			criticalCount++
		}
	}

	if criticalCount > 0 {
		return false, false // Rejected
	}

	if analysis.RiskScore.GreaterThan(decimal.NewFromInt(70)) || len(analysis.Violations) > 2 {
		return false, true // Requires review
	}

	if analysis.RiskScore.LessThan(decimal.NewFromInt(30)) && len(analysis.Violations) == 0 {
		return true, false // Approved
	}

	return false, true // Borderline - requires review
}

func (res *RiskEngineService) generateRecommendations(analysis *TradeRiskAnalysis, tx *models.Transaction) {
	// Size recommendation
	if analysis.PortfolioImpact.GreaterThan(decimal.NewFromFloat(0.1)) {
		suggestedSize := tx.Quantity.Mul(decimal.NewFromFloat(0.1).Div(analysis.PortfolioImpact))
		analysis.SuggestedSize = suggestedSize
	}

	// Hedge recommendation
	if analysis.ConcentrationImpact.GreaterThan(decimal.NewFromFloat(0.3)) {
		analysis.HedgeRecommendation = "Consider hedging with inverse ETF or options to reduce concentration risk"
	}
}

func (res *RiskEngineService) updateTransactionRiskStatus(tx *models.Transaction, analysis *TradeRiskAnalysis) {
	violationsJSON, _ := json.Marshal(analysis.Violations)

	updates := map[string]interface{}{
		"risk_approved":   analysis.Approved,
		"requires_review": analysis.RequiresReview,
		"risk_violations": violationsJSON,
		"risk_score":      int(analysis.RiskScore.IntPart()),
	}

	res.db.Model(tx).Updates(updates)
}

func (res *RiskEngineService) createRiskAlerts(tx *models.Transaction, analysis *TradeRiskAnalysis) {
	for _, violation := range analysis.Violations {
		if violation.Severity == "CRITICAL" || violation.Severity == "VIOLATION" {
			alert := &models.Alert{
				PortfolioID: tx.PortfolioID,
				AlertType:   "RISK_VIOLATION",
				Severity:    violation.Severity,
				Title:       fmt.Sprintf("Risk Violation: %s", violation.Type),
				Description: violation.Description,
				Source:      "RISK_ENGINE",
				Status:      "ACTIVE",
				TriggeredBy: models.JSON{
					"transaction_id": tx.ID,
					"symbol":         tx.Symbol,
					"violation":      violation,
				},
			}

			res.alertService.CreateAlert(alert)
		}
	}
}

// MonitorPortfolioRisk continuously monitors portfolio risk metrics
func (res *RiskEngineService) MonitorPortfolioRisk(portfolioID uuid.UUID) error {
	// Get portfolio
	var portfolio models.Portfolio
	if err := res.db.Preload("Positions").First(&portfolio, portfolioID).Error; err != nil {
		return fmt.Errorf("portfolio not found: %w", err)
	}

	// Get thresholds
	thresholds, err := res.getOrCreateThresholds(portfolioID)
	if err != nil {
		return err
	}

	// Calculate current VaR using enhanced method
	returns := []float64{} // Mock returns - would need real price data
	for range portfolio.Positions {
		// Generate mock returns for each position
		for i := 0; i < 30; i++ {
			ret := (rand.Float64() - 0.5) * 0.04 // ±2% daily change
			returns = append(returns, ret)
		}
	}

	var currentVaR95 float64
	if len(returns) > 0 {
		currentVaR95 = res.varCalculator.CalculateFilteredHistoricalVaR(returns, 0.95)
	} else {
		currentVaR95 = portfolio.TotalValue.InexactFloat64() * 0.05 // 5% fallback
	}

	varValue := decimal.NewFromFloat(currentVaR95)

	// Check VaR against thresholds
	if varValue.GreaterThan(thresholds.MaxVaR95) {
		res.alertService.CreateRiskBreachAlert(
			portfolioID,
			"VAR",
			varValue.InexactFloat64(),
			thresholds.MaxVaR95.InexactFloat64(),
		)
	}

	// Calculate liquidity using enhanced method
	liquidityMetrics, err := res.liquidityCalc.CalculateAdvancedLiquidityRisk(portfolio.Positions)
	if err != nil {
		return err
	}

	liquidityValue := decimal.NewFromFloat(liquidityMetrics.LiquidityScore)

	// Check liquidity against thresholds
	if liquidityValue.LessThan(thresholds.MinLiquidityRatio) {
		res.alertService.CreateRiskBreachAlert(
			portfolioID,
			"LIQUIDITY",
			liquidityValue.InexactFloat64(),
			thresholds.MinLiquidityRatio.InexactFloat64(),
		)
	}

	// Broadcast updates via Redis
	ctx := context.Background()
	update := map[string]interface{}{
		"portfolio_id": portfolioID,
		"var":          varValue.InexactFloat64(),
		"liquidity":    liquidityValue.InexactFloat64(),
		"timestamp":    time.Now().Unix(),
	}

	updateJSON, _ := json.Marshal(update)
	database.GetRedis().Publish(ctx, "risk_updates", updateJSON)

	return nil
}

// VaRCalculationRequest contains parameters for VaR calculation
type VaRCalculationRequest struct {
	PortfolioID     uuid.UUID `json:"portfolio_id"`
	ConfidenceLevel float64   `json:"confidence_level"`
	TimeHorizon     int       `json:"time_horizon"`
	Method          string    `json:"method"`
}

// VaRResult contains the calculated VaR and related metrics
type VaRResult struct {
	PortfolioID     uuid.UUID       `json:"portfolio_id"`
	VaRValue        decimal.Decimal `json:"var_value"`
	VaRPercentage   decimal.Decimal `json:"var_percentage"`
	ConfidenceLevel decimal.Decimal `json:"confidence_level"`
	TimeHorizon     int             `json:"time_horizon"`
	Method          string          `json:"method"`
	PortfolioValue  decimal.Decimal `json:"portfolio_value"`
	CalculatedAt    time.Time       `json:"calculated_at"`
	Status          string          `json:"status"`
	Threshold       decimal.Decimal `json:"threshold"`
}

// CalculateVaR calculates Value at Risk for a portfolio
func (res *RiskEngineService) CalculateVaR(req VaRCalculationRequest) (*VaRResult, error) {
	// Get portfolio and positions
	var portfolio models.Portfolio
	if err := res.db.Preload("Positions").First(&portfolio, req.PortfolioID).Error; err != nil {
		return nil, fmt.Errorf("portfolio not found: %w", err)
	}

	// Handle empty portfolio
	if len(portfolio.Positions) == 0 || portfolio.TotalValue.IsZero() {
		return &VaRResult{
			PortfolioID:     req.PortfolioID,
			VaRValue:        decimal.Zero,
			VaRPercentage:   decimal.Zero,
			ConfidenceLevel: decimal.NewFromFloat(req.ConfidenceLevel),
			TimeHorizon:     req.TimeHorizon,
			Method:          req.Method,
			PortfolioValue:  portfolio.TotalValue,
			CalculatedAt:    time.Now(),
			Status:          "SAFE",
			Threshold:       decimal.Zero,
		}, nil
	}

	// Generate mock price history for positions
	priceHistory := make(map[string][]float64)
	for _, position := range portfolio.Positions {
		// Generate simple mock price history (30 days)
		prices := make([]float64, 30)
		basePrice := position.CurrentPrice.InexactFloat64()
		for i := 0; i < 30; i++ {
			// Simple random walk
			change := (rand.Float64() - 0.5) * 0.04 // ±2% daily change
			if i == 0 {
				prices[i] = basePrice
			} else {
				prices[i] = prices[i-1] * (1 + change)
			}
		}
		priceHistory[position.Symbol] = prices
	}

	// Use enhanced VaR calculation
	returns := []float64{}
	for _, prices := range priceHistory {
		if len(prices) > 1 {
			for i := 1; i < len(prices); i++ {
				ret := (prices[i] - prices[i-1]) / prices[i-1]
				returns = append(returns, ret)
			}
		}
	}

	var currentVaR95 float64
	if len(returns) > 0 {
		currentVaR95 = res.varCalculator.CalculateFilteredHistoricalVaR(returns, req.ConfidenceLevel)
	} else {
		// Fallback to simple calculation
		varValue := portfolio.TotalValue.Mul(decimal.NewFromFloat(0.05)) // 5% VaR
		threshold := portfolio.TotalValue.Mul(decimal.NewFromFloat(0.08))

		status := "SAFE"
		if varValue.GreaterThan(threshold) {
			status = "CRITICAL"
		} else if varValue.GreaterThan(threshold.Mul(decimal.NewFromFloat(0.75))) {
			status = "WARNING"
		}

		return &VaRResult{
			PortfolioID:     req.PortfolioID,
			VaRValue:        varValue,
			VaRPercentage:   decimal.NewFromFloat(5.0),
			ConfidenceLevel: decimal.NewFromFloat(req.ConfidenceLevel),
			TimeHorizon:     req.TimeHorizon,
			Method:          "fallback",
			PortfolioValue:  portfolio.TotalValue,
			CalculatedAt:    time.Now(),
			Status:          status,
			Threshold:       threshold,
		}, nil
	}

	// Convert to service result format
	varValue := decimal.NewFromFloat(currentVaR95)
	threshold := portfolio.TotalValue.Mul(decimal.NewFromFloat(0.08))

	status := "SAFE"
	if varValue.GreaterThan(threshold) {
		status = "CRITICAL"
	} else if varValue.GreaterThan(threshold.Mul(decimal.NewFromFloat(0.75))) {
		status = "WARNING"
	}

	var varPercentage decimal.Decimal
	if portfolio.TotalValue.IsZero() {
		varPercentage = decimal.NewFromInt(0)
	} else {
		varPercentage = varValue.Div(portfolio.TotalValue).Mul(decimal.NewFromInt(100))
	}

	return &VaRResult{
		PortfolioID:     req.PortfolioID,
		VaRValue:        varValue,
		VaRPercentage:   varPercentage,
		ConfidenceLevel: decimal.NewFromFloat(req.ConfidenceLevel),
		TimeHorizon:     req.TimeHorizon,
		Method:          req.Method,
		PortfolioValue:  portfolio.TotalValue,
		CalculatedAt:    time.Now(),
		Status:          status,
		Threshold:       threshold,
	}, nil
}

// CalculateLiquidityRisk calculates liquidity risk for a portfolio
func (res *RiskEngineService) CalculateLiquidityRisk(portfolioID uuid.UUID) (*LiquidityResult, error) {
	// Get portfolio and positions
	var portfolio models.Portfolio
	if err := res.db.Preload("Positions").First(&portfolio, portfolioID).Error; err != nil {
		return nil, fmt.Errorf("portfolio not found: %w", err)
	}

	// Handle empty portfolio
	if len(portfolio.Positions) == 0 || portfolio.TotalValue.IsZero() {
		return &LiquidityResult{
			PortfolioID:     portfolioID,
			LiquidityRatio:  decimal.NewFromFloat(1.0), // 100% liquid (cash equivalent)
			LiquidityScore:  "HIGH",
			DaysToLiquidate: decimal.Zero,
			RiskAssessment:  "LOW_RISK",
			CalculatedAt:    time.Now(),
		}, nil
	}

	// Use the enhanced liquidity calculator
	liquidityMetrics, err := res.liquidityCalc.CalculateAdvancedLiquidityRisk(portfolio.Positions)
	if err != nil {
		// Fallback calculation based on position liquidity tags
		totalValue := decimal.Zero
		highLiquid := decimal.Zero

		for _, position := range portfolio.Positions {
			totalValue = totalValue.Add(position.MarketValue)
			if position.Liquidity == "HIGH" || position.Liquidity == "" {
				highLiquid = highLiquid.Add(position.MarketValue)
			}
		}

		liquidityRatio := decimal.NewFromFloat(0.5) // Default 50%
		if !totalValue.IsZero() {
			liquidityRatio = highLiquid.Div(totalValue)
		}

		riskAssessment := "MEDIUM_RISK"
		if liquidityRatio.GreaterThan(decimal.NewFromFloat(0.7)) {
			riskAssessment = "LOW_RISK"
		} else if liquidityRatio.LessThan(decimal.NewFromFloat(0.3)) {
			riskAssessment = "HIGH_RISK"
		}

		return &LiquidityResult{
			PortfolioID:     portfolioID,
			LiquidityRatio:  liquidityRatio,
			LiquidityScore:  "MEDIUM",
			DaysToLiquidate: decimal.NewFromFloat(3.0),
			RiskAssessment:  riskAssessment,
			CalculatedAt:    time.Now(),
		}, nil
	}

	// Convert to service result format
	liquidityRatio := decimal.NewFromFloat(liquidityMetrics.LiquidityScore)

	riskAssessment := "LOW_RISK"
	if liquidityMetrics.LiquidityScore < 0.3 {
		riskAssessment = "HIGH_RISK"
	} else if liquidityMetrics.LiquidityScore < 0.7 {
		riskAssessment = "MEDIUM_RISK"
	}

	// Calculate average time to liquidate from the metrics
	avgTimeToLiquidate := 0.0
	count := 0
	for key, time := range liquidityMetrics.TimeToLiquidate {
		if key != "" {
			avgTimeToLiquidate += time
			count++
		}
	}
	if count > 0 {
		avgTimeToLiquidate /= float64(count)
	} else {
		avgTimeToLiquidate = 3.0 // Default
	}

	return &LiquidityResult{
		PortfolioID:     portfolioID,
		LiquidityRatio:  liquidityRatio,
		LiquidityScore:  "HIGH", // Will be determined by score
		DaysToLiquidate: decimal.NewFromFloat(avgTimeToLiquidate),
		RiskAssessment:  riskAssessment,
		CalculatedAt:    time.Now(),
	}, nil
}

// CheckPositionLimits checks position size limits
func (res *RiskEngineService) CheckPositionLimits(portfolioID uuid.UUID, maxLimitPercent float64) (*PositionLimitResult, error) {
	// Get portfolio and positions
	var portfolio models.Portfolio
	if err := res.db.Preload("Positions").First(&portfolio, portfolioID).Error; err != nil {
		return nil, fmt.Errorf("portfolio not found: %w", err)
	}

	violations := []PositionViolation{}
	maxLimit := decimal.NewFromFloat(maxLimitPercent)

	for _, position := range portfolio.Positions {
		positionPercent := position.MarketValue.Div(portfolio.TotalValue).Mul(decimal.NewFromInt(100))
		if positionPercent.GreaterThan(maxLimit) {
			violations = append(violations, PositionViolation{
				Symbol:         position.Symbol,
				CurrentPercent: positionPercent,
				MaxPercent:     maxLimit,
				ExcessPercent:  positionPercent.Sub(maxLimit),
				MarketValue:    position.MarketValue,
				Severity:       "MAJOR",
			})
		}
	}

	status := "COMPLIANT"
	if len(violations) > 0 {
		status = "VIOLATION"
	}

	return &PositionLimitResult{
		PortfolioID:     portfolioID,
		MaxLimit:        maxLimit,
		Violations:      violations,
		ComplianceScore: decimal.NewFromInt(100).Sub(decimal.NewFromInt(int64(len(violations) * 10))),
		Status:          status,
		TotalPositions:  len(portfolio.Positions),
		CalculatedAt:    time.Now(),
	}, nil
}
