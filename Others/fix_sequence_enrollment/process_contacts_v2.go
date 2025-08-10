// Modified processSequenceContacts to handle the new flow
func (s *SequenceTriggerProcessor) processSequenceContacts(deviceLoads map[string]DeviceLoad) (int, error) {
	// STEP 1: Activate any pending contacts whose time has come
	activateQuery := `
		UPDATE sequence_contacts 
		SET status = 'active'
		WHERE status = 'pending' 
		AND next_trigger_time <= $1
		RETURNING id, contact_phone, current_step, current_trigger
	`
	
	activateRows, err := s.db.Query(activateQuery, time.Now())
	if err != nil {
		logrus.Errorf("Failed to activate pending contacts: %v", err)
	} else {
		defer activateRows.Close()
		activatedCount := 0
		for activateRows.Next() {
			var id, phone, trigger string
			var step int
			activateRows.Scan(&id, &phone, &step, &trigger)
			activatedCount++
			logrus.Infof("ACTIVATED: Step %d for %s (trigger: %s) - ready to send", 
				step, phone, trigger)
		}
		if activatedCount > 0 {
			logrus.Infof("Total activated: %d contacts moved from pending → active", activatedCount)
		}
	}
	
	// STEP 2: Process all active contacts (send messages)
	query := `
		SELECT 
			sc.id, sc.sequence_id, sc.contact_phone, sc.contact_name,
			sc.current_trigger, sc.current_step,
			ss.content, ss.message_type, ss.media_url,
			ss.next_trigger, ss.trigger_delay_hours,
			COALESCE(sc.assigned_device_id, l.device_id) as preferred_device_id,
			COALESCE(ss.min_delay_seconds, 5) as min_delay_seconds,
			COALESCE(ss.max_delay_seconds, 15) as max_delay_seconds,
			l.user_id
		FROM sequence_contacts sc
		JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
		JOIN sequences s ON s.id = sc.sequence_id
		LEFT JOIN leads l ON l.phone = sc.contact_phone
		WHERE sc.status = 'active'
			AND s.is_active = true
			AND sc.processing_device_id IS NULL
		ORDER BY sc.next_trigger_time ASC
		LIMIT $1
	`
	
	rows, err := s.db.Query(query, s.batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get contacts for processing: %w", err)
	}
	defer rows.Close()
	
	// Log how many active contacts we found
	var activeCount int
	s.db.QueryRow("SELECT COUNT(*) FROM sequence_contacts WHERE status = 'active'").Scan(&activeCount)
	logrus.Infof("Found %d active contacts ready to process", activeCount)
	
	// Process in parallel with worker pool
	jobs := make(chan contactJob, s.batchSize)
	results := make(chan bool, s.batchSize)
	
	// Start workers
	numWorkers := 50
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for job := range jobs {
				success := s.processContact(job, deviceLoads)
				if success {
					logrus.Debugf("Worker %d: Successfully processed %s step %d", 
						workerID, job.phone, job.currentStep)
				}
				results <- success
			}
		}(i)
	}
	
	// Queue jobs
	go func() {
		jobCount := 0
		for rows.Next() {
			var job contactJob
			if err := rows.Scan(&job.contactID, &job.sequenceID, &job.phone, &job.name,
				&job.currentTrigger, &job.currentStep, &job.messageText, &job.messageType,
				&job.mediaURL, &job.nextTrigger, &job.delayHours, &job.preferredDevice,
				&job.minDelaySeconds, &job.maxDelaySeconds, &job.userID); err != nil {
				logrus.Errorf("Error scanning job: %v", err)
				continue
			}
			jobs <- job
			jobCount++
		}
		close(jobs)
		logrus.Infof("Queued %d jobs for processing", jobCount)
	}()
	
	// Wait for completion
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Count successes
	processedCount := 0
	for success := range results {
		if success {
			processedCount++
		}
	}
	
	return processedCount, nil
}

// Simplified updateContactProgress - just mark as completed
func (s *SequenceTriggerProcessor) updateContactProgress(contactID string, nextTrigger sql.NullString, delayHours int) error {
	// Mark current step as completed
	query := `
		UPDATE sequence_contacts 
		SET status = 'completed', 
			processing_device_id = NULL,
			processing_started_at = NULL,
			completed_at = NOW()
		WHERE id = $1
	`
	
	result, err := s.db.Exec(query, contactID)
	if err != nil {
		return fmt.Errorf("failed to mark contact as completed: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		logrus.Warnf("No rows updated for contact ID: %s", contactID)
		return nil
	}
	
	// Get contact info for logging
	var phone string
	var step int
	s.db.QueryRow(`
		SELECT contact_phone, current_step 
		FROM sequence_contacts 
		WHERE id = $1
	`, contactID).Scan(&phone, &step)
	
	logrus.Infof("COMPLETED: Step %d for %s marked as completed", step, phone)
	
	// Check if ALL steps are completed for this contact
	var pendingCount int
	err = s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM sequence_contacts 
		WHERE sequence_id = (SELECT sequence_id FROM sequence_contacts WHERE id = $1)
		AND contact_phone = (SELECT contact_phone FROM sequence_contacts WHERE id = $1)
		AND status IN ('pending', 'active')
	`, contactID).Scan(&pendingCount)
	
	if err == nil && pendingCount == 0 {
		// All steps completed - sequence is done
		var currentTrigger string
		s.db.QueryRow(`
			SELECT current_trigger FROM sequence_contacts WHERE id = $1
		`, contactID).Scan(&currentTrigger)
		
		// Remove trigger from lead
		s.removeCompletedTriggerFromLead(phone, currentTrigger)
		logrus.Infof("SEQUENCE COMPLETE: All steps finished for %s - removed trigger %s", 
			phone, currentTrigger)
	} else {
		logrus.Infof("Progress: %s has %d remaining steps", phone, pendingCount)
	}
	
	return nil
}
