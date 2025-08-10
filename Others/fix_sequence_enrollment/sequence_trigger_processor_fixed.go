// Modified enrollContactInSequence to create ALL steps at once
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
	cumulativeDelay := 0 // Track total delay hours
	
	for i, step := range steps {
		// Calculate next_trigger_time based on cumulative delays
		var nextTriggerTime time.Time
		var status string
		
		if i == 0 {
			// First step - process immediately
			nextTriggerTime = currentTime
			status = "active"
		} else {
			// Subsequent steps - add cumulative delay
			nextTriggerTime = currentTime.Add(time.Duration(cumulativeDelay) * time.Hour)
			status = "pending" // Not active yet
		}
		
		// Add current step's delay for next iteration
		if step.TriggerDelayHours > 0 {
			cumulativeDelay += step.TriggerDelayHours
		}
		
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
			status,              // status - only first is active
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
		
		logrus.Infof("Created step %d for %s - status: %s, trigger_time: %v", 
			step.DayNumber, lead.Phone, status, nextTriggerTime.Format("2006-01-02 15:04:05"))
	}
	
	logrus.Infof("Successfully enrolled %s with all %d steps", lead.Phone, len(steps))
	return nil
}
