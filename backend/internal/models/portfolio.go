package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Portfolio struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key" json:"id"`
	UserID      uuid.UUID       `gorm:"type:uuid;not null" json:"user_id"`
	Name        string          `gorm:"not null" json:"name"`
	Description string          `json:"description"`
	TotalValue  decimal.Decimal `gorm:"type:decimal(20,2)" json:"total_value"`
	Currency    string          `gorm:"default:'USD'" json:"currency"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`

	// Relations
	User      User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Positions []Position `gorm:"foreignKey:PortfolioID" json:"positions,omitempty"`
}

type Position struct {
	ID           uuid.UUID       `gorm:"type:uuid;primary_key" json:"id"`
	PortfolioID  uuid.UUID       `gorm:"type:uuid;not null" json:"portfolio_id"`
	Symbol       string          `gorm:"not null" json:"symbol"`
	Quantity     decimal.Decimal `gorm:"type:decimal(20,8)" json:"quantity"`
	AveragePrice decimal.Decimal `gorm:"type:decimal(20,8)" json:"average_price"`
	CurrentPrice decimal.Decimal `gorm:"type:decimal(20,8)" json:"current_price"`
	MarketValue  decimal.Decimal `gorm:"type:decimal(20,2)" json:"market_value"`
	PnL          decimal.Decimal `gorm:"type:decimal(20,2)" json:"pnl"`
	PnLPercent   decimal.Decimal `gorm:"type:decimal(10,4)" json:"pnl_percent"`
	Weight       decimal.Decimal `gorm:"type:decimal(10,4)" json:"weight"` // Position weight in portfolio
	AssetType    string          `gorm:"not null" json:"asset_type"`       // STOCK, BOND, COMMODITY, etc.
	Liquidity    string          `gorm:"default:'HIGH'" json:"liquidity"`  // HIGH, MEDIUM, LOW
	UpdatedAt    time.Time       `json:"updated_at"`
}

func (p *Portfolio) BeforeCreate(tx *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}

func (p *Position) BeforeCreate(tx *gorm.DB) error {
	p.ID = uuid.New()
	return nil
}
