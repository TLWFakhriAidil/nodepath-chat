// CORRECTED: processSequenceContacts - Only process ACTIVE contacts with time <= NOW
func (s *SequenceTriggerProcessor) processSequenceContacts(deviceLoads map[string]DeviceLoad) (int, error) {
	// STEP 1: Activate pending contacts that are ready
	// When a step is completed, activate the next pending step if its time has come
	activateQuery := `
		UPDATE sequence_contacts 
		SET status = 'active'
		WHERE status = 'pending' 
		AND next_trigger_time <= $1
		RETURNING id, contact_phone, current_step, current_trigger, next_trigger_time
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
			var triggerTime time.Time
			activateRows.Scan(&id, &phone, &step, &trigger, &triggerTime)
			activatedCount++
			logrus.Infof("✅ ACTIVATED: Step %d for %s (trigger: %s) - was scheduled for %v", 
				step, phone, trigger, triggerTime.Format("15:04:05"))
		}
		if activatedCount > 0 {
			logrus.Infof("Total activated: %d contacts moved from pending → active", activatedCount)
		}
	}
	
	// STEP 2: Process ACTIVE contacts where next_trigger_time <= NOW
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
	
	rows, err := s.db.Query(query, time.Now(), s.batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get contacts for processing: %w", err)
	}
	defer rows.Close()
	
	// Debug: Log what we're about to process
	var readyCount int
	s.db.QueryRow(`
		SELECT COUNT(*) 
		FROM sequence_contacts sc
		JOIN sequences s ON s.id = sc.sequence_id
		WHERE sc.status = 'active' 
		AND sc.next_trigger_time <= $1
		AND s.is_active = true
	`, time.Now()).Scan(&readyCount)
	
	logrus.Infof("📤 Found %d ACTIVE contacts ready to send (next_trigger_time <= NOW)", readyCount)
	
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
					logrus.Debugf("Worker %d: Sent message for %s step %d", 
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
			var triggerTime time.Time
			if err := rows.Scan(&job.contactID, &job.sequenceID, &job.phone, &job.name,
				&job.currentTrigger, &job.currentStep, &job.messageText, &job.messageType,
				&job.mediaURL, &job.nextTrigger, &job.delayHours, &job.preferredDevice,
				&job.minDelaySeconds, &job.maxDelaySeconds, &job.userID, &triggerTime); err != nil {
				logrus.Errorf("Error scanning job: %v", err)
				continue
			}
			
			logrus.Infof("📨 Queueing: %s step %d (was scheduled for %v)", 
				job.phone, job.currentStep, triggerTime.Format("15:04:05"))
			
			jobs <- job
			jobCount++
		}
		close(jobs)
		
		if jobCount > 0 {
			logrus.Infof("Queued %d messages for sending", jobCount)
		}
	}()
	
	// Wait for completion
	go func() {
		wg.Wait()
		close(results)
	}()
	
	// Count successes
	processedCount := 0
	failedCount := 0
	for success := range results {
		if success {
			processedCount++
		} else {
			failedCount++
		}
	}
	
	if processedCount > 0 || failedCount > 0 {
		logrus.Infof("✅ Processing complete: %d sent, %d failed", processedCount, failedCount)
	}
	
	return processedCount, nil
}

// No changes needed to updateContactProgress - it stays the same
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
	
	logrus.Infof("✅ COMPLETED: Step %d for %s", step, phone)
	
	// Check remaining steps
	var pendingCount, activeCount int
	s.db.QueryRow(`
		SELECT 
			COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending,
			COUNT(CASE WHEN status = 'active' THEN 1 END) as active
		FROM sequence_contacts 
		WHERE sequence_id = (SELECT sequence_id FROM sequence_contacts WHERE id = $1)
		AND contact_phone = (SELECT contact_phone FROM sequence_contacts WHERE id = $1)
	`, contactID).Scan(&pendingCount, &activeCount)
	
	if pendingCount == 0 && activeCount == 0 {
		// All steps completed
		var currentTrigger string
		s.db.QueryRow(`
			SELECT current_trigger FROM sequence_contacts WHERE id = $1
		`, contactID).Scan(&currentTrigger)
		
		// Remove trigger from lead
		s.removeCompletedTriggerFromLead(phone, currentTrigger)
		logrus.Infof("🎉 SEQUENCE COMPLETE: All steps finished for %s", phone)
	} else {
		logrus.Infof("📊 Progress: %s has %d pending + %d active steps remaining", 
			phone, pendingCount, activeCount)
	}
	
	return nil
}
