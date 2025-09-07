package services

import (
	"log"

	"github.com/Taf0711/financial-risk-monitor/internal/websocket"
)

// RealtimeAlertService provides real-time alert functionality
type RealtimeAlertService struct {
	hub            *websocket.SimpleHub
	tradingService *TradingService
	riskService    *AdvancedRiskService
	isRunning      bool
}

// NewRealtimeAlertService creates a new real-time alert service
func NewRealtimeAlertService(hub *websocket.SimpleHub, tradingService *TradingService, riskService *AdvancedRiskService) *RealtimeAlertService {
	return &RealtimeAlertService{
		hub:            hub,
		tradingService: tradingService,
		riskService:    riskService,
		isRunning:      false,
	}
}

// Start begins the real-time alert monitoring
func (s *RealtimeAlertService) Start() {
	if s.isRunning {
		return
	}

	s.isRunning = true
	log.Println("RealtimeAlertService: Starting real-time alert monitoring")

	// Start monitoring in a goroutine
	go s.monitorAlerts()
}

// Stop stops the real-time alert monitoring
func (s *RealtimeAlertService) Stop() {
	s.isRunning = false
	log.Println("RealtimeAlertService: Stopping real-time alert monitoring")
}

// monitorAlerts runs the main monitoring loop
func (s *RealtimeAlertService) monitorAlerts() {
	// This would contain the main monitoring logic
	// For now, it's just a placeholder
	log.Println("RealtimeAlertService: Alert monitoring loop started")
}

// SendAlert sends an alert to connected WebSocket clients
func (s *RealtimeAlertService) SendAlert(alertType string, message string, data interface{}) error {
	alert := map[string]interface{}{
		"type":    "alert",
		"subtype": alertType,
		"message": message,
		"data":    data,
	}

	return s.hub.BroadcastToAll(alert)
}

// IsRunning returns whether the service is currently running
func (s *RealtimeAlertService) IsRunning() bool {
	return s.isRunning
}
