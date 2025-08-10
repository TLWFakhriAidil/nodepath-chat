package rest

import (
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// WebhookLeadRequest represents the incoming webhook payload
type WebhookLeadRequest struct {
	Name         string `json:"name"`
	Phone        string `json:"phone"`
	TargetStatus string `json:"target_status"`
	DeviceID     string `json:"device_id"`
	UserID       string `json:"user_id"`
	Niche        string `json:"niche"`
	Trigger      string `json:"trigger"`
	DeviceName   string `json:"device_name"`  // New field
	Platform     string `json:"platform"`     // New field
}

// InitWebhookLead initializes the webhook endpoint for creating leads
func InitWebhookLead(app *fiber.App) {
	// Public webhook endpoint (no auth middleware)
	app.Post("/webhook/lead/create", CreateLeadWebhook)
}

// CreateLeadWebhook handles the webhook request to create a lead
func CreateLeadWebhook(c *fiber.Ctx) error {
	// Parse request body
	var request WebhookLeadRequest
	if err := c.BodyParser(&request); err != nil {
		logrus.Error("Webhook Lead: Failed to parse request body - ", err)
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
			Results: map[string]interface{}{
				"error": err.Error(),
			},
		})
	}

	// Log the incoming request for debugging
	logrus.Info("Webhook Lead: Received request - ", request)
	
	// Create a unique key based on phone + user_id to prevent duplicates
	// This helps if the webhook is called multiple times quickly
	dedupeKey := request.Phone + "_" + request.UserID
	logrus.Info("Webhook Lead: Processing request with dedupe key: ", dedupeKey)

	// Basic validation - only check required fields
	if request.Name == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "VALIDATION_ERROR",
			Message: "Name is required",
		})
	}
	
	if request.Phone == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "VALIDATION_ERROR",
			Message: "Phone is required",
		})
	}
	
	if request.UserID == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "VALIDATION_ERROR",
			Message: "User ID is required",
		})
	}
	
	if request.DeviceID == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "VALIDATION_ERROR",
			Message: "Device ID is required",
		})
	}

	// Determine device name first
	deviceName := request.DeviceName
	if deviceName == "" {
		// Generate device name based on device_id
		if _, uuidErr := uuid.Parse(request.DeviceID); uuidErr != nil {
			// Non-UUID, use first 6 chars as prefix
			prefix := request.DeviceID
			if len(prefix) > 6 {
				prefix = prefix[:6]
			}
			deviceName = "Device-" + prefix
		} else {
			// UUID device_id, create timestamp-based name
			deviceName = "Device-" + time.Now().Format("20060102150405")
		}
	}
	
	logrus.Info("Webhook Lead: Using device name: ", deviceName)
	
	// First check if a device with this user_id and device_name already exists
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByUserAndName(request.UserID, deviceName)
	
	// If device doesn't exist by name, check by JID as fallback
	if err != nil || device == nil {
		device, err = userRepo.GetDeviceByUserAndJID(request.UserID, request.DeviceID)
	}
	
	// If device exists, update JID; otherwise create new
	if device != nil {
		// Device exists - just update the JID
		logrus.Info("Webhook Lead: Device found, updating JID from ", device.JID, " to ", request.DeviceID)
		
		updateQuery := `
			UPDATE user_devices 
			SET jid = ?, platform = ?, updated_at = ?, status = 'online'
			WHERE id = ?
		`
		_, err = userRepo.GetDB().Exec(updateQuery, request.DeviceID, request.Platform, time.Now(), device.ID)
		if err != nil {
			logrus.Error("Webhook Lead: Failed to update device JID - ", err)
		} else {
			logrus.Info("Webhook Lead: Successfully updated device JID")
		}
	} else {
		// Device doesn't exist, create it
		logrus.Info("Webhook Lead: Device not found, creating new device - ", request.DeviceID)
		
		// Handle device ID based on whether it's a valid UUID or not
		var deviceID string
		deviceJID := request.DeviceID // JID always stores the full original
		
		if _, err := uuid.Parse(request.DeviceID); err != nil {
			// Not a valid UUID, generate new UUID
			deviceID = uuid.New().String()
			logrus.Info("Webhook Lead: Non-UUID device_id. Generated UUID: ", deviceID)
		} else {
			// It's a valid UUID, use it
			deviceID = request.DeviceID
			logrus.Info("Webhook Lead: Valid UUID device_id, using as-is")
		}
		
		// Create new device
		newDevice := &models.UserDevice{
			ID:               deviceID,
			UserID:           request.UserID,
			DeviceName:       deviceName,
			Phone:            "", // null/empty
			JID:              deviceJID, // Always the full original device_id
			Status:           "online",
			LastSeen:         time.Now(),
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
			MinDelaySeconds:  5,
			MaxDelaySeconds:  15,
			Platform:         request.Platform,
		}
		
		// Create device in database
		err = userRepo.CreateDevice(newDevice)
		if err != nil {
			logrus.Error("Webhook Lead: Failed to create device - ", err)
			return c.Status(500).JSON(utils.ResponseData{
				Status:  500,
				Code:    "DEVICE_CREATE_FAILED",
				Message: "Failed to create device",
				Results: map[string]interface{}{
					"error": err.Error(),
					"hint": "Check if device with same name already exists for this user",
				},
			})
		}
		
		logrus.Info("Webhook Lead: Device created successfully - ID: ", newDevice.ID, ", JID: ", newDevice.JID)
		device = newDevice
	}

	// Create lead object with direct field mapping
	lead := &models.Lead{
		ID:           uuid.New().String(),
		Name:         request.Name,
		Phone:        request.Phone,
		Niche:        request.Niche,
		Trigger:      request.Trigger,
		TargetStatus: request.TargetStatus,
		DeviceID:     device.ID, // Use the actual device ID (UUID)
		UserID:       request.UserID,
		Platform:     request.Platform, // Add platform field
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Check if lead with same device_id, user_id, phone AND niche already exists
	leadRepo := repository.GetLeadRepository()
	existingLead, err := leadRepo.GetLeadByDeviceUserPhoneNiche(device.ID, request.UserID, request.Phone, request.Niche)
	if err == nil && existingLead != nil {
		// Lead already exists with same device, user, phone AND niche - skip creation
		logrus.Info("Webhook Lead: Lead already exists with same device_id, user_id, phone, and niche. Skipping creation.")
		
		// Return success but indicate it was a duplicate
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "DUPLICATE_SKIPPED",
			Message: "Lead already exists with same device, user, phone, and niche",
			Results: map[string]interface{}{
				"lead_id":       existingLead.ID,
				"name":          existingLead.Name,
				"phone":         existingLead.Phone,
				"niche":         existingLead.Niche,
				"trigger":       existingLead.Trigger,
				"target_status": existingLead.TargetStatus,
				"device_id":     device.ID,
				"device_jid":    device.JID,
				"device_name":   device.DeviceName,
				"user_id":       existingLead.UserID,
				"platform":      existingLead.Platform,
				"duplicate":     true,
				"created_at":    existingLead.CreatedAt,
				"updated_at":    existingLead.UpdatedAt,
			},
		})
	}

	// No duplicate found, create new lead
	err = leadRepo.CreateLead(lead)
	if err != nil {
		logrus.Error("Webhook Lead: Failed to create lead - ", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "CREATE_FAILED",
			Message: "Failed to create lead",
			Results: map[string]interface{}{
				"error": err.Error(),
			},
		})
	}

	logrus.Info("Webhook Lead: Successfully created lead - ", lead.ID)

	// Return success response with all the data that was saved
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Lead created successfully",
		Results: map[string]interface{}{
			"lead_id":       lead.ID,
			"name":          lead.Name,
			"phone":         lead.Phone,
			"niche":         lead.Niche,
			"trigger":       lead.Trigger,
			"target_status": lead.TargetStatus,
			"device_id":     device.ID, // Return the actual UUID used
			"device_jid":    device.JID, // Return the JID (original device_id)
			"user_id":       lead.UserID,
			"platform":      lead.Platform,
			"device_created": device.CreatedAt.After(time.Now().Add(-1 * time.Minute)), // Device created in last minute
			"created_at":    lead.CreatedAt,
		},
	})
}
