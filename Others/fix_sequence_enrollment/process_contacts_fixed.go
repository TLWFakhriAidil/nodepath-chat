// Modified processSequenceContacts to handle pending → active transition
func (s *SequenceTriggerProcessor) processSequenceContacts(deviceLoads map[string]DeviceLoad) (int, error) {
	// FIRST: Activate any pending contacts that are ready
	activateQuery := `
		UPDATE sequence_contacts 
		SET status = 'active'
		WHERE status = 'pending' 
		AND next_trigger_time <= $1
		RETURNING id, contact_phone, current_step
	`
	
	rows, err := s.db.Query(activateQuery, time.Now())
	if err != nil {
		logrus.Errorf("Failed to activate pending contacts: %v", err)
	} else {
		activatedCount := 0
		for rows.Next() {
			var id, phone string
			var step int
			rows.Scan(&id, &phone, &step)
			activatedCount++
			logrus.Infof("Activated step %d for contact %s", step, phone)
		}
		rows.Close()
		if activatedCount > 0 {
			logrus.Infof("Activated %d pending contacts", activatedCount)
		}
	}
	
	// THEN: Process active contacts as before
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
	
	// Continue with existing processing logic...
}

// Modified updateContactProgress - just mark as completed
func (s *SequenceTriggerProcessor) updateContactProgress(contactID string, nextTrigger sql.NullString, delayHours int) error {
	// Simply mark current contact record as completed
	query := `
		UPDATE sequence_contacts 
		SET status = 'completed', 
			processing_device_id = NULL,
			processing_started_at = NULL,
			completed_at = NOW()
		WHERE id = $1
	`
	
	_, err := s.db.Exec(query, contactID)
	if err != nil {
		return fmt.Errorf("failed to mark contact as completed: %w", err)
	}
	
	logrus.Infof("Marked contact %s as completed", contactID)
	
	// Check if this was the last step
	var isLastStep bool
	err = s.db.QueryRow(`
		SELECT NOT EXISTS (
			SELECT 1 FROM sequence_contacts 
			WHERE sequence_id = (SELECT sequence_id FROM sequence_contacts WHERE id = $1)
			AND contact_phone = (SELECT contact_phone FROM sequence_contacts WHERE id = $1)
			AND status IN ('pending', 'active')
		)
	`, contactID).Scan(&isLastStep)
	
	if err == nil && isLastStep {
		// All steps completed - remove trigger from lead
		var phone, currentTrigger string
		err := s.db.QueryRow(`
			SELECT contact_phone, current_trigger 
			FROM sequence_contacts 
			WHERE id = $1
		`, contactID).Scan(&phone, &currentTrigger)
		
		if err == nil {
			s.removeCompletedTriggerFromLead(phone, currentTrigger)
			logrus.Infof("Sequence fully completed for %s - removed trigger %s", phone, currentTrigger)
		}
	}
	
	return nil
}
