package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Taf0711/financial-risk-monitor/internal/marketdata"
	"github.com/Taf0711/financial-risk-monitor/internal/risk/calculator"
)

// AlpacaProvider implements the MarketDataProvider interface for Alpaca Markets
type AlpacaProvider struct {
	client    *http.Client
	baseURL   string
	apiKey    string
	apiSecret string
}

// NewAlpacaProvider creates a new AlpacaProvider
func NewAlpacaProvider(apiKey, apiSecret string) *AlpacaProvider {
	// Use data API for market data (free tier)
	baseURL := "https://data.alpaca.markets"

	return &AlpacaProvider{
		client:    &http.Client{Timeout: 10 * time.Second},
		baseURL:   baseURL,
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

// GetProviderName returns the name of the provider
func (a *AlpacaProvider) GetProviderName() string {
	return "Alpaca Markets"
}

// makeRequest makes an authenticated request to Alpaca API
func (a *AlpacaProvider) makeRequest(endpoint string) (*http.Response, error) {
	req, err := http.NewRequest("GET", a.baseURL+endpoint, nil)
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	req.Header.Set("APCA-API-KEY-ID", a.apiKey)
	req.Header.Set("APCA-API-SECRET-KEY", a.apiSecret)
	req.Header.Set("Content-Type", "application/json")

	return a.client.Do(req)
}

// GetHistoricalData fetches historical data for a given symbol and period
func (a *AlpacaProvider) GetHistoricalData(symbol string, period string) (*marketdata.HistoricalData, error) {
	timeframe, start, end := a.parsePeriod(period)
	// Use Alpaca data API for market data
	endpoint := fmt.Sprintf("/v2/stocks/%s/bars?timeframe=%s&start=%s&end=%s&limit=1000",
		symbol, timeframe, start, end)

	resp, err := a.makeRequest(endpoint)
	if err != nil {
		return nil, &marketdata.MarketDataError{
			Provider: a.GetProviderName(),
			Code:     marketdata.ErrCodeNetworkError,
			Message:  "Failed to fetch historical data",
			Symbol:   symbol,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &marketdata.MarketDataError{
			Provider: a.GetProviderName(),
			Code:     marketdata.ErrCodeProviderDown,
			Message:  fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(bodyBytes)),
			Symbol:   symbol,
		}
	}

	var alpacaResponse AlpacaHistoricalResponse
	if err := json.NewDecoder(resp.Body).Decode(&alpacaResponse); err != nil {
		return nil, &marketdata.MarketDataError{
			Provider: a.GetProviderName(),
			Code:     marketdata.ErrCodeNoData,
			Message:  "Failed to parse historical data response",
			Symbol:   symbol,
		}
	}

	if len(alpacaResponse.Bars) == 0 {
		return nil, &marketdata.MarketDataError{
			Provider: a.GetProviderName(),
			Code:     marketdata.ErrCodeNoData,
			Message:  "No historical data found",
			Symbol:   symbol,
		}
	}

	return a.parseHistoricalData(symbol, period, &alpacaResponse)
}

// GetRealtimeQuote fetches a real-time quote for a given symbol
func (a *AlpacaProvider) GetRealtimeQuote(symbol string) (*marketdata.Quote, error) {
	// Try trades endpoint first for actual last trade price
	tradeEndpoint := fmt.Sprintf("/v2/stocks/%s/trades/latest", symbol)

	resp, err := a.makeRequest(tradeEndpoint)
	if err == nil && resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()

		var tradeResponse AlpacaTradeResponse
		if json.NewDecoder(resp.Body).Decode(&tradeResponse) == nil && tradeResponse.Trade.Price > 0 {
			return a.parseTrade(symbol, &tradeResponse.Trade), nil
		}
	}
	if resp != nil {
		resp.Body.Close()
	}

	// Fallback to quotes endpoint if trades don't work
	quoteEndpoint := fmt.Sprintf("/v2/stocks/%s/quotes/latest", symbol)

	resp, err = a.makeRequest(quoteEndpoint)
	if err != nil {
		return nil, &marketdata.MarketDataError{
			Provider: a.GetProviderName(),
			Code:     marketdata.ErrCodeNetworkError,
			Message:  "Failed to fetch real-time quote",
			Symbol:   symbol,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, &marketdata.MarketDataError{
			Provider: a.GetProviderName(),
			Code:     marketdata.ErrCodeProviderDown,
			Message:  fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(bodyBytes)),
			Symbol:   symbol,
		}
	}

	var quoteResponse AlpacaQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&quoteResponse); err != nil {
		return nil, &marketdata.MarketDataError{
			Provider: a.GetProviderName(),
			Code:     marketdata.ErrCodeNoData,
			Message:  "Failed to parse real-time quote response",
			Symbol:   symbol,
		}
	}

	if quoteResponse.Quote.BidPrice == 0 && quoteResponse.Quote.AskPrice == 0 {
		return nil, &marketdata.MarketDataError{
			Provider: a.GetProviderName(),
			Code:     marketdata.ErrCodeNoData,
			Message:  "No real-time quote found",
			Symbol:   symbol,
		}
	}

	return a.parseQuote(symbol, &quoteResponse.Quote), nil
}

