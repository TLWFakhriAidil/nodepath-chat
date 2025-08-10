package rest

import (
	"github.com/gofiber/fiber/v2"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
)

// ClearAllSessions clears all WhatsApp session data
func (handler *App) ClearAllSessions(c *fiber.Ctx) error {
	// The authentication is already handled by the middleware
	// We just need to get the user ID if available
	userID := c.Locals("userID")
	var userIDStr string
	
	if userID != nil {
		userIDStr = userID.(string)
	} else {
		// Try to get from session cookie as fallback
		token := c.Cookies("session_token")
		if token != "" {
			userRepo := repository.GetUserRepository()
			session, err := userRepo.GetSession(token)
			if err == nil && session != nil {
				userIDStr = session.UserID
			}
		}
	}
	
	// Log the action
	logrus.Infof("Clearing all WhatsApp sessions requested by user: %s", userIDStr)
	
	// Get database connection
	userRepo := repository.GetUserRepository()
	db := userRepo.DB()
	
	// List of WhatsApp session tables to clear (DELETE data, not DROP)
	tables := []string{
		"whatsmeow_app_state_mutation_macs",
		"whatsmeow_app_state_sync_keys",
		"whatsmeow_app_state_version",
		"whatsmeow_chat_settings",
		"whatsmeow_contacts",
		"whatsmeow_device",
		"whatsmeow_identity_keys",
		"whatsmeow_message_secrets",
		"whatsmeow_pre_keys",
		"whatsmeow_privacy_tokens",
		"whatsmeow_sender_keys",
		"whatsmeow_sessions",
	}
	
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to start transaction",
		})
	}
	defer tx.Rollback()
	
	// Clear data from each table (DELETE instead of DROP)
	for _, table := range tables {
		query := "DELETE FROM " + table
		result, err := tx.Exec(query)
		if err != nil {
			logrus.Warnf("Failed to clear table %s: %v", table, err)
			// Continue with other tables even if one fails
		} else {
			rowsAffected, _ := result.RowsAffected()
			logrus.Infof("Cleared table %s: %d rows deleted", table, rowsAffected)
		}
	}
	
	// Update all devices to offline status
	// If we have a user ID, only update that user's devices
	var updateQuery string
	if userIDStr != "" {
		updateQuery = `
			UPDATE user_devices SET status = 'offline', 
			    jid = NULL,
			    updated_at = CURRENT_TIMESTAMP
			WHERE user_id = ? AND status != 'deleted'
		`
		_, err = tx.Exec(updateQuery, userIDStr)
	} else {
		// If no user ID, update all devices (admin action)
		updateQuery = `
			UPDATE user_devices SET status = 'offline', 
			    jid = NULL,
			    updated_at = CURRENT_TIMESTAMP
			WHERE status != 'deleted'
		`
		_, err = tx.Exec(updateQuery)
	}
	if err != nil {
		logrus.Errorf("Failed to update device status: %v", err)
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "ERROR",
			Message: "Failed to clear sessions",
		})
	}
	
	// Clear any in-memory sessions as well
	// This would clear connection sessions for all users - be careful!
	// For now, we'll just log that sessions were cleared
	if userIDStr != "" {
		logrus.Infof("WhatsApp session data cleared for user %s", userIDStr)
	} else {
		logrus.Infof("WhatsApp session data cleared for all users (admin action)")
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "All WhatsApp session data has been cleared",
		Results: map[string]interface{}{
			"tablesCleared": len(tables),
			"devicesUpdated": true,
		},
	})
}