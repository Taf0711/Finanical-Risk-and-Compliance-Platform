package marketdata

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

// Service manages multiple market data providers
type Service struct {
	providers   map[string]MarketDataProvider
	primary     string
	fallbacks   []string
	cache       *redis.Client
	cacheTTL    time.Duration
	rateLimiter map[string]*RateLimiter
	mu          sync.RWMutex
}

// RateLimiter manages rate limiting for providers
type RateLimiter struct {
	tokens     chan struct{}
	refillRate time.Duration
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, refillRate time.Duration) *RateLimiter {
	rl := &RateLimiter{
		tokens:     make(chan struct{}, limit),
		refillRate: refillRate,
		lastRefill: time.Now(),
	}

	// Fill initial tokens
	for i := 0; i < limit; i++ {
		rl.tokens <- struct{}{}
	}

	// Start refill goroutine only if refillRate is positive
	if refillRate > 0 {
		go rl.refillTokens()
	}

	return rl
}

// Allow checks if a request is allowed
func (rl *RateLimiter) Allow() bool {
	select {
	case <-rl.tokens:
		return true
	default:
		return false
	}
}

// refillTokens periodically refills the token bucket
func (rl *RateLimiter) refillTokens() {
	ticker := time.NewTicker(rl.refillRate)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
			// Bucket is full
		}
	}
}

// ServiceConfig holds configuration for the market data service
type ServiceConfig struct {
	PrimaryProvider   string
	FallbackProviders []string
	CacheTTL          time.Duration
	RateLimits        map[string]RateLimitConfig
}

// RateLimitConfig holds rate limiting configuration for a provider
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstLimit        int
}

// NewService creates a new market data service
func NewService(config *ServiceConfig, redisClient *redis.Client) *Service {
	service := &Service{
		providers:   make(map[string]MarketDataProvider),
		primary:     config.PrimaryProvider,
		fallbacks:   config.FallbackProviders,
		cache:       redisClient,
		cacheTTL:    config.CacheTTL,
		rateLimiter: make(map[string]*RateLimiter),
	}

	// Initialize rate limiters
	for provider, config := range config.RateLimits {
		if config.RequestsPerMinute > 0 && config.BurstLimit > 0 {
			refillRate := time.Duration(60*1000/config.RequestsPerMinute) * time.Millisecond
			if refillRate > 0 {
				service.rateLimiter[provider] = NewRateLimiter(config.BurstLimit, refillRate)
			}
		}
	}

	return service
}

// RegisterProvider registers a market data provider
func (s *Service) RegisterProvider(name string, provider MarketDataProvider) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.providers[name] = provider
	log.Printf("Registered market data provider: %s", name)
}

// GetHistoricalData fetches historical data with fallback support
func (s *Service) GetHistoricalData(symbol string, period string) (*HistoricalData, error) {
	// Try cache first
	if cached := s.getCachedHistoricalData(symbol, period); cached != nil {
		return cached, nil
	}

	// Try primary provider first
	if data, err := s.getHistoricalDataFromProvider(s.primary, symbol, period); err == nil {
		s.cacheHistoricalData(symbol, period, data)
		return data, nil
	}

	// Try fallback providers
	for _, providerName := range s.fallbacks {
		if data, err := s.getHistoricalDataFromProvider(providerName, symbol, period); err == nil {
			s.cacheHistoricalData(symbol, period, data)
			return data, nil
		}
	}

	return nil, &MarketDataError{
		Provider: "Service",
		Code:     ErrCodeNoData,
		Message:  "All providers failed to fetch historical data",
		Symbol:   symbol,
	}
}

// GetRealtimeQuote fetches real-time quote with fallback support
func (s *Service) GetRealtimeQuote(symbol string) (*Quote, error) {
	// Try cache first (with shorter TTL for real-time data)
	if cached := s.getCachedQuote(symbol); cached != nil {
		return cached, nil
	}

	// Try primary provider first
	if quote, err := s.getQuoteFromProvider(s.primary, symbol); err == nil {
		s.cacheQuote(symbol, quote)
		return quote, nil
	}

	// Try fallback providers
	for _, providerName := range s.fallbacks {
		if quote, err := s.getQuoteFromProvider(providerName, symbol); err == nil {
			s.cacheQuote(symbol, quote)
			return quote, nil
		}
	}

	return nil, &MarketDataError{
		Provider: "Service",
		Code:     ErrCodeNoData,
		Message:  "All providers failed to fetch quote",
		Symbol:   symbol,
	}
}

