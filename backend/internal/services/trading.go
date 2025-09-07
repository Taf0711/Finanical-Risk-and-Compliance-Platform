package services

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

// TradingService is a lightweight stub for broker interactions (Alpaca removed)
type TradingService struct {
	db *gorm.DB
}

type OrderRequest struct {
	Symbol      string  `json:"symbol" validate:"required"`
	Quantity    float64 `json:"quantity" validate:"required,gt=0"`
	Side        string  `json:"side" validate:"required,oneof=buy sell"`
	OrderType   string  `json:"order_type" validate:"required,oneof=market limit stop stop_limit"`
	TimeInForce string  `json:"time_in_force" validate:"required,oneof=day gtc ioc fok"`
	LimitPrice  float64 `json:"limit_price,omitempty"`
	StopPrice   float64 `json:"stop_price,omitempty"`
	PortfolioID string  `json:"portfolio_id" validate:"required"`
}

type OrderResponse struct {
	ID             string    `json:"id"`
	Symbol         string    `json:"symbol"`
	Quantity       string    `json:"quantity"`
	Side           string    `json:"side"`
	OrderType      string    `json:"order_type"`
	TimeInForce    string    `json:"time_in_force"`
	LimitPrice     string    `json:"limit_price,omitempty"`
	StopPrice      string    `json:"stop_price,omitempty"`
	Status         string    `json:"status"`
	FilledQuantity string    `json:"filled_quantity"`
	FilledPrice    string    `json:"filled_price,omitempty"`
	SubmittedAt    time.Time `json:"submitted_at"`
	FilledAt       time.Time `json:"filled_at,omitempty"`
}

type AccountInfo struct {
	ID             string `json:"id"`
	Currency       string `json:"currency"`
	Cash           string `json:"cash"`
	PortfolioValue string `json:"portfolio_value"`
	BuyingPower    string `json:"buying_power"`
}

type Position struct {
	Symbol       string `json:"symbol"`
	Quantity     string `json:"quantity"`
	CurrentPrice string `json:"current_price"`
}

// NewTradingServiceStub returns a stubbed trading service without external broker clients
func NewTradingServiceStub() *TradingService {
	return &TradingService{db: database.GetDB()}
}

// PlaceOrder simulates placing an order and immediately fills it (stub)
func (s *TradingService) PlaceOrder(req OrderRequest) (*OrderResponse, error) {
	// Validate portfolio ownership
	var portfolio models.Portfolio
	if err := s.db.Where("id = ?", req.PortfolioID).First(&portfolio).Error; err != nil {
		return nil, errors.New("portfolio not found")
	}

	qty := decimal.NewFromFloat(req.Quantity)
	now := time.Now()

	response := &OrderResponse{
		ID:             uuid.New().String(),
		Symbol:         req.Symbol,
		Quantity:       fmt.Sprintf("%.2f", req.Quantity),
		Side:           req.Side,
		OrderType:      req.OrderType,
		TimeInForce:    req.TimeInForce,
		Status:         "filled",
		FilledQuantity: fmt.Sprintf("%.2f", req.Quantity),
		FilledPrice:    fmt.Sprintf("%.2f", qty.InexactFloat64()),
		SubmittedAt:    now,
		FilledAt:       now,
	}

	// Store simulated transaction in DB
	go func() {
		executedAt := now
		transaction := models.Transaction{
			PortfolioID:     portfolio.ID,
			TransactionType: req.Side,
			Symbol:          req.Symbol,
			Quantity:        qty,
			Price:           decimal.NewFromFloat(qty.InexactFloat64()),
			Amount:          decimal.NewFromFloat(qty.InexactFloat64()).Mul(qty),
			Currency:        "USD",
			Status:          "filled",
			ExecutedAt:      &executedAt,
			KYCVerified:     true,
			AMLChecked:      true,
			RiskScore:       50,
		}
		if err := s.db.Create(&transaction).Error; err != nil {
			log.Printf("Error storing simulated transaction: %v", err)
		}
	}()

	return response, nil
}

// GetOrders returns stubbed orders (from DB transactions)
func (s *TradingService) GetOrders(status string, limit int) ([]OrderResponse, error) {
	var txs []models.Transaction
	query := s.db
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Limit(limit).Find(&txs).Error; err != nil {
		return nil, err
	}

	var res []OrderResponse
	for _, t := range txs {
		res = append(res, OrderResponse{
			ID:             t.ID.String(),
			Symbol:         t.Symbol,
			Quantity:       t.Quantity.String(),
			Side:           t.TransactionType,
			Status:         t.Status,
			FilledQuantity: t.Quantity.String(),
			SubmittedAt:    time.Now(),
		})
	}
	return res, nil
}

// GetOrder returns a single order (stub)
func (s *TradingService) GetOrder(orderID string) (*OrderResponse, error) {
	var tx models.Transaction
	if err := s.db.Where("id = ?", orderID).First(&tx).Error; err != nil {
		return nil, fmt.Errorf("order not found: %v", err)
	}
	return &OrderResponse{
		ID:             tx.ID.String(),
		Symbol:         tx.Symbol,
		Quantity:       tx.Quantity.String(),
		Side:           tx.TransactionType,
		Status:         tx.Status,
		FilledQuantity: tx.Quantity.String(),
		SubmittedAt:    time.Now(),
	}, nil
}

// CancelOrder simulates cancelling an order (stub)
func (s *TradingService) CancelOrder(orderID string) error {
	// For stub, simply mark transaction cancelled if exists
	return s.db.Model(&models.Transaction{}).Where("id = ?", orderID).Update("status", "cancelled").Error
}

// GetAccount returns a stubbed account
func (s *TradingService) GetAccount() (*AccountInfo, error) {
	return &AccountInfo{
		ID:             "stub-account",
		Currency:       "USD",
		Cash:           "10000",
		PortfolioValue: "100000",
		BuyingPower:    "50000",
	}, nil
}

// GetPositions returns positions from portfolio (stub)
func (s *TradingService) GetPositions() ([]Position, error) {
	// Return empty slice - positions are managed in portfolios
	return []Position{}, nil
}

// GetPosition returns a stubbed position
func (s *TradingService) GetPosition(symbol string) (*Position, error) {
	return &Position{Symbol: symbol, Quantity: "0", CurrentPrice: "0"}, nil
}

// ClosePosition simulates closing a position
func (s *TradingService) ClosePosition(symbol string) (*OrderResponse, error) {
	// Not implemented in stub
	return nil, fmt.Errorf("ClosePosition not supported in stub")
}