// GetMultipleQuotes fetches multiple quotes concurrently
func (a *AlpacaProvider) GetMultipleQuotes(symbols []string) (map[string]*marketdata.Quote, error) {
	quotes := make(map[string]*marketdata.Quote)

	// For simplicity, fetch quotes sequentially to avoid rate limits
	// In production, you might want to use Alpaca's batch endpoint or concurrent requests with rate limiting
	for _, symbol := range symbols {
		quote, err := a.GetRealtimeQuote(symbol)
		if err != nil {
			continue // Skip failed quotes, don't fail the entire request
		}
		quotes[symbol] = quote
	}

	return quotes, nil
}

// GetCompanyInfo fetches basic company information
func (a *AlpacaProvider) GetCompanyInfo(symbol string) (*marketdata.CompanyInfo, error) {
	// Alpaca doesn't have a direct company info endpoint in the free tier
	// We'll return basic info with the symbol
	return &marketdata.CompanyInfo{
		Symbol:   symbol,
		Name:     symbol, // Placeholder
		Exchange: "US",   // Alpaca covers US markets
		Currency: "USD",
		Country:  "US",
		Provider: a.GetProviderName(),
	}, nil
}

// ValidateSymbol checks if a symbol is valid by trying to get a quote
func (a *AlpacaProvider) ValidateSymbol(symbol string) bool {
	_, err := a.GetRealtimeQuote(symbol)
	return err == nil
}

// parsePeriod converts a human-readable period string to Alpaca API parameters
func (a *AlpacaProvider) parsePeriod(period string) (string, string, string) {
	now := time.Now()
	var start time.Time
	var timeframe string

	switch strings.ToLower(period) {
	case "1d":
		start = now.AddDate(0, 0, -1)
		timeframe = "1Min"
	case "5d":
		start = now.AddDate(0, 0, -5)
		timeframe = "5Min"
	case "1mo":
		start = now.AddDate(0, -1, 0)
		timeframe = "1Hour"
	case "3mo":
		start = now.AddDate(0, -3, 0)
		timeframe = "1Hour"
	case "6mo":
		start = now.AddDate(0, -6, 0)
		timeframe = "1Day"
	case "1y":
		start = now.AddDate(-1, 0, 0)
		timeframe = "1Day"
	case "2y":
		start = now.AddDate(-2, 0, 0)
		timeframe = "1Day"
	case "5y":
		start = now.AddDate(-5, 0, 0)
		timeframe = "1Day"
	default:
		// Default to 1 year
		start = now.AddDate(-1, 0, 0)
		timeframe = "1Day"
	}

	return timeframe, start.Format("2006-01-02T15:04:05Z"), now.Format("2006-01-02T15:04:05Z")
}

