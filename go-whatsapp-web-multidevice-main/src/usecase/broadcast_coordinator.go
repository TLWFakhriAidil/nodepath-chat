package usecase

import (
	"fmt"
	"time"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/sirupsen/logrus"
)

// BroadcastCoordinator prevents overlap between campaigns and sequences
type BroadcastCoordinator struct {
	minGapMinutes int
}

// NewBroadcastCoordinator creates a new coordinator
func NewBroadcastCoordinator() *BroadcastCoordinator {
	return &BroadcastCoordinator{
		minGapMinutes: 30, // Minimum 30 minutes between any broadcasts
	}
}

// CanStartBroadcast checks if a new broadcast (campaign or sequence) can start
func (bc *BroadcastCoordinator) CanStartBroadcast(userID string, broadcastType string) (bool, string, error) {
	db := database.GetDB()
	
	// Check if any campaign is currently running
	var activeCampaigns int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM campaigns 
		WHERE user_id = ? 
		AND status IN ('triggered', 'processing')
	`, userID).Scan(&activeCampaigns)
	
	if err != nil {
		return false, "", err
	}
	
	if activeCampaigns > 0 {
		return false, fmt.Sprintf("Cannot start %s: %d campaign(s) currently running", broadcastType, activeCampaigns), nil
	}
	
	// Check if any sequence is actively sending
	var activeSequences int
	err = db.QueryRow(`
		SELECT COUNT(DISTINCT s.id) 
		FROM sequences s
		JOIN sequence_contacts sc ON s.id = sc.sequence_id
		WHERE s.user_id = ? 
		AND s.status = 'active'
		AND sc.status = 'active'
		AND EXISTS (
			SELECT 1 FROM broadcast_messages bm 
			WHERE bm.sequence_id = s.id::text 
			AND bm.status IN ('pending', 'queued')
			AND bm.created_at > DATE_SUB(NOW(), INTERVAL 30 MINUTE)
		)
	`, userID).Scan(&activeSequences)
	
	if err == nil && activeSequences > 0 {
		return false, fmt.Sprintf("Cannot start %s: %d sequence(s) currently sending messages", broadcastType, activeSequences), nil
	}
	
	// Check device availability
	var availableDevices int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM user_devices 
		WHERE user_id = ? 
		AND status IN ('connected', 'online')
	`, userID).Scan(&availableDevices)
	
	if err == nil && availableDevices == 0 {
		return false, "No connected devices available", nil
	}
	
	// Check recent broadcast completion
	var lastBroadcastTime *time.Time
	err = db.QueryRow(`
		SELECT MAX(last_activity) FROM (
			-- Last campaign activity
			SELECT MAX(COALESCE(updated_at, created_at)) AS last_activity
			FROM campaigns 
			WHERE user_id = ? 
			AND status IN ('finished', 'failed')
			
			UNION ALL
			
			-- Last sequence message activity  
			SELECT MAX(bm.updated_at) AS last_activity
			FROM broadcast_messages bm
			JOIN sequences s ON bm.sequence_id = s.id::text
			WHERE s.user_id = ?
			AND bm.status IN ('sent', 'failed')
		) recent_activity
	`, userID).Scan(&lastBroadcastTime)
	
	if err == nil && lastBroadcastTime != nil {
		timeSinceLastBroadcast := time.Since(*lastBroadcastTime)
		if timeSinceLastBroadcast < time.Duration(bc.minGapMinutes)*time.Minute {
			waitTime := time.Duration(bc.minGapMinutes)*time.Minute - timeSinceLastBroadcast
			return false, fmt.Sprintf("Please wait %d minutes before starting new %s", int(waitTime.Minutes()), broadcastType), nil
		}
	}
	
	return true, "", nil
}

// LockBroadcast prevents other broadcasts from starting
func (bc *BroadcastCoordinator) LockBroadcast(userID string, broadcastType string, broadcastID string) error {
	db := database.GetDB()
	
	// Create a broadcast lock record
	_, err := db.Exec(`
		INSERT INTO broadcast_locks(user_id, broadcast_type, broadcast_id, locked_at)
		VALUES (?, ?, ?, NOW())
		ON CONFLICT (user_id) DO UPDATE 
		SET broadcast_type = ?, broadcast_id = ?, locked_at = NOW()
	`, userID, broadcastType, broadcastID)
	
	if err != nil {
		// Table might not exist, try to create it
		_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS broadcast_locks (
				user_id VARCHAR(255) PRIMARY KEY,
				broadcast_type VARCHAR(50),
				broadcast_id VARCHAR(255),
				locked_at TIMESTAMP DEFAULT NOW()
			)
		`)
		
		// Retry the insert
		if err == nil {
			_, err = db.Exec(`
				INSERT INTO broadcast_locks(user_id, broadcast_type, broadcast_id, locked_at)
				VALUES (?, ?, ?, NOW())
			`, userID, broadcastType, broadcastID)
		}
	}
	
	return err
}

// UnlockBroadcast releases the broadcast lock
func (bc *BroadcastCoordinator) UnlockBroadcast(userID string) error {
	db := database.GetDB()
	
	_, err := db.Exec(`
		DELETE FROM broadcast_locks 
		WHERE user_id = ?
	`, userID)
	
	return err
}

// GetCurrentBroadcast returns info about currently running broadcast
func (bc *BroadcastCoordinator) GetCurrentBroadcast(userID string) (broadcastType string, broadcastID string, err error) {
	db := database.GetDB()
	
	err = db.QueryRow(`
		SELECT broadcast_type, broadcast_id 
		FROM broadcast_locks 
		WHERE user_id = ? 
		AND locked_at > DATE_SUB(NOW(), INTERVAL 2 HOUR)
	`, userID).Scan(&broadcastType, &broadcastID)
	
	return
}

// CleanupStaleLocks removes locks older than 2 hours
func (bc *BroadcastCoordinator) CleanupStaleLocks() {
	db := database.GetDB()
	
	result, err := db.Exec(`
		DELETE FROM broadcast_locks 
		WHERE locked_at < DATE_SUB(NOW(), INTERVAL 2 HOUR)
	`)
	
	if err == nil {
		if rowsAffected, _ := result.RowsAffected(); rowsAffected > 0 {
			logrus.Infof("Cleaned up %d stale broadcast locks", rowsAffected)
		}
	}
}

// StartBroadcastCoordinator starts the coordinator background process
func StartBroadcastCoordinator() {
	coordinator := NewBroadcastCoordinator()
	
	// Cleanup stale locks periodically
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			coordinator.CleanupStaleLocks()
		}
	}
}
