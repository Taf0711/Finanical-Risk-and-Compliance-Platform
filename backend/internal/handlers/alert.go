package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"github.com/Taf0711/financial-risk-monitor/internal/services"
)

type AlertHandler struct {
	alertService *services.AlertService
}

func NewAlertHandler() *AlertHandler {
	return &AlertHandler{
		alertService: services.NewAlertService(),
	}
}

// GetAlerts returns all alerts
func (h *AlertHandler) GetAlerts(c *fiber.Ctx) error {
	status := c.Query("status", "")
	severity := c.Query("severity", "")
	limit := 100 // Default limit

	alerts, err := h.alertService.GetAlerts(status, severity, limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve alerts",
		})
	}

	return c.JSON(alerts)
}

// GetActiveAlerts returns only active alerts
func (h *AlertHandler) GetActiveAlerts(c *fiber.Ctx) error {
	alerts, err := h.alertService.GetActiveAlerts()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve active alerts",
		})
	}

	return c.JSON(alerts)
}

// GetAlert returns a specific alert
func (h *AlertHandler) GetAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	alertUUID, err := uuid.Parse(alertID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid alert ID",
		})
	}

	alert, err := h.alertService.GetAlertByID(alertUUID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Alert not found",
		})
	}

	return c.JSON(alert)
}

// AcknowledgeAlert acknowledges an alert
func (h *AlertHandler) AcknowledgeAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	alertUUID, err := uuid.Parse(alertID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid alert ID",
		})
	}

	userID := c.Locals("user_id").(string)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	err = h.alertService.AcknowledgeAlert(alertUUID, userUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to acknowledge alert",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Alert acknowledged successfully",
	})
}

// ResolveAlert resolves an alert
func (h *AlertHandler) ResolveAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	alertUUID, err := uuid.Parse(alertID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid alert ID",
		})
	}

	userID := c.Locals("user_id").(string)
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req struct {
		Resolution string `json:"resolution"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	err = h.alertService.ResolveAlert(alertUUID, userUUID, req.Resolution)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to resolve alert",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Alert resolved successfully",
	})
}

// DeleteAlert deletes an alert
func (h *AlertHandler) DeleteAlert(c *fiber.Ctx) error {
	alertID := c.Params("id")
	alertUUID, err := uuid.Parse(alertID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid alert ID",
		})
	}

	err = h.alertService.DeleteAlert(alertUUID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete alert",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Alert deleted successfully",
	})
}
