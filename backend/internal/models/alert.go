package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Alert struct {
	ID             uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	PortfolioID    uuid.UUID  `gorm:"type:uuid;not null" json:"portfolio_id"`
	AlertType      string     `gorm:"not null" json:"alert_type"` // RISK_BREACH, COMPLIANCE_VIOLATION, SUSPICIOUS_ACTIVITY
	Severity       string     `gorm:"not null" json:"severity"`   // LOW, MEDIUM, HIGH, CRITICAL
	Title          string     `gorm:"not null" json:"title"`
	Description    string     `json:"description"`
	Source         string     `json:"source"`                         // VAR_CALCULATOR, POSITION_LIMIT_CHECKER, AML_CHECKER, etc.
	Status         string     `gorm:"default:'ACTIVE'" json:"status"` // ACTIVE, ACKNOWLEDGED, RESOLVED, DISMISSED
	TriggeredBy    JSON       `gorm:"type:jsonb" json:"triggered_by"` // Details of what triggered the alert
	Resolution     string     `json:"resolution"`
	AcknowledgedBy *uuid.UUID `gorm:"type:uuid" json:"acknowledged_by"`
	AcknowledgedAt *time.Time `json:"acknowledged_at"`
	ResolvedBy     *uuid.UUID `gorm:"type:uuid" json:"resolved_by"`
	ResolvedAt     *time.Time `json:"resolved_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Relations
	Portfolio Portfolio `gorm:"foreignKey:PortfolioID" json:"portfolio,omitempty"`
}

func (a *Alert) BeforeCreate(tx *gorm.DB) error {
	a.ID = uuid.New()
	return nil
}
