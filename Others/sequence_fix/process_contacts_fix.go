// MODIFIED PROCESS CONTACTS - NEW LOGIC
// Find earliest PENDING step and check if time has arrived

func (s *SequenceTriggerProcessor) processContactsReadyForMessages(deviceLoads map[string]DeviceLoad) (int, error) {
    // NEW QUERY: Find PENDING steps with earliest next_trigger_time
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
        return 0, fmt.Errorf("failed to get earliest pending contacts: %w", err)
    }
    defer rows.Close()

    processedCount := 0
    
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
        
        // Check if it's time to send this message
        now := time.Now()
        
        if triggerTime.After(now) {
            // Not time yet - mark as ACTIVE so we track it
            logrus.Infof("⏰ Step %d for %s not ready yet (triggers at %v, %v remaining)", 
                job.currentStep, job.phone, triggerTime.Format("15:04:05"), 
                time.Until(triggerTime))
            
            // Update to active status
            _, err = s.db.Exec(`
                UPDATE sequence_contacts 
                SET status = 'active', updated_at = NOW()
                WHERE id = $1 AND status = 'pending'
            `, job.contactID)
            
            if err != nil {
                logrus.Errorf("Failed to activate contact: %v", err)
            }
            continue
        }
        
        // Time has arrived! Process the message
        logrus.Infof("✅ Time reached for %s step %d - sending message", 
            job.phone, job.currentStep)
        
        // Check device
        deviceID := job.preferredDevice.String
        if deviceID == "" {
            logrus.Warnf("No assigned device for contact %s - skipping", job.phone)
            continue
        }
        
        // Create broadcast message
        broadcastMsg := domainBroadcast.BroadcastMessage{
            UserID:         job.userID,
            DeviceID:       deviceID,
            SequenceID:     &job.sequenceID,
            SequenceStepID: &sequenceStepID,  // Track which step
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
        
        if job.mediaURL.Valid && job.mediaURL.String != "" {
            broadcastMsg.MediaURL = job.mediaURL.String
            broadcastMsg.ImageURL = job.mediaURL.String
        }
        
        // Queue to database
        broadcastRepo := repository.GetBroadcastRepository()
        if err := broadcastRepo.QueueMessage(broadcastMsg); err != nil {
            logrus.Errorf("Failed to queue sequence message for %s: %v", job.phone, err)
            continue
        }
        
        // Mark this step as completed
        _, err = s.db.Exec(`
            UPDATE sequence_contacts 
            SET status = 'completed', 
                completed_at = NOW(),
                processing_device_id = $1
            WHERE id = $2
        `, deviceID, job.contactID)
        
        if err != nil {
            logrus.Errorf("Failed to mark contact as completed: %v", err)
        }
        
        processedCount++
        logrus.Infof("📤 Queued message for %s step %d", job.phone, job.currentStep)
    }
    
    return processedCount, nil
}