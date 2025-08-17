package main

// Fix for sequence duplicate messages
// This fix addresses two main issues:
// 1. Race condition in QueueMessage - multiple processes checking at same time
// 2. ProcessDailySequenceMessages creating duplicates when run concurrently

// SOLUTION 1: Update QueueMessage to use transaction with row locking
// In broadcast_repository.go, replace the QueueMessage function:

/*
func (r *BroadcastRepository) QueueMessage(msg domainBroadcast.BroadcastMessage) error {
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	
	// START TRANSACTION for atomic operation
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// For SEQUENCES: Use FOR UPDATE to lock rows during check
	if msg.SequenceStepID != nil && *msg.SequenceStepID != "" {
		duplicateCheck := `
			SELECT COUNT(*) 
			FROM broadcast_messages 
			WHERE sequence_stepid = ? 
			AND recipient_phone = ? 
			AND device_id = ?
			AND status IN ('pending', 'sent', 'queued', 'processing')
			FOR UPDATE
		`
		
		var count int
		err := tx.QueryRow(duplicateCheck, *msg.SequenceStepID, msg.RecipientPhone, msg.DeviceID).Scan(&count)
		if err != nil {
			logrus.Warnf("Error checking sequence duplicates: %v", err)
		} else if count > 0 {
			logrus.Infof("Skipping duplicate sequence message for %s - sequence_step %s already exists", 
				msg.RecipientPhone, *msg.SequenceStepID)
			return nil // Skip duplicate
		}
	}
	
	// ... rest of the function using tx instead of r.db ...
	
	// At the end, commit the transaction
	return tx.Commit()
}
*/

// SOLUTION 2: Add distributed lock for ProcessDailySequenceMessages
// This prevents multiple instances from processing same sequence simultaneously

/*
import (
	"sync"
	"time"
)

var sequenceProcessingMutex sync.Mutex
var sequenceLastProcessed = make(map[string]time.Time)

func (cts *CampaignTriggerService) ProcessDailySequenceMessages() error {
	// Prevent concurrent execution
	sequenceProcessingMutex.Lock()
	defer sequenceProcessingMutex.Unlock()
	
	// Check if we recently processed (within last 30 seconds)
	lastRun, exists := sequenceLastProcessed["daily"]
	if exists && time.Since(lastRun) < 30*time.Second {
		logrus.Info("Skipping ProcessDailySequenceMessages - ran recently")
		return nil
	}
	sequenceLastProcessed["daily"] = time.Now()
	
	// ... rest of the function ...
}
*/

// SOLUTION 3: Use INSERT IGNORE with unique key
// Add this to your database:
/*
ALTER TABLE broadcast_messages 
ADD UNIQUE KEY unique_sequence_step_message (
    sequence_stepid, 
    recipient_phone, 
    device_id
);
*/

// Then modify the INSERT to use INSERT IGNORE:
/*
query := `
	INSERT IGNORE INTO broadcast_messages(id, user_id, device_id, campaign_id, sequence_id, sequence_stepid, recipient_phone, recipient_name,
	 message_type, content, media_url, status, scheduled_at, created_at, group_id, group_order)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`
*/
