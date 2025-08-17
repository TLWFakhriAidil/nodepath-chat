// CORRECTED: enrollContactInSequence - First step ACTIVE, others PENDING
func (s *SequenceTriggerProcessor) enrollContactInSequence(sequenceID string, lead models.Lead, trigger string) error {
	// Get ALL steps for the sequence
	stepsQuery := `
		SELECT id, day_number, trigger, next_trigger, trigger_delay_hours 
		FROM sequence_steps 
		WHERE sequence_id = $1 
		ORDER BY day_number ASC
	`
	
	rows, err := s.db.Query(stepsQuery, sequenceID)
	if err != nil {
		return fmt.Errorf("failed to get sequence steps: %w", err)
	}
	defer rows.Close()
	
	var steps []struct {
		ID                string
		DayNumber         int
		Trigger           string
		NextTrigger       sql.NullString
		TriggerDelayHours int
	}
	
	// Collect all steps
	for rows.Next() {
		var step struct {
			ID                string
			DayNumber         int
			Trigger           string
			NextTrigger       sql.NullString
			TriggerDelayHours int
		}
		
		err := rows.Scan(&step.ID, &step.DayNumber, &step.Trigger, 
			&step.NextTrigger, &step.TriggerDelayHours)
		if err != nil {
			continue
		}
		
		steps = append(steps, step)
	}
	
	if len(steps) == 0 {
		return fmt.Errorf("no steps found for sequence %s", sequenceID)
	}
	
	logrus.Infof("Enrolling contact %s in sequence %s - creating ALL %d steps", 
		lead.Phone, sequenceID, len(steps))
	
	// Create records for ALL steps
	currentTime := time.Now()
	var previousTriggerTime time.Time
	
	for i, step := range steps {
		// Calculate next_trigger_time based on previous step
		var nextTriggerTime time.Time
		var status string
		
		if i == 0 {
			// FIRST STEP: ACTIVE with 5 minute delay
			nextTriggerTime = currentTime.Add(5 * time.Minute)
			status = "active" // ACTIVE, not pending!
			logrus.Infof("Step 1: ACTIVE - will send at %v (NOW + 5 minutes)", 
				nextTriggerTime.Format("15:04:05"))
		} else {
			// Subsequent steps - PENDING with calculated time
			if step.TriggerDelayHours > 0 {
				nextTriggerTime = previousTriggerTime.Add(time.Duration(step.TriggerDelayHours) * time.Hour)
			} else {
				// If no delay specified, default to 24 hours
				nextTriggerTime = previousTriggerTime.Add(24 * time.Hour)
			}
			status = "pending"
			
			logrus.Infof("Step %d: PENDING - will activate at %v (previous + %d hours)", 
				step.DayNumber, 
				nextTriggerTime.Format("2006-01-02 15:04:05"),
				step.TriggerDelayHours)
		}
		
		// Store this trigger time for next iteration
		previousTriggerTime = nextTriggerTime
		
		insertQuery := `
			INSERT INTO sequence_contacts (
				sequence_id, contact_phone, contact_name, 
				current_step, status, current_trigger,
				next_trigger_time, sequence_stepid, assigned_device_id, user_id
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			ON CONFLICT (sequence_id, contact_phone, sequence_stepid) DO NOTHING
		`
		
		_, err = s.db.Exec(insertQuery, 
			sequenceID,          // sequence_id
			lead.Phone,          // contact_phone
			lead.Name,           // contact_name
			step.DayNumber,      // current_step
			status,              // status - FIRST is active, others pending
			step.Trigger,        // current_trigger
			nextTriggerTime,     // next_trigger_time
			step.ID,             // sequence_stepid
			lead.DeviceID,       // assigned_device_id
			lead.UserID,         // user_id
		)
		
		if err != nil {
			logrus.Warnf("Failed to create step %d for contact %s: %v", 
				step.DayNumber, lead.Phone, err)
			continue
		}
		
		logrus.Infof("Created step %d for %s - status: %s, trigger: %s, time: %v", 
			step.DayNumber, lead.Phone, status, step.Trigger, 
			nextTriggerTime.Format("2006-01-02 15:04:05"))
	}
	
	// Log summary
	logrus.Infof("✅ Enrollment complete for %s:", lead.Phone)
	logrus.Infof("  - Step 1 (ACTIVE): Will send in 5 minutes")
	logrus.Infof("  - Steps 2-%d (PENDING): Scheduled up to %v", 
		len(steps), previousTriggerTime.Format("2006-01-02 15:04:05"))
	
	return nil
}
