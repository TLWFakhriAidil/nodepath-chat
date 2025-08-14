package usecase

import (
	"database/sql"
	"time"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/sirupsen/logrus"
)

// StartCampaignStatusMonitor monitors campaign progress and updates status
func StartCampaignStatusMonitor() {
	logrus.Info("Starting campaign status monitor...")
	
	ticker := time.NewTicker(10 * time.Second) // Check every 10 seconds
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			updateCampaignStatuses()
		}
	}
}

func updateCampaignStatuses() {
	db := database.GetDB()
	
	// Find ALL campaigns that need status check
	rows, err := db.Query(`
		SELECT DISTINCT c.id, c.title, c.status
		FROM campaigns c
		WHERE c.status IN ('pending', 'triggered', 'processing')
		AND (
			-- Has messages
			EXISTS (
				SELECT 1 FROM broadcast_messages bm 
				WHERE bm.campaign_id = c.id
			)
			OR 
			-- Should have been triggered by now
			(c.status = 'pending' AND c.scheduled_at <= CURRENT_TIMESTAMP)
		)
	`)
	if err != nil {
		logrus.Errorf("Failed to get campaigns for status update: %v", err)
		return
	}
	defer rows.Close()
	
	for rows.Next() {
		var campaignID int
		var title, currentStatus string
		if err := rows.Scan(&campaignID, &title, &currentStatus); err != nil {
			continue
		}
		
		// Get message statistics
		var total, pending, queued, sent, failed, skipped int
		var oldestQueuedMinutes sql.NullFloat64
		
		err := db.QueryRow(`
			SELECT COUNT(*) AS total,
				COUNT(CASE WHEN status = 'pending' THEN 1 END) AS pending,
				COUNT(CASE WHEN status = 'queued' THEN 1 END) AS queued,
				COUNT(CASE WHEN status = 'sent' THEN 1 END) AS sent,
				COUNT(CASE WHEN status = 'failed' THEN 1 END) AS failed,
				COUNT(CASE WHEN status = 'skipped' THEN 1 END) AS skipped,
				EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - MIN(CASE WHEN status = 'queued' THEN updated_at END)))/60 AS oldest_queued
			FROM broadcast_messages
			WHERE campaign_id = ?
		`, campaignID).Scan(&total, &pending, &queued, &sent, &failed, &skipped, &oldestQueuedMinutes)
		
		if err != nil {
			logrus.Errorf("Failed to get message stats for campaign %d: %v", campaignID, err)
			continue
		}
		
		// Check for stuck queued messages
		if queued > 0 && oldestQueuedMinutes.Valid && oldestQueuedMinutes.Float64 > 5 {
			logrus.Warnf("Campaign %d has %d messages stuck in queued state for %.0f minutes", 
				campaignID, queued, oldestQueuedMinutes.Float64)
		}
		
		// Get device count for this campaign
		var deviceCount int
		err = db.QueryRow(`
			SELECT COUNT(DISTINCT device_id) 
			FROM broadcast_messages 
			WHERE campaign_id = ?
		`, campaignID).Scan(&deviceCount)
		
		if err != nil {
			deviceCount = 1 // Default to 1 if error
		}
		
		// Determine new status
		var newStatus string
		
		if total == 0 && currentStatus == "pending" {
			// No messages created yet but should have been triggered
			var scheduledAt time.Time
			err = db.QueryRow(`SELECT scheduled_at FROM campaigns WHERE id = ?`, campaignID).Scan(&scheduledAt)
			if err == nil && scheduledAt.Before(time.Now()) {
				// Campaign should have been triggered but wasn't - mark as failed
				newStatus = "failed"
				logrus.Errorf("Campaign %d should have been triggered at %v but has no messages", 
					campaignID, scheduledAt)
			}
		} else if total == 0 {
			// No messages, keep current status
			continue
		} else if currentStatus == "triggered" && (sent > 0 || failed > 0) {
			// First worker has started processing
			newStatus = "processing"
		} else if pending == 0 && queued == 0 {
			// All messages have been processed
			if sent > 0 {
				newStatus = "finished"
			} else if failed == total || skipped == total {
				newStatus = "failed"
			} else {
				newStatus = "finished"
			}
		} else if currentStatus == "processing" && queued > 0 && oldestQueuedMinutes.Valid && oldestQueuedMinutes.Float64 > 10 {
			// Messages stuck in queue for too long - might indicate a problem
			logrus.Warnf("Campaign %d might be stuck - %d messages queued for over 10 minutes", 
				campaignID, queued)
		}
		
		// Update status if changed
		if newStatus != "" && newStatus != currentStatus {
			_, err = db.Exec(`
				UPDATE campaigns SET status = ?, updated_at = CURRENT_TIMESTAMP 
				WHERE id = ?
			`, newStatus, campaignID)
			
			if err != nil {
				logrus.Errorf("Failed to update campaign %d status to %s: %v", 
					campaignID, newStatus, err)
			} else {
				// Calculate progress percentage
				processed := sent + failed + skipped
				progress := 0
				if total > 0 {
					progress = (processed * 100) / total
				}
				
				logrus.Infof("Campaign '%s' (ID: %d) status: %s → %s | Progress: %d%% (%d/%d) | Devices: %d | Sent: %d, Failed: %d, Skipped: %d", 
					title, campaignID, currentStatus, newStatus, progress, processed, total, deviceCount, sent, failed, skipped)
				
				// If campaign finished or failed, unlock broadcast
				if newStatus == "finished" || newStatus == "failed" {
					coordinator := NewBroadcastCoordinator()
					var userID string
					err = db.QueryRow(`SELECT user_id FROM campaigns WHERE id = ?`, campaignID).Scan(&userID)
					if err == nil {
						coordinator.UnlockBroadcast(userID)
						logrus.Infof("Broadcast lock released for user %s after campaign %d %s", userID, campaignID, newStatus)
					}
				}
			}
		}
	}
}
