package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

type AlertGeneratorService struct {
	db          *gorm.DB
	redisClient *redis.Client
	riskService *RiskEngineService
}

func NewAlertGeneratorService() *AlertGeneratorService {
	return &AlertGeneratorService{
		db:          database.GetDB(),
		redisClient: database.GetRedis(),
		riskService: NewRiskEngineService(),
	}
}

// MonitorPortfolioRisks continuously monitors portfolios and generates alerts
func (a *AlertGeneratorService) MonitorPortfolioRisks() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.checkAllPortfoliosForRiskAlerts()
		}
	}
}

// checkAllPortfoliosForRiskAlerts checks all active portfolios for risk threshold breaches
func (a *AlertGeneratorService) checkAllPortfoliosForRiskAlerts() {
	var portfolios []models.Portfolio
	if err := a.db.Find(&portfolios).Error; err != nil {
		return
	}

	for _, portfolio := range portfolios {
		a.checkPortfolioRisks(portfolio.ID)
	}
}

// checkPortfolioRisks performs risk analysis and generates alerts if needed
func (a *AlertGeneratorService) checkPortfolioRisks(portfolioID uuid.UUID) {
	// Check VaR
	varReq := VaRCalculationRequest{
		PortfolioID:     portfolioID,
		ConfidenceLevel: 0.95,
		TimeHorizon:     1,
		Method:          "historical",
	}

	varResult, err := a.riskService.CalculateVaR(varReq)
	if err == nil && (varResult.Status == "WARNING" || varResult.Status == "CRITICAL") {
		a.generateVaRAlert(varResult)
	}

	// Check Liquidity Risk
	liquidityResult, err := a.riskService.CalculateLiquidityRisk(portfolioID)
	if err == nil && (liquidityResult.RiskAssessment == "MEDIUM_RISK" || liquidityResult.RiskAssessment == "HIGH_RISK") {
		a.generateLiquidityAlert(liquidityResult)
	}

	// Check Position Limits
	positionResult, err := a.riskService.CheckPositionLimits(portfolioID, 25.0)
	if err == nil && len(positionResult.Violations) > 0 {
		a.generatePositionLimitAlert(positionResult)
	}

	// Check for AML flags (mock implementation)
	a.checkForAMLAlerts(portfolioID)
}

// generateVaRAlert creates a VaR threshold breach alert
func (a *AlertGeneratorService) generateVaRAlert(varResult *VaRResult) {
	// Check if similar alert exists in last 10 minutes to avoid spam
	if a.alertExists(varResult.PortfolioID, "RISK_BREACH", 10*time.Minute) {
		return
	}

	severity := "MEDIUM"
	if varResult.Status == "CRITICAL" {
		severity = "HIGH"
	}

	title := fmt.Sprintf("VaR Limit %s", varResult.Status)
	description := fmt.Sprintf("Portfolio VaR of $%.2f (%.2f%%) %s threshold of $%.2f",
		varResult.VaRValue.InexactFloat64(),
		varResult.VaRPercentage.InexactFloat64(),
		map[string]string{"WARNING": "approaching", "CRITICAL": "exceeds"}[varResult.Status],
		varResult.Threshold.InexactFloat64())

	alert := models.Alert{
		PortfolioID: varResult.PortfolioID,
		AlertType:   "RISK_BREACH",
		Severity:    severity,
		Title:       title,
		Description: description,
		Source:      "VAR_CALCULATOR",
		Status:      "ACTIVE",
		TriggeredBy: models.JSON{
			"var_value":        varResult.VaRValue,
			"var_percentage":   varResult.VaRPercentage,
			"threshold":        varResult.Threshold,
			"confidence_level": varResult.ConfidenceLevel,
			"method":           varResult.Method,
		},
	}

	a.storeAndBroadcastAlert(alert)
}

