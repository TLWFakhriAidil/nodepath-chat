package rest

import (
	"fmt"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/rest/middleware"
	"github.com/gofiber/fiber/v2"
)

func (rest *App) CheckDeviceWorkerStatus(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Device ID is required",
		})
	}
	
	// Get user ID from context
	userID, ok := middleware.GetUserFromContext(c)
	if !ok {
		return c.JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
	}
	
	// Verify device ownership
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil || device.UserID != userID {
		return c.JSON(utils.ResponseData{
			Status:  403,
			Code:    "FORBIDDEN",
			Message: "Device not found or access denied",
		})
	}
	
	// Get broadcast manager
	manager := broadcast.GetBroadcastManager()
	
	// Get worker status
	status, exists := manager.GetWorkerStatus(deviceID)
	
	if !exists {
		return c.JSON(utils.ResponseData{
			Status:  200,
			Code:    "SUCCESS",
			Message: "Worker status",
			Results: map[string]interface{}{
				"device_id": deviceID,
				"device_name": device.DeviceName,
				"worker_exists": false,
				"status": "no_worker",
				"message": "No worker running for this device. Worker will start automatically when messages are queued.",
			},
		})
	}
	
	// Get current campaign/sequence info
	db := database.GetDB()
	var currentCampaign, currentSequence map[string]interface{}
	
	// Check for active campaign messages (with user validation)
	var campaignTitle, campaignStatus string
	var campaignID int
	err = db.QueryRow(`
		SELECT DISTINCT c.id, c.title, c.status 
		FROM broadcast_messages bm 
		JOIN campaigns c ON bm.campaign_id = c.id 
		WHERE bm.device_id = ? AND bm.status IN ('pending', 'processing') 
		AND c.user_id = ?
		ORDER BY bm.created_at DESC 
		LIMIT 1
	`, deviceID, userID).Scan(&campaignID, &campaignTitle, &campaignStatus)
	
	if err == nil {
		currentCampaign = map[string]interface{}{
			"id": campaignID,
			"name": campaignTitle,
			"status": campaignStatus,
		}
	}
	
	// Check for active sequence messages (with user validation)
	var sequenceID, sequenceName, sequenceStatus string
	err = db.QueryRow(`
		SELECT DISTINCT s.id, s.name, s.status 
		FROM broadcast_messages bm 
		JOIN sequences s ON bm.sequence_id = s.id 
		WHERE bm.device_id = ? AND bm.status IN ('pending', 'processing') 
		AND s.user_id = ?
		ORDER BY bm.created_at DESC 
		LIMIT 1
	`, deviceID, userID).Scan(&sequenceID, &sequenceName, &sequenceStatus)
	
	if err == nil {
		currentSequence = map[string]interface{}{
			"id": sequenceID,
			"name": sequenceName,
			"status": sequenceStatus,
		}
	}
	
	// Worker exists, return detailed status
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS", 
		Message: "Worker status",
		Results: map[string]interface{}{
			"device_id": deviceID,
			"worker_exists": true,
			"status": status.Status,
			"queue_size": status.QueueSize,
			"processed_count": status.ProcessedCount,
			"failed_count": status.FailedCount,
			"last_activity": status.LastActivity,
			"is_active": status.Status == "active" || status.Status == "processing",
			"current_campaign": currentCampaign,
			"current_sequence": currentSequence,
			"message": func() string {
				switch status.Status {
				case "active", "processing":
					return "✅ Worker is active and processing messages"
				case "idle":
					return "💤 Worker is idle, waiting for messages"
				case "error":
					return "❌ Worker encountered an error"
				default:
					return "❓ Worker status: " + status.Status
				}
			}(),
		},
	})
}

// CheckAllWorkersStatus returns status for all workers for user's devices only
func (rest *App) CheckAllWorkersStatus(c *fiber.Ctx) error {
	// Get user ID from context
	userID, ok := middleware.GetUserFromContext(c)
	if !ok {
		return c.JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "User not authenticated",
		})
	}
	
	// Get user's devices
	userRepo := repository.GetUserRepository()
	devices, err := userRepo.GetUserDevices(userID)
	if err != nil {
		return c.JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to get user devices",
		})
	}
	
	// Get broadcast manager
	manager := broadcast.GetBroadcastManager()
	
	// Get worker statuses for user's devices only
	statuses := make([]domainBroadcast.WorkerStatus, 0)
	for _, device := range devices {
		if status, exists := manager.GetWorkerStatus(device.ID); exists {
			statuses = append(statuses, status)
		}
	}
	
	// Count statistics
	totalWorkers := len(statuses)
	activeWorkers := 0
	idleWorkers := 0
	errorWorkers := 0
	totalQueued := 0
	totalProcessed := 0
	totalFailed := 0
	
	for _, status := range statuses {
		totalQueued += status.QueueSize
		totalProcessed += status.ProcessedCount
		totalFailed += status.FailedCount
		
		switch status.Status {
		case "active", "processing":
			activeWorkers++
		case "idle":
			idleWorkers++
		case "error":
			errorWorkers++
		}
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "All workers status",
		Results: map[string]interface{}{
			"summary": map[string]interface{}{
				"total_workers": totalWorkers,
				"active_workers": activeWorkers,
				"idle_workers": idleWorkers,
				"error_workers": errorWorkers,
				"total_queued": totalQueued,
				"total_processed": totalProcessed,
				"total_failed": totalFailed,
			},
			"workers": statuses,
			"message": func() string {
				if totalWorkers == 0 {
					return "No workers currently running"
				} else if activeWorkers > 0 {
					return fmt.Sprintf("✅ %d workers active out of %d total", activeWorkers, totalWorkers)
				} else {
					return fmt.Sprintf("💤 All %d workers are idle", totalWorkers)
				}
			}(),
		},
	})
}
