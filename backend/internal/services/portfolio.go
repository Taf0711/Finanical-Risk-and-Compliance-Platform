package services

import (
	"errors"

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
	err := s.db.Where("user_id = ?", userID).Find(&portfolios).Error
	return portfolios, err
}

// GetPortfolio returns a specific portfolio by ID, ensuring it belongs to the user
func (s *PortfolioService) GetPortfolio(portfolioID, userID uuid.UUID) (*models.Portfolio, error) {
	var portfolio models.Portfolio
	err := s.db.Where("id = ? AND user_id = ?", portfolioID, userID).
		Preload("Positions").
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
