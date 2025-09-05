package handlers

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/Taf0711/financial-risk-monitor/internal/marketdata"
)

// MarketDataHandler handles market data API requests
type MarketDataHandler struct {
	service *marketdata.Service
}

// NewMarketDataHandler creates a new market data handler
func NewMarketDataHandler(service *marketdata.Service) *MarketDataHandler {
	return &MarketDataHandler{
		service: service,
	}
}

// GetHistoricalData handles GET /api/v1/marketdata/historical/{symbol}
func (h *MarketDataHandler) GetHistoricalData(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	period := c.Query("period", "1y") // Default to 1 year

	data, err := h.service.GetHistoricalData(symbol, period)
	if err != nil {
		if mdErr, ok := err.(*marketdata.MarketDataError); ok {
			switch mdErr.Code {
			case marketdata.ErrCodeInvalidSymbol:
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			case marketdata.ErrCodeNoData:
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			case marketdata.ErrCodeRateLimit:
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			}
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch historical data",
		})
	}

	return c.JSON(data)
}

// GetRealtimeQuote handles GET /api/v1/marketdata/quote/{symbol}
func (h *MarketDataHandler) GetRealtimeQuote(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	quote, err := h.service.GetRealtimeQuote(symbol)
	if err != nil {
		if mdErr, ok := err.(*marketdata.MarketDataError); ok {
			switch mdErr.Code {
			case marketdata.ErrCodeInvalidSymbol:
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			case marketdata.ErrCodeNoData:
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			case marketdata.ErrCodeRateLimit:
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			}
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch quote",
		})
	}

	return c.JSON(quote)
}

// GetMultipleQuotes handles GET /api/v1/marketdata/quotes?symbols=AAPL,GOOGL,MSFT
func (h *MarketDataHandler) GetMultipleQuotes(c *fiber.Ctx) error {
	symbolsParam := c.Query("symbols")
	if symbolsParam == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbols parameter is required",
		})
	}

	symbols := strings.Split(symbolsParam, ",")
	if len(symbols) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "At least one symbol is required",
		})
	}

	// Limit number of symbols to prevent abuse
	if len(symbols) > 50 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Maximum 50 symbols allowed per request",
		})
	}

	// Clean symbols (trim whitespace)
	for i, symbol := range symbols {
		symbols[i] = strings.TrimSpace(symbol)
	}

	quotes, err := h.service.GetMultipleQuotes(symbols)
	if err != nil {
		if mdErr, ok := err.(*marketdata.MarketDataError); ok {
			switch mdErr.Code {
			case marketdata.ErrCodeInvalidSymbol:
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			case marketdata.ErrCodeRateLimit:
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			}
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch quotes",
		})
	}

	return c.JSON(fiber.Map{
		"quotes": quotes,
		"count":  len(quotes),
	})
}

// GetCompanyInfo handles GET /api/v1/marketdata/company/{symbol}
func (h *MarketDataHandler) GetCompanyInfo(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	info, err := h.service.GetCompanyInfo(symbol)
	if err != nil {
		if mdErr, ok := err.(*marketdata.MarketDataError); ok {
			switch mdErr.Code {
			case marketdata.ErrCodeInvalidSymbol:
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			case marketdata.ErrCodeNoData:
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			case marketdata.ErrCodeRateLimit:
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			default:
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": mdErr.Message,
					"code":  mdErr.Code,
				})
			}
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch company info",
		})
	}

	return c.JSON(info)
}

// GetProviderStatus handles GET /api/v1/marketdata/status
func (h *MarketDataHandler) GetProviderStatus(c *fiber.Ctx) error {
	status := h.service.GetProviderStatus()

	return c.JSON(fiber.Map{
		"providers": status,
		"timestamp": time.Now().Unix(),
	})
}

// SearchSymbols handles GET /api/v1/marketdata/search?query=apple
func (h *MarketDataHandler) SearchSymbols(c *fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Query parameter is required",
		})
	}

	// For now, return a simple response
	// In the future, we can implement symbol search across providers
	return c.JSON(fiber.Map{
		"message": "Symbol search not yet implemented",
		"query":   query,
	})
}

// ValidateSymbol handles GET /api/v1/marketdata/validate/{symbol}
func (h *MarketDataHandler) ValidateSymbol(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol parameter is required",
		})
	}

	// Try to get a quote to validate the symbol
	_, err := h.service.GetRealtimeQuote(symbol)
	if err != nil {
		if mdErr, ok := err.(*marketdata.MarketDataError); ok {
			if mdErr.Code == marketdata.ErrCodeInvalidSymbol || mdErr.Code == marketdata.ErrCodeNoData {
				return c.JSON(fiber.Map{
					"symbol": symbol,
					"valid":  false,
					"reason": mdErr.Message,
				})
			}
		}
		// If it's a network error or similar, we can't determine validity
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to validate symbol",
		})
	}

	return c.JSON(fiber.Map{
		"symbol": symbol,
		"valid":  true,
	})
}

// GetSupportedPeriods handles GET /api/v1/marketdata/periods
func (h *MarketDataHandler) GetSupportedPeriods(c *fiber.Ctx) error {
	periods := []fiber.Map{
		{"value": "1d", "label": "1 Day", "description": "1 minute intervals"},
		{"value": "5d", "label": "5 Days", "description": "5 minute intervals"},
		{"value": "1mo", "label": "1 Month", "description": "Daily intervals"},
		{"value": "3mo", "label": "3 Months", "description": "Daily intervals"},
		{"value": "6mo", "label": "6 Months", "description": "Daily intervals"},
		{"value": "1y", "label": "1 Year", "description": "Daily intervals"},
		{"value": "2y", "label": "2 Years", "description": "Daily intervals"},
		{"value": "5y", "label": "5 Years", "description": "Daily intervals"},
		{"value": "10y", "label": "10 Years", "description": "Daily intervals"},
		{"value": "max", "label": "Maximum", "description": "All available data"},
	}

	return c.JSON(fiber.Map{
		"periods": periods,
	})
}
