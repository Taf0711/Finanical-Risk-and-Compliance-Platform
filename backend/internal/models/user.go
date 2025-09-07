package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	Email     string    `gorm:"unique;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	FirstName string    `gorm:"not null" json:"first_name"`
	LastName  string    `gorm:"not null" json:"last_name"`
	Role      string    `gorm:"not null;default:'analyst'" json:"role"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`

	// Alpaca Connect fields
	AlpacaAccountID    string    `gorm:"column:alpaca_account_id;index" json:"alpaca_account_id,omitempty"`
	AlpacaAccessToken  string    `gorm:"column:alpaca_access_token" json:"-"`
	AlpacaRefreshToken string    `gorm:"column:alpaca_refresh_token" json:"-"`
	AlpacaExpiresAt    time.Time `gorm:"column:alpaca_expires_at" json:"alpaca_expires_at,omitempty"`
	AlpacaIsLinked     bool      `gorm:"column:alpaca_is_linked;default:false" json:"alpaca_is_linked"`

	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	u.ID = uuid.New()
	return nil
}
