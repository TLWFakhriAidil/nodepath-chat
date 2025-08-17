# Sequence Duplicate Messages Fix

## Problem
Sequences are sending duplicate messages to the same recipient. The same message is being sent multiple times with different greetings.

## Root Cause Analysis

1. **No Worker ID Usage**: The `processing_worker_id` is NULL for all messages, meaning the atomic locking mechanism is not being used
2. **Race Condition**: Multiple processes can check for duplicates simultaneously, both find none, then both create messages
3. **Duplicate Check Not Atomic**: The current duplicate check in `QueueMessage` is not atomic

## Solution

### 1. Database-Level Fix (IMMEDIATE)

Add a unique constraint to prevent duplicates at the database level:

```sql
-- Remove existing duplicates first
DELETE bm1 FROM broadcast_messages bm1
INNER JOIN broadcast_messages bm2 
WHERE bm1.sequence_stepid = bm2.sequence_stepid 
AND bm1.recipient_phone = bm2.recipient_phone 
AND bm1.device_id = bm2.device_id
AND bm1.created_at > bm2.created_at;

-- Add unique constraint
ALTER TABLE broadcast_messages 
ADD UNIQUE INDEX unique_sequence_message (
    sequence_stepid, 
    recipient_phone, 
    device_id
);
```

### 2. Code-Level Fix

Update `QueueMessage` in `broadcast_repository.go` to use transactions:

```go
func (r *BroadcastRepository) QueueMessage(msg domainBroadcast.BroadcastMessage) error {
    if msg.ID == "" {
        msg.ID = uuid.New().String()
    }
    
    // Use transaction for atomic operation
    tx, err := r.db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Check for duplicates with row locking
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
            return nil
        }
    }
    
    // ... rest of the function using tx instead of r.db ...
    
    // Commit the transaction
    return tx.Commit()
}
```

### 3. Prevent Concurrent Processing

Add a check in `ProcessDailySequenceMessages` to prevent concurrent execution:

```go
var sequenceProcessingLock sync.Mutex
var lastSequenceProcess time.Time

func (cts *CampaignTriggerService) ProcessDailySequenceMessages() error {
    // Prevent concurrent execution
    if !sequenceProcessingLock.TryLock() {
        logrus.Info("ProcessDailySequenceMessages already running, skipping")
        return nil
    }
    defer sequenceProcessingLock.Unlock()
    
    // Check if we recently processed
    if time.Since(lastSequenceProcess) < 30*time.Second {
        logrus.Info("ProcessDailySequenceMessages ran recently, skipping")
        return nil
    }
    lastSequenceProcess = time.Now()
    
    // ... rest of the function ...
}
```

## Testing

After applying the fix:
1. Check that no duplicate messages are created
2. Monitor the logs for "Skipping duplicate" messages
3. Verify that the unique constraint prevents duplicates at the database level
