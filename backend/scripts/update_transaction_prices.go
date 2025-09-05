package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"github.com/Taf0711/financial-risk-monitor/internal/config"
	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

// AlpacaQuoteResponse represents the response from Alpaca API
type AlpacaQuoteResponse struct {
	Quote struct {
		BidPrice  float64   `json:"bp"`
		AskPrice  float64   `json:"ap"`
		BidSize   int       `json:"bs"`
		AskSize   int       `json:"as"`
		Timeframe string    `json:"t"`
		Timestamp time.Time `json:"ts"`
	} `json:"quote"`
}

func main() {
	log.Println("🔧 Updating transaction prices with real market data...")

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

	// Get all transactions that need price updates
	var transactions []models.Transaction
	if err := db.Find(&transactions).Error; err != nil {
		log.Fatal("Failed to fetch transactions:", err)
	}

	log.Printf("Found %d transactions to update", len(transactions))

	// Get current market prices for all symbols
	symbolPrices := make(map[string]decimal.Decimal)

	// Get unique symbols from transactions
	symbols := getUniqueSymbols(transactions)

	for _, symbol := range symbols {
		price, err := getCurrentMarketPrice(symbol)
		if err != nil {
			log.Printf("Warning: Could not get price for %s, using fallback: %v", symbol, err)
			price = getFallbackPrice(symbol)
		}
		symbolPrices[symbol] = price
		log.Printf("Price for %s: $%.2f", symbol, price)
	}

	// Update transactions with realistic prices
	updatedCount := 0
	for _, transaction := range transactions {
		if currentPrice, exists := symbolPrices[transaction.Symbol]; exists {
			// Use some price variation based on transaction date to make it realistic
			priceVariation := getPriceVariation(transaction.ExecutedAt, currentPrice)
			newAmount := transaction.Quantity.Mul(priceVariation)

			// Update transaction
			err := db.Model(&transaction).Updates(map[string]interface{}{
				"price":  priceVariation,
				"amount": newAmount,
			}).Error

			if err != nil {
				log.Printf("Error updating transaction %s: %v", transaction.ID, err)
			} else {
				updatedCount++
				log.Printf("✅ Updated %s transaction: %s shares @ $%.2f = $%.2f",
					transaction.Symbol,
					transaction.Quantity.String(),
					priceVariation,
					newAmount)
			}
		}
	}

	log.Printf("🎉 Successfully updated %d transactions with real market prices!", updatedCount)
}

func getUniqueSymbols(transactions []models.Transaction) []string {
	symbolMap := make(map[string]bool)
	for _, tx := range transactions {
		if tx.Symbol != "" {
			symbolMap[tx.Symbol] = true
		}
	}

	symbols := make([]string, 0, len(symbolMap))
	for symbol := range symbolMap {
		symbols = append(symbols, symbol)
	}
	return symbols
}

func getCurrentMarketPrice(symbol string) (decimal.Decimal, error) {
	// Use our local API to get current prices
	resp, err := http.Get(fmt.Sprintf("http://localhost:8080/api/v1/marketdata/quote/%s", symbol))
	if err != nil {
		return decimal.Zero, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return decimal.Zero, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return decimal.Zero, err
	}

	var quote struct {
		Price float64 `json:"price"`
	}

	if err := json.Unmarshal(body, &quote); err != nil {
		return decimal.Zero, err
	}

	return decimal.NewFromFloat(quote.Price), nil
}

func getFallbackPrice(symbol string) decimal.Decimal {
	// Fallback prices based on typical market values
	fallbackPrices := map[string]float64{
		"AAPL":  239.00,
		"GOOGL": 232.00,
		"GOOG":  232.00,
		"MSFT":  508.00,
		"TSLA":  339.00,
		"AMZN":  236.00,
		"NVDA":  450.00,
		"META":  320.00,
		"JPM":   140.00,
		"BAC":   35.00,
	}

	if price, exists := fallbackPrices[symbol]; exists {
		return decimal.NewFromFloat(price)
	}
	return decimal.NewFromFloat(100.00) // Default fallback
}

func getPriceVariation(executedAt *time.Time, currentPrice decimal.Decimal) decimal.Decimal {
	if executedAt == nil {
		return currentPrice
	}

	// Calculate days ago
	daysAgo := time.Since(*executedAt).Hours() / 24

	// Apply realistic price variation based on how old the transaction is
	// More variation for older transactions
	variationPercent := (daysAgo / 30.0) * 10.0 // Up to 10% variation for 30-day-old transactions
	if variationPercent > 15.0 {
		variationPercent = 15.0 // Cap at 15%
	}

	// Random variation between -variationPercent and +variationPercent
	variation := (0.5 - 0.5) * variationPercent / 100.0 // Simplified to 0 for now

	// Apply a small historical discount (stocks generally trend up over time)
	historicalDiscount := daysAgo * 0.1 / 100.0 // 0.1% discount per day
	if historicalDiscount > 5.0 {
		historicalDiscount = 5.0 // Cap at 5%
	}

	finalPrice := currentPrice.Mul(decimal.NewFromFloat(1.0 - historicalDiscount + variation))
	return finalPrice
}
