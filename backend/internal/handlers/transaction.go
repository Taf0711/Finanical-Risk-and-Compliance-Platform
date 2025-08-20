package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

type TransactionHandler struct {
	// Add transaction service when implemented
}

func NewTransactionHandler() *TransactionHandler {
	return &TransactionHandler{}
}

type CreateTransactionRequest struct {
	PortfolioID     string  `json:"portfolio_id" validate:"required"`
	TransactionType string  `json:"transaction_type" validate:"required"`
	Symbol          string  `json:"symbol"`
	Quantity        float64 `json:"quantity"`
	Price           float64 `json:"price"`
	Currency        string  `json:"currency"`
	ExecutedAt      string  `json:"executed_at"`
	Notes           string  `json:"notes"`
}

type UpdateTransactionStatusRequest struct {
	Status string `json:"status" validate:"required"`
}

// GetTransactions returns all transactions
func (h *TransactionHandler) GetTransactions(c *fiber.Ctx) error {
	var transactions []models.Transaction

	if err := database.GetDB().Find(&transactions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve transactions",
		})
	}

	return c.JSON(transactions)
}

// CreateTransaction creates a new transaction
func (h *TransactionHandler) CreateTransaction(c *fiber.Ctx) error {
	var req CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	portfolioID, err := uuid.Parse(req.PortfolioID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid portfolio ID",
		})
	}

	transaction := models.Transaction{
		PortfolioID:     portfolioID,
		TransactionType: req.TransactionType,
		Symbol:          req.Symbol,
		Quantity:        decimal.NewFromFloat(req.Quantity),
		Price:           decimal.NewFromFloat(req.Price),
		Amount:          decimal.NewFromFloat(req.Quantity * req.Price),
		Currency:        req.Currency,
		Status:          "PENDING",
		Notes:           req.Notes,
	}

	if req.ExecutedAt != "" {
		if executedAt, err := time.Parse(time.RFC3339, req.ExecutedAt); err == nil {
			transaction.ExecutedAt = &executedAt
		}
	}

	if req.Currency == "" {
		transaction.Currency = "USD"
	}

	if err := database.GetDB().Create(&transaction).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create transaction",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message":     "Transaction created successfully",
		"transaction": transaction,
	})
}

// GetTransaction returns a specific transaction
func (h *TransactionHandler) GetTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	transactionID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid transaction ID",
		})
	}

	var transaction models.Transaction
	if err := database.GetDB().First(&transaction, transactionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Transaction not found",
		})
	}

	return c.JSON(transaction)
}

// UpdateTransaction updates a transaction
func (h *TransactionHandler) UpdateTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	transactionID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid transaction ID",
		})
	}

	var req CreateTransactionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var transaction models.Transaction
	if err := database.GetDB().First(&transaction, transactionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Transaction not found",
		})
	}

	// Update fields
	if req.Symbol != "" {
		transaction.Symbol = req.Symbol
	}
	if req.Quantity != 0 {
		transaction.Quantity = decimal.NewFromFloat(req.Quantity)
	}
	if req.Price != 0 {
		transaction.Price = decimal.NewFromFloat(req.Price)
		transaction.Amount = transaction.Quantity.Mul(transaction.Price)
	}
	if req.Notes != "" {
		transaction.Notes = req.Notes
	}

	if err := database.GetDB().Save(&transaction).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update transaction",
		})
	}

	return c.JSON(fiber.Map{
		"message":     "Transaction updated successfully",
		"transaction": transaction,
	})
}

// DeleteTransaction deletes a transaction
func (h *TransactionHandler) DeleteTransaction(c *fiber.Ctx) error {
	id := c.Params("id")
	transactionID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid transaction ID",
		})
	}

	var transaction models.Transaction
	if err := database.GetDB().First(&transaction, transactionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Transaction not found",
		})
	}

	if err := database.GetDB().Delete(&transaction).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete transaction",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Transaction deleted successfully",
	})
}

// UpdateTransactionStatus updates the status of a transaction
func (h *TransactionHandler) UpdateTransactionStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	transactionID, err := uuid.Parse(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid transaction ID",
		})
	}

	var req UpdateTransactionStatusRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var transaction models.Transaction
	if err := database.GetDB().First(&transaction, transactionID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Transaction not found",
		})
	}

	transaction.Status = req.Status
	if req.Status == "COMPLETED" {
		now := time.Now()
		transaction.ExecutedAt = &now
	}

	if err := database.GetDB().Save(&transaction).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update transaction status",
		})
	}

	return c.JSON(fiber.Map{
		"message":     "Transaction status updated successfully",
		"transaction": transaction,
	})
}