// GetMultipleQuotes fetches multiple quotes efficiently
func (s *Service) GetMultipleQuotes(symbols []string) (map[string]*Quote, error) {
	if len(symbols) == 0 {
		return make(map[string]*Quote), nil
	}

	// Check cache for all symbols
	quotes := make(map[string]*Quote)
	uncachedSymbols := make([]string, 0)

	for _, symbol := range symbols {
		if cached := s.getCachedQuote(symbol); cached != nil {
			quotes[symbol] = cached
		} else {
			uncachedSymbols = append(uncachedSymbols, symbol)
		}
	}

	// If all symbols were cached, return
	if len(uncachedSymbols) == 0 {
		return quotes, nil
	}

	// Try to fetch uncached symbols from primary provider
	if newQuotes, err := s.getMultipleQuotesFromProvider(s.primary, uncachedSymbols); err == nil {
		for symbol, quote := range newQuotes {
			quotes[symbol] = quote
			s.cacheQuote(symbol, quote)
		}
		return quotes, nil
	}

	// Try fallback providers for remaining symbols
	for _, providerName := range s.fallbacks {
		if newQuotes, err := s.getMultipleQuotesFromProvider(providerName, uncachedSymbols); err == nil {
			for symbol, quote := range newQuotes {
				if _, exists := quotes[symbol]; !exists {
					quotes[symbol] = quote
					s.cacheQuote(symbol, quote)
				}
			}
		}
	}

	return quotes, nil
}

// GetCompanyInfo fetches company information
func (s *Service) GetCompanyInfo(symbol string) (*CompanyInfo, error) {
	// Try cache first
	if cached := s.getCachedCompanyInfo(symbol); cached != nil {
		return cached, nil
	}

	// Try primary provider first
	if info, err := s.getCompanyInfoFromProvider(s.primary, symbol); err == nil {
		s.cacheCompanyInfo(symbol, info)
		return info, nil
	}

	// Try fallback providers
	for _, providerName := range s.fallbacks {
		if info, err := s.getCompanyInfoFromProvider(providerName, symbol); err == nil {
			s.cacheCompanyInfo(symbol, info)
			return info, nil
		}
	}

	return nil, &MarketDataError{
		Provider: "Service",
		Code:     ErrCodeNoData,
		Message:  "All providers failed to fetch company info",
		Symbol:   symbol,
	}
}

// GetProviderStatus returns the status of all providers
func (s *Service) GetProviderStatus() map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := make(map[string]bool)
	for name := range s.providers {
		// Simple health check - try to validate a known symbol
		provider := s.providers[name]
		status[name] = provider.ValidateSymbol("AAPL")
	}

	return status
}

// Helper methods for provider interaction with rate limiting

func (s *Service) getHistoricalDataFromProvider(providerName, symbol, period string) (*HistoricalData, error) {
	s.mu.RLock()
	provider, exists := s.providers[providerName]
	s.mu.RUnlock()

	if !exists {
		return nil, &MarketDataError{
			Provider: providerName,
			Code:     "PROVIDER_NOT_FOUND",
			Message:  "Provider not registered",
		}
	}

	// Check rate limit
	if rl, exists := s.rateLimiter[providerName]; exists && !rl.Allow() {
		return nil, &MarketDataError{
			Provider: providerName,
			Code:     ErrCodeRateLimit,
			Message:  "Rate limit exceeded",
		}
	}

	return provider.GetHistoricalData(symbol, period)
}

func (s *Service) getQuoteFromProvider(providerName, symbol string) (*Quote, error) {
	s.mu.RLock()
	provider, exists := s.providers[providerName]
	s.mu.RUnlock()

	if !exists {
		return nil, &MarketDataError{
			Provider: providerName,
			Code:     "PROVIDER_NOT_FOUND",
			Message:  "Provider not registered",
		}
	}

	// Check rate limit
	if rl, exists := s.rateLimiter[providerName]; exists && !rl.Allow() {
		return nil, &MarketDataError{
			Provider: providerName,
			Code:     ErrCodeRateLimit,
			Message:  "Rate limit exceeded",
		}
	}

	return provider.GetRealtimeQuote(symbol)
}

