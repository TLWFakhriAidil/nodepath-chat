// Modified enrollContactInSequence - Step 1 starts in 5 minutes, each step builds on previous
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
			// First step - starts in 5 minutes
			nextTriggerTime = currentTime.Add(5 * time.Minute)
			status = "pending" // Will become active when time comes
			logrus.Infof("Step 1: Setting trigger time to NOW + 5 minutes = %v", 
				nextTriggerTime.Format("15:04:05"))
		} else {
			// Subsequent steps - previous trigger time + current step delay
			if step.TriggerDelayHours > 0 {
				nextTriggerTime = previousTriggerTime.Add(time.Duration(step.TriggerDelayHours) * time.Hour)
			} else {
				// If no delay specified, default to 24 hours
				nextTriggerTime = previousTriggerTime.Add(24 * time.Hour)
			}
			status = "pending"
			
			logrus.Infof("Step %d: Previous trigger %v + %d hours = %v", 
				step.DayNumber, 
				previousTriggerTime.Format("15:04:05"),
				step.TriggerDelayHours,
				nextTriggerTime.Format("2006-01-02 15:04:05"))
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
			status,              // status - all start as pending
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
		
		logrus.Infof("Created step %d for %s - trigger: %s, time: %v", 
			step.DayNumber, lead.Phone, step.Trigger, 
			nextTriggerTime.Format("2006-01-02 15:04:05"))
	}
	
	// Log summary
	if len(steps) > 0 {
		logrus.Infof("Enrollment complete for %s:", lead.Phone)
		logrus.Infof("  - First message in: 5 minutes")
		logrus.Infof("  - Last message at: %v", previousTriggerTime.Format("2006-01-02 15:04:05"))
		logrus.Infof("  - Total duration: %v", previousTriggerTime.Sub(currentTime.Add(5*time.Minute)))
	}
	
	return nil
}

// Example of what gets created:
// Phone: 60123456789
// Step 1: trigger_delay_hours = 0,  next_trigger = NOW + 5 min        (e.g., 10:05)
// Step 2: trigger_delay_hours = 24, next_trigger = 10:05 + 24h       (e.g., Day 2 10:05)
// Step 3: trigger_delay_hours = 48, next_trigger = Day 2 10:05 + 48h (e.g., Day 4 10:05)
// Step 4: trigger_delay_hours = 72, next_trigger = Day 4 10:05 + 72h (e.g., Day 7 10:05)
// Step 5: trigger_delay_hours = 96, next_trigger = Day 7 10:05 + 96h (e.g., Day 11 10:05)
