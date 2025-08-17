package usecase

import (
	"time"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/sirupsen/logrus"
)

// StartMessageCleanupJob starts a job to clean up stuck messages
func StartMessageCleanupJob() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			cleanupStuckMessages()
		}
	}
}

// cleanupStuckMessages resets messages stuck in processing for more than 5 minutes
func cleanupStuckMessages() {
	db := database.GetDB()
	
	// Reset messages stuck in processing
	result, err := db.Exec(`
		UPDATE broadcast_messages 
		SET processing_worker_id = NULL,
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
