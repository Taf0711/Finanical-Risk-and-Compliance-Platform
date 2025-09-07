package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

type PortfolioService struct {
	db *gorm.DB
}

func NewPortfolioService() *PortfolioService {
	return &PortfolioService{
		db: database.GetDB(),
	}
}

type CreatePortfolioRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
	Currency    string `json:"currency"`
}

type UpdatePortfolioRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// GetUserPortfolios returns all portfolios for a specific user
func (s *PortfolioService) GetUserPortfolios(userID uuid.UUID) ([]models.Portfolio, error) {
	var portfolios []models.Portfolio
	err := s.db.Preload("User").Where("user_id = ?", userID).Find(&portfolios).Error
	return portfolios, err
}

// GetPortfolio returns a specific portfolio by ID, ensuring it belongs to the user
func (s *PortfolioService) GetPortfolio(portfolioID, userID uuid.UUID) (*models.Portfolio, error) {
	var portfolio models.Portfolio
	err := s.db.Where("id = ? AND user_id = ?", portfolioID, userID).
		Preload("Positions").
		Preload("User").
		First(&portfolio).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("portfolio not found")
		}
		return nil, err
	}

	return &portfolio, nil
}

// CreatePortfolio creates a new portfolio for a user
func (s *PortfolioService) CreatePortfolio(userID uuid.UUID, req CreatePortfolioRequest) (*models.Portfolio, error) {
	portfolio := models.Portfolio{
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
		Currency:    req.Currency,
		TotalValue:  decimal.Zero,
	}

	if portfolio.Currency == "" {
		portfolio.Currency = "USD"
	}

	err := s.db.Create(&portfolio).Error
	if err != nil {
		return nil, err
	}

	return &portfolio, nil
}

// UpdatePortfolio updates an existing portfolio
func (s *PortfolioService) UpdatePortfolio(portfolioID, userID uuid.UUID, req UpdatePortfolioRequest) (*models.Portfolio, error) {
	var portfolio models.Portfolio

	// Check if portfolio exists and belongs to user
	err := s.db.Where("id = ? AND user_id = ?", portfolioID, userID).First(&portfolio).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("portfolio not found")
		}
		return nil, err
	}

	// Update fields
	if req.Name != "" {
		portfolio.Name = req.Name
	}
	if req.Description != "" {
		portfolio.Description = req.Description
	}

	err = s.db.Save(&portfolio).Error
	if err != nil {
		return nil, err
	}

	return &portfolio, nil
}

// DeletePortfolio deletes a portfolio and all its positions
func (s *PortfolioService) DeletePortfolio(portfolioID, userID uuid.UUID) error {
	// Check if portfolio exists and belongs to user
	var portfolio models.Portfolio
	err := s.db.Where("id = ? AND user_id = ?", portfolioID, userID).First(&portfolio).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("portfolio not found")
		}
		return err
	}

	// Delete all positions first (cascade delete)
	err = s.db.Where("portfolio_id = ?", portfolioID).Delete(&models.Position{}).Error
	if err != nil {
		return err
	}

	// Delete the portfolio
	err = s.db.Delete(&portfolio).Error
	return err
}

// GetPortfolioPositions returns all positions for a portfolio
func (s *PortfolioService) GetPortfolioPositions(portfolioID, userID uuid.UUID) ([]models.Position, error) {
	// First verify the portfolio belongs to the user
	var portfolio models.Portfolio
	err := s.db.Where("id = ? AND user_id = ?", portfolioID, userID).First(&portfolio).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("portfolio not found")
		}
		return nil, err
	}

	var positions []models.Position
	err = s.db.Where("portfolio_id = ?", portfolioID).Find(&positions).Error
	return positions, err
}

// CalculatePortfolioValue recalculates the total value of a portfolio
func (s *PortfolioService) CalculatePortfolioValue(portfolioID uuid.UUID) error {
	var positions []models.Position
	err := s.db.Where("portfolio_id = ?", portfolioID).Find(&positions).Error
	if err != nil {
		return err
	}

	totalValue := decimal.Zero
	for _, position := range positions {
		totalValue = totalValue.Add(position.MarketValue)
	}

	// Update portfolio total value
	err = s.db.Model(&models.Portfolio{}).
		Where("id = ?", portfolioID).
		Update("total_value", totalValue).Error

	return err
}

