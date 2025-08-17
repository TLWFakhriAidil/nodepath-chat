package usecase

import (
	"time"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/sirupsen/logrus"
)

// StartQueuedMessageCleaner handles stuck messages in queued state
func StartQueuedMessageCleaner() {
	logrus.Info("Starting queued message cleaner...")
	
	ticker := time.NewTicker(60 * time.Second) // Check every minute
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			cleanStuckMessages()
		}
	}
}

func cleanStuckMessages() {
	db := database.GetDB()
	var rowsAffected int64
	
	// 1. Move old 'queued' messages back to 'pending' (stuck for more than 5 minutes)
	result, err := db.Exec(`
		UPDATE broadcast_messages 
		SET ` + "status" + ` = 'pending', 
		    updated_at = CURRENT_TIMESTAMP
		WHERE ` + "status" + ` = 'queued' 
		AND updated_at < DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 5 MINUTE)
	`)
	
	if err != nil {
		logrus.Errorf("Failed to reset stuck queued messages: %v", err)
	} else {
		rowsAffected, _ = result.RowsAffected()
		if rowsAffected > 0 {
			logrus.Infof("Reset %d stuck queued messages back to pending", rowsAffected)
		}
	}
	
	// 2. Reset timeout messages for retry
	result, err = db.Exec(`
		UPDATE broadcast_messages 
		SET ` + "status" + ` = 'pending', 
		    error_message = NULL,
		    updated_at = CURRENT_TIMESTAMP
		WHERE ` + "status" + ` = 'sent' 
		AND error_message LIKE '%timeout - could not be delivered within 12 hours%'
	`)
	
	if err != nil {
		logrus.Errorf("Failed to reset timeout messages: %v", err)
	} else {
		rowsAffected, _ = result.RowsAffected()
		if rowsAffected > 0 {
			logrus.Infof("Reset %d timeout messages for retry", rowsAffected)
		}
	}
	
	// 3. Clean messages that have been queued for more than 12 hours (as fallback)
	result, err = db.Exec(`
		UPDATE broadcast_messages SET ` + "status" + ` = 'failed', 
		    updated_at = CURRENT_TIMESTAMP,
		    error_message = 'Message timeout - could not be delivered within 12 hours'
		WHERE ` + "status" + ` = 'queued' 
		AND updated_at < DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 12 HOUR)
	`)
	
	if err != nil {
		logrus.Errorf("Failed to clean stuck messages: %v", err)
		return
	}
	
	rowsAffected, _ = result.RowsAffected()
	if rowsAffected > 0 {
		logrus.Warnf("Marked %d stuck queued messages as failed after 12 hour timeout", rowsAffected)
	}
	
	// Log current queue status
	var queuedCount int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM broadcast_messages 
		WHERE ` + "status" + ` = 'queued'
	`).Scan(&queuedCount)
	
	if err == nil && queuedCount > 0 {
		logrus.Debugf("Currently %d messages in queue", queuedCount)
	}
}
