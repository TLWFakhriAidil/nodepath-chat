# Go Code Fixes for Sequence System

## 1. Update contactJob struct to include sequenceStepID

In `sequence_trigger_processor.go`, update the contactJob struct:

```go
// contactJob represents a job for processing a contact message
type contactJob struct {
	contactID        string
	sequenceID       string
	sequenceStepID   string  // ADD THIS LINE
	phone            string
	name             string
	currentTrigger   string
	currentStep      int
	messageText      string
	messageType      string
	mediaURL         sql.NullString
	nextTrigger      sql.NullString
	delayHours       int
	preferredDevice  sql.NullString
	minDelaySeconds  int
	maxDelaySeconds  int
	userID           string
}
```

## 2. Update processSequenceContacts query to include sequence_stepid

```go
// In processSequenceContacts function, update the query:
query := `
	SELECT 
		sc.id, sc.sequence_id, sc.contact_phone, sc.contact_name,
		sc.current_trigger, sc.current_step,
		ss.content, ss.message_type, ss.media_url,
		ss.next_trigger, ss.trigger_delay_hours,
		COALESCE(sc.assigned_device_id, l.device_id) as preferred_device_id,
		COALESCE(ss.min_delay_seconds, 5) as min_delay_seconds,
		COALESCE(ss.max_delay_seconds, 15) as max_delay_seconds,
		l.user_id,
		sc.next_trigger_time,
		sc.sequence_stepid  -- ADD THIS LINE
	FROM sequence_contacts sc
	JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
	JOIN sequences s ON s.id = sc.sequence_id
	LEFT JOIN leads l ON l.phone = sc.contact_phone
	WHERE sc.status = 'active'
		AND s.is_active = true
		AND sc.next_trigger_time <= $1
		AND sc.processing_device_id IS NULL
	ORDER BY sc.next_trigger_time ASC
	LIMIT $2
`

// And update the Scan to include it:
if err := rows.Scan(&job.contactID, &job.sequenceID, &job.phone, &job.name,
	&job.currentTrigger, &job.currentStep, &job.messageText, &job.messageType,
	&job.mediaURL, &job.nextTrigger, &job.delayHours, &job.preferredDevice,
	&job.minDelaySeconds, &job.maxDelaySeconds, &job.userID, &triggerTime,
	&job.sequenceStepID); err != nil {  // ADD sequenceStepID HERE
	logrus.Errorf("Error scanning job: %v", err)
	continue
}
```

## 3. Update processContact to include sequence_stepid in broadcast message

```go
// In processContact function, update broadcast message creation:
broadcastMsg := domainBroadcast.BroadcastMessage{
	UserID:         job.userID,
	DeviceID:       deviceID,
	SequenceID:     &job.sequenceID,
	SequenceStepID: &job.sequenceStepID,  // ADD THIS LINE
	RecipientPhone: job.phone,
	RecipientName:  job.name,
	Message:        job.messageText,
	Content:        job.messageText,
	Type:           job.messageType,
	MinDelay:       job.minDelaySeconds,
	MaxDelay:       job.maxDelaySeconds,
	ScheduledAt:    time.Now(),
	Status:         "pending",
}
```

## 4. Update BroadcastMessage struct in domains/broadcast

In `domains/broadcast/broadcast.go`, add the SequenceStepID field:

```go
type BroadcastMessage struct {
	ID             string
	UserID         string
	DeviceID       string
	CampaignID     *string
	SequenceID     *string
	SequenceStepID *string  // ADD THIS LINE
	RecipientPhone string
	RecipientName  string
	MessageType    string
	Message        string
	Content        string
	MediaURL       string
	ImageURL       string
	Type           string
	MinDelay       int
	MaxDelay       int
	Status         string
	ScheduledAt    time.Time
	GroupID        *string
	GroupOrder     *int
}
```

## 5. Update QueueMessage in broadcast_repository.go

```go
// Update the INSERT query to include sequence_stepid:
query := `
	INSERT INTO broadcast_messages 
	(id, user_id, device_id, campaign_id, sequence_id, sequence_stepid, recipient_phone, recipient_name,
	 message_type, content, media_url, status, scheduled_at, created_at, group_id, group_order)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