// parseHistoricalData converts Alpaca response to our format
func (a *AlpacaProvider) parseHistoricalData(symbol, period string, response *AlpacaHistoricalResponse) (*marketdata.HistoricalData, error) {
	var dataPoints []marketdata.PricePoint

	for _, bar := range response.Bars {
		timestamp, err := time.Parse(time.RFC3339, bar.Timestamp)
		if err != nil {
			continue // Skip invalid timestamps
		}

		dataPoints = append(dataPoints, marketdata.PricePoint{
			Date:     timestamp,
			Open:     bar.Open,
			High:     bar.High,
			Low:      bar.Low,
			Close:    bar.Close,
			Volume:   bar.Volume,
			AdjClose: bar.Close, // Alpaca doesn't provide adjusted close in bars
		})
	}

	return &marketdata.HistoricalData{
		Symbol:     symbol,
		Period:     period,
		DataPoints: dataPoints,
		Metadata: &marketdata.DataMetadata{
			Provider:    a.GetProviderName(),
			LastUpdated: time.Now(),
			TimeZone:    "America/New_York",
			Currency:    "USD",
			Exchange:    "US",
		},
	}, nil
}

// parseQuote converts Alpaca quote response to our format
func (a *AlpacaProvider) parseQuote(symbol string, alpacaQuote *AlpacaQuote) *marketdata.Quote {
	// Calculate mid price from bid/ask
	midPrice := (alpacaQuote.BidPrice + alpacaQuote.AskPrice) / 2

	timestamp, _ := time.Parse(time.RFC3339, alpacaQuote.Timestamp)

	return &marketdata.Quote{
		Symbol:        symbol,
		Price:         midPrice,
		Change:        0, // Not available in quote response
		ChangePercent: 0, // Not available in quote response
		Volume:        0, // Not available in quote response
		Open:          0, // Not available in quote response
		DayHigh:       0, // Not available in quote response
		DayLow:        0, // Not available in quote response
		PreviousClose: 0, // Not available in quote response
		Timestamp:     timestamp,
		IsMarketOpen:  true, // Assume market is open
		Provider:      a.GetProviderName(),
	}
}

// parseTrade converts Alpaca trade response to our format
func (a *AlpacaProvider) parseTrade(symbol string, alpacaTrade *AlpacaTrade) *marketdata.Quote {
	timestamp, _ := time.Parse(time.RFC3339, alpacaTrade.Timestamp)

	log.Printf("Alpaca trade data for %s: Price=%.2f, Size=%d, Timestamp=%s",
		symbol, alpacaTrade.Price, alpacaTrade.Size, alpacaTrade.Timestamp)

	return &marketdata.Quote{
		Symbol:        symbol,
		Price:         alpacaTrade.Price, // Use actual trade price
		Change:        0,                 // Will need to calculate from previous close
		ChangePercent: 0,                 // Will need to calculate from previous close
		Volume:        alpacaTrade.Size,  // Use trade size as volume indicator
		Open:          0,                 // Not available in trade response
		DayHigh:       0,                 // Not available in trade response
		DayLow:        0,                 // Not available in trade response
		PreviousClose: 0,                 // Not available in trade response
		Timestamp:     timestamp,
		IsMarketOpen:  true, // Assume market is open if we get trade data
		Provider:      a.GetProviderName(),
	}
}

// Alpaca API response structures
type AlpacaHistoricalResponse struct {
	Bars          []AlpacaBar `json:"bars"`
	Symbol        string      `json:"symbol"`
	NextPageToken string      `json:"next_page_token,omitempty"`
}

type AlpacaBar struct {
	Timestamp string  `json:"t"`
	Open      float64 `json:"o"`
	High      float64 `json:"h"`
	Low       float64 `json:"l"`
	Close     float64 `json:"c"`
	Volume    int64   `json:"v"`
}

