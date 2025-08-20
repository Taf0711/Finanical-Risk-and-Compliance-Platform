package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/Taf0711/financial-risk-monitor/internal/services"
)

type PortfolioHandler struct {
	portfolioService *services.PortfolioService
}

func NewPortfolioHandler() *PortfolioHandler {
	return &PortfolioHandler{
		portfolioService: services.NewPortfolioService(),
	}
}

// GetPortfolios returns all portfolios for a user
func (h *PortfolioHandler) GetPortfolios(c *fiber.Ctx) error {
	userID := c.Locals("user_id").(string)

	portfolios, err := h.portfolioService.GetUserPortfolios(uuid.MustParse(userID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch portfolios",
		})
	}

	return c.JSON(portfolios)
}

// GetPortfolio returns a specific portfolio
func (h *PortfolioHandler) GetPortfolio(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	userID := c.Locals("user_id").(string)

	portfolio, err := h.portfolioService.GetPortfolio(uuid.MustParse(portfolioID), uuid.MustParse(userID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Portfolio not found",
		})
	}

	return c.JSON(portfolio)
}

// CreatePortfolio creates a new portfolio
func (h *PortfolioHandler) CreatePortfolio(c *fiber.Ctx) error {
	var req struct {
		Name        string `json:"name" validate:"required"`
		Description string `json:"description"`
		Currency    string `json:"currency"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	userID := c.Locals("user_id").(string)

	createReq := services.CreatePortfolioRequest{
		Name:        req.Name,
		Description: req.Description,
		Currency:    req.Currency,
	}

	portfolio, err := h.portfolioService.CreatePortfolio(uuid.MustParse(userID), createReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create portfolio",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(portfolio)
}

// UpdatePortfolio updates a portfolio
func (h *PortfolioHandler) UpdatePortfolio(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	userID := c.Locals("user_id").(string)

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	updateReq := services.UpdatePortfolioRequest{
		Name:        req.Name,
		Description: req.Description,
	}

	portfolio, err := h.portfolioService.UpdatePortfolio(
		uuid.MustParse(portfolioID),
		uuid.MustParse(userID),
		updateReq,
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update portfolio",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Portfolio updated successfully",
		"data":    portfolio,
	})
}

// DeletePortfolio deletes a portfolio
func (h *PortfolioHandler) DeletePortfolio(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	userID := c.Locals("user_id").(string)

	err := h.portfolioService.DeletePortfolio(
		uuid.MustParse(portfolioID),
		uuid.MustParse(userID),
	)

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete portfolio",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Portfolio deleted successfully",
	})
}

// GetPositions returns all positions for a portfolio
func (h *PortfolioHandler) GetPositions(c *fiber.Ctx) error {
	portfolioID := c.Params("id")
	userID := c.Locals("user_id").(string)

	positions, err := h.portfolioService.GetPortfolioPositions(uuid.MustParse(portfolioID), uuid.MustParse(userID))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Portfolio not found or access denied",
		})
	}

	return c.JSON(positions)
}

// AddPosition adds a position to a portfolio
func (h *PortfolioHandler) AddPosition(c *fiber.Ctx) error {
	// TODO: Implement position addition
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "Position addition not yet implemented",
	})
}

// UpdatePosition updates a position in a portfolio
func (h *PortfolioHandler) UpdatePosition(c *fiber.Ctx) error {
	// TODO: Implement position update
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "Position update not yet implemented",
	})
}

// DeletePosition deletes a position from a portfolio
func (h *PortfolioHandler) DeletePosition(c *fiber.Ctx) error {
	// TODO: Implement position deletion
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"error": "Position deletion not yet implemented",
	})
}
