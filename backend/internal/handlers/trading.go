package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"

	"github.com/Taf0711/financial-risk-monitor/internal/services"
)

type TradingHandler struct {
	tradingService *services.TradingService
}

func NewTradingHandler(tradingService *services.TradingService) *TradingHandler {
	return &TradingHandler{
		tradingService: tradingService,
	}
}

// PlaceOrder handles POST /api/v1/trading/orders
func (h *TradingHandler) PlaceOrder(c *fiber.Ctx) error {
	var req services.OrderRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate order
	if err := h.tradingService.ValidateOrder(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Place order
	order, err := h.tradingService.PlaceOrder(req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"order":   order,
	})
}

// GetOrders handles GET /api/v1/trading/orders
func (h *TradingHandler) GetOrders(c *fiber.Ctx) error {
	status := c.Query("status")
	limitStr := c.Query("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	orders, err := h.tradingService.GetOrders(status, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"orders":  orders,
		"count":   len(orders),
	})
}

// GetOrder handles GET /api/v1/trading/orders/:id
func (h *TradingHandler) GetOrder(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	order, err := h.tradingService.GetOrder(orderID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"order":   order,
	})
}

// CancelOrder handles DELETE /api/v1/trading/orders/:id
func (h *TradingHandler) CancelOrder(c *fiber.Ctx) error {
	orderID := c.Params("id")
	if orderID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order ID is required",
		})
	}

	err := h.tradingService.CancelOrder(orderID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Order cancelled successfully",
	})
}

// GetAccount handles GET /api/v1/trading/account
func (h *TradingHandler) GetAccount(c *fiber.Ctx) error {
	account, err := h.tradingService.GetAccount()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"account": account,
	})
}

// GetPositions handles GET /api/v1/trading/positions
func (h *TradingHandler) GetPositions(c *fiber.Ctx) error {
	positions, err := h.tradingService.GetPositions()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"positions": positions,
		"count":     len(positions),
	})
}

// GetPosition handles GET /api/v1/trading/positions/:symbol
func (h *TradingHandler) GetPosition(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol is required",
		})
	}

	position, err := h.tradingService.GetPosition(symbol)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"position": position,
	})
}

// ClosePosition handles DELETE /api/v1/trading/positions/:symbol
func (h *TradingHandler) ClosePosition(c *fiber.Ctx) error {
	symbol := c.Params("symbol")
	if symbol == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Symbol is required",
		})
	}

	order, err := h.tradingService.ClosePosition(symbol)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Position closed successfully",
		"order":   order,
	})
}

// GetTradingStatus handles GET /api/v1/trading/status
func (h *TradingHandler) GetTradingStatus(c *fiber.Ctx) error {
	account, err := h.tradingService.GetAccount()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Get positions count
	positions, err := h.tradingService.GetPositions()
	if err != nil {
		positions = []services.Position{} // Default to empty if error
	}

	// Get recent orders
	orders, err := h.tradingService.GetOrders("", 10)
	if err != nil {
		orders = []services.OrderResponse{} // Default to empty if error
	}

	return c.JSON(fiber.Map{
		"success": true,
		"status": fiber.Map{
			"account_status":     account.Status,
			"trading_blocked":    account.TradingBlocked,
			"pattern_day_trader": account.PatternDayTrader,
			"buying_power":       account.BuyingPower,
			"cash":               account.Cash,
			"portfolio_value":    account.PortfolioValue,
			"positions_count":    len(positions),
			"recent_orders":      len(orders),
			"day_trade_count":    account.DayTradeCount,
		},
	})
}

// ValidateOrder handles POST /api/v1/trading/orders/validate
func (h *TradingHandler) ValidateOrder(c *fiber.Ctx) error {
	var req services.OrderRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate order
	err := h.tradingService.ValidateOrder(req)
	if err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"valid":   false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"valid":   true,
		"message": "Order is valid and ready to submit",
	})
}
