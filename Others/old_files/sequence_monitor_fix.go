# Update the monitorBroadcastResults function to properly sync status

# This code should be added to the monitorBroadcastResults() function in sequence_trigger_processor.go

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
					AND bm.sequence_stepid = sc.sequence_stepid  -- NEW: Match specific step
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
				SET status = 'sent',
					completed_at = NOW()
				FROM broadcast_messages bm
				WHERE bm.sequence_id = sc.sequence_id
					AND bm.recipient_phone = sc.contact_phone
					AND bm.sequence_stepid = sc.sequence_stepid  -- NEW: Match specific step
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
			
			// NEW: Handle failed sequences - mark entire sequence as failed after 3 failures
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

# Also update the processContact function to include sequence_stepid when queuing:

// In processContact() function, update the broadcast message creation:
broadcastMsg := domainBroadcast.BroadcastMessage{
	UserID:         job.userID,
	DeviceID:       deviceID,
	SequenceID:     &job.sequenceID,
	SequenceStepID: &job.stepID,  // NEW: Add this field
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
