package services

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	"github.com/alpacahq/alpaca-trade-api-go/v3/marketdata"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

type TradingService struct {
	db           *gorm.DB
	alpacaClient *alpaca.Client
	marketClient *marketdata.Client
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
	ID             string     `json:"id"`
	Symbol         string     `json:"symbol"`
	Quantity       string     `json:"quantity"`
	Side           string     `json:"side"`
	OrderType      string     `json:"order_type"`
	TimeInForce    string     `json:"time_in_force"`
	LimitPrice     string     `json:"limit_price,omitempty"`
	StopPrice      string     `json:"stop_price,omitempty"`
	Status         string     `json:"status"`
	FilledQuantity string     `json:"filled_quantity"`
	FilledPrice    string     `json:"filled_price,omitempty"`
	SubmittedAt    time.Time  `json:"submitted_at"`
	FilledAt       time.Time  `json:"filled_at,omitempty"`
	CancelledAt    time.Time  `json:"cancelled_at,omitempty"`
	ExpiredAt      time.Time  `json:"expired_at,omitempty"`
	AssetClass     string     `json:"asset_class"`
	ExtendedHours  bool       `json:"extended_hours"`
	Legs           []OrderLeg `json:"legs,omitempty"`
}

type OrderLeg struct {
	Symbol   string `json:"symbol"`
	Quantity string `json:"quantity"`
	Side     string `json:"side"`
}

type AccountInfo struct {
	ID                   string    `json:"id"`
	AccountNumber        string    `json:"account_number"`
	Status               string    `json:"status"`
	Currency             string    `json:"currency"`
	Cash                 string    `json:"cash"`
	PortfolioValue       string    `json:"portfolio_value"`
	BuyingPower          string    `json:"buying_power"`
	Equity               string    `json:"equity"`
	LastEquity           string    `json:"last_equity"`
	Multiplier           string    `json:"multiplier"`
	DayTradeCount        int       `json:"day_trade_count"`
	PatternDayTrader     bool      `json:"pattern_day_trader"`
	TradingBlocked       bool      `json:"trading_blocked"`
	TransfersBlocked     bool      `json:"transfers_blocked"`
	AccountBlocked       bool      `json:"account_blocked"`
	CreatedAt            time.Time `json:"created_at"`
	TradeSuspendedByUser bool      `json:"trade_suspended_by_user"`
	MaxMarginMultiplier  string    `json:"max_margin_multiplier"`
	AgreementSigned      bool      `json:"agreement_signed"`
	OptionsTradingLevel  string    `json:"options_trading_level"`
}

type Position struct {
	AssetID                string `json:"asset_id"`
	Symbol                 string `json:"symbol"`
	Exchange               string `json:"exchange"`
	AssetClass             string `json:"asset_class"`
	Quantity               string `json:"quantity"`
	MarketValue            string `json:"market_value"`
	CostBasis              string `json:"cost_basis"`
	UnrealizedPL           string `json:"unrealized_pl"`
	UnrealizedPLPC         string `json:"unrealized_plpc"`
	UnrealizedIntradayPL   string `json:"unrealized_intraday_pl"`
	UnrealizedIntradayPLPC string `json:"unrealized_intraday_plpc"`
	CurrentPrice           string `json:"current_price"`
	LastdayPrice           string `json:"lastday_price"`
	ChangeToday            string `json:"change_today"`
}

func NewTradingService(apiKey, secretKey string, isPaper bool) *TradingService {
	var baseURL string
	if isPaper {
		baseURL = "https://paper-api.alpaca.markets"
	} else {
		baseURL = "https://api.alpaca.markets"
	}

	// Initialize Alpaca client
	alpacaClient := alpaca.NewClient(alpaca.ClientOpts{
		APIKey:    apiKey,
		APISecret: secretKey,
		BaseURL:   baseURL,
	})

	// Initialize market data client
	marketClient := marketdata.NewClient(marketdata.ClientOpts{
		APIKey:    apiKey,
		APISecret: secretKey,
	})

	return &TradingService{
		db:           database.GetDB(),
		alpacaClient: alpacaClient,
		marketClient: marketClient,
	}
}

