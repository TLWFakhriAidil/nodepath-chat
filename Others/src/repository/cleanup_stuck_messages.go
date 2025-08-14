package repository

import (
	"time"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/sirupsen/logrus"
)

// CleanupStuckMessages resets messages that have been stuck in processing for too long
func CleanupStuckMessages() {
	db := database.GetDB()
	
	// Reset messages stuck in processing for more than 5 minutes
	result, err := db.Exec(`
		UPDATE broadcast_messages 
		SET processing_worker_id = NULL,
			processing_started_at = NULL,
			status = 'pending'
		WHERE status = 'processing'
		AND processing_started_at < DATE_SUB(NOW(), INTERVAL 5 MINUTE)
	`)
	
	if err != nil {
		logrus.Errorf("Failed to cleanup stuck messages: %v", err)
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		logrus.Infof("Reset %d stuck messages back to pending", rowsAffected)
	}
}

// StartCleanupWorker starts a background worker that cleans up stuck messages
func StartCleanupWorker() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			CleanupStuckMessages()
		}
	}
}