// generateLiquidityAlert creates a liquidity risk alert
func (a *AlertGeneratorService) generateLiquidityAlert(liquidityResult *LiquidityResult) {
	if a.alertExists(liquidityResult.PortfolioID, "LIQUIDITY_RISK", 15*time.Minute) {
		return
	}

	severity := "MEDIUM"
	if liquidityResult.RiskAssessment == "HIGH_RISK" {
		severity = "HIGH"
	}

	title := "Liquidity Risk Detected"
	description := fmt.Sprintf("Portfolio liquidity ratio of %.2f%% indicates %s. Estimated %s days to liquidate.",
		liquidityResult.LiquidityRatio.Mul(decimal.NewFromInt(100)).InexactFloat64(),
		map[string]string{
			"HIGH_RISK":   "high liquidity risk",
			"MEDIUM_RISK": "moderate liquidity risk",
		}[liquidityResult.RiskAssessment],
		liquidityResult.DaysToLiquidate.String())

	alert := models.Alert{
		PortfolioID: liquidityResult.PortfolioID,
		AlertType:   "LIQUIDITY_RISK",
		Severity:    severity,
		Title:       title,
		Description: description,
		Source:      "LIQUIDITY_CALCULATOR",
		Status:      "ACTIVE",
		TriggeredBy: models.JSON{
			"liquidity_ratio":     liquidityResult.LiquidityRatio,
			"liquidity_score":     liquidityResult.LiquidityScore,
			"days_to_liquidate":   liquidityResult.DaysToLiquidate,
			"risk_assessment":     liquidityResult.RiskAssessment,
			"liquidity_breakdown": liquidityResult.LiquidityBreakdown,
		},
	}

	a.storeAndBroadcastAlert(alert)
}

// generatePositionLimitAlert creates position concentration alerts
func (a *AlertGeneratorService) generatePositionLimitAlert(positionResult *PositionLimitResult) {
	if a.alertExists(positionResult.PortfolioID, "COMPLIANCE_VIOLATION", 5*time.Minute) {
		return
	}

	// Find the most severe violation
	maxSeverity := "MINOR"
	var criticalViolations []PositionViolation

	for _, violation := range positionResult.Violations {
		if violation.Severity == "CRITICAL" {
			criticalViolations = append(criticalViolations, violation)
			maxSeverity = "CRITICAL"
		} else if violation.Severity == "MAJOR" && maxSeverity != "CRITICAL" {
			maxSeverity = "MAJOR"
		}
	}

	severity := "MEDIUM"
	if maxSeverity == "CRITICAL" {
		severity = "HIGH"
	}

	title := "Position Limit Breach"
	description := fmt.Sprintf("%d position(s) exceed the %.1f%% concentration limit",
		len(positionResult.Violations),
		positionResult.MaxLimit.InexactFloat64())

	if len(criticalViolations) > 0 {
		violation := criticalViolations[0]
		description += fmt.Sprintf(". %s: %.2f%% (excess: %.2f%%)",
			violation.Symbol,
			violation.CurrentPercent.InexactFloat64(),
			violation.ExcessPercent.InexactFloat64())
	}

	alert := models.Alert{
		PortfolioID: positionResult.PortfolioID,
		AlertType:   "COMPLIANCE_VIOLATION",
		Severity:    severity,
		Title:       title,
		Description: description,
		Source:      "POSITION_LIMIT_CHECKER",
		Status:      "ACTIVE",
		TriggeredBy: models.JSON{
			"max_limit":        positionResult.MaxLimit,
			"violations_count": len(positionResult.Violations),
			"compliance_score": positionResult.ComplianceScore,
			"violations":       positionResult.Violations,
		},
	}

	a.storeAndBroadcastAlert(alert)
}

// checkForAMLAlerts simulates AML transaction monitoring
func (a *AlertGeneratorService) checkForAMLAlerts(portfolioID uuid.UUID) {
	// Get recent transactions for this portfolio
	var transactions []models.Transaction
	cutoff := time.Now().Add(-24 * time.Hour)

	if err := a.db.Where("portfolio_id = ? AND created_at > ?", portfolioID, cutoff).
		Find(&transactions).Error; err != nil {
		return
	}

	for _, tx := range transactions {
		// Check for large transactions (> $10,000)
		if tx.Amount.GreaterThan(decimal.NewFromInt(10000)) && !tx.AMLChecked {
			a.generateAMLAlert(tx)
		}

		// Check for rapid transactions (velocity check)
		if a.detectHighVelocity(portfolioID, cutoff) {
			a.generateVelocityAlert(portfolioID)
			break // Only generate one velocity alert per check
		}
	}
}