// FundTransactionRequest represents a fund deposit or withdrawal request
type FundTransactionRequest struct {
	Amount      decimal.Decimal `json:"amount" validate:"required,gt=0"`
	Type        string          `json:"type" validate:"required,oneof=deposit withdrawal"`
	Description string          `json:"description"`
	Method      string          `json:"method" validate:"required,oneof=bank_transfer wire credit_card"`
}

// FundTransactionResponse represents the response for fund operations
type FundTransactionResponse struct {
	TransactionID uuid.UUID       `json:"transaction_id"`
	PortfolioID   uuid.UUID       `json:"portfolio_id"`
	Amount        decimal.Decimal `json:"amount"`
	Type          string          `json:"type"`
	Status        string          `json:"status"`
	Method        string          `json:"method"`
	Description   string          `json:"description"`
	ProcessedAt   time.Time       `json:"processed_at"`
	NewBalance    decimal.Decimal `json:"new_balance"`
}

// AddFunds adds funds to a portfolio (simulated for demo purposes)
func (s *PortfolioService) AddFunds(portfolioID, userID uuid.UUID, req FundTransactionRequest) (*FundTransactionResponse, error) {
	// Verify portfolio ownership
	var portfolio models.Portfolio
	err := s.db.Where("id = ? AND user_id = ?", portfolioID, userID).First(&portfolio).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("portfolio not found")
		}
		return nil, err
	}

	// Validate request
	if req.Type != "deposit" {
		return nil, errors.New("invalid transaction type for adding funds")
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("amount must be greater than zero")
	}

	// Create a cash position or update existing one
	var cashPosition models.Position
	err = s.db.Where("portfolio_id = ? AND symbol = ?", portfolioID, "CASH").First(&cashPosition).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create new cash position
		cashPosition = models.Position{
			PortfolioID:  portfolioID,
			Symbol:       "CASH",
			AssetType:    "CASH",
			Quantity:     req.Amount,
			AveragePrice: decimal.NewFromInt(1), // $1 per unit for cash
			CurrentPrice: decimal.NewFromInt(1),
			MarketValue:  req.Amount,
			Liquidity:    "HIGH",
		}
		err = s.db.Create(&cashPosition).Error
	} else if err == nil {
		// Update existing cash position
		cashPosition.Quantity = cashPosition.Quantity.Add(req.Amount)
		cashPosition.MarketValue = cashPosition.MarketValue.Add(req.Amount)
		err = s.db.Save(&cashPosition).Error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update cash position: %w", err)
	}

	// Create transaction record
	transaction := models.Transaction{
		PortfolioID:     portfolioID,
		TransactionType: "DEPOSIT",
		Symbol:          "CASH",
		Quantity:        req.Amount,
		Price:           decimal.NewFromInt(1),
		Amount:          req.Amount,
		Currency:        portfolio.Currency,
		Status:          "COMPLETED",
		ExecutedAt:      &[]time.Time{time.Now()}[0],
		Notes:           req.Description,
		KYCVerified:     true,
		AMLChecked:      true,
		RiskScore:       10, // Low risk for cash deposits
	}

	err = s.db.Create(&transaction).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Update portfolio total value
	err = s.CalculatePortfolioValue(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to update portfolio value: %w", err)
	}

	// Get updated portfolio
	err = s.db.First(&portfolio, portfolioID).Error
	if err != nil {
		return nil, err
	}

	return &FundTransactionResponse{
		TransactionID: transaction.ID,
		PortfolioID:   portfolioID,
		Amount:        req.Amount,
		Type:          "deposit",
		Status:        "completed",
		Method:        req.Method,
		Description:   req.Description,
		ProcessedAt:   time.Now(),
		NewBalance:    portfolio.TotalValue,
	}, nil
}

