package mock

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
	"github.com/Taf0711/financial-risk-monitor/internal/services"
	"github.com/Taf0711/financial-risk-monitor/internal/websocket"
)

type MockDataGenerator struct {
	hub          *websocket.Hub
	simpleHub    interface{} // We'll use interface{} to avoid import cycle
	redisClient  *redis.Client
	riskService  *services.RiskEngineService
	alertService *services.AlertService
	symbols      []string
	prices       map[string]float64
}

func NewMockDataGenerator(hub *websocket.Hub) *MockDataGenerator {
	return &MockDataGenerator{
		hub:          hub,
		redisClient:  database.GetRedis(),
		riskService:  services.NewRiskEngineService(),
		alertService: services.NewAlertService(),
		symbols: []string{
			"AAPL", "GOOGL", "MSFT", "AMZN", "TSLA",
			"JPM", "BAC", "GS", "MS", "WFC",
			"BTC", "ETH", "GOLD", "SILVER", "OIL",
		},
		prices: map[string]float64{
			"AAPL":   150.00,
			"GOOGL":  2800.00,
			"MSFT":   300.00,
			"AMZN":   3300.00,
			"TSLA":   800.00,
			"JPM":    140.00,
			"BAC":    35.00,
			"GS":     350.00,
			"MS":     90.00,
			"WFC":    45.00,
			"BTC":    45000.00,
			"ETH":    3000.00,
			"GOLD":   1800.00,
			"SILVER": 25.00,
			"OIL":    75.00,
		},
	}
}

// SetSimpleHub sets the simple hub for broadcasting
func (m *MockDataGenerator) SetSimpleHub(hub interface{}) {
	m.simpleHub = hub
}

// broadcastMessage sends message to both hubs
func (m *MockDataGenerator) broadcastMessage(message websocket.Message) {
	// Try to broadcast to original hub
	if m.hub != nil {
		if err := m.hub.BroadcastToAll(message); err != nil {
			log.Printf("Warning: Failed to broadcast to hub: %v", err)
		}
	}

	// Try to broadcast to simple hub using interface method
	if m.simpleHub != nil {
		// Type assertion to call BroadcastToAll
		if simpleHub, ok := m.simpleHub.(interface{ BroadcastToAll(interface{}) error }); ok {
			if err := simpleHub.BroadcastToAll(message); err != nil {
				log.Printf("Warning: Failed to broadcast to simple hub: %v", err)
			}
		}
	}
}

func (m *MockDataGenerator) Start() {
	log.Println("Starting mock data generator...")

	// Generate price updates
	go m.generatePriceUpdates()

	// Generate transactions
	go m.generateTransactions()

	// Generate risk metrics
	go m.generateRiskMetrics()

	// Generate alerts
	go m.generateAlerts()
}

func (m *MockDataGenerator) generatePriceUpdates() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			updates := make(map[string]interface{})

			for symbol, basePrice := range m.prices {
				// Random walk with mean reversion
				change := (rand.Float64() - 0.5) * 0.02 // ±1% change
				newPrice := basePrice * (1 + change)

				// Mean reversion
				if newPrice > basePrice*1.1 {
					newPrice = basePrice * 1.09
				} else if newPrice < basePrice*0.9 {
					newPrice = basePrice * 0.91
				}

				m.prices[symbol] = newPrice
				updates[symbol] = map[string]interface{}{
					"price":     newPrice,
					"change":    change * 100,
					"timestamp": time.Now().Unix(),
				}
			}

			// Broadcast price updates
			log.Printf("Generated price updates: %+v", updates)
			message := websocket.Message{
				Type: "price_update",
				Data: updates,
			}

			// Broadcast to all hubs
			m.broadcastMessage(message)

			// Store in Redis
			ctx := context.Background()
			for symbol, price := range m.prices {
				key := fmt.Sprintf("price:%s", symbol)
				m.redisClient.Set(ctx, key, price, 5*time.Minute)
			}
		}
	}
}

func (m *MockDataGenerator) generateTransactions() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Generate random transaction
			transaction := m.createMockTransaction()

			// Skip if empty transaction (failed to get portfolio)
			if transaction.ID == uuid.Nil {
				continue
			}

			// Check if it triggers AML flags
			if transaction.Amount.GreaterThan(decimal.NewFromInt(10000)) {
				m.generateAMLAlert(transaction)
			}

			// Broadcast transaction
			log.Printf("Generated transaction: %s %s %s @ %s", transaction.TransactionType, transaction.Symbol, transaction.Quantity.String(), transaction.Price.String())
			message := websocket.Message{
				Type: "new_transaction",
				Data: map[string]interface{}{
					"transaction": transaction,
					"timestamp":   time.Now().Unix(),
				},
			}

			// Broadcast to all hubs
			m.broadcastMessage(message)
		}
	}
}

