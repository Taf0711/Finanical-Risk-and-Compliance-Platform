// backend/internal/models/risk_thresholds.go
package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// RiskThresholds defines the risk limits for portfolios
type RiskThresholds struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	PortfolioID uuid.UUID `gorm:"type:uuid;not null" json:"portfolio_id"`

	// VaR Limits
	MaxVaR95 decimal.Decimal `gorm:"type:decimal(20,8)" json:"max_var_95"`
	MaxVaR99 decimal.Decimal `gorm:"type:decimal(20,8)" json:"max_var_99"`

	// Position Limits
	MaxPositionSize        decimal.Decimal `gorm:"type:decimal(10,4)" json:"max_position_size"` // % of portfolio
	MaxSingleAssetExposure decimal.Decimal `gorm:"type:decimal(10,4)" json:"max_single_asset_exposure"`
	MaxSectorExposure      decimal.Decimal `gorm:"type:decimal(10,4)" json:"max_sector_exposure"`

	// Risk Metrics Limits
	MinLiquidityRatio decimal.Decimal `gorm:"type:decimal(10,4)" json:"min_liquidity_ratio"`
	MaxLeverage       decimal.Decimal `gorm:"type:decimal(10,4)" json:"max_leverage"`
	MaxConcentration  decimal.Decimal `gorm:"type:decimal(10,4)" json:"max_concentration"`

	// Loss Limits
	MaxDailyLoss  decimal.Decimal `gorm:"type:decimal(10,4)" json:"max_daily_loss"` // % of portfolio
	MaxWeeklyLoss decimal.Decimal `gorm:"type:decimal(10,4)" json:"max_weekly_loss"`
	MaxDrawdown   decimal.Decimal `gorm:"type:decimal(10,4)" json:"max_drawdown"`

	// Stop Loss Rules
	RequireStopLoss     bool            `gorm:"default:true" json:"require_stop_loss"`
	MaxStopLossDistance decimal.Decimal `gorm:"type:decimal(10,4)" json:"max_stop_loss_distance"` // Max % from entry

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Portfolio Portfolio `gorm:"foreignKey:PortfolioID" json:"portfolio,omitempty"`
}

func (rt *RiskThresholds) BeforeCreate(tx *gorm.DB) error {
	rt.ID = uuid.New()
	return nil
}

// GetDefaultThresholds returns default risk thresholds for a new portfolio
func GetDefaultThresholds(portfolioID uuid.UUID) *RiskThresholds {
	return &RiskThresholds{
		PortfolioID:            portfolioID,
		MaxVaR95:               decimal.NewFromFloat(0.05), // 5% of portfolio
		MaxVaR99:               decimal.NewFromFloat(0.10), // 10% of portfolio
		MaxPositionSize:        decimal.NewFromFloat(0.25), // 25% max per position
		MaxSingleAssetExposure: decimal.NewFromFloat(0.30), // 30% max per asset
		MaxSectorExposure:      decimal.NewFromFloat(0.40), // 40% max per sector
		MinLiquidityRatio:      decimal.NewFromFloat(0.30), // 30% min liquidity
		MaxLeverage:            decimal.NewFromFloat(2.0),  // 2x leverage max
		MaxConcentration:       decimal.NewFromFloat(0.35), // 35% Herfindahl index
		MaxDailyLoss:           decimal.NewFromFloat(0.03), // 3% daily loss limit
		MaxWeeklyLoss:          decimal.NewFromFloat(0.07), // 7% weekly loss limit
		MaxDrawdown:            decimal.NewFromFloat(0.15), // 15% max drawdown
		RequireStopLoss:        true,
		MaxStopLossDistance:    decimal.NewFromFloat(0.05), // 5% max stop distance
	}
}
