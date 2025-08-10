// Fix for updateContactProgress in sequence_trigger_processor.go
// This fix changes the activation logic from step number based to time based

// REPLACE THIS FUNCTION in src/usecase/sequence_trigger_processor.go

// updateContactProgress completes current step and activates next step based on earliest time
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
	
	// Step 2: Find and activate the next pending step based on EARLIEST next_trigger_time
	// FIXED: Now using next_trigger_time instead of current_step
	activateNextQuery := `
		UPDATE sequence_contacts 
		SET status = 'active'
		WHERE sequence_id = $1 
		AND contact_phone = $2
		AND status = 'pending'
		AND next_trigger_time = (
			SELECT MIN(next_trigger_time)  -- FIXED: Use earliest time, not lowest step number
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