func (m *MockDataGenerator) createMockTransaction() models.Transaction {
	symbol := m.symbols[rand.Intn(len(m.symbols))]
	quantity := decimal.NewFromFloat(rand.Float64() * 100)
	price := decimal.NewFromFloat(m.prices[symbol])
	amount := quantity.Mul(price)

	transactionTypes := []string{"BUY", "SELL"}
	transactionType := transactionTypes[rand.Intn(len(transactionTypes))]

	// Get actual portfolio ID from database
	var portfolios []models.Portfolio
	if err := database.GetDB().Find(&portfolios).Error; err != nil || len(portfolios) == 0 {
		// Fallback to a default portfolio ID if database query fails
		log.Printf("Warning: failed to fetch portfolios for transaction: %v", err)
		return models.Transaction{} // Return empty transaction
	}

	selectedPortfolio := portfolios[rand.Intn(len(portfolios))]

	return models.Transaction{
		ID:              uuid.New(),
		PortfolioID:     selectedPortfolio.ID,
		TransactionType: transactionType,
		Symbol:          symbol,
		Quantity:        quantity,
		Price:           price,
		Amount:          amount,
		Currency:        "USD",
		Status:          "COMPLETED",
		ExecutedAt:      &time.Time{},
		KYCVerified:     rand.Float64() > 0.1, // 90% verified
		AMLChecked:      rand.Float64() > 0.2, // 80% checked
		RiskScore:       rand.Intn(100),
		CreatedAt:       time.Now(),
	}
}

func (m *MockDataGenerator) generateRiskMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get existing portfolios to generate metrics for
			var portfolios []models.Portfolio
			if err := database.GetDB().Find(&portfolios).Error; err != nil {
				log.Printf("Warning: failed to fetch portfolios: %v", err)
				continue
			}

			if len(portfolios) == 0 {
				log.Println("No portfolios found, skipping risk metric generation")
				continue
			}

			// Pick a random portfolio
			portfolio := portfolios[rand.Intn(len(portfolios))]

			// Calculate actual VaR using RiskService
			varReq := services.VaRCalculationRequest{
				PortfolioID:     portfolio.ID,
				TimeHorizon:     1,
				ConfidenceLevel: 95.0,
				Method:          "historical_simulation",
			}
			varMetric, err := m.riskService.CalculateVaR(varReq)
			if err != nil {
				log.Printf("Warning: failed to calculate VaR for portfolio %s: %v", portfolio.ID, err)
			}

			// Calculate actual Liquidity using RiskService
			liquidityMetric, err := m.riskService.CalculateLiquidityRisk(portfolio.ID)
			if err != nil {
				log.Printf("Warning: failed to calculate liquidity for portfolio %s: %v", portfolio.ID, err)
			}

			// Skip this iteration if both metrics are nil (empty portfolio)
			if varMetric == nil && liquidityMetric == nil {
				log.Printf("Skipping risk metric generation for empty portfolio %s", portfolio.ID)
				continue
			}

			// Check if we need to generate alerts for breaches
			if varMetric != nil && varMetric.Status != "SAFE" {
				err := m.alertService.CreateRiskBreachAlert(
					portfolio.ID,
					"VAR",
					varMetric.VaRValue.InexactFloat64(),
					varMetric.Threshold.InexactFloat64(),
				)
				if err != nil {
					log.Printf("Warning: failed to create VaR alert: %v", err)
				}
			}

			if liquidityMetric != nil && liquidityMetric.RiskAssessment != "LOW_RISK" {
				err := m.alertService.CreateRiskBreachAlert(
					portfolio.ID,
					"LIQUIDITY_RATIO",
					liquidityMetric.LiquidityRatio.InexactFloat64(),
					decimal.NewFromFloat(0.3).InexactFloat64(), // Default threshold
				)
				if err != nil {
					log.Printf("Warning: failed to create liquidity alert: %v", err)
				}
			}

			// Broadcast risk metrics
			var varStr, varStatus, liquidityStr, liquidityStatus string
			if varMetric != nil {
				varStr = varMetric.VaRValue.String()
				varStatus = varMetric.Status
			} else {
				varStr = "N/A"
				varStatus = "N/A"
			}
			if liquidityMetric != nil {
				liquidityStr = liquidityMetric.LiquidityRatio.String()
				liquidityStatus = liquidityMetric.RiskAssessment
			} else {
				liquidityStr = "N/A"
				liquidityStatus = "N/A"
			}

			log.Printf("Generated risk metrics for portfolio %s - VaR: %s (%s), Liquidity: %s (%s)",
				portfolio.ID, varStr, varStatus, liquidityStr, liquidityStatus)

			message := websocket.Message{
				Type: "risk_update",
				Data: map[string]interface{}{
					"portfolio_id": portfolio.ID,
					"var":          varMetric,
					"liquidity":    liquidityMetric,
					"timestamp":    time.Now().Unix(),
				},
			}

			m.broadcastMessage(message)
		}
	}
}

