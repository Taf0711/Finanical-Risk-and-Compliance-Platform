package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type RiskMetric struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key" json:"id"`
	PortfolioID     uuid.UUID       `gorm:"type:uuid;not null" json:"portfolio_id"`
	MetricType      string          `gorm:"not null" json:"metric_type"` // VAR, LIQUIDITY_RATIO, SHARPE_RATIO, etc.
	Value           decimal.Decimal `gorm:"type:decimal(20,8)" json:"value"`
	Threshold       decimal.Decimal `gorm:"type:decimal(20,8)" json:"threshold"`
	Status          string          `gorm:"not null" json:"status"` // SAFE, WARNING, CRITICAL
	CalculatedAt    time.Time       `json:"calculated_at"`
	TimeHorizon     int             `json:"time_horizon"`                              // in days
	ConfidenceLevel decimal.Decimal `gorm:"type:decimal(5,4)" json:"confidence_level"` // For VaR
	Details         JSON            `gorm:"type:jsonb" json:"details"`                 // Additional metric-specific data

	// Relations
	Portfolio Portfolio `gorm:"foreignKey:PortfolioID" json:"portfolio,omitempty"`
}

// Historical risk data for trend analysis
type RiskHistory struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key" json:"id"`
	PortfolioID uuid.UUID       `gorm:"type:uuid;not null" json:"portfolio_id"`
	MetricType  string          `gorm:"not null" json:"metric_type"`
	Value       decimal.Decimal `gorm:"type:decimal(20,8)" json:"value"`
	RecordedAt  time.Time       `json:"recorded_at"`
}

func (r *RiskMetric) BeforeCreate(tx *gorm.DB) error {
	r.ID = uuid.New()
	r.CalculatedAt = time.Now()
	return nil
}

func (r *RiskHistory) BeforeCreate(tx *gorm.DB) error {
	r.ID = uuid.New()
	return nil
}