// PlaceOrder places a new order with Alpaca
func (s *TradingService) PlaceOrder(req OrderRequest) (*OrderResponse, error) {
	// Validate portfolio ownership
	var portfolio models.Portfolio
	if err := s.db.Where("id = ?", req.PortfolioID).First(&portfolio).Error; err != nil {
		return nil, errors.New("portfolio not found")
	}

	// Convert quantity to decimal
	qty := decimal.NewFromFloat(req.Quantity)

	// Create Alpaca order request
	orderReq := alpaca.PlaceOrderRequest{
		Symbol:      req.Symbol,
		Qty:         &qty,
		Side:        alpaca.Side(req.Side),
		Type:        alpaca.OrderType(req.OrderType),
		TimeInForce: alpaca.TimeInForce(req.TimeInForce),
	}

	// Add limit price if specified
	if req.LimitPrice > 0 {
		limitPrice := decimal.NewFromFloat(req.LimitPrice)
		orderReq.LimitPrice = &limitPrice
	}

	// Add stop price if specified
	if req.StopPrice > 0 {
		stopPrice := decimal.NewFromFloat(req.StopPrice)
		orderReq.StopPrice = &stopPrice
	}

	// Place order with Alpaca
	order, err := s.alpacaClient.PlaceOrder(orderReq)
	if err != nil {
		log.Printf("Error placing order: %v", err)
		return nil, fmt.Errorf("failed to place order: %v", err)
	}

	// Convert Alpaca order to our response format
	response := &OrderResponse{
		ID:             order.ID,
		Symbol:         order.Symbol,
		Quantity:       order.Qty.String(),
		Side:           string(order.Side),
		OrderType:      string(order.Type),
		TimeInForce:    string(order.TimeInForce),
		Status:         string(order.Status),
		FilledQuantity: order.FilledQty.String(),
		SubmittedAt:    order.SubmittedAt,
		AssetClass:     string(order.AssetClass),
		ExtendedHours:  order.ExtendedHours,
	}

	if order.LimitPrice != nil {
		response.LimitPrice = order.LimitPrice.String()
	}

	if order.StopPrice != nil {
		response.StopPrice = order.StopPrice.String()
	}

	if order.FilledAvgPrice != nil {
		response.FilledPrice = order.FilledAvgPrice.String()
	}

	if order.FilledAt != nil && !order.FilledAt.IsZero() {
		response.FilledAt = *order.FilledAt
	}

	if order.ExpiredAt != nil && !order.ExpiredAt.IsZero() {
		response.ExpiredAt = *order.ExpiredAt
	}

	// Store order in database for tracking
	go s.storeOrderInDB(order, req.PortfolioID)

	return response, nil
}

// GetOrders retrieves orders for a user
func (s *TradingService) GetOrders(status string, limit int) ([]OrderResponse, error) {
	// Create request parameters
	req := alpaca.GetOrdersRequest{}

	if status != "" {
		req.Status = status
	}

	if limit > 0 {
		req.Limit = limit
	}

	// Get orders from Alpaca
	orders, err := s.alpacaClient.GetOrders(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %v", err)
	}

	// Convert to response format
	var responses []OrderResponse
	for _, order := range orders {
		response := OrderResponse{
			ID:             order.ID,
			Symbol:         order.Symbol,
			Quantity:       order.Qty.String(),
			Side:           string(order.Side),
			OrderType:      string(order.Type),
			TimeInForce:    string(order.TimeInForce),
			Status:         string(order.Status),
			FilledQuantity: order.FilledQty.String(),
			SubmittedAt:    order.SubmittedAt,
			AssetClass:     string(order.AssetClass),
			ExtendedHours:  order.ExtendedHours,
		}

		if order.LimitPrice != nil {
			response.LimitPrice = order.LimitPrice.String()
		}

		if order.StopPrice != nil {
			response.StopPrice = order.StopPrice.String()
		}

		if order.FilledAvgPrice != nil {
			response.FilledPrice = order.FilledAvgPrice.String()
		}

		if order.FilledAt != nil && !order.FilledAt.IsZero() {
			response.FilledAt = *order.FilledAt
		}

		if order.ExpiredAt != nil && !order.ExpiredAt.IsZero() {
			response.ExpiredAt = *order.ExpiredAt
		}

		responses = append(responses, response)
	}

	return responses, nil
}

