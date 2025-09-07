package providers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Taf0711/financial-risk-monitor/internal/marketdata"
)

// AlphaVantageProvider implements MarketDataProvider using Alpha Vantage
type AlphaVantageProvider struct {
	client  *http.Client
	baseURL string
	apiKey  string
}

func NewAlphaVantageProvider(apiKey string) *AlphaVantageProvider {
	return &AlphaVantageProvider{
		client:  &http.Client{Timeout: 15 * time.Second},
		baseURL: "https://www.alphavantage.co/query",
		apiKey:  apiKey,
	}
}

func (a *AlphaVantageProvider) GetProviderName() string {
	return "AlphaVantage"
}

func (a *AlphaVantageProvider) makeRequest(params url.Values) (*http.Response, error) {
	params.Set("apikey", a.apiKey)
	reqURL := a.baseURL + "?" + params.Encode()
	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return a.client.Do(req)
}

// GetRealtimeQuote uses GLOBAL_QUOTE
func (a *AlphaVantageProvider) GetRealtimeQuote(symbol string) (*marketdata.Quote, error) {
	params := url.Values{}
	params.Set("function", "GLOBAL_QUOTE")
	params.Set("symbol", symbol)

	resp, err := a.makeRequest(params)
	if err != nil {
		return nil, &marketdata.MarketDataError{Provider: a.GetProviderName(), Code: marketdata.ErrCodeNetworkError, Message: err.Error(), Symbol: symbol}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &marketdata.MarketDataError{Provider: a.GetProviderName(), Code: marketdata.ErrCodeProviderDown, Message: string(body), Symbol: symbol}
	}

	var body struct {
		Quote map[string]string `json:"Global Quote"`
	}
	raw, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, &marketdata.MarketDataError{Provider: a.GetProviderName(), Code: marketdata.ErrCodeNoData, Message: "failed to parse quote", Symbol: symbol}
	}
	if len(body.Quote) == 0 {
		return nil, &marketdata.MarketDataError{Provider: a.GetProviderName(), Code: marketdata.ErrCodeNoData, Message: "no quote returned", Symbol: symbol}
	}

	priceStr := body.Quote["05. price"]
	var price float64
	fmt.Sscanf(priceStr, "%f", &price)

	return &marketdata.Quote{
		Symbol:    symbol,
		Price:     price,
		Timestamp: time.Now(),
		Provider:  a.GetProviderName(),
	}, nil
}

// GetHistoricalData uses TIME_SERIES_DAILY
func (a *AlphaVantageProvider) GetHistoricalData(symbol string, period string) (*marketdata.HistoricalData, error) {
	params := url.Values{}
	params.Set("function", "TIME_SERIES_DAILY_ADJUSTED")
	params.Set("symbol", symbol)
	params.Set("outputsize", "compact")

	resp, err := a.makeRequest(params)
	if err != nil {
		return nil, &marketdata.MarketDataError{Provider: a.GetProviderName(), Code: marketdata.ErrCodeNetworkError, Message: err.Error(), Symbol: symbol}
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, &marketdata.MarketDataError{Provider: a.GetProviderName(), Code: marketdata.ErrCodeProviderDown, Message: string(body), Symbol: symbol}
	}

	raw, _ := io.ReadAll(resp.Body)
	var parsed map[string]interface{}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, &marketdata.MarketDataError{Provider: a.GetProviderName(), Code: marketdata.ErrCodeNoData, Message: "failed to parse historical", Symbol: symbol}
	}

	// look for Time Series (Daily)
	var series map[string]interface{}
	for k, v := range parsed {
		if strings.HasPrefix(k, "Time Series") {
			series = v.(map[string]interface{})
			break
		}
	}
	if series == nil || len(series) == 0 {
		return nil, &marketdata.MarketDataError{Provider: a.GetProviderName(), Code: marketdata.ErrCodeNoData, Message: "no historical data", Symbol: symbol}
	}

	var points []marketdata.PricePoint
	for dateStr, v := range series {
		day := v.(map[string]interface{})
		var open, high, low, close float64
		fmt.Sscanf(day["1. open"].(string), "%f", &open)
		fmt.Sscanf(day["2. high"].(string), "%f", &high)
		fmt.Sscanf(day["3. low"].(string), "%f", &low)
		fmt.Sscanf(day["4. close"].(string), "%f", &close)

		t, _ := time.Parse("2006-01-02", dateStr)
		points = append(points, marketdata.PricePoint{Date: t, Open: open, High: high, Low: low, Close: close, Volume: 0})
	}

	return &marketdata.HistoricalData{Symbol: symbol, Period: period, DataPoints: points, Metadata: &marketdata.DataMetadata{Provider: a.GetProviderName(), LastUpdated: time.Now()}}, nil
}

func (a *AlphaVantageProvider) GetMultipleQuotes(symbols []string) (map[string]*marketdata.Quote, error) {
	quotes := make(map[string]*marketdata.Quote)
	for _, s := range symbols {
		q, err := a.GetRealtimeQuote(s)
		if err != nil {
			continue
		}
		quotes[s] = q
	}
	return quotes, nil
}

func (a *AlphaVantageProvider) GetCompanyInfo(symbol string) (*marketdata.CompanyInfo, error) {
	// AlphaVantage has limited company info; return basic
	return &marketdata.CompanyInfo{Symbol: symbol, Name: symbol, Exchange: "US", Currency: "USD", Provider: a.GetProviderName()}, nil
}

func (a *AlphaVantageProvider) ValidateSymbol(symbol string) bool {
	_, err := a.GetRealtimeQuote(symbol)
	return err == nil
}