type AlpacaQuoteResponse struct {
	Quote AlpacaQuote `json:"quote"`
}

type AlpacaQuote struct {
	Timestamp string  `json:"t"`
	BidPrice  float64 `json:"bp"`
	BidSize   int64   `json:"bs"`
	AskPrice  float64 `json:"ap"`
	AskSize   int64   `json:"as"`
}

type AlpacaTradeResponse struct {
	Trade AlpacaTrade `json:"trade"`
}

type AlpacaTrade struct {
	Timestamp string  `json:"t"`
	Price     float64 `json:"p"`
	Size      int64   `json:"s"`
}

// Implement risk calculator MarketDataProvider interface
func (a *AlpacaProvider) GetAverageDailyVolume(symbol string) float64 {
	// In a real implementation, this would fetch volume data from Alpaca
	// For now, return realistic estimates based on symbol
	volumes := map[string]float64{
		"AAPL":   75000000,
		"GOOGL":  30000000,
		"MSFT":   45000000,
		"AMZN":   35000000,
		"TSLA":   85000000,
		"JPM":    15000000,
		"BAC":    45000000,
		"GS":     3000000,
		"MS":     8000000,
		"WFC":    25000000,
		"BTC":    25000,
		"ETH":    150000,
		"GOLD":   200000,
		"SILVER": 500000,
		"OIL":    300000,
	}

	if vol, exists := volumes[symbol]; exists {
		return vol
	}
	return 1000000 // Default volume
}

func (a *AlpacaProvider) GetBidAskSpread(symbol string) float64 {
	// Return realistic spreads as percentages
	spreads := map[string]float64{
		"AAPL":   0.00005, // 0.005% for large cap stocks
		"GOOGL":  0.00008,
		"MSFT":   0.00006,
		"AMZN":   0.00007,
		"TSLA":   0.0001, // Higher spread for more volatile stocks
		"JPM":    0.0001,
		"BAC":    0.0002,
		"GS":     0.0003,
		"MS":     0.0002,
		"WFC":    0.0002,
		"BTC":    0.0005, // 0.05% for major crypto
		"ETH":    0.0008,
		"GOLD":   0.001, // 0.1% for commodities
		"SILVER": 0.002,
		"OIL":    0.0015,
	}

	if spread, exists := spreads[symbol]; exists {
		return spread
	}
	return 0.0005 // Default spread
}

func (a *AlpacaProvider) GetMarketDepth(symbol string) *calculator.MarketDepth {
	// In production, this would fetch real order book data
	// For now, return realistic mock order book
	basePrice := 100.0
	if symbol == "AAPL" {
		basePrice = 175.0
	} else if symbol == "GOOGL" {
		basePrice = 2800.0
	} else if symbol == "BTC" {
		basePrice = 45000.0
	}

	return &calculator.MarketDepth{
		BidLevels: []calculator.PriceLevel{
			{Price: basePrice - 0.05, Quantity: 5000, Orders: 25},
			{Price: basePrice - 0.10, Quantity: 8000, Orders: 40},
			{Price: basePrice - 0.15, Quantity: 12000, Orders: 60},
		},
		AskLevels: []calculator.PriceLevel{
			{Price: basePrice + 0.05, Quantity: 4500, Orders: 22},
			{Price: basePrice + 0.10, Quantity: 7500, Orders: 35},
			{Price: basePrice + 0.15, Quantity: 11000, Orders: 55},
		},
		Timestamp: time.Now(),
	}
}

func (a *AlpacaProvider) GetMarketCap(symbol string) float64 {
	// Return market caps in USD
	marketCaps := map[string]float64{
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
		"BTC":    900000000000,   // $900B
		"ETH":    450000000000,   // $450B
		"GOLD":   13000000000000, // $13T
		"SILVER": 1500000000000,  // $1.5T
		"OIL":    5000000000000,  // $5T
	}

	if cap, exists := marketCaps[symbol]; exists {
		return cap
	}
	return 5000000000 // $5B default
}
