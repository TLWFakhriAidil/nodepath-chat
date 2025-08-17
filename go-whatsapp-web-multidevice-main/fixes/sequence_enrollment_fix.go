package usecase

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"your-module/src/models"
)

// SequenceDebugFix is a temporary function to debug and fix sequence enrollment issues
func (s *SequenceTriggerProcessor) SequenceDebugFix() error {
	logrus.Info("=== SEQUENCE DEBUG FIX STARTED ===")
	
	// First, let's check what's in sequence_steps
	query := `
		SELECT s.name, ss.id, ss.day_number, ss.trigger, ss.next_trigger, ss.trigger_delay_hours
		FROM sequence_steps ss
		JOIN sequences s ON s.id = ss.sequence_id
		WHERE s.is_active = true
		ORDER BY s.name, ss.day_number
	`
	
	rows, err := s.db.Query(query)
	if err != nil {
		return fmt.Errorf("failed to query sequence steps: %w", err)
	}
	defer rows.Close()
	
	logrus.Info("Current sequence steps:")
	for rows.Next() {
		var seqName, stepID, trigger string
		var dayNumber, delayHours int
		var nextTrigger sql.NullString
		
		if err := rows.Scan(&seqName, &stepID, &dayNumber, &trigger, &nextTrigger, &delayHours); err != nil {
			continue
		}
		
		logrus.Infof("Sequence: %s, Day: %d, StepID: %s, Trigger: %s -> %s (delay: %dh)", 
			seqName, dayNumber, stepID, trigger, nextTrigger.String, delayHours)
	}
	
	// Clean up all sequence_contacts
	logrus.Info("Cleaning up sequence_contacts table...")
	_, err = s.db.Exec("DELETE FROM sequence_contacts")
	if err != nil {
		return fmt.Errorf("failed to clean sequence_contacts: %w", err)
	}
	
	logrus.Info("=== SEQUENCE DEBUG FIX COMPLETED ===")
	return nil
}

// Fixed enrollment function with proper current_step assignment
func (s *SequenceTriggerProcessor) enrollContactInSequenceFixed(sequenceID string, lead models.Lead, trigger string) error {
	// First, get all steps for this sequence
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
		ID              string
		DayNumber       int
		Trigger         string
		NextTrigger     sql.NullString
		TriggerDelayHours int
	}
	
	for rows.Next() {
		var step struct {
			ID              string
			DayNumber       int
			Trigger         string
			NextTrigger     sql.NullString
			TriggerDelayHours int
		}
		if err := rows.Scan(&step.ID, &step.DayNumber, &step.Trigger, &step.NextTrigger, &step.TriggerDelayHours); err != nil {
			continue
		}
		steps = append(steps, step)
	}
	
	if len(steps) == 0 {
		return fmt.Errorf("no steps found for sequence %s", sequenceID)
	}
	
	// Debug log
	logrus.Infof("Enrolling %s in sequence %s with %d steps", lead.Phone, sequenceID, len(steps))
	
	// Now create one sequence_contacts record for each step
	insertQuery := `
		INSERT INTO sequence_contacts (
			sequence_id, contact_phone, contact_name, 
			current_step, status, current_trigger,
			next_trigger_time, sequence_stepid, assigned_device_id, user_id
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (sequence_id, contact_phone, sequence_stepid) DO NOTHING
	`
	
	enrolledCount := 0
	currentTime := time.Now()
	
	for i, step := range steps {
		// Calculate when this step should be processed
		var nextTriggerTime time.Time
		if i == 0 {
			// First step processes immediately
			nextTriggerTime = currentTime
		} else {
			// Calculate based on cumulative delays from previous steps
			totalHours := 0
			for j := 0; j < i; j++ {
				totalHours += steps[j].TriggerDelayHours
			}
			nextTriggerTime = currentTime.Add(time.Duration(totalHours) * time.Hour)
		}
		
		// Determine status - first step is 'active', others are 'pending'
		status := "pending"
		if i == 0 {
			status = "active"
		}
		
		// IMPORTANT: Use the actual step number (i+1) instead of day_number if day_number is wrong
		currentStep := i + 1  // This ensures step 1, 2, 3, 4 instead of all being 4
		if step.DayNumber > 0 && step.DayNumber <= len(steps) {
			currentStep = step.DayNumber  // Use day_number if it's valid
		}
		
		logrus.Debugf("Enrolling step %d: current_step=%d, status=%s, trigger=%s, next_time=%v", 
			i+1, currentStep, status, step.Trigger, nextTriggerTime)
		
		_, err := s.db.Exec(insertQuery, 
			sequenceID,          // sequence_id
			lead.Phone,          // contact_phone
			lead.Name,           // contact_name
			currentStep,         // current_step (FIXED)
			status,              // status
			step.Trigger,        // current_trigger
			nextTriggerTime,     // next_trigger_time
			step.ID,             // sequence_stepid
			lead.DeviceID,       // assigned_device_id
			lead.UserID,         // user_id
		)
		
		if err != nil {
			logrus.Warnf("Failed to enroll contact %s for step %d: %v", lead.Phone, currentStep, err)
			continue
		}
		
		enrolledCount++
	}
	
	if enrolledCount > 0 {
		logrus.Infof("Successfully enrolled %s in sequence %s with %d steps", lead.Phone, sequenceID, enrolledCount)
	}
	
	return nil
}
