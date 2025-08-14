// processSequenceContacts processes contacts using PENDING-FIRST approach
func (s *SequenceTriggerProcessor) processSequenceContacts(deviceLoads map[string]DeviceLoad) (int, error) {
	// NEW LOGIC: Find earliest PENDING step for each contact
	query := `
		WITH earliest_pending AS (
			SELECT DISTINCT ON (sc.sequence_id, sc.contact_phone)
				sc.id, sc.sequence_id, sc.contact_phone, sc.contact_name,
				sc.current_trigger, sc.current_step,
				ss.content, ss.message_type, ss.media_url,
				ss.next_trigger, ss.trigger_delay_hours,
				sc.assigned_device_id as preferred_device_id,
				COALESCE(ss.min_delay_seconds, 5) as min_delay_seconds,
				COALESCE(ss.max_delay_seconds, 15) as max_delay_seconds,
				sc.user_id,
				sc.next_trigger_time,
				sc.sequence_stepid
			FROM sequence_contacts sc
			JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
			JOIN sequences s ON s.id = sc.sequence_id
			WHERE sc.status = 'pending'
				AND s.is_active = true
				AND sc.processing_device_id IS NULL
			ORDER BY sc.sequence_id, sc.contact_phone, sc.next_trigger_time ASC
		)
		SELECT * FROM earliest_pending
		ORDER BY next_trigger_time ASC
		LIMIT $1
	`

	rows, err := s.db.Query(query, s.batchSize)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending contacts: %w", err)
	}
	defer rows.Close()

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
				success := s.processContactWithNewLogic(job, deviceLoads)
				results <- success
			}
		}(i)
	}

	// Queue jobs
	go func() {
		for rows.Next() {
			var job contactJob
			var triggerTime time.Time
			var sequenceStepID string
			
			if err := rows.Scan(&job.contactID, &job.sequenceID, &job.phone, &job.name,
				&job.currentTrigger, &job.currentStep, &job.messageText, &job.messageType,
				&job.mediaURL, &job.nextTrigger, &job.delayHours, &job.preferredDevice,
				&job.minDelaySeconds, &job.maxDelaySeconds, &job.userID, &triggerTime,
				&sequenceStepID); err != nil {
				logrus.Errorf("Error scanning job: %v", err)
				continue
			}
			
			job.sequenceStepID = sequenceStepID
			job.nextTriggerTime = triggerTime
			jobs <- job
		}
		close(jobs)
	}()

	// Wait for completion
	wg.Wait()
	close(results)

	// Count results
	processedCount := 0
	for success := range results {
		if success {
			processedCount++
		}
	}

	return processedCount, nil
}