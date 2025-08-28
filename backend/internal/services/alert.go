package services

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/Taf0711/financial-risk-monitor/internal/database"
	"github.com/Taf0711/financial-risk-monitor/internal/models"
)

type AlertService struct {
	db *gorm.DB
}

func NewAlertService() *AlertService {
	return &AlertService{
		db: database.GetDB(),
	}
}

// CreateAlert creates a new alert
func (s *AlertService) CreateAlert(alert *models.Alert) error {
	return s.db.Create(alert).Error
}

// GetAlerts returns all alerts with optional filtering
func (s *AlertService) GetAlerts(status string, severity string, limit int) ([]models.Alert, error) {
	var alerts []models.Alert
	query := s.db.Preload("Portfolio", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, user_id, name, description, total_value, currency, created_at, updated_at")
	})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if severity != "" {
		query = query.Where("severity = ?", severity)
	}

	err := query.Order("created_at DESC").Limit(limit).Find(&alerts).Error
	return alerts, err
}

// GetAlertByID returns a specific alert by ID
func (s *AlertService) GetAlertByID(alertID uuid.UUID) (*models.Alert, error) {
	var alert models.Alert
	err := s.db.Preload("Portfolio", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, user_id, name, description, total_value, currency, created_at, updated_at")
	}).First(&alert, alertID).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// AcknowledgeAlert acknowledges an alert
func (s *AlertService) AcknowledgeAlert(alertID uuid.UUID, userID uuid.UUID) error {
	return s.db.Model(&models.Alert{}).Where("id = ?", alertID).Updates(map[string]interface{}{
		"status":          "ACKNOWLEDGED",
		"acknowledged_by": userID,
		"acknowledged_at": time.Now(),
		"updated_at":      time.Now(),
	}).Error
}

// ResolveAlert resolves an alert
func (s *AlertService) ResolveAlert(alertID uuid.UUID, userID uuid.UUID, resolution string) error {
	return s.db.Model(&models.Alert{}).Where("id = ?", alertID).Updates(map[string]interface{}{
		"status":      "RESOLVED",
		"resolved_by": userID,
		"resolved_at": time.Now(),
		"resolution":  resolution,
		"updated_at":  time.Now(),
	}).Error
}

// DeleteAlert deletes an alert
func (s *AlertService) DeleteAlert(alertID uuid.UUID) error {
	return s.db.Delete(&models.Alert{}, alertID).Error
}

// CreateRiskBreachAlert creates an alert for risk threshold breach
func (s *AlertService) CreateRiskBreachAlert(portfolioID uuid.UUID, metricType string, currentValue, threshold float64) error {
	var severity string
	breachRatio := currentValue / threshold

	switch {
	case breachRatio >= 2.0:
		severity = "CRITICAL"
	case breachRatio >= 1.2:
		severity = "HIGH"
	default:
		severity = "MEDIUM"
	}

	alert := &models.Alert{
		PortfolioID: portfolioID,
		AlertType:   "RISK_BREACH",
		Severity:    severity,
		Title:       fmt.Sprintf("%s Threshold Breached", metricType),
		Description: fmt.Sprintf("%s of %.2f exceeds threshold of %.2f (%.1f%% breach)", metricType, currentValue, threshold, (breachRatio-1)*100),
		Source:      fmt.Sprintf("%s_CALCULATOR", metricType),
		Status:      "ACTIVE",
		TriggeredBy: models.JSON{
			"metric_type":   metricType,
			"current_value": currentValue,
			"threshold":     threshold,
			"breach_ratio":  breachRatio,
		},
	}

	return s.CreateAlert(alert)
}

// CreateComplianceAlert creates an alert for compliance violations
func (s *AlertService) CreateComplianceAlert(portfolioID uuid.UUID, violationType string, details map[string]interface{}) error {
	var severity, title, description string

	switch violationType {
	case "POSITION_LIMIT":
		severity = "HIGH"
		title = "Position Limit Violation"
		description = "Single position exceeds maximum allowed percentage"
	case "KYC_AML":
		severity = "CRITICAL"
		title = "KYC/AML Violation"
		description = "Suspicious transaction activity detected"
	case "LIQUIDITY_RISK":
		severity = "MEDIUM"
		title = "Liquidity Risk Alert"
		description = "Portfolio liquidity below acceptable threshold"
	default:
		severity = "MEDIUM"
		title = "Compliance Alert"
		description = "Compliance rule violation detected"
	}

	alert := &models.Alert{
		PortfolioID: portfolioID,
		AlertType:   "COMPLIANCE_VIOLATION",
		Severity:    severity,
		Title:       title,
		Description: description,
		Source:      fmt.Sprintf("%s_CHECKER", violationType),
		Status:      "ACTIVE",
		TriggeredBy: models.JSON(details),
	}

	return s.CreateAlert(alert)
}

// CreateSuspiciousActivityAlert creates an alert for suspicious trading activity
func (s *AlertService) CreateSuspiciousActivityAlert(portfolioID uuid.UUID, activityType string, details map[string]interface{}) error {
	alert := &models.Alert{
		PortfolioID: portfolioID,
		AlertType:   "SUSPICIOUS_ACTIVITY",
		Severity:    "HIGH",
		Title:       "Suspicious Activity Detected",
		Description: fmt.Sprintf("Suspicious %s activity detected", activityType),
		Source:      "PATTERN_DETECTOR",
		Status:      "ACTIVE",
		TriggeredBy: models.JSON(details),
	}

	return s.CreateAlert(alert)
}

// GetActiveAlerts returns all active alerts
func (s *AlertService) GetActiveAlerts() ([]models.Alert, error) {
	return s.GetAlerts("ACTIVE", "", 100)
}

// GetAlertsByPortfolio returns alerts for a specific portfolio
func (s *AlertService) GetAlertsByPortfolio(portfolioID uuid.UUID) ([]models.Alert, error) {
	var alerts []models.Alert
	err := s.db.Where("portfolio_id = ?", portfolioID).Preload("Portfolio", func(db *gorm.DB) *gorm.DB {
		return db.Select("id, user_id, name, description, total_value, currency, created_at, updated_at")
	}).Order("created_at DESC").Find(&alerts).Error
	return alerts, err
}

// GetCriticalAlerts returns all critical severity alerts
func (s *AlertService) GetCriticalAlerts() ([]models.Alert, error) {
	return s.GetAlerts("", "CRITICAL", 50)
}

// CleanupOldAlerts removes old resolved alerts based on retention policy
func (s *AlertService) CleanupOldAlerts(daysToKeep int) error {
	cutoffDate := time.Now().AddDate(0, 0, -daysToKeep)
	return s.db.Where("status IN (?, ?) AND created_at < ?", "RESOLVED", "DISMISSED", cutoffDate).Delete(&models.Alert{}).Error
}

// GetAlertStats returns statistics about alerts
func (s *AlertService) GetAlertStats() (map[string]int, error) {
	var stats []struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
	}

	err := s.db.Model(&models.Alert{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&stats).Error

	if err != nil {
		return nil, err
	}

	result := make(map[string]int)
	for _, stat := range stats {
		result[stat.Status] = stat.Count
	}

	return result, nil
}
