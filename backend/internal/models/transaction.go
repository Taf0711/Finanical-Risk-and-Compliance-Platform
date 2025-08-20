package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Transaction struct {
	ID              uuid.UUID       `gorm:"type:uuid;primary_key" json:"id"`
	PortfolioID     uuid.UUID       `gorm:"type:uuid;not null" json:"portfolio_id"`
	TransactionType string          `gorm:"not null" json:"transaction_type"` // BUY, SELL, DEPOSIT, WITHDRAWAL
	Symbol          string          `json:"symbol"`
	Quantity        decimal.Decimal `gorm:"type:decimal(20,8)" json:"quantity"`
	Price           decimal.Decimal `gorm:"type:decimal(20,8)" json:"price"`
	Amount          decimal.Decimal `gorm:"type:decimal(20,2)" json:"amount"`
	Currency        string          `gorm:"default:'USD'" json:"currency"`
	Status          string          `gorm:"default:'PENDING'" json:"status"` // PENDING, COMPLETED, FAILED, CANCELLED
	ExecutedAt      *time.Time      `json:"executed_at"`
	Notes           string          `json:"notes"`

	// Compliance fields
	KYCVerified     bool   `gorm:"default:false" json:"kyc_verified"`
	AMLChecked      bool   `gorm:"default:false" json:"aml_checked"`
	RiskScore       int    `json:"risk_score"` // 0-100
	ComplianceNotes string `json:"compliance_notes"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Portfolio Portfolio `gorm:"foreignKey:PortfolioID" json:"portfolio,omitempty"`
}

func (t *Transaction) BeforeCreate(tx *gorm.DB) error {
	t.ID = uuid.New()
	return nil
}
