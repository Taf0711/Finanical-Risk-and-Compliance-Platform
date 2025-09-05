package marketdata

import (
	"fmt"
	"time"
)

// MarketDataProvider interface for all market data providers
type MarketDataProvider interface {
	GetHistoricalData(symbol string, period string) (*HistoricalData, error)
	GetRealtimeQuote(symbol string) (*Quote, error)
	GetMultipleQuotes(symbols []string) (map[string]*Quote, error)
	GetCompanyInfo(symbol string) (*CompanyInfo, error)
	ValidateSymbol(symbol string) bool
	GetProviderName() string
}

// HistoricalData represents historical price data
type HistoricalData struct {
	Symbol     string        `json:"symbol"`
	Period     string        `json:"period"`
	DataPoints []PricePoint  `json:"data_points"`
	Metadata   *DataMetadata `json:"metadata"`
}

// PricePoint represents a single price data point
type PricePoint struct {
	Date     time.Time `json:"date"`
	Open     float64   `json:"open"`
	High     float64   `json:"high"`
	Low      float64   `json:"low"`
	Close    float64   `json:"close"`
	Volume   int64     `json:"volume"`
	AdjClose float64   `json:"adj_close"`
}

// Quote represents real-time quote data
type Quote struct {
	Symbol           string    `json:"symbol"`
	Price            float64   `json:"price"`
	Change           float64   `json:"change"`
	ChangePercent    float64   `json:"change_percent"`
	Volume           int64     `json:"volume"`
	MarketCap        int64     `json:"market_cap,omitempty"`
	PreviousClose    float64   `json:"previous_close"`
	Open             float64   `json:"open"`
	DayHigh          float64   `json:"day_high"`
	DayLow           float64   `json:"day_low"`
	FiftyTwoWeekHigh float64   `json:"fifty_two_week_high,omitempty"`
	FiftyTwoWeekLow  float64   `json:"fifty_two_week_low,omitempty"`
	Timestamp        time.Time `json:"timestamp"`
	IsMarketOpen     bool      `json:"is_market_open"`
	Provider         string    `json:"provider"`
}

// CompanyInfo represents company information
type CompanyInfo struct {
	Symbol      string `json:"symbol"`
	Name        string `json:"name"`
	Exchange    string `json:"exchange"`
	Currency    string `json:"currency"`
	Country     string `json:"country"`
	Sector      string `json:"sector,omitempty"`
	Industry    string `json:"industry,omitempty"`
	Description string `json:"description,omitempty"`
	Website     string `json:"website,omitempty"`
	MarketCap   int64  `json:"market_cap,omitempty"`
	Provider    string `json:"provider"`
}

// DataMetadata contains metadata about the data
type DataMetadata struct {
	Provider    string    `json:"provider"`
	LastUpdated time.Time `json:"last_updated"`
	TimeZone    string    `json:"time_zone"`
	Currency    string    `json:"currency"`
	Exchange    string    `json:"exchange"`
}

// MarketDataError represents errors from market data providers
type MarketDataError struct {
	Provider string `json:"provider"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Symbol   string `json:"symbol,omitempty"`
}

func (e *MarketDataError) Error() string {
	if e.Symbol != "" {
		return fmt.Sprintf("[%s] %s: %s (symbol: %s)", e.Provider, e.Code, e.Message, e.Symbol)
	}
	return fmt.Sprintf("[%s] %s: %s", e.Provider, e.Code, e.Message)
}

// Common error codes
const (
	ErrCodeInvalidSymbol = "INVALID_SYMBOL"
	ErrCodeRateLimit     = "RATE_LIMIT"
	ErrCodeProviderDown  = "PROVIDER_DOWN"
	ErrCodeNoData        = "NO_DATA"
	ErrCodeInvalidPeriod = "INVALID_PERIOD"
	ErrCodeNetworkError  = "NETWORK_ERROR"
)
