// processContactWithNewLogic handles a single contact with time-based logic
func (s *SequenceTriggerProcessor) processContactWithNewLogic(job contactJob, deviceLoads map[string]DeviceLoad) bool {
	now := time.Now()
	
	// Check if it's time to send this message
	if job.nextTriggerTime.After(now) {
		// Not time yet - mark as ACTIVE to track it
		timeRemaining := time.Until(job.nextTriggerTime)
		logrus.Infof("⏰ Step %d for %s not ready (triggers in %v at %v)", 
			job.currentStep, job.phone, timeRemaining, 
			job.nextTriggerTime.Format("15:04:05"))
		
		// Update to active status so we know it's next in line
		result, err := s.db.Exec(`
			UPDATE sequence_contacts 
			SET status = 'active', updated_at = NOW()
			WHERE id = $1 AND status = 'pending'
		`, job.contactID)
		
		if err != nil {
			logrus.Errorf("Failed to activate contact %s: %v", job.contactID, err)
			return false
		}
		
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			logrus.Debugf("Marked step %d for %s as ACTIVE (next in line)", 
				job.currentStep, job.phone)
		}
		
		return false // Not processed, just marked active
	}
	
	// Time has arrived! Send the message
	logrus.Infof("✅ Time reached for %s step %d - processing message", 
		job.phone, job.currentStep)
	
	// Check device
	deviceID := job.preferredDevice.String
	if deviceID == "" {
		logrus.Warnf("No assigned device for contact %s - skipping", job.phone)
		return false
	}
	
	// Create broadcast message
	broadcastMsg := domainBroadcast.BroadcastMessage{
		UserID:         job.userID,
		DeviceID:       deviceID,
		SequenceID:     &job.sequenceID,
		SequenceStepID: &job.sequenceStepID,
		RecipientPhone: job.phone,
		RecipientName:  job.name,
		Message:        job.messageText,
		Content:        job.messageText,
		Type:           job.messageType,
		MinDelay:       job.minDelaySeconds,
		MaxDelay:       job.maxDelaySeconds,
		ScheduledAt:    now,
		Status:         "pending",
	}
	
	if job.mediaURL.Valid && job.mediaURL.String != "" {
		broadcastMsg.MediaURL = job.mediaURL.String
		broadcastMsg.ImageURL = job.mediaURL.String
	}
	
	// Queue to database
	broadcastRepo := repository.GetBroadcastRepository()
	if err := broadcastRepo.QueueMessage(broadcastMsg); err != nil {
		logrus.Errorf("Failed to queue sequence message for %s: %v", job.phone, err)
		
		// Mark as failed
		s.db.Exec(`
			UPDATE sequence_contacts 
			SET status = 'failed', 
				last_error = $1,
				updated_at = NOW()
			WHERE id = $2
		`, err.Error(), job.contactID)
		
		return false
	}
	
	// Mark this step as completed
	result, err := s.db.Exec(`
		UPDATE sequence_contacts 
		SET status = 'completed', 
			completed_at = NOW(),
			processing_device_id = $1,
			updated_at = NOW()
		WHERE id = $2
	`, deviceID, job.contactID)
	
	if err != nil {
		logrus.Errorf("Failed to mark contact as completed: %v", err)
		return false
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logrus.Warnf("No rows updated when marking contact %s as completed", job.contactID)
		return false
	}
	
	logrus.Infof("📤 Successfully queued and completed step %d for %s", 
		job.currentStep, job.phone)
	
	// Check if this was the last step
	if !job.nextTrigger.Valid || job.nextTrigger.String == "" {
		logrus.Infof("🎉 Sequence complete for %s - no more steps", job.phone)
		
		// Remove trigger from lead if it's complete
		s.removeCompletedTriggerFromLead(job.phone, job.currentTrigger)
	}
	
	return true
}

// Add these fields to the contactJob struct:
type contactJob struct {
	contactID        string
	sequenceID       string
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
	sequenceStepID   string    // Added
	nextTriggerTime  time.Time // Added
}