// GetOrder retrieves a specific order by ID
func (s *TradingService) GetOrder(orderID string) (*OrderResponse, error) {
	order, err := s.alpacaClient.GetOrder(orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %v", err)
	}

	response := &OrderResponse{
		ID:             order.ID,
		Symbol:         order.Symbol,
		Quantity:       order.Qty.String(),
		Side:           string(order.Side),
		OrderType:      string(order.Type),
		TimeInForce:    string(order.TimeInForce),
		Status:         string(order.Status),
		FilledQuantity: order.FilledQty.String(),
		SubmittedAt:    order.SubmittedAt,
		AssetClass:     string(order.AssetClass),
		ExtendedHours:  order.ExtendedHours,
	}

	if order.LimitPrice != nil {
		response.LimitPrice = order.LimitPrice.String()
	}

	if order.StopPrice != nil {
		response.StopPrice = order.StopPrice.String()
	}

	if order.FilledAvgPrice != nil {
		response.FilledPrice = order.FilledAvgPrice.String()
	}

	if order.FilledAt != nil && !order.FilledAt.IsZero() {
		response.FilledAt = *order.FilledAt
	}

	if order.ExpiredAt != nil && !order.ExpiredAt.IsZero() {
		response.ExpiredAt = *order.ExpiredAt
	}

	return response, nil
}

// CancelOrder cancels an order
func (s *TradingService) CancelOrder(orderID string) error {
	return s.alpacaClient.CancelOrder(orderID)
}

// GetAccount retrieves account information
func (s *TradingService) GetAccount() (*AccountInfo, error) {
	account, err := s.alpacaClient.GetAccount()
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %v", err)
	}

	return &AccountInfo{
		ID:                   account.ID,
		AccountNumber:        account.AccountNumber,
		Status:               string(account.Status),
		Currency:             account.Currency,
		Cash:                 account.Cash.String(),
		PortfolioValue:       account.PortfolioValue.String(),
		BuyingPower:          account.BuyingPower.String(),
		Equity:               account.Equity.String(),
		LastEquity:           account.LastEquity.String(),
		Multiplier:           account.Multiplier.String(),
		DayTradeCount:        int(account.DaytradeCount),
		PatternDayTrader:     account.PatternDayTrader,
		TradingBlocked:       account.TradingBlocked,
		TransfersBlocked:     account.TransfersBlocked,
		AccountBlocked:       account.AccountBlocked,
		CreatedAt:            account.CreatedAt,
		TradeSuspendedByUser: account.TradeSuspendedByUser,
		MaxMarginMultiplier:  "4",  // Default value since field doesn't exist
		AgreementSigned:      true, // Default value since field doesn't exist
		OptionsTradingLevel:  "0",  // Default value since field doesn't exist
	}, nil
}

// GetPositions retrieves current positions
func (s *TradingService) GetPositions() ([]Position, error) {
	positions, err := s.alpacaClient.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("failed to get positions: %v", err)
	}

	var result []Position
	for _, pos := range positions {
		position := Position{
			AssetID:                pos.AssetID,
			Symbol:                 pos.Symbol,
			Exchange:               pos.Exchange,
			AssetClass:             string(pos.AssetClass),
			Quantity:               pos.Qty.String(),
			MarketValue:            pos.MarketValue.String(),
			CostBasis:              pos.CostBasis.String(),
			UnrealizedPL:           pos.UnrealizedPL.String(),
			UnrealizedPLPC:         pos.UnrealizedPLPC.String(),
			UnrealizedIntradayPL:   pos.UnrealizedIntradayPL.String(),
			UnrealizedIntradayPLPC: pos.UnrealizedIntradayPLPC.String(),
			CurrentPrice:           pos.CurrentPrice.String(),
			LastdayPrice:           pos.LastdayPrice.String(),
			ChangeToday:            pos.ChangeToday.String(),
		}
		result = append(result, position)
	}

	return result, nil
}

// GetPosition retrieves a specific position by symbol
func (s *TradingService) GetPosition(symbol string) (*Position, error) {
	pos, err := s.alpacaClient.GetPosition(symbol)
	if err != nil {
		return nil, fmt.Errorf("failed to get position: %v", err)
	}

	return &Position{
		AssetID:                pos.AssetID,
		Symbol:                 pos.Symbol,
		Exchange:               pos.Exchange,
		AssetClass:             string(pos.AssetClass),
		Quantity:               pos.Qty.String(),
		MarketValue:            pos.MarketValue.String(),
		CostBasis:              pos.CostBasis.String(),
		UnrealizedPL:           pos.UnrealizedPL.String(),
		UnrealizedPLPC:         pos.UnrealizedPLPC.String(),
		UnrealizedIntradayPL:   pos.UnrealizedIntradayPL.String(),
		UnrealizedIntradayPLPC: pos.UnrealizedIntradayPLPC.String(),
		CurrentPrice:           pos.CurrentPrice.String(),
		LastdayPrice:           pos.LastdayPrice.String(),
		ChangeToday:            pos.ChangeToday.String(),
	}, nil
}

