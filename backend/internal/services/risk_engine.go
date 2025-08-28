// backend/internal/services/risk_engine.go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/calculator"
)

type RiskEngineService struct {
	db            *gorm.DB
	alertService  *AlertService
	varCalculator *calculator.VaRCalculator
	liquidityCalc *calculator.LiquidityCalculator
}

func NewRiskEngineService() *RiskEngineService {
	return &RiskEngineService{
		db:            database.GetDB(),
		alertService:  NewAlertService(),
		varCalculator: calculator.NewVaRCalculator(100000),    // Default portfolio value
		liquidityCalc: calculator.NewLiquidityCalculator(nil), // Will need mock provider
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
	currentVaRResult, err := res.varCalculator.CalculateVaR(portfolio.Positions, priceHistory, 1)
	if err != nil {
		return nil, err
	}

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
	liquidityResult, err := res.liquidityCalc.CalculateLiquidity(portfolio.Positions, portfolio.TotalValue.InexactFloat64())
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

	// Calculate current VaR
	priceHistory := make(map[string][]float64) // Mock price history
	varResult, err := res.varCalculator.CalculateVaR(portfolio.Positions, priceHistory, 1)
	if err != nil {
		return err
	}

	varValue := decimal.NewFromFloat(varResult.VaR95)

	// Check VaR against thresholds
	if varValue.GreaterThan(thresholds.MaxVaR95) {
		res.alertService.CreateRiskBreachAlert(
			portfolioID,
			"VAR",
			varValue.InexactFloat64(),
			thresholds.MaxVaR95.InexactFloat64(),
		)
	}

	// Calculate liquidity
	liquidityResult, err := res.liquidityCalc.CalculateLiquidity(portfolio.Positions, portfolio.TotalValue.InexactFloat64())
	if err != nil {
		return err
	}

	liquidityValue := decimal.NewFromFloat(liquidityResult.LiquidityRatio)

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

	// Use the calculator
	priceHistory := make(map[string][]float64) // Mock price history
	calcResult, err := res.varCalculator.CalculateVaR(portfolio.Positions, priceHistory, req.TimeHorizon)
	if err != nil {
		return nil, err
	}

	// Convert to service result format
	varValue := decimal.NewFromFloat(calcResult.VaR95)
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
		VaRPercentage:   varValue.Div(portfolio.TotalValue).Mul(decimal.NewFromInt(100)),
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

	// Use the calculator
	calcResult, err := res.liquidityCalc.CalculateLiquidity(portfolio.Positions, portfolio.TotalValue.InexactFloat64())
	if err != nil {
		return nil, err
	}

	// Convert to service result format
	liquidityRatio := decimal.NewFromFloat(calcResult.LiquidityRatio)

	riskAssessment := "LOW_RISK"
	if calcResult.LiquidityRatio < 0.3 {
		riskAssessment = "HIGH_RISK"
	} else if calcResult.LiquidityRatio < 0.7 {
		riskAssessment = "MEDIUM_RISK"
	}

	return &LiquidityResult{
		PortfolioID:     portfolioID,
		LiquidityRatio:  liquidityRatio,
		LiquidityScore:  calcResult.LiquidityHealth,
		DaysToLiquidate: decimal.NewFromFloat(calcResult.NormalMarketDays),
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
