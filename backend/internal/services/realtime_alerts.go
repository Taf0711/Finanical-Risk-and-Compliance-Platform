package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
	wsHandler "github.com/Taf0711/financial-risk-monitor/internal/websocket"
)

// RealtimeAlertService manages real-time alerts and notifications
type RealtimeAlertService struct {
	db             *gorm.DB
	redis          *redis.Client
	wsHub          *wsHandler.SimpleHub
	tradingService *TradingService
	riskService    *AdvancedRiskService
	alertRules     map[string]*AlertRule
	activeMonitors map[string]*PortfolioMonitor
	mu             sync.RWMutex
	stopChan       chan bool
}

// AlertRule defines conditions for triggering alerts
type AlertRule struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Type          AlertType              `json:"type"`
	Conditions    []AlertCondition       `json:"conditions"`
	Actions       []AlertAction          `json:"actions"`
	Enabled       bool                   `json:"enabled"`
	Cooldown      time.Duration          `json:"cooldown"`
	LastTriggered time.Time              `json:"last_triggered"`
	Metadata      map[string]interface{} `json:"metadata"`
}

type AlertType string

const (
	AlertTypeRiskBreach       AlertType = "risk_breach"
	AlertTypePriceMovement    AlertType = "price_movement"
	AlertTypeVolumeSpike      AlertType = "volume_spike"
	AlertTypePositionLimit    AlertType = "position_limit"
	AlertTypeLiquidityRisk    AlertType = "liquidity_risk"
	AlertTypeVolatilitySpike  AlertType = "volatility_spike"
	AlertTypeCorrelationBreak AlertType = "correlation_break"
	AlertTypeDrawdown         AlertType = "drawdown"
)

type AlertCondition struct {
	Field     string      `json:"field"`
	Operator  string      `json:"operator"`
	Value     interface{} `json:"value"`
	Timeframe string      `json:"timeframe"`
}

type AlertAction struct {
	Type       string                 `json:"type"`
	Parameters map[string]interface{} `json:"parameters"`
}

// PortfolioMonitor tracks a specific portfolio for alert conditions
type PortfolioMonitor struct {
	PortfolioID  string
	UserID       string
	Rules        []*AlertRule
	LastUpdate   time.Time
	PriceHistory map[string][]PricePoint
	RiskHistory  []RiskSnapshot
	mu           sync.RWMutex
}

type PricePoint struct {
	Price     float64   `json:"price"`
	Volume    int64     `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}

type RiskSnapshot struct {
	VaR           float64   `json:"var"`
	Volatility    float64   `json:"volatility"`
	Concentration float64   `json:"concentration"`
	Liquidity     float64   `json:"liquidity"`
	Timestamp     time.Time `json:"timestamp"`
}

// RealtimeAlert represents an alert that needs to be sent
type RealtimeAlert struct {
	ID          string                 `json:"id"`
	PortfolioID string                 `json:"portfolio_id"`
	UserID      string                 `json:"user_id"`
	Type        AlertType              `json:"type"`
	Severity    string                 `json:"severity"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Actions     []string               `json:"suggested_actions"`
}

func NewRealtimeAlertService(
	wsHub *wsHandler.SimpleHub,
	tradingService *TradingService,
	riskService *AdvancedRiskService,
) *RealtimeAlertService {
	service := &RealtimeAlertService{
		db:             database.GetDB(),
		redis:          database.GetRedis(),
		wsHub:          wsHub,
		tradingService: tradingService,
		riskService:    riskService,
		alertRules:     make(map[string]*AlertRule),
		activeMonitors: make(map[string]*PortfolioMonitor),
		stopChan:       make(chan bool),
	}

	// Initialize default alert rules
	service.initializeDefaultRules()

	return service
}

// Start begins monitoring for real-time alerts
func (s *RealtimeAlertService) Start() {
	log.Println("Starting realtime alert service...")

	// Start monitoring goroutines
	go s.monitorPortfolios()
	go s.monitorPriceMovements()
	go s.monitorRiskMetrics()

	log.Println("Realtime alert service started")
}

// Stop stops all monitoring
func (s *RealtimeAlertService) Stop() {
	log.Println("Stopping realtime alert service...")
	close(s.stopChan)
}

// AddPortfolioMonitor adds a portfolio to real-time monitoring
func (s *RealtimeAlertService) AddPortfolioMonitor(portfolioID, userID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.activeMonitors[portfolioID]; exists {
		return nil // Already monitoring
	}

	monitor := &PortfolioMonitor{
		PortfolioID:  portfolioID,
		UserID:       userID,
		Rules:        s.getApplicableRules(portfolioID),
		LastUpdate:   time.Now(),
		PriceHistory: make(map[string][]PricePoint),
		RiskHistory:  make([]RiskSnapshot, 0),
	}

	s.activeMonitors[portfolioID] = monitor
	log.Printf("Added portfolio monitor for %s", portfolioID)

	return nil
}