// WithdrawFunds withdraws funds from a portfolio
func (s *PortfolioService) WithdrawFunds(portfolioID, userID uuid.UUID, req FundTransactionRequest) (*FundTransactionResponse, error) {
	// Verify portfolio ownership
	var portfolio models.Portfolio
	err := s.db.Where("id = ? AND user_id = ?", portfolioID, userID).First(&portfolio).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("portfolio not found")
		}
		return nil, err
	}

	// Validate request
	if req.Type != "withdrawal" {
		return nil, errors.New("invalid transaction type for withdrawing funds")
	}

	if req.Amount.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("amount must be greater than zero")
	}

	// Check if there's enough cash available
	var cashPosition models.Position
	err = s.db.Where("portfolio_id = ? AND symbol = ?", portfolioID, "CASH").First(&cashPosition).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("insufficient cash balance")
		}
		return nil, err
	}

	if cashPosition.Quantity.LessThan(req.Amount) {
		return nil, errors.New("insufficient cash balance for withdrawal")
	}

	// Update cash position
	cashPosition.Quantity = cashPosition.Quantity.Sub(req.Amount)
	cashPosition.MarketValue = cashPosition.MarketValue.Sub(req.Amount)

	// If cash position becomes zero, we can keep it or delete it
	if cashPosition.Quantity.IsZero() {
		err = s.db.Delete(&cashPosition).Error
	} else {
		err = s.db.Save(&cashPosition).Error
	}

	if err != nil {
		return nil, fmt.Errorf("failed to update cash position: %w", err)
	}

	// Create transaction record
	transaction := models.Transaction{
		PortfolioID:     portfolioID,
		TransactionType: "WITHDRAWAL",
		Symbol:          "CASH",
		Quantity:        req.Amount,
		Price:           decimal.NewFromInt(1),
		Amount:          req.Amount,
		Currency:        portfolio.Currency,
		Status:          "COMPLETED",
		ExecutedAt:      &[]time.Time{time.Now()}[0],
		Notes:           req.Description,
		KYCVerified:     true,
		AMLChecked:      true,
		RiskScore:       15, // Slightly higher risk for withdrawals
	}

	err = s.db.Create(&transaction).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create transaction record: %w", err)
	}

	// Update portfolio total value
	err = s.CalculatePortfolioValue(portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to update portfolio value: %w", err)
	}

	// Get updated portfolio
	err = s.db.First(&portfolio, portfolioID).Error
	if err != nil {
		return nil, err
	}

	return &FundTransactionResponse{
		TransactionID: transaction.ID,
		PortfolioID:   portfolioID,
		Amount:        req.Amount,
		Type:          "withdrawal",
		Status:        "completed",
		Method:        req.Method,
		Description:   req.Description,
		ProcessedAt:   time.Now(),
		NewBalance:    portfolio.TotalValue,
	}, nil
}

// GetCashBalance returns the available cash balance in a portfolio
func (s *PortfolioService) GetCashBalance(portfolioID, userID uuid.UUID) (decimal.Decimal, error) {
	// Verify portfolio ownership
	var portfolio models.Portfolio
	err := s.db.Where("id = ? AND user_id = ?", portfolioID, userID).First(&portfolio).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return decimal.Zero, errors.New("portfolio not found")
		}
		return decimal.Zero, err
	}

	var cashPosition models.Position
	err = s.db.Where("portfolio_id = ? AND symbol = ?", portfolioID, "CASH").First(&cashPosition).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return decimal.Zero, nil // No cash position means zero balance
		}
		return decimal.Zero, err
	}

	return cashPosition.Quantity, nil
}

// GetTransactionHistory returns transaction history for a portfolio
func (s *PortfolioService) GetTransactionHistory(portfolioID, userID uuid.UUID, limit int) ([]models.Transaction, error) {
	// Verify portfolio ownership
	var portfolio models.Portfolio
	err := s.db.Where("id = ? AND user_id = ?", portfolioID, userID).First(&portfolio).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("portfolio not found")
		}
		return nil, err
	}

	var transactions []models.Transaction
	query := s.db.Where("portfolio_id = ?", portfolioID).Order("executed_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err = query.Find(&transactions).Error
	return transactions, err
}
