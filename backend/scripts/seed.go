package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"

	"github.com/Taf0711/financial-risk-monitor/internal/config"
	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Initialize database
	if err := database.InitPostgres(&cfg.Database); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	db := database.GetDB()

	// Create demo users
	users := []models.User{
		{
			Email:     "admin@example.com",
			Password:  hashPassword("admin123"),
			FirstName: "Admin",
			LastName:  "User",
			Role:      "admin",
			IsActive:  true,
		},
		{
			Email:     "analyst@example.com",
			Password:  hashPassword("analyst123"),
			FirstName: "Risk",
			LastName:  "Analyst",
			Role:      "analyst",
			IsActive:  true,
		},
		{
			Email:     "trader@example.com",
			Password:  hashPassword("trader123"),
			FirstName: "John",
			LastName:  "Trader",
			Role:      "trader",
			IsActive:  true,
		},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Failed to create user %s: %v", user.Email, err)
			continue
		}
		log.Printf("Created user: %s", user.Email)

		// Create portfolios for each user
		portfolios := []models.Portfolio{
			{
				UserID:      user.ID,
				Name:        "Growth Portfolio",
				Description: "High growth technology stocks",
				TotalValue:  decimal.NewFromFloat(100000),
				Currency:    "USD",
			},
			{
				UserID:      user.ID,
				Name:        "Conservative Portfolio",
				Description: "Low risk bonds and blue chips",
				TotalValue:  decimal.NewFromFloat(250000),
				Currency:    "USD",
			},
		}

		for _, portfolio := range portfolios {
			if err := db.Create(&portfolio).Error; err != nil {
				log.Printf("Failed to create portfolio: %v", err)
				continue
			}
			log.Printf("Created portfolio: %s", portfolio.Name)

			// Create positions
			positions := generatePositions(portfolio.ID)
			for _, position := range positions {
				if err := db.Create(&position).Error; err != nil {
					log.Printf("Failed to create position: %v", err)
				}
			}

			// Create transactions
			transactions := generateTransactions(portfolio.ID)
			for _, transaction := range transactions {
				if err := db.Create(&transaction).Error; err != nil {
					log.Printf("Failed to create transaction: %v", err)
				}
			}

			// Create risk metrics
			metrics := generateRiskMetrics(portfolio.ID)
			for _, metric := range metrics {
				if err := db.Create(&metric).Error; err != nil {
					log.Printf("Failed to create risk metric: %v", err)
				}
			}
		}
	}

	log.Println("Database seeding completed!")
}

func hashPassword(password string) string {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashedPassword)
}

func generatePositions(portfolioID uuid.UUID) []models.Position {
	symbols := []string{"AAPL", "GOOGL", "MSFT", "AMZN", "TSLA", "JPM", "BAC"}
	positions := []models.Position{}

	for _, symbol := range symbols {
		quantity := decimal.NewFromFloat(rand.Float64() * 100)
		avgPrice := decimal.NewFromFloat(100 + rand.Float64()*200)
		currentPrice := avgPrice.Mul(decimal.NewFromFloat(0.9 + rand.Float64()*0.3))
		marketValue := quantity.Mul(currentPrice)

		positions = append(positions, models.Position{
			PortfolioID:  portfolioID,
			Symbol:       symbol,
			Quantity:     quantity,
			AveragePrice: avgPrice,
			CurrentPrice: currentPrice,
			MarketValue:  marketValue,
			PnL:          marketValue.Sub(quantity.Mul(avgPrice)),
			PnLPercent:   currentPrice.Sub(avgPrice).Div(avgPrice).Mul(decimal.NewFromInt(100)),
			Weight:       decimal.NewFromFloat(rand.Float64() * 0.2), // 0-20% weight
			AssetType:    "STOCK",
			Liquidity:    getLiquidity(),
		})
	}

	return positions
}

func getLiquidity() string {
	r := rand.Float64()
	switch {
	case r > 0.7:
		return "HIGH"
	case r > 0.3:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func generateTransactions(portfolioID uuid.UUID) []models.Transaction {
	transactions := []models.Transaction{}
	symbols := []string{"AAPL", "GOOGL", "MSFT", "AMZN", "TSLA"}
	types := []string{"BUY", "SELL"}

	for i := 0; i < 20; i++ {
		now := time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour)
		transactions = append(transactions, models.Transaction{
			PortfolioID:     portfolioID,
			TransactionType: types[rand.Intn(2)],
			Symbol:          symbols[rand.Intn(len(symbols))],
			Quantity:        decimal.NewFromFloat(rand.Float64() * 50),
			Price:           decimal.NewFromFloat(100 + rand.Float64()*200),
			Amount:          decimal.NewFromFloat(1000 + rand.Float64()*10000),
			Currency:        "USD",
			Status:          "COMPLETED",
			ExecutedAt:      &now,
			KYCVerified:     true,
			AMLChecked:      true,
			RiskScore:       rand.Intn(30),
			CreatedAt:       now,
		})
	}

	return transactions
}

func generateRiskMetrics(portfolioID uuid.UUID) []models.RiskMetric {
	return []models.RiskMetric{
		{
			PortfolioID:     portfolioID,
			MetricType:      "VAR",
			Value:           decimal.NewFromFloat(50000 + rand.Float64()*30000),
			Threshold:       decimal.NewFromFloat(75000),
			Status:          "WARNING",
			CalculatedAt:    time.Now(),
			TimeHorizon:     1,
			ConfidenceLevel: decimal.NewFromFloat(0.95),
		},
		{
			PortfolioID:  portfolioID,
			MetricType:   "LIQUIDITY_RATIO",
			Value:        decimal.NewFromFloat(0.6 + rand.Float64()*0.3),
			Threshold:    decimal.NewFromFloat(0.3),
			Status:       "SAFE",
			CalculatedAt: time.Now(),
		},
	}
}