`

// Update the Exec to include sequence_stepid:
var sequenceStepID interface{}
if msg.SequenceStepID != nil {
	sequenceStepID = *msg.SequenceStepID
} else {
	sequenceStepID = nil
}

_, err := r.db.Exec(query, msg.ID, userID, msg.DeviceID, campaignID,
	sequenceID, sequenceStepID, msg.RecipientPhone, msg.RecipientName, 
	msg.Type, msg.Content, msg.MediaURL, "pending", msg.ScheduledAt, 
	time.Now(), groupID, groupOrder)
```

## 6. Complete the monitorBroadcastResults function

```go
func (s *SequenceTriggerProcessor) monitorBroadcastResults() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Update FAILED messages
			failQuery := `
				UPDATE sequence_contacts sc
				SET status = 'failed',
					last_error = bm.error_message,
					retry_count = sc.retry_count + 1
				FROM broadcast_messages bm
				WHERE bm.sequence_id = sc.sequence_id
					AND bm.recipient_phone = sc.contact_phone
					AND bm.sequence_stepid = sc.sequence_stepid
					AND bm.status = 'failed'
					AND sc.status = 'active'
					AND sc.processing_device_id IS NOT NULL
					AND bm.created_at > NOW() - INTERVAL '5 minutes'
			`
			
			failResult, err := s.db.Exec(failQuery)
			if err == nil {
				if affected, _ := failResult.RowsAffected(); affected > 0 {
					logrus.Warnf("Marked %d sequence contacts as failed due to broadcast failures", affected)
				}
			}
			
			// NEW: Update SUCCESSFUL messages
			successQuery := `
				UPDATE sequence_contacts sc
				SET status = 'sent'
				FROM broadcast_messages bm
				WHERE bm.sequence_id = sc.sequence_id
					AND bm.recipient_phone = sc.contact_phone
					AND bm.sequence_stepid = sc.sequence_stepid
					AND bm.status = 'sent'
					AND sc.status = 'active'
					AND sc.processing_device_id IS NOT NULL
					AND bm.sent_at > NOW() - INTERVAL '5 minutes'
			`
			
			successResult, err := s.db.Exec(successQuery)
			if err == nil {
				if affected, _ := successResult.RowsAffected(); affected > 0 {
					logrus.Infof("Marked %d sequence contacts as sent due to successful broadcasts", affected)
				}
			}
			
			// Mark entire sequence as failed after 3 failures
			markFailedQuery := `
				UPDATE sequence_contacts
				SET status = 'sequence_failed'
				WHERE sequence_id IN (
					SELECT sequence_id
					FROM sequence_contacts
					WHERE status = 'failed'
					AND retry_count >= 3
					GROUP BY sequence_id, contact_phone
				)
				AND status IN ('pending', 'active')
			`
			
			s.db.Exec(markFailedQuery)
			
		case <-s.stopChan:
			return
		}
	}
}
```

## 7. Update updateContactProgress to handle sent status

```go
// In updateContactProgress, check if message was sent before marking completed:
func (s *SequenceTriggerProcessor) updateContactProgress(contactID string, nextTrigger sql.NullString, delayHours int) error {
	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// First check if the message was actually sent
	var messageStatus string
	err = tx.QueryRow(`
		SELECT bm.status 
		FROM sequence_contacts sc
		JOIN broadcast_messages bm ON bm.sequence_id = sc.sequence_id 
			AND bm.recipient_phone = sc.contact_phone 
			AND bm.sequence_stepid = sc.sequence_stepid
		WHERE sc.id = $1
		ORDER BY bm.created_at DESC
		LIMIT 1
	`, contactID).Scan(&messageStatus)
	
	// Only mark as completed if message was sent
	if err == nil && messageStatus == "sent" {
		// Continue with existing logic to mark completed and activate next
		// ... existing code ...
	} else {
		// Message not sent yet, don't progress
		return nil
	}
}
```

These changes will ensure:
1. Each broadcast message is linked to its specific sequence step
2. The monitoring function properly syncs status between tables
3. Sequences only progress when messages are actually sent
4. Failed sequences can be identified and handled appropriately