// RemovePortfolioMonitor removes a portfolio from monitoring
func (s *RealtimeAlertService) RemovePortfolioMonitor(portfolioID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.activeMonitors, portfolioID)
	log.Printf("Removed portfolio monitor for %s", portfolioID)
}

// CreateAlertRule creates a new custom alert rule
func (s *RealtimeAlertService) CreateAlertRule(rule *AlertRule) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.alertRules[rule.ID] = rule

	// Store in database
	ruleJSON, _ := json.Marshal(rule)
	key := "alert_rule:" + rule.ID
	return s.redis.Set(context.Background(), key, ruleJSON, 0).Err()
}

func (s *RealtimeAlertService) monitorPortfolios() {
	ticker := time.NewTicker(30 * time.Second) // Check every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkAllPortfolios()
		case <-s.stopChan:
			return
		}
	}
}

func (s *RealtimeAlertService) monitorPriceMovements() {
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkPriceMovements()
		case <-s.stopChan:
			return
		}
	}
}

func (s *RealtimeAlertService) monitorRiskMetrics() {
	ticker := time.NewTicker(2 * time.Minute) // Check every 2 minutes
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.checkRiskMetrics()
		case <-s.stopChan:
			return
		}
	}
}

func (s *RealtimeAlertService) checkAllPortfolios() {
	s.mu.RLock()
	monitors := make([]*PortfolioMonitor, 0, len(s.activeMonitors))
	for _, monitor := range s.activeMonitors {
		monitors = append(monitors, monitor)
	}
	s.mu.RUnlock()

	for _, monitor := range monitors {
		s.checkPortfolioAlerts(monitor)
	}
}

func (s *RealtimeAlertService) checkPortfolioAlerts(monitor *PortfolioMonitor) {
	// Get current positions
	positions, err := s.tradingService.GetPositions()
	if err != nil {
		log.Printf("Error getting positions for portfolio %s: %v", monitor.PortfolioID, err)
		return
	}

	// Check each alert rule
	for _, rule := range monitor.Rules {
		if !rule.Enabled || time.Since(rule.LastTriggered) < rule.Cooldown {
			continue
		}

		if s.evaluateAlertRule(rule, positions, monitor) {
			s.triggerAlert(rule, monitor, positions)
		}
	}
}

func (s *RealtimeAlertService) checkPriceMovements() {
	// This would integrate with your market data service
	// For now, we'll simulate price movement alerts
	s.mu.RLock()
	monitors := make([]*PortfolioMonitor, 0, len(s.activeMonitors))
	for _, monitor := range s.activeMonitors {
		monitors = append(monitors, monitor)
	}
	s.mu.RUnlock()

	for _, monitor := range monitors {
		positions, err := s.tradingService.GetPositions()
		if err != nil {
			continue
		}

		for _, position := range positions {
			s.checkSymbolPriceMovement(monitor, position.Symbol)
		}
	}
}

func (s *RealtimeAlertService) checkSymbolPriceMovement(monitor *PortfolioMonitor, symbol string) {
	// Get current price (this would come from your market data service)
	// For now, simulate a price movement check

	monitor.mu.RLock()
	history, exists := monitor.PriceHistory[symbol]
	monitor.mu.RUnlock()

	if !exists || len(history) < 2 {
		return
	}

	current := history[len(history)-1]
	previous := history[len(history)-2]

	changePercent := (current.Price - previous.Price) / previous.Price * 100

	// Check for significant price movements
	if abs(changePercent) > 5.0 { // 5% movement
		severity := "medium"
		if abs(changePercent) > 10.0 {
			severity = "high"
		}

		alert := &RealtimeAlert{
			ID:          generateAlertID(),
			PortfolioID: monitor.PortfolioID,
			UserID:      monitor.UserID,
			Type:        AlertTypePriceMovement,
			Severity:    severity,
			Title:       "Significant Price Movement",
			Message:     fmt.Sprintf("%s moved %.2f%% in the last period", symbol, changePercent),
			Data: map[string]interface{}{
				"symbol":         symbol,
				"change_percent": changePercent,
				"current_price":  current.Price,
				"previous_price": previous.Price,
			},
			Timestamp: time.Now(),
			Actions:   []string{"Review position", "Consider rebalancing", "Monitor closely"},
		}

		s.sendAlert(alert)
	}
}

