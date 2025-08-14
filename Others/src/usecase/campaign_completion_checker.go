package usecase

import (
	"time"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
)

// StartCampaignCompletionChecker monitors campaigns and updates their status correctly
func StartCampaignCompletionChecker() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()

	logrus.Info("Campaign completion checker started")

	for {
		select {
		case <-ticker.C:
			checkCampaignCompletions()
		}
	}
}

func checkCampaignCompletions() {
	db := database.GetDB()
	campaignRepo := repository.GetCampaignRepository()

	// Get all campaigns that are marked as 'triggered' 
	query := `
		SELECT DISTINCT c.id, c.title
		FROM campaigns c
		WHERE c.status = 'triggered'
	`

	rows, err := db.Query(query)
	if err != nil {
		logrus.Errorf("Failed to get triggered campaigns: %v", err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var campaignID int
		var title string
		if err := rows.Scan(&campaignID, &title); err != nil {
			continue
		}

		// Check if campaign has any pending messages
		var pendingCount int
		checkQuery := `
			SELECT COUNT(*) 
			FROM broadcast_messages 
			WHERE campaign_id = ? AND status = 'pending'
		`
		err := db.QueryRow(checkQuery, campaignID).Scan(&pendingCount)
		if err != nil {
			continue
		}

		// If no pending messages, check if it should be marked as completed
		if pendingCount == 0 {
			// Check total messages sent and failed
			var sentCount, failedCount int
			statsQuery := `
				SELECT COUNT(CASE WHEN status = 'sent' THEN 1 END) AS sent,
					COUNT(CASE WHEN status = 'failed' THEN 1 END) AS failed
				FROM broadcast_messages 
				WHERE campaign_id = ?
			`
			err := db.QueryRow(statsQuery, campaignID).Scan(&sentCount, &failedCount)
			if err != nil {
				continue
			}

			// Only mark as completed if there are actually sent/failed messages
			if sentCount > 0 || failedCount > 0 {
				if failedCount > 0 && sentCount == 0 {
					// All failed
					campaignRepo.UpdateCampaignStatus(campaignID, "failed")
					logrus.Infof("Campaign '%s' marked as failed: %d messages failed", title, failedCount)
				} else {
					// Completed
					campaignRepo.UpdateCampaignStatus(campaignID, "completed")
					logrus.Infof("Campaign '%s' marked as completed: %d sent, %d failed", title, sentCount, failedCount)
				}
			}
		} else {
			logrus.Debugf("Campaign '%s' still has %d pending messages", title, pendingCount)
		}
	}
}
