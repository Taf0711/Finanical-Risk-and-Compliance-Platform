package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

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
	users := createUsers(db)
	log.Printf("Created %d users", len(users))

	// Create portfolios and positions for each user
	for _, user := range users {
		portfolios := createPortfoliosForUser(db, user)
		log.Printf("Created %d portfolios for user %s", len(portfolios), user.Email)

		for _, portfolio := range portfolios {
			positions := createPositionsForPortfolio(db, portfolio)
			log.Printf("Created %d positions for portfolio %s", len(positions), portfolio.Name)

			transactions := createTransactionsForPortfolio(db, portfolio)
			log.Printf("Created %d transactions for portfolio %s", len(transactions), portfolio.Name)

			// Update portfolio total value
			updatePortfolioValue(db, portfolio.ID)
		}
	}

	log.Println("✅ Demo data seeding completed successfully!")
	log.Println("🔐 Login credentials:")
	log.Println("   Email: demo@example.com")
	log.Println("   Password: password123")
	log.Println("")
	log.Println("📊 You can now test the risk calculation endpoints!")
}

func createUsers(db *gorm.DB) []models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	users := []models.User{
		{
			Email:     "demo@example.com",
			Password:  string(hashedPassword),
			FirstName: "Demo",
			LastName:  "User",
			Role:      "analyst",
			IsActive:  true,
		},
		{
			Email:     "trader@example.com",
			Password:  string(hashedPassword),
			FirstName: "John",
			LastName:  "Trader",
			Role:      "trader",
			IsActive:  true,
		},
	}

	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			log.Printf("User %s may already exist: %v", users[i].Email, err)
		}
	}

	return users
}

func createPortfoliosForUser(db *gorm.DB, user models.User) []models.Portfolio {
	portfolios := []models.Portfolio{
		{
			UserID:      user.ID,
			Name:        "Growth Portfolio",
			Description: "High growth technology stocks with risk monitoring",
			TotalValue:  decimal.Zero, // Will be calculated
			Currency:    "USD",
		},
		{
			UserID:      user.ID,
			Name:        "Conservative Portfolio",
			Description: "Low risk bonds and blue chips with compliance checking",
			TotalValue:  decimal.Zero, // Will be calculated
			Currency:    "USD",
		},
	}

	for i := range portfolios {
		if err := db.Create(&portfolios[i]).Error; err != nil {
			log.Printf("Portfolio creation error: %v", err)
		}
	}

	return portfolios
}

func createPositionsForPortfolio(db *gorm.DB, portfolio models.Portfolio) []models.Position {
	// Create diverse positions with different liquidity levels
	positionsData := []struct {
		Symbol       string
		Quantity     float64
		AveragePrice float64
		CurrentPrice float64
		AssetType    string
		Liquidity    string
	}{
		{"AAPL", 100, 150.00, 155.50, "STOCK", "HIGH"},
		{"GOOGL", 50, 2800.00, 2850.00, "STOCK", "HIGH"},
		{"MSFT", 75, 300.00, 310.00, "STOCK", "HIGH"},
		{"TSLA", 30, 800.00, 790.00, "STOCK", "HIGH"},
		{"JPM", 200, 140.00, 142.50, "STOCK", "HIGH"},
		{"BOND1", 1000, 100.00, 102.00, "BOND", "MEDIUM"},
		{"REIT1", 500, 50.00, 48.00, "REIT", "MEDIUM"},
		{"PRIV1", 100, 1000.00, 1050.00, "PRIVATE", "LOW"},
	}

	var positions []models.Position

	for _, data := range positionsData {
		marketValue := decimal.NewFromFloat(data.Quantity * data.CurrentPrice)
		avgPrice := decimal.NewFromFloat(data.AveragePrice)
		currentPrice := decimal.NewFromFloat(data.CurrentPrice)
		quantity := decimal.NewFromFloat(data.Quantity)

		pnl := marketValue.Sub(quantity.Mul(avgPrice))
		pnlPercent := pnl.Div(quantity.Mul(avgPrice)).Mul(decimal.NewFromInt(100))

		// Calculate weight as a percentage (will be updated later when portfolio total is calculated)
		weight := decimal.NewFromFloat(100.0 / float64(len(positionsData))) // Equal weighting for demo

		position := models.Position{
			PortfolioID:  portfolio.ID,
			Symbol:       data.Symbol,
			Quantity:     quantity,
			AveragePrice: avgPrice,
			CurrentPrice: currentPrice,
			MarketValue:  marketValue,
			PnL:          pnl,
			PnLPercent:   pnlPercent,
			Weight:       weight,
			AssetType:    data.AssetType,
			Liquidity:    data.Liquidity,
		}

		if err := db.Create(&position).Error; err != nil {
			log.Printf("Position creation error: %v", err)
		} else {
			positions = append(positions, position)
		}
	}

	return positions
}

func createTransactionsForPortfolio(db *gorm.DB, portfolio models.Portfolio) []models.Transaction {
	symbols := []string{"AAPL", "GOOGL", "MSFT", "TSLA", "JPM"}
	transactionTypes := []string{"BUY", "SELL"}

	var transactions []models.Transaction

	// Create 20 sample transactions
	for i := 0; i < 20; i++ {
		symbol := symbols[rand.Intn(len(symbols))]
		txType := transactionTypes[rand.Intn(len(transactionTypes))]
		quantity := decimal.NewFromFloat(10 + rand.Float64()*90) // 10-100 shares
		price := decimal.NewFromFloat(100 + rand.Float64()*200)  // $100-300 per share
		amount := quantity.Mul(price)

		// Create transaction from 1-30 days ago
		executedAt := time.Now().Add(-time.Duration(rand.Intn(30)) * 24 * time.Hour)

		transaction := models.Transaction{
			PortfolioID:     portfolio.ID,
			TransactionType: txType,
			Symbol:          symbol,
			Quantity:        quantity,
			Price:           price,
			Amount:          amount,
			Currency:        "USD",
			Status:          "COMPLETED",
			ExecutedAt:      &executedAt,
			KYCVerified:     true,
			AMLChecked:      true,
			RiskScore:       rand.Intn(30), // Low risk scores
		}

		if err := db.Create(&transaction).Error; err != nil {
			log.Printf("Transaction creation error: %v", err)
		} else {
			transactions = append(transactions, transaction)
		}
	}

	return transactions
}

func updatePortfolioValue(db *gorm.DB, portfolioID interface{}) {
	// Calculate total portfolio value from positions
	var positions []models.Position
	if err := db.Where("portfolio_id = ?", portfolioID).Find(&positions).Error; err != nil {
		log.Printf("Error fetching positions: %v", err)
		return
	}

	totalValue := decimal.Zero
	for _, position := range positions {
		totalValue = totalValue.Add(position.MarketValue)
	}

	// Update portfolio total value
	if err := db.Model(&models.Portfolio{}).Where("id = ?", portfolioID).Update("total_value", totalValue).Error; err != nil {
		log.Printf("Error updating portfolio value: %v", err)
	}

	log.Printf("Updated portfolio %v total value to $%s", portfolioID, totalValue.String())
}
