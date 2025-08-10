// MODIFIED SEQUENCE TRIGGER PROCESSOR - NEW LOGIC
// All steps start as PENDING
// Worker finds earliest pending and processes based on time

// Key changes in enrollContactInSequence function:
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
    
    logrus.Infof("Enrolling contact %s in sequence %s - creating ALL %d steps as PENDING", 
        lead.Phone, sequenceID, len(steps))
    
    // Create records for ALL steps as PENDING
    currentTime := time.Now()
    var previousTriggerTime time.Time
    
    for i, step := range steps {
        // Calculate next_trigger_time based on previous step
        var nextTriggerTime time.Time
        
        if i == 0 {
            // FIRST STEP: Set to trigger in 5 minutes
            nextTriggerTime = currentTime.Add(5 * time.Minute)
            logrus.Infof("Step 1: PENDING - will trigger at %v (NOW + 5 minutes)", 
                nextTriggerTime.Format("15:04:05"))
        } else {
            // Subsequent steps - calculate based on delay
            if step.TriggerDelayHours > 0 {
                nextTriggerTime = previousTriggerTime.Add(time.Duration(step.TriggerDelayHours) * time.Hour)
            } else {
                // If no delay specified, default to 24 hours
                nextTriggerTime = previousTriggerTime.Add(24 * time.Hour)
            }
            
            logrus.Infof("Step %d: PENDING - will trigger at %v (previous + %d hours)", 
                i+1, nextTriggerTime.Format("15:04:05"), step.TriggerDelayHours)
        }
        
        // ALL STEPS START AS PENDING
        status := "pending"
        
        // Insert the sequence_contact record
        insertQuery := `
            INSERT INTO sequence_contacts 
            (id, sequence_id, sequence_stepid, contact_phone, contact_name, 
             current_trigger, current_step, status, created_at, 
             next_trigger_time, user_id, assigned_device_id) 
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
        `
        
        contactID := uuid.New().String()
        _, err = s.db.Exec(insertQuery, 
            contactID, sequenceID, step.ID, lead.Phone, lead.Name,
            step.Trigger, step.DayNumber, status, currentTime,
            nextTriggerTime, lead.UserID, lead.DeviceID)
            
        if err != nil {
            logrus.Errorf("Failed to insert sequence contact: %v", err)
            continue
        }
        
        // Update for next iteration
        previousTriggerTime = nextTriggerTime
    }
    
    logrus.Infof("✅ Created %d PENDING steps for %s", len(steps), lead.Phone)
    return nil
}