func (s *RealtimeAlertService) checkRiskMetrics() {
	s.mu.RLock()
	monitors := make([]*PortfolioMonitor, 0, len(s.activeMonitors))
	for _, monitor := range s.activeMonitors {
		monitors = append(monitors, monitor)
	}
	s.mu.RUnlock()

	for _, monitor := range monitors {
		// Calculate current risk metrics
		riskMetrics, err := s.riskService.CalculateAdvancedRisk(monitor.PortfolioID)
		if err != nil {
			continue
		}

		// Check for risk threshold breaches
		s.checkRiskThresholds(monitor, riskMetrics)

		// Update risk history
		monitor.mu.Lock()
		snapshot := RiskSnapshot{
			VaR:           float64(riskMetrics.VaR.Value1Day.InexactFloat64()),
			Volatility:    riskMetrics.VolatilityMetrics.PortfolioVolatility,
			Concentration: riskMetrics.ConcentrationRisk.HerfindahlIndex,
			Liquidity:     riskMetrics.LiquidityRisk.OverallScore,
			Timestamp:     time.Now(),
		}
		monitor.RiskHistory = append(monitor.RiskHistory, snapshot)

		// Keep only last 100 snapshots
		if len(monitor.RiskHistory) > 100 {
			monitor.RiskHistory = monitor.RiskHistory[1:]
		}
		monitor.mu.Unlock()
	}
}

func (s *RealtimeAlertService) checkRiskThresholds(monitor *PortfolioMonitor, metrics *RiskMetrics) {
	// VaR threshold check
	varThreshold := 100000.0 // $100k VaR threshold
	if metrics.VaR.Value1Day.InexactFloat64() > varThreshold {
		alert := &RealtimeAlert{
			ID:          generateAlertID(),
			PortfolioID: monitor.PortfolioID,
			UserID:      monitor.UserID,
			Type:        AlertTypeRiskBreach,
			Severity:    "high",
			Title:       "VaR Threshold Breach",
			Message:     fmt.Sprintf("Portfolio VaR (%.0f) exceeds threshold (%.0f)", metrics.VaR.Value1Day.InexactFloat64(), varThreshold),
			Data: map[string]interface{}{
				"current_var":  metrics.VaR.Value1Day.InexactFloat64(),
				"threshold":    varThreshold,
				"breach_ratio": metrics.VaR.Value1Day.InexactFloat64() / varThreshold,
			},
			Timestamp: time.Now(),
			Actions:   []string{"Reduce position sizes", "Add hedging instruments", "Review risk limits"},
		}
		s.sendAlert(alert)
	}

	// Concentration risk check
	if metrics.ConcentrationRisk.MaxSinglePosition > 0.25 { // 25% concentration
		alert := &RealtimeAlert{
			ID:          generateAlertID(),
			PortfolioID: monitor.PortfolioID,
			UserID:      monitor.UserID,
			Type:        AlertTypePositionLimit,
			Severity:    "medium",
			Title:       "High Position Concentration",
			Message:     fmt.Sprintf("Single position concentration (%.1f%%) exceeds recommended limit", metrics.ConcentrationRisk.MaxSinglePosition*100),
			Data: map[string]interface{}{
				"concentration": metrics.ConcentrationRisk.MaxSinglePosition,
				"threshold":     0.25,
			},
			Timestamp: time.Now(),
			Actions:   []string{"Diversify holdings", "Reduce large positions", "Add uncorrelated assets"},
		}
		s.sendAlert(alert)
	}

	// Liquidity risk check
	if metrics.LiquidityRisk.OverallScore < 50 { // Low liquidity score
		alert := &RealtimeAlert{
			ID:          generateAlertID(),
			PortfolioID: monitor.PortfolioID,
			UserID:      monitor.UserID,
			Type:        AlertTypeLiquidityRisk,
			Severity:    "medium",
			Title:       "Low Portfolio Liquidity",
			Message:     fmt.Sprintf("Portfolio liquidity score (%.0f) is below recommended level", metrics.LiquidityRisk.OverallScore),
			Data: map[string]interface{}{
				"liquidity_score":   metrics.LiquidityRisk.OverallScore,
				"days_to_liquidate": metrics.LiquidityRisk.DaysToLiquidate,
			},
			Timestamp: time.Now(),
			Actions:   []string{"Add liquid positions", "Reduce illiquid holdings", "Monitor market conditions"},
		}
		s.sendAlert(alert)
	}
}

func (s *RealtimeAlertService) evaluateAlertRule(rule *AlertRule, positions []Position, monitor *PortfolioMonitor) bool {
	// Simplified rule evaluation logic
	// In production, this would be more sophisticated

	for _, condition := range rule.Conditions {
		if !s.evaluateCondition(condition, positions, monitor) {
			return false // All conditions must be true
		}
	}

	return true
}

