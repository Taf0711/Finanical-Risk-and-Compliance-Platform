package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// RiskMetric represents calculated risk metrics for a portfolio
type RiskMetric struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key" json:"id"`
	PortfolioID     uuid.UUID       `gorm:"type:uuid;not null" json:"portfolio_id"`
	MetricType      string          `gorm:"type:varchar(50);not null" json:"metric_type"`
	Value           decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"value"`
	Threshold       decimal.Decimal `gorm:"type:decimal(20,8)" json:"threshold"`
	Status          string          `gorm:"type:varchar(20);not null" json:"status"`
	CalculatedAt    time.Time       `gorm:"default:CURRENT_TIMESTAMP" json:"calculated_at"`
	TimeHorizon     int             `json:"time_horizon"`
	ConfidenceLevel decimal.Decimal `gorm:"type:decimal(5,4)" json:"confidence_level"`
	Details         JSON            `gorm:"type:jsonb" json:"details"`

	// Relationships
	Portfolio Portfolio `gorm:"foreignKey:PortfolioID" json:"portfolio,omitempty"`
}

func (r *RiskMetric) BeforeCreate(tx *gorm.DB) error {
	r.ID = uuid.New()
	return nil
}

// RiskHistory represents historical risk data
type RiskHistory struct {
	ID          uuid.UUID       `gorm:"type:uuid;primary_key" json:"id"`
	PortfolioID uuid.UUID       `gorm:"type:uuid;not null" json:"portfolio_id"`
	MetricType  string          `gorm:"type:varchar(50);not null" json:"metric_type"`
	Value       decimal.Decimal `gorm:"type:decimal(20,8);not null" json:"value"`
	RecordedAt  time.Time       `gorm:"default:CURRENT_TIMESTAMP" json:"recorded_at"`

	// Relationships
	Portfolio Portfolio `gorm:"foreignKey:PortfolioID" json:"portfolio,omitempty"`
}

func (r *RiskHistory) BeforeCreate(tx *gorm.DB) error {
	r.ID = uuid.New()
	return nil
}

// TradeRiskAnalysis represents risk analysis for a potential or executed trade
type TradeRiskAnalysis struct {
	TradeID  string  `json:"trade_id"`
	Symbol   string  `json:"symbol"`
	Side     string  `json:"side"` // BUY, SELL
	Quantity float64 `json:"quantity"`
	Price    float64 `json:"price"`

	// Risk Metrics
	PositionRisk        float64 `json:"position_risk"`    // $ at risk for this position
	PortfolioImpact     float64 `json:"portfolio_impact"` // % impact on portfolio
	ConcentrationImpact float64 `json:"concentration_impact"`
	LiquidityImpact     float64 `json:"liquidity_impact"`

	// Risk Checks
	Violations     []RiskViolation `json:"violations"`
	RiskScore      float64         `json:"risk_score"` // 0-100, higher = riskier
	Approved       bool            `json:"approved"`
	RequiresReview bool            `json:"requires_review"`

	// Recommendations
	SuggestedStopLoss   float64 `json:"suggested_stop_loss,omitempty"`
	SuggestedSize       float64 `json:"suggested_size,omitempty"`
	HedgeRecommendation string  `json:"hedge_recommendation,omitempty"`
}

// RiskViolation represents a specific risk limit breach
type RiskViolation struct {
	Type         string  `json:"type"`     // POSITION_SIZE, VAR_LIMIT, CONCENTRATION, etc.
	Severity     string  `json:"severity"` // WARNING, VIOLATION, CRITICAL
	Description  string  `json:"description"`
	CurrentValue float64 `json:"current_value"`
	Limit        float64 `json:"limit"`
	Impact       float64 `json:"impact"` // % over limit
}
