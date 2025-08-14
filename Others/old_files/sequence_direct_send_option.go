// Alternative: Direct Sequence Message Sending (without broadcast_messages table)
// This shows how to modify processContact to send directly

// Option 1: Modify processContact to send directly
func (s *SequenceTriggerProcessor) processContactDirect(job contactJob, deviceLoads map[string]DeviceLoad) bool {
	// ... existing validation code ...
	
	// Instead of queueing to broadcast_messages, send directly
	
	// 1. Get WhatsApp client or platform details
	device, err := s.userRepo.GetDeviceByID(deviceID)
	if err != nil {
		logrus.Errorf("Failed to get device: %v", err)
		return false
	}
	
	// 2. Prepare message with anti-spam
	messageSender := broadcast.NewWhatsAppMessageSender()
	
	// Create a temporary broadcast message object for anti-spam processing
	tempMsg := &domainBroadcast.BroadcastMessage{
		UserID:         job.userID,
		DeviceID:       deviceID,
		SequenceID:     &job.sequenceID,
		RecipientPhone: job.phone,
		RecipientName:  job.name,
		Type:           job.messageType,
		Message:        job.messageText,
		ImageURL:       job.mediaURL.String,
	}
	
	// 3. Send directly
	err = messageSender.SendMessage(deviceID, tempMsg)
	if err != nil {
		logrus.Errorf("Failed to send sequence message directly: %v", err)
		
		// Update sequence_contacts to mark as failed
		s.db.Exec(`
			UPDATE sequence_contacts 
			SET status = 'failed',
				error_message = $1,
				processing_device_id = NULL,
				completed_at = NOW()
			WHERE id = $2
		`, err.Error(), job.contactID)
		
		return false
	}
	
	// 4. Log to sequence_send_logs for tracking (optional)
	s.db.Exec(`
		INSERT INTO sequence_send_logs (
			sequence_id, contact_phone, step_number, 
			device_id, sent_at, status
		) VALUES ($1, $2, $3, $4, NOW(), 'sent')
	`, job.sequenceID, job.phone, job.currentStep, deviceID)
	
	// 5. Update progress
	if err := s.updateContactProgress(job.contactID, job.nextTrigger, job.delayHours); err != nil {
		logrus.Errorf("Failed to update contact progress: %v", err)
		return false
	}
	
	return true
}

// Option 2: Add a configuration flag to choose between queuing and direct sending
type SequenceTriggerProcessor struct {
	// ... existing fields ...
	directSend bool  // If true, bypass broadcast_messages
}

// In your config or environment:
// SEQUENCE_DIRECT_SEND=true  # Send directly without queuing

// Option 3: Hybrid approach - use broadcast_messages only for rate limiting
// This gives you the benefits of both approaches:

func (s *SequenceTriggerProcessor) processContactHybrid(job contactJob, deviceLoads map[string]DeviceLoad) bool {
	// Check device load
	load := deviceLoads[deviceID]
	
	// If device is under threshold, send directly
	if load.MessagesHour < 50 && load.CurrentProcessing < 5 {
		return s.processContactDirect(job, deviceLoads)
	}
	
	// Otherwise, queue to broadcast_messages for rate limiting
	return s.processContactQueued(job, deviceLoads)
}

/* PROS of Direct Sending:
 * - Simpler flow
 * - Less database writes
 * - Immediate sending
 * - Less complexity
 *
 * CONS of Direct Sending:
 * - No centralized message tracking
 * - Harder to implement rate limiting
 * - No unified retry mechanism
 * - Device load balancing more complex
 * - Can't easily pause/resume
 * - No message history in one place
 *
 * RECOMMENDATION:
 * Keep using broadcast_messages for these benefits:
 * 1. Unified message queue for all types
 * 2. Better observability (all messages in one table)
 * 3. Easy to implement rate limiting
 * 4. Retry logic in one place
 * 5. Can pause campaigns/sequences by updating status
 * 6. Analytics and reporting easier
 */
