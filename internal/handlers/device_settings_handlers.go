package handlers

import (
	"nodepath-chat/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// GetDeviceSettings retrieves all device settings
func (h *Handlers) GetDeviceSettings(c *fiber.Ctx) error {
	settings, err := h.deviceSettingsService.GetAll()
	if err != nil {
		logrus.WithError(err).Error("Failed to get device settings")
		return h.errorResponse(c, 500, "Failed to retrieve device settings")
	}

	return h.successResponse(c, settings)
}

// GetDeviceSettingsById retrieves a device setting by ID
func (h *Handlers) GetDeviceSettingsById(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return h.errorResponse(c, 400, "Device setting ID is required")
	}

	setting, err := h.deviceSettingsService.GetByID(id)
	if err != nil {
		logrus.WithError(err).Error("Failed to get device setting")
		if err.Error() == "device setting not found" {
			return h.errorResponse(c, 404, "Device setting not found")
		}
		return h.errorResponse(c, 500, "Failed to retrieve device setting")
	}

	return h.successResponse(c, setting)
}

// CreateDeviceSettings creates a new device setting
func (h *Handlers) CreateDeviceSettings(c *fiber.Ctx) error {
	var req models.CreateDeviceSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	// Validate required fields
	if req.IDDevice == "" {
		return h.errorResponse(c, 400, "ID Device is required")
	}
	if req.IDERP == "" {
		return h.errorResponse(c, 400, "ID ERP is required")
	}
	if req.IDAdmin == "" {
		return h.errorResponse(c, 400, "ID Admin is required")
	}
	// DeviceID is optional - it will be generated later if not provided

	setting, err := h.deviceSettingsService.Create(&req)
	if err != nil {
		logrus.WithError(err).Error("Failed to create device setting")
		return h.errorResponse(c, 500, "Failed to create device setting")
	}

	return h.successMessageResponse(c, "Device setting created successfully", setting)
}

// UpdateDeviceSettings updates an existing device setting
func (h *Handlers) UpdateDeviceSettings(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return h.errorResponse(c, 400, "Device setting ID is required")
	}

	var req models.UpdateDeviceSettingsRequest
	if err := c.BodyParser(&req); err != nil {
		return h.errorResponse(c, 400, "Invalid request body")
	}

	setting, err := h.deviceSettingsService.Update(id, &req)
	if err != nil {
		logrus.WithError(err).Error("Failed to update device setting")
		if err.Error() == "device setting not found" {
			return h.errorResponse(c, 404, "Device setting not found")
		}
		return h.errorResponse(c, 500, "Failed to update device setting")
	}

	return h.successMessageResponse(c, "Device setting updated successfully", setting)
}

// DeleteDeviceSettings deletes a device setting
func (h *Handlers) DeleteDeviceSettings(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return h.errorResponse(c, 400, "Device setting ID is required")
	}

	err := h.deviceSettingsService.Delete(id)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete device setting")
		if err.Error() == "device setting not found" {
			return h.errorResponse(c, 404, "Device setting not found")
		}
		return h.errorResponse(c, 500, "Failed to delete device setting")
	}

	return h.successMessageResponse(c, "Device setting deleted successfully", nil)
}