// generateAMLAlert creates AML-related alerts
func (a *AlertGeneratorService) generateAMLAlert(transaction models.Transaction) {
	if a.alertExists(transaction.PortfolioID, "SUSPICIOUS_ACTIVITY", 1*time.Hour) {
		return
	}

	alert := models.Alert{
		PortfolioID: transaction.PortfolioID,
		AlertType:   "SUSPICIOUS_ACTIVITY",
		Severity:    "HIGH",
		Title:       "Large Transaction Detected",
		Description: fmt.Sprintf("Transaction of $%.2f exceeds AML monitoring threshold ($10,000). Symbol: %s, Type: %s",
			transaction.Amount.InexactFloat64(),
			transaction.Symbol,
			transaction.TransactionType),
		Source: "AML_CHECKER",
		Status: "ACTIVE",
		TriggeredBy: models.JSON{
			"transaction_id": transaction.ID,
			"amount":         transaction.Amount,
			"symbol":         transaction.Symbol,
			"type":           transaction.TransactionType,
			"threshold":      10000,
		},
	}

	a.storeAndBroadcastAlert(alert)

	// Mark transaction as AML checked
	a.db.Model(&transaction).Update("aml_checked", true)
}

// generateVelocityAlert creates high-frequency trading alerts
func (a *AlertGeneratorService) generateVelocityAlert(portfolioID uuid.UUID) {
	if a.alertExists(portfolioID, "SUSPICIOUS_ACTIVITY", 30*time.Minute) {
		return
	}

	alert := models.Alert{
		PortfolioID: portfolioID,
		AlertType:   "SUSPICIOUS_ACTIVITY",
		Severity:    "MEDIUM",
		Title:       "High Transaction Velocity",
		Description: "Unusually high number of transactions detected in the last 24 hours. This may indicate suspicious trading patterns.",
		Source:      "VELOCITY_CHECKER",
		Status:      "ACTIVE",
		TriggeredBy: models.JSON{
			"time_window": "24h",
			"threshold":   10,
		},
	}

	a.storeAndBroadcastAlert(alert)
}

// detectHighVelocity checks if there are too many transactions in a time period
func (a *AlertGeneratorService) detectHighVelocity(portfolioID uuid.UUID, since time.Time) bool {
	var count int64
	a.db.Model(&models.Transaction{}).
		Where("portfolio_id = ? AND created_at > ?", portfolioID, since).
		Count(&count)

	return count > 10 // More than 10 transactions in 24 hours
}

// alertExists checks if a similar alert already exists to prevent spam
func (a *AlertGeneratorService) alertExists(portfolioID uuid.UUID, alertType string, within time.Duration) bool {
	var count int64
	cutoff := time.Now().Add(-within)

	a.db.Model(&models.Alert{}).
		Where("portfolio_id = ? AND alert_type = ? AND status = 'ACTIVE' AND created_at > ?",
			portfolioID, alertType, cutoff).
		Count(&count)

	return count > 0
}

// storeAndBroadcastAlert saves alert to database and broadcasts via WebSocket
func (a *AlertGeneratorService) storeAndBroadcastAlert(alert models.Alert) {
	// Save to database
	if err := a.db.Create(&alert).Error; err != nil {
		return
	}

	// Cache in Redis
	ctx := context.Background()
	alertJSON, _ := json.Marshal(alert)
	key := fmt.Sprintf("alert:%s", alert.ID)
	a.redisClient.Set(ctx, key, alertJSON, 24*time.Hour)
	a.redisClient.SAdd(ctx, "active_alerts", alert.ID.String())

	// Broadcast via WebSocket (publish to Redis channel)
	a.redisClient.Publish(ctx, "alerts_channel", alertJSON)

	fmt.Printf("🚨 Alert Generated: %s - %s (Severity: %s)\n",
		alert.AlertType, alert.Title, alert.Severity)
}
