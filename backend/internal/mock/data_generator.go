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
	"github.com/Taf0711/financial-risk-monitor/internal/websocket"
)

type MockDataGenerator struct {
	hub         *websocket.Hub
	redisClient *redis.Client
	symbols     []string
	prices      map[string]float64
}

func NewMockDataGenerator(hub *websocket.Hub) *MockDataGenerator {
	return &MockDataGenerator{
		hub:         hub,
		redisClient: database.GetRedis(),
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
			message := websocket.Message{
				Type: "price_update",
				Data: updates,
			}

			m.hub.BroadcastToAll(message)

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

			// Check if it triggers AML flags
			if transaction.Amount.GreaterThan(decimal.NewFromInt(10000)) {
				m.generateAMLAlert(transaction)
			}

			// Broadcast transaction
			message := websocket.Message{
				Type: "new_transaction",
				Data: map[string]interface{}{
					"transaction": transaction,
					"timestamp":   time.Now().Unix(),
				},
			}

			m.hub.BroadcastToAll(message)
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

	return models.Transaction{
		ID:              uuid.New(),
		PortfolioID:     uuid.New(), // Mock portfolio ID
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
			// Generate VaR metric
			varValue := 50000 + rand.Float64()*50000 // $50k - $100k
			varMetric := models.RiskMetric{
				ID:              uuid.New(),
				PortfolioID:     uuid.New(),
				MetricType:      "VAR",
				Value:           decimal.NewFromFloat(varValue),
				Threshold:       decimal.NewFromFloat(75000),
				Status:          m.getRiskStatus(varValue, 75000),
				CalculatedAt:    time.Now(),
				TimeHorizon:     1,
				ConfidenceLevel: decimal.NewFromFloat(0.95),
			}

			// Generate Liquidity Ratio
			liquidityRatio := rand.Float64() // 0-100%
			liquidityMetric := models.RiskMetric{
				ID:           uuid.New(),
				PortfolioID:  uuid.New(),
				MetricType:   "LIQUIDITY_RATIO",
				Value:        decimal.NewFromFloat(liquidityRatio),
				Threshold:    decimal.NewFromFloat(0.3),
				Status:       m.getLiquidityStatus(liquidityRatio),
				CalculatedAt: time.Now(),
			}

			// Broadcast risk metrics
			message := websocket.Message{
				Type: "risk_update",
				Data: map[string]interface{}{
					"var":       varMetric,
					"liquidity": liquidityMetric,
					"timestamp": time.Now().Unix(),
				},
			}

			m.hub.BroadcastToAll(message)
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

	for {
		select {
		case <-ticker.C:
			// Randomly generate an alert
			if rand.Float64() > 0.7 { // 30% chance
				alertType := alertTypes[rand.Intn(len(alertTypes))]

				alert := models.Alert{
					ID:          uuid.New(),
					PortfolioID: uuid.New(),
					AlertType:   alertType.Type,
					Severity:    alertType.Severity,
					Title:       alertType.Title,
					Description: alertType.Description,
					Source:      alertType.Source,
					Status:      "ACTIVE",
					CreatedAt:   time.Now(),
				}

				// Broadcast alert
				message := websocket.Message{
					Type: "new_alert",
					Data: map[string]interface{}{
						"alert":     alert,
						"timestamp": time.Now().Unix(),
					},
				}

				m.hub.BroadcastToAll(message)

				// Store in Redis
				ctx := context.Background()
				alertJSON, _ := json.Marshal(alert)
				key := fmt.Sprintf("alert:%s", alert.ID)
				m.redisClient.Set(ctx, key, alertJSON, 24*time.Hour)
			}
		}
	}
}

func (m *MockDataGenerator) generateAMLAlert(transaction models.Transaction) {
	alert := models.Alert{
		ID:          uuid.New(),
		PortfolioID: transaction.PortfolioID,
		AlertType:   "SUSPICIOUS_ACTIVITY",
		Severity:    "HIGH",
		Title:       "Large Transaction Detected",
		Description: fmt.Sprintf("Transaction of %s exceeds AML threshold", transaction.Amount),
		Source:      "AML_CHECKER",
		Status:      "ACTIVE",
		TriggeredBy: map[string]interface{}{
			"transaction_id": transaction.ID,
			"amount":         transaction.Amount,
			"symbol":         transaction.Symbol,
		},
		CreatedAt: time.Now(),
	}

	message := websocket.Message{
		Type: "aml_alert",
		Data: map[string]interface{}{
			"alert":       alert,
			"transaction": transaction,
			"timestamp":   time.Now().Unix(),
		},
	}

	m.hub.BroadcastToAll(message)
}
