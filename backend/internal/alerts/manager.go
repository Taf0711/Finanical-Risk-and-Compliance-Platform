package alerts

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

type AlertManager struct {
	db          *gorm.DB
	redisClient *redis.Client
}

func NewAlertManager() *AlertManager {
	return &AlertManager{
		db:          database.GetDB(),
		redisClient: database.GetRedis(),
	}
}

// CreateAlert creates a new alert and stores it in both DB and Redis
func (am *AlertManager) CreateAlert(alert *models.Alert) error {
	// Save to database
	if err := am.db.Create(alert).Error; err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	// Cache in Redis for real-time access
	ctx := context.Background()
	alertJSON, err := json.Marshal(alert)
	if err != nil {
		return err
	}

	// Store in Redis with expiration
	key := fmt.Sprintf("alert:%s", alert.ID)
	am.redisClient.Set(ctx, key, alertJSON, 24*time.Hour)

	// Add to active alerts set
	am.redisClient.SAdd(ctx, "active_alerts", alert.ID.String())

	// Publish to WebSocket channel
	am.redisClient.Publish(ctx, "alerts_channel", alertJSON)

	return nil
}

// GetActiveAlerts retrieves all active alerts for a portfolio
func (am *AlertManager) GetActiveAlerts(portfolioID uuid.UUID) ([]models.Alert, error) {
	var alerts []models.Alert

	err := am.db.Where("portfolio_id = ? AND status = ?", portfolioID, "ACTIVE").
		Order("created_at DESC").
		Find(&alerts).Error

	return alerts, err
}

// AcknowledgeAlert marks an alert as acknowledged
func (am *AlertManager) AcknowledgeAlert(alertID, userID uuid.UUID) error {
	now := time.Now()

	err := am.db.Model(&models.Alert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":          "ACKNOWLEDGED",
			"acknowledged_by": userID,
			"acknowledged_at": now,
		}).Error

	if err != nil {
		return err
	}

	// Update Redis cache
	ctx := context.Background()
	am.redisClient.SRem(ctx, "active_alerts", alertID.String())

	return nil
}

// ResolveAlert marks an alert as resolved
func (am *AlertManager) ResolveAlert(alertID, userID uuid.UUID, resolution string) error {
	now := time.Now()

	err := am.db.Model(&models.Alert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":      "RESOLVED",
			"resolution":  resolution,
			"resolved_by": userID,
			"resolved_at": now,
		}).Error

	if err != nil {
		return err
	}

	// Remove from Redis
	ctx := context.Background()
	key := fmt.Sprintf("alert:%s", alertID)
	am.redisClient.Del(ctx, key)
	am.redisClient.SRem(ctx, "active_alerts", alertID.String())

	return nil
}

// CleanupOldAlerts removes alerts older than specified days
func (am *AlertManager) CleanupOldAlerts(days int) error {
	cutoff := time.Now().AddDate(0, 0, -days)

	return am.db.Where("created_at < ? AND status IN ?", cutoff, []string{"RESOLVED", "DISMISSED"}).
		Delete(&models.Alert{}).Error
}
