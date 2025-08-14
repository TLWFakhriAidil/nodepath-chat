// CORRECTED: updateContactProgress - Complete current step AND activate next step
func (s *SequenceTriggerProcessor) updateContactProgress(contactID string, nextTrigger sql.NullString, delayHours int) error {
	// Start transaction to ensure atomic updates
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Step 1: Mark current step as completed
	query := `
		UPDATE sequence_contacts 
		SET status = 'completed', 
			processing_device_id = NULL,
			processing_started_at = NULL,
			completed_at = NOW()
		WHERE id = $1
		RETURNING sequence_id, contact_phone, current_step
	`
	
	var sequenceID, phone string
	var currentStep int
	err = tx.QueryRow(query, contactID).Scan(&sequenceID, &phone, &currentStep)
	if err != nil {
		return fmt.Errorf("failed to mark contact as completed: %w", err)
	}
	
	logrus.Infof("✅ COMPLETED: Step %d for %s", currentStep, phone)
	
	// Step 2: Find and activate the next pending step for this contact
	activateNextQuery := `
		UPDATE sequence_contacts 
		SET status = 'active'
		WHERE sequence_id = $1 
		AND contact_phone = $2
		AND status = 'pending'
		AND current_step = (
			SELECT MIN(current_step) 
			FROM sequence_contacts 
			WHERE sequence_id = $1 
			AND contact_phone = $2 
			AND status = 'pending'
		)
		RETURNING current_step, current_trigger, next_trigger_time
	`
	
	var nextStep int
	var nextStepTrigger string
	var nextTriggerTime time.Time
	
	err = tx.QueryRow(activateNextQuery, sequenceID, phone).Scan(&nextStep, &nextStepTrigger, &nextTriggerTime)
	if err == sql.ErrNoRows {
		// No more pending steps - sequence is complete
		logrus.Infof("🎉 SEQUENCE COMPLETE: All steps finished for %s", phone)
		
		// Get the trigger to remove from lead
		var trigger string
		tx.QueryRow("SELECT current_trigger FROM sequence_contacts WHERE id = $1", contactID).Scan(&trigger)
		
		// Commit transaction before removing trigger
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}
		
		// Remove trigger from lead
		s.removeCompletedTriggerFromLead(phone, trigger)
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to activate next step: %w", err)
	}
	
	// Successfully activated next step
	timeUntilNext := time.Until(nextTriggerTime)
	if timeUntilNext > 0 {
		logrus.Infof("⚡ ACTIVATED: Step %d for %s (trigger: %s) - will process in %v at %v", 
			nextStep, phone, nextStepTrigger, timeUntilNext.Round(time.Second), 
			nextTriggerTime.Format("15:04:05"))
	} else {
		logrus.Infof("⚡ ACTIVATED: Step %d for %s (trigger: %s) - ready to process NOW!", 
			nextStep, phone, nextStepTrigger)
	}
	
	// Check how many steps remain
	var remainingCount int
	tx.QueryRow(`
		SELECT COUNT(*) 
		FROM sequence_contacts 
		WHERE sequence_id = $1 
		AND contact_phone = $2 
		AND status = 'pending'
	`, sequenceID, phone).Scan(&remainingCount)
	
	logrus.Infof("📊 Progress: %s completed step %d → activated step %d (with %d more pending)", 
		phone, currentStep, nextStep, remainingCount)
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// SIMPLIFIED: processSequenceContacts - No need for separate activation query
func (s *SequenceTriggerProcessor) processSequenceContacts(deviceLoads map[string]DeviceLoad) (int, error) {
	// ONLY process ACTIVE contacts where next_trigger_time <= NOW
	// No need for separate activation - it happens after each message is sent
	
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
			sc.next_trigger_time
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
	
	// Debug: Show what's ready to process
	var readyCount int
	s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM sequence_contacts sc
		JOIN sequences s ON s.id = sc.sequence_id
		WHERE sc.status = 'active' 
		AND sc.next_trigger_time <= $1
		AND s.is_active = true
		AND sc.processing_device_id IS NULL
	`, time.Now()).Scan(&readyCount)
	
	if readyCount > 0 {
		logrus.Infof("📤 Found %d ACTIVE contacts ready to send now", readyCount)
	}
	
	rows, err := s.db.Query(query, time.Now(), s.batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get contacts for processing: %w", err)
	}
	defer rows.Close()
	
	// Process messages...
	// [Rest of the processing code stays the same]
	
	return processedCount, nil
}