func (m *MockDataGenerator) getRiskStatus(value, threshold float64) string {
	ratio := value / threshold
	switch {
	case ratio < 0.8:
		return "SAFE"
	case ratio < 1.0:
		return "WARNING"
	default:
		return "CRITICAL"
	}
}

func (m *MockDataGenerator) getLiquidityStatus(ratio float64) string {
	switch {
	case ratio >= 0.7:
		return "SAFE"
	case ratio >= 0.3:
		return "WARNING"
	default:
		return "CRITICAL"
	}
}

func (m *MockDataGenerator) generateAlerts() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Get existing portfolios to generate alerts for
			var portfolios []models.Portfolio
			if err := database.GetDB().Find(&portfolios).Error; err != nil {
				log.Printf("Warning: failed to fetch portfolios: %v", err)
				continue
			}

			if len(portfolios) == 0 {
				log.Println("No portfolios found, skipping alert generation")
				continue
			}

			// Randomly generate an alert (30% chance)
			if rand.Float64() > 0.7 {
				portfolio := portfolios[rand.Intn(len(portfolios))]

				alertTypes := []struct {
					Type        string
					Severity    string
					Title       string
					Description string
					Source      string
				}{
					{
						Type:        "RISK_BREACH",
						Severity:    "HIGH",
						Title:       "VaR Limit Exceeded",
						Description: "Portfolio Value at Risk exceeds threshold",
						Source:      "VAR_CALCULATOR",
					},
					{
						Type:        "COMPLIANCE_VIOLATION",
						Severity:    "CRITICAL",
						Title:       "Position Limit Breach",
						Description: "Single position exceeds 25% of portfolio",
						Source:      "POSITION_LIMIT_CHECKER",
					},
					{
						Type:        "SUSPICIOUS_ACTIVITY",
						Severity:    "MEDIUM",
						Title:       "Unusual Trading Pattern",
						Description: "High frequency trading detected",
						Source:      "PATTERN_DETECTOR",
					},
				}

				alertType := alertTypes[rand.Intn(len(alertTypes))]

				alert := &models.Alert{
					PortfolioID: portfolio.ID,
					AlertType:   alertType.Type,
					Severity:    alertType.Severity,
					Title:       alertType.Title,
					Description: alertType.Description,
					Source:      alertType.Source,
					Status:      "ACTIVE",
					TriggeredBy: models.JSON{
						"mock_generated": true,
						"portfolio_name": portfolio.Name,
					},
				}

				// Store alert in database using AlertService
				err := m.alertService.CreateAlert(alert)
				if err != nil {
					log.Printf("Warning: failed to create alert: %v", err)
					continue
				}

				// Broadcast alert
				log.Printf("Generated alert for portfolio %s: %s - %s", portfolio.ID, alert.Severity, alert.Title)
				message := websocket.Message{
					Type: "new_alert",
					Data: map[string]interface{}{
						"alert":     alert,
						"timestamp": time.Now().Unix(),
					},
				}

				m.broadcastMessage(message)

				// Store in Redis for caching
				ctx := context.Background()
				alertJSON, _ := json.Marshal(alert)
				key := fmt.Sprintf("alert:%s", alert.ID)
				m.redisClient.Set(ctx, key, alertJSON, 24*time.Hour)
			}
		}
	}
}

func (m *MockDataGenerator) generateAMLAlert(transaction models.Transaction) {
	alert := &models.Alert{
		PortfolioID: transaction.PortfolioID,
		AlertType:   "SUSPICIOUS_ACTIVITY",
		Severity:    "HIGH",
		Title:       "Large Transaction Detected",
		Description: fmt.Sprintf("Transaction of %s exceeds AML threshold", transaction.Amount),
		Source:      "AML_CHECKER",
		Status:      "ACTIVE",
		TriggeredBy: models.JSON{
			"transaction_id": transaction.ID,
			"amount":         transaction.Amount,
			"symbol":         transaction.Symbol,
			"mock_generated": true,
		},
	}

	// Store alert in database using AlertService
	err := m.alertService.CreateAlert(alert)
	if err != nil {
		log.Printf("Warning: failed to create AML alert: %v", err)
		return
	}

	log.Printf("Generated AML alert for transaction: %s %s", transaction.Amount.String(), transaction.Symbol)
	message := websocket.Message{
		Type: "aml_alert",
		Data: map[string]interface{}{
			"alert":       alert,
			"transaction": transaction,
			"timestamp":   time.Now().Unix(),
		},
	}

	m.broadcastMessage(message)
}