func (s *Service) getMultipleQuotesFromProvider(providerName string, symbols []string) (map[string]*Quote, error) {
	s.mu.RLock()
	provider, exists := s.providers[providerName]
	s.mu.RUnlock()

	if !exists {
		return nil, &MarketDataError{
			Provider: providerName,
			Code:     "PROVIDER_NOT_FOUND",
			Message:  "Provider not registered",
		}
	}

	// Check rate limit
	if rl, exists := s.rateLimiter[providerName]; exists && !rl.Allow() {
		return nil, &MarketDataError{
			Provider: providerName,
			Code:     ErrCodeRateLimit,
			Message:  "Rate limit exceeded",
		}
	}

	return provider.GetMultipleQuotes(symbols)
}

func (s *Service) getCompanyInfoFromProvider(providerName, symbol string) (*CompanyInfo, error) {
	s.mu.RLock()
	provider, exists := s.providers[providerName]
	s.mu.RUnlock()

	if !exists {
		return nil, &MarketDataError{
			Provider: providerName,
			Code:     "PROVIDER_NOT_FOUND",
			Message:  "Provider not registered",
		}
	}

	// Check rate limit
	if rl, exists := s.rateLimiter[providerName]; exists && !rl.Allow() {
		return nil, &MarketDataError{
			Provider: providerName,
			Code:     ErrCodeRateLimit,
			Message:  "Rate limit exceeded",
		}
	}

	return provider.GetCompanyInfo(symbol)
}

// Cache methods

func (s *Service) getCachedHistoricalData(symbol, period string) *HistoricalData {
	if s.cache == nil {
		return nil
	}

	key := fmt.Sprintf("historical:%s:%s", symbol, period)
	val, err := s.cache.Get(context.Background(), key).Result()
	if err != nil {
		return nil
	}

	var data HistoricalData
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return nil
	}

	return &data
}

func (s *Service) cacheHistoricalData(symbol, period string, data *HistoricalData) {
	if s.cache == nil {
		return
	}

	key := fmt.Sprintf("historical:%s:%s", symbol, period)
	val, err := json.Marshal(data)
	if err != nil {
		return
	}

	// Cache historical data for longer (1 hour)
	s.cache.Set(context.Background(), key, val, time.Hour)
}

func (s *Service) getCachedQuote(symbol string) *Quote {
	if s.cache == nil {
		return nil
	}

	key := fmt.Sprintf("quote:%s", symbol)
	val, err := s.cache.Get(context.Background(), key).Result()
	if err != nil {
		return nil
	}

	var quote Quote
	if err := json.Unmarshal([]byte(val), &quote); err != nil {
		return nil
	}

	return &quote
}

func (s *Service) cacheQuote(symbol string, quote *Quote) {
	if s.cache == nil {
		return
	}

	key := fmt.Sprintf("quote:%s", symbol)
	val, err := json.Marshal(quote)
	if err != nil {
		return
	}

	// Cache quotes for shorter time (5 minutes for real-time data)
	s.cache.Set(context.Background(), key, val, 5*time.Minute)
}

func (s *Service) getCachedCompanyInfo(symbol string) *CompanyInfo {
	if s.cache == nil {
		return nil
	}

	key := fmt.Sprintf("company:%s", symbol)
	val, err := s.cache.Get(context.Background(), key).Result()
	if err != nil {
		return nil
	}

	var info CompanyInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return nil
	}

	return &info
}

func (s *Service) cacheCompanyInfo(symbol string, info *CompanyInfo) {
	if s.cache == nil {
		return
	}

	key := fmt.Sprintf("company:%s", symbol)
	val, err := json.Marshal(info)
	if err != nil {
		return
	}

	// Cache company info for longer (24 hours)
	s.cache.Set(context.Background(), key, val, 24*time.Hour)
}
