package main

// COMPREHENSIVE FIX FOR SEQUENCES AND CAMPAIGNS
// This ensures A-Z flow works perfectly without duplicates

// 1. FIX QueueMessage to include 'processing' status in duplicate checks
// In src/repository/broadcast_repository.go:

/*
// For SEQUENCES: Check based on sequence_stepid, recipient_phone, device_id
if msg.SequenceStepID != nil && *msg.SequenceStepID != "" {
    duplicateCheck := `
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE sequence_stepid = ? 
        AND recipient_phone = ? 
        AND device_id = ?
        AND status IN ('pending', 'sent', 'queued', 'processing')  // ADD 'processing'
    `
    
    var count int
    err := r.db.QueryRow(duplicateCheck, *msg.SequenceStepID, msg.RecipientPhone, msg.DeviceID).Scan(&count)
    if err != nil {
        logrus.Warnf("Error checking sequence duplicates: %v", err)
    } else if count > 0 {
        logrus.Infof("Skipping duplicate sequence message for %s - sequence_step %s already exists", 
            msg.RecipientPhone, *msg.SequenceStepID)
        return nil // Skip duplicate
    }
}

// For CAMPAIGNS: Check based on campaign_id, recipient_phone, device_id
if msg.CampaignID != nil && *msg.CampaignID > 0 {
    duplicateCheck := `
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE campaign_id = ? 
        AND recipient_phone = ? 
        AND device_id = ?
        AND status IN ('pending', 'sent', 'queued', 'processing')  // ADD 'processing'
    `
    
    var count int
    err := r.db.QueryRow(duplicateCheck, *msg.CampaignID, msg.RecipientPhone, msg.DeviceID).Scan(&count)
    if err != nil {
        logrus.Warnf("Error checking campaign duplicates: %v", err)
    } else if count > 0 {
        logrus.Infof("Skipping duplicate campaign message for %s - campaign %d already exists", 
            msg.RecipientPhone, *msg.CampaignID)
        return nil // Skip duplicate
    }
}
*/

// 2. ENSURE GetPendingMessagesAndLock is used everywhere
// Check these files use GetPendingMessagesAndLock NOT GetPendingMessages:
// - src/usecase/optimized_broadcast_processor.go ✓ (just fixed)
// - src/usecase/broadcast_worker_processor.go (if exists)

// 3. ADD UNIQUE CONSTRAINTS AT DATABASE LEVEL (CRITICAL!)
/*
-- For sequences: unique on (sequence_stepid, recipient_phone, device_id)
ALTER TABLE broadcast_messages 
ADD UNIQUE INDEX IF NOT EXISTS unique_sequence_message (
    sequence_stepid, 
    recipient_phone, 
    device_id
);

-- For campaigns: unique on (campaign_id, recipient_phone, device_id)
ALTER TABLE broadcast_messages 
ADD UNIQUE INDEX IF NOT EXISTS unique_campaign_message (
    campaign_id, 
    recipient_phone, 
    device_id
);
*/

// 4. FIX ProcessDailySequenceMessages duplicate check
// In src/usecase/campaign_trigger.go, the duplicate check should include 'processing':
/*
err = db.QueryRow(`
    SELECT COUNT(*) FROM broadcast_messages 
    WHERE sequence_stepid = ? 
    AND recipient_phone = ? 
    AND device_id = ?
    AND status IN ('pending', 'processing', 'queued', 'sent')  // Include all statuses
`, nextStep.ID, contact.ContactPhone, device.ID).Scan(&existingCount)
*/

// 5. ENSURE WORKER ID IS BEING SET
// The GetPendingMessagesAndLock function must be called, not GetPendingMessages
