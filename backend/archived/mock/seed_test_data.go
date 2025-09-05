// Seed test data for development and testing
package mock

import (
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

// SeedTestData creates sample portfolios and positions for testing
func SeedTestData() error {
	db := database.GetDB()

	// Create a test user first
	// Hash the password properly
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return err
	}

	user := models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
		Password:  string(hashedPassword),
		Role:      "analyst",
		IsActive:  true,
	}

	// Check if user already exists
	var existingUser models.User
	if err := db.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
		user = existingUser // Use existing user
		log.Printf("Using existing test user: %s", user.Email)
	} else {
		// Create new user
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Warning: Could not create test user: %v", err)
			// Try to find any existing user
			if err := db.First(&user).Error; err != nil {
				return err
			}
		} else {
			log.Printf("Created test user: %s", user.Email)
		}
	}

	// Create sample portfolios
	portfolios := []models.Portfolio{
		{
			ID:          uuid.New(),
			UserID:      user.ID,
			Name:        "Tech Growth Portfolio",
			Description: "High-growth technology stocks",
			TotalValue:  decimal.NewFromFloat(500000),
			Currency:    "USD",
		},
		{
			ID:          uuid.New(),
			UserID:      user.ID,
			Name:        "Diversified Portfolio",
			Description: "Mixed asset allocation portfolio",
			TotalValue:  decimal.NewFromFloat(750000),
			Currency:    "USD",
		},
		{
			ID:          uuid.New(),
			UserID:      user.ID,
			Name:        "Crypto Portfolio",
			Description: "Cryptocurrency investments",
			TotalValue:  decimal.NewFromFloat(250000),
			Currency:    "USD",
		},
	}

	// Create portfolios if they don't exist
	for i, portfolio := range portfolios {
		var existing models.Portfolio
		if err := db.Where("name = ? AND user_id = ?", portfolio.Name, user.ID).First(&existing).Error; err != nil {
			if err := db.Create(&portfolio).Error; err != nil {
				log.Printf("Warning: Could not create portfolio %s: %v", portfolio.Name, err)
				continue
			}
			log.Printf("Created portfolio: %s", portfolio.Name)
		} else {
			portfolios[i] = existing // Use existing portfolio
			log.Printf("Using existing portfolio: %s", existing.Name)
		}
	}

	// Create positions for each portfolio
	positions := [][]models.Position{
		// Tech Growth Portfolio positions
		{
			{
				ID:           uuid.New(),
				PortfolioID:  portfolios[0].ID,
				Symbol:       "AAPL",
				Quantity:     decimal.NewFromFloat(1000),
				AveragePrice: decimal.NewFromFloat(150.00),
				CurrentPrice: decimal.NewFromFloat(155.00),
				MarketValue:  decimal.NewFromFloat(155000),
				PnL:          decimal.NewFromFloat(5000),
				PnLPercent:   decimal.NewFromFloat(3.33),
				Weight:       decimal.NewFromFloat(31.0),
				AssetType:    "STOCK",
				Liquidity:    "HIGH",
			},
			{
				ID:           uuid.New(),
				PortfolioID:  portfolios[0].ID,
				Symbol:       "GOOGL",
				Quantity:     decimal.NewFromFloat(100),
				AveragePrice: decimal.NewFromFloat(2800.00),
				CurrentPrice: decimal.NewFromFloat(2850.00),
				MarketValue:  decimal.NewFromFloat(285000),
				PnL:          decimal.NewFromFloat(5000),
				PnLPercent:   decimal.NewFromFloat(1.79),
				Weight:       decimal.NewFromFloat(57.0),
				AssetType:    "STOCK",
				Liquidity:    "HIGH",
			},
			{
				ID:           uuid.New(),
				PortfolioID:  portfolios[0].ID,
				Symbol:       "MSFT",
				Quantity:     decimal.NewFromFloat(200),
				AveragePrice: decimal.NewFromFloat(300.00),
				CurrentPrice: decimal.NewFromFloat(305.00),
				MarketValue:  decimal.NewFromFloat(61000),
				PnL:          decimal.NewFromFloat(1000),
				PnLPercent:   decimal.NewFromFloat(1.67),
				Weight:       decimal.NewFromFloat(12.2),
				AssetType:    "STOCK",
				Liquidity:    "HIGH",
			},
		},
		// Diversified Portfolio positions
		{
			{
				ID:           uuid.New(),
				PortfolioID:  portfolios[1].ID,
				Symbol:       "AAPL",
				Quantity:     decimal.NewFromFloat(500),
				AveragePrice: decimal.NewFromFloat(150.00),
				CurrentPrice: decimal.NewFromFloat(155.00),
				MarketValue:  decimal.NewFromFloat(77500),
				PnL:          decimal.NewFromFloat(2500),
				PnLPercent:   decimal.NewFromFloat(3.33),
				Weight:       decimal.NewFromFloat(10.33),
				AssetType:    "STOCK",
				Liquidity:    "HIGH",
			},
			{
				ID:           uuid.New(),
				PortfolioID:  portfolios[1].ID,
				Symbol:       "JPM",
				Quantity:     decimal.NewFromFloat(1000),
				AveragePrice: decimal.NewFromFloat(140.00),
				CurrentPrice: decimal.NewFromFloat(142.00),
				MarketValue:  decimal.NewFromFloat(142000),
				PnL:          decimal.NewFromFloat(2000),
				PnLPercent:   decimal.NewFromFloat(1.43),
				Weight:       decimal.NewFromFloat(18.93),
				AssetType:    "STOCK",
				Liquidity:    "HIGH",
			},
			{
				ID:           uuid.New(),
				PortfolioID:  portfolios[1].ID,
				Symbol:       "GOLD",
				Quantity:     decimal.NewFromFloat(100),
				AveragePrice: decimal.NewFromFloat(1800.00),
				CurrentPrice: decimal.NewFromFloat(1950.00),
				MarketValue:  decimal.NewFromFloat(195000),
				PnL:          decimal.NewFromFloat(15000),
				PnLPercent:   decimal.NewFromFloat(8.33),
				Weight:       decimal.NewFromFloat(26.0),
				AssetType:    "COMMODITY",
				Liquidity:    "MEDIUM",
			},
			{
				ID:           uuid.New(),
				PortfolioID:  portfolios[1].ID,
				Symbol:       "BTC",
				Quantity:     decimal.NewFromFloat(2.0),
				AveragePrice: decimal.NewFromFloat(40000.00),
				CurrentPrice: decimal.NewFromFloat(45000.00),
				MarketValue:  decimal.NewFromFloat(90000),
				PnL:          decimal.NewFromFloat(10000),
				PnLPercent:   decimal.NewFromFloat(12.5),
				Weight:       decimal.NewFromFloat(12.0),
				AssetType:    "CRYPTO",
				Liquidity:    "MEDIUM",
			},
		},
		// Crypto Portfolio positions
		{
			{
				ID:           uuid.New(),
				PortfolioID:  portfolios[2].ID,
				Symbol:       "BTC",
				Quantity:     decimal.NewFromFloat(3.5),
				AveragePrice: decimal.NewFromFloat(35000.00),
				CurrentPrice: decimal.NewFromFloat(45000.00),
				MarketValue:  decimal.NewFromFloat(157500),
				PnL:          decimal.NewFromFloat(35000),
				PnLPercent:   decimal.NewFromFloat(28.57),
				Weight:       decimal.NewFromFloat(63.0),
				AssetType:    "CRYPTO",
				Liquidity:    "MEDIUM",
			},
			{
				ID:           uuid.New(),
				PortfolioID:  portfolios[2].ID,
				Symbol:       "ETH",
				Quantity:     decimal.NewFromFloat(30),
				AveragePrice: decimal.NewFromFloat(2000.00),
				CurrentPrice: decimal.NewFromFloat(2500.00),
				MarketValue:  decimal.NewFromFloat(75000),
				PnL:          decimal.NewFromFloat(15000),
				PnLPercent:   decimal.NewFromFloat(25.0),
				Weight:       decimal.NewFromFloat(30.0),
				AssetType:    "CRYPTO",
				Liquidity:    "MEDIUM",
			},
			{
				ID:           uuid.New(),
				PortfolioID:  portfolios[2].ID,
				Symbol:       "ADA",
				Quantity:     decimal.NewFromFloat(50000),
				AveragePrice: decimal.NewFromFloat(0.30),
				CurrentPrice: decimal.NewFromFloat(0.35),
				MarketValue:  decimal.NewFromFloat(17500),
				PnL:          decimal.NewFromFloat(2500),
				PnLPercent:   decimal.NewFromFloat(16.67),
				Weight:       decimal.NewFromFloat(7.0),
				AssetType:    "CRYPTO",
				Liquidity:    "LOW",
			},
		},
	}

	// Create positions for each portfolio
	for i, portfolioPositions := range positions {
		for _, position := range portfolioPositions {
			var existing models.Position
			if err := db.Where("portfolio_id = ? AND symbol = ?", position.PortfolioID, position.Symbol).First(&existing).Error; err != nil {
				position.UpdatedAt = time.Now()
				if err := db.Create(&position).Error; err != nil {
					log.Printf("Warning: Could not create position %s for portfolio %s: %v", position.Symbol, portfolios[i].Name, err)
					continue
				}
				log.Printf("Created position %s for portfolio %s", position.Symbol, portfolios[i].Name)
			} else {
				log.Printf("Using existing position %s for portfolio %s", existing.Symbol, portfolios[i].Name)
			}
		}
	}

	log.Printf("Test data seeding completed successfully")
	return nil
}