func (s *RealtimeAlertService) evaluateCondition(condition AlertCondition, positions []Position, monitor *PortfolioMonitor) bool {
	// Simplified condition evaluation
	// This would be expanded based on your specific needs

	switch condition.Field {
	case "portfolio_value":
		totalValue := 0.0
		for _, pos := range positions {
			dec, _ := decimal.NewFromString(pos.MarketValue)
			value, _ := dec.Float64()
			totalValue += value
		}
		return s.compareValues(totalValue, condition.Operator, condition.Value)

	case "position_count":
		return s.compareValues(float64(len(positions)), condition.Operator, condition.Value)

	default:
		return false
	}
}

func (s *RealtimeAlertService) compareValues(actual float64, operator string, expected interface{}) bool {
	expectedFloat, ok := expected.(float64)
	if !ok {
		return false
	}

	switch operator {
	case ">":
		return actual > expectedFloat
	case ">=":
		return actual >= expectedFloat
	case "<":
		return actual < expectedFloat
	case "<=":
		return actual <= expectedFloat
	case "==":
		return actual == expectedFloat
	case "!=":
		return actual != expectedFloat
	default:
		return false
	}
}

func (s *RealtimeAlertService) triggerAlert(rule *AlertRule, monitor *PortfolioMonitor, positions []Position) {
	alert := &RealtimeAlert{
		ID:          generateAlertID(),
		PortfolioID: monitor.PortfolioID,
		UserID:      monitor.UserID,
		Type:        rule.Type,
		Severity:    "medium", // Could be derived from rule
		Title:       rule.Name,
		Message:     fmt.Sprintf("Alert rule '%s' has been triggered", rule.Name),
		Data: map[string]interface{}{
			"rule_id":         rule.ID,
			"positions_count": len(positions),
		},
		Timestamp: time.Now(),
		Actions:   []string{"Review rule conditions", "Take appropriate action"},
	}

	s.sendAlert(alert)

	// Update last triggered time
	rule.LastTriggered = time.Now()
}

func (s *RealtimeAlertService) sendAlert(alert *RealtimeAlert) {
	// Parse portfolio ID to UUID
	portfolioUUID, err := uuid.Parse(alert.PortfolioID)
	if err != nil {
		log.Printf("Invalid portfolio ID format: %v", err)
		return
	}

	// Store alert in database
	dbAlert := models.Alert{
		PortfolioID: portfolioUUID,
		AlertType:   string(alert.Type),
		Severity:    alert.Severity,
		Title:       alert.Title,
		Description: alert.Message,
		Source:      "realtime_monitor",
		Status:      "ACTIVE",
		TriggeredBy: alert.Data,
	}

	if err := s.db.Create(&dbAlert).Error; err != nil {
		log.Printf("Error storing alert: %v", err)
	}

	// Send via WebSocket
	wsMessage := map[string]interface{}{
		"type": "new_alert",
		"data": alert,
	}

	if s.wsHub != nil {
		s.wsHub.BroadcastToAll(wsMessage)
	}

	// Execute alert actions
	for _, action := range alert.Actions {
		s.executeAlertAction(action, alert)
	}

	log.Printf("Alert sent: %s - %s", alert.Type, alert.Title)
}

func (s *RealtimeAlertService) executeAlertAction(action string, alert *RealtimeAlert) {
	// Implement various alert actions
	switch action {
	case "email":
		// Send email notification
		log.Printf("Would send email alert for: %s", alert.Title)
	case "sms":
		// Send SMS notification
		log.Printf("Would send SMS alert for: %s", alert.Title)
	case "webhook":
		// Call webhook
		log.Printf("Would call webhook for: %s", alert.Title)
	default:
		// Log action
		log.Printf("Alert action '%s' for: %s", action, alert.Title)
	}
}

func (s *RealtimeAlertService) initializeDefaultRules() {
	// High VaR rule
	highVarRule := &AlertRule{
		ID:   "high_var_rule",
		Name: "High Value at Risk",
		Type: AlertTypeRiskBreach,
		Conditions: []AlertCondition{
			{
				Field:     "var_1day",
				Operator:  ">",
				Value:     50000.0,
				Timeframe: "1d",
			},
		},
		Actions: []AlertAction{
			{
				Type: "websocket",
				Parameters: map[string]interface{}{
					"severity": "high",
				},
			},
		},
		Enabled:  true,
		Cooldown: 1 * time.Hour,
	}

	s.alertRules[highVarRule.ID] = highVarRule

	// Add more default rules as needed
	log.Println("Initialized default alert rules")
}

func (s *RealtimeAlertService) getApplicableRules(portfolioID string) []*AlertRule {
	var rules []*AlertRule

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, rule := range s.alertRules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}

	return rules
}

// Helper functions
func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