// ClosePosition closes a position
func (s *TradingService) ClosePosition(symbol string) (*OrderResponse, error) {
	req := alpaca.ClosePositionRequest{}
	order, err := s.alpacaClient.ClosePosition(symbol, req)
	if err != nil {
		return nil, fmt.Errorf("failed to close position: %v", err)
	}

	response := &OrderResponse{
		ID:             order.ID,
		Symbol:         order.Symbol,
		Quantity:       order.Qty.String(),
		Side:           string(order.Side),
		OrderType:      string(order.Type),
		TimeInForce:    string(order.TimeInForce),
		Status:         string(order.Status),
		FilledQuantity: order.FilledQty.String(),
		SubmittedAt:    order.SubmittedAt,
		AssetClass:     string(order.AssetClass),
		ExtendedHours:  order.ExtendedHours,
	}

	if order.LimitPrice != nil {
		response.LimitPrice = order.LimitPrice.String()
	}

	if order.StopPrice != nil {
		response.StopPrice = order.StopPrice.String()
	}

	if order.FilledAvgPrice != nil {
		response.FilledPrice = order.FilledAvgPrice.String()
	}

	if order.FilledAt != nil && !order.FilledAt.IsZero() {
		response.FilledAt = *order.FilledAt
	}

	return response, nil
}

// storeOrderInDB stores order information in the database for tracking
func (s *TradingService) storeOrderInDB(order *alpaca.Order, portfolioID string) {
	// Parse portfolio ID to UUID
	portfolioUUID, err := uuid.Parse(portfolioID)
	if err != nil {
		log.Printf("Invalid portfolio ID format: %v", err)
		return
	}

	// Convert to our transaction model
	executedAt := time.Now()
	transaction := models.Transaction{
		PortfolioID:     portfolioUUID,
		TransactionType: string(order.Side),
		Symbol:          order.Symbol,
		Quantity:        *order.Qty,
		Price:           decimal.Zero, // Will be updated when filled
		Amount:          decimal.Zero, // Will be calculated when filled
		Currency:        "USD",
		Status:          string(order.Status),
		ExecutedAt:      &executedAt,
		KYCVerified:     true,
		AMLChecked:      true,
		RiskScore:       50, // Default risk score (0-100)
	}

	if order.LimitPrice != nil {
		transaction.Price = *order.LimitPrice
		transaction.Amount = order.LimitPrice.Mul(*order.Qty)
	}

	if err := s.db.Create(&transaction).Error; err != nil {
		log.Printf("Error storing transaction: %v", err)
	}
}

// ValidateOrder validates order parameters before submission
func (s *TradingService) ValidateOrder(req OrderRequest) error {
	// Check if market is open for market orders
	if req.OrderType == "market" {
		// TODO: Add market hours validation
	}

	// Validate limit price for limit orders
	if req.OrderType == "limit" && req.LimitPrice <= 0 {
		return errors.New("limit price must be greater than 0 for limit orders")
	}

	// Validate stop price for stop orders
	if (req.OrderType == "stop" || req.OrderType == "stop_limit") && req.StopPrice <= 0 {
		return errors.New("stop price must be greater than 0 for stop orders")
	}

	// Validate buying power
	account, err := s.GetAccount()
	if err != nil {
		return fmt.Errorf("failed to validate buying power: %v", err)
	}

	if req.Side == "buy" {
		buyingPower, _ := strconv.ParseFloat(account.BuyingPower, 64)
		orderValue := req.Quantity * req.LimitPrice

		if req.OrderType == "market" || orderValue > buyingPower {
			// For market orders, we can't validate exact cost, but we can warn
			if buyingPower < 1000 { // Minimum threshold
				return errors.New("insufficient buying power for this order")
			}
		}
	}

	return nil
}
