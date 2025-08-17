#!/usr/bin/env python3
import re

# Read the file
with open('src/usecase/sequence_trigger_processor.go', 'r', encoding='utf-8') as f:
    content = f.read()

# Add context import if not present
if '"context"' not in content:
    content = content.replace(
        'import (\n\t"database/sql"',
        'import (\n\t"context"\n\t"database/sql"'
    )
    print("Added context import")

# Find and replace the updateContactProgress function
# First, let's find where it starts
start_pattern = r'// updateContactProgress completes current step and activates next step\nfunc \(s \*SequenceTriggerProcessor\) updateContactProgress'
match = re.search(start_pattern, content)

if match:
    print(f"Found function at position {match.start()}")
    
    # Find the end of the function by counting braces
    start_pos = match.start()
    brace_count = 0
    in_func = False
    end_pos = start_pos
    
    for i in range(start_pos, len(content)):
        if content[i] == '{':
            brace_count += 1
            in_func = True
        elif content[i] == '}':
            brace_count -= 1
            if in_func and brace_count == 0:
                end_pos = i + 1
                break
    
    print(f"Function ends at position {end_pos}")
    
    # Replace with new function
    new_func = '''// updateContactProgress completes current step and activates next step
// Optimized for 3000 concurrent devices with proper locking
func (s *SequenceTriggerProcessor) updateContactProgress(contactID string, nextTrigger sql.NullString, delayHours int) error {
	// Start transaction with proper isolation level for concurrent access
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	
	// Step 1: Mark current step as completed with row-level lock
	query := `
		UPDATE sequence_contacts 
		SET status = 'completed', 
			completed_at = NOW()
		WHERE id = $1 AND status = 'active'
		RETURNING sequence_id, contact_phone, current_step
	`
	
	var sequenceID, phone string
	var currentStep int
	err = tx.QueryRow(query, contactID).Scan(&sequenceID, &phone, &currentStep)
	if err == sql.ErrNoRows {
		// Already processed by another device
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to mark contact as completed: %w", err)
	}
	
	logrus.Infof("COMPLETED: Step %d for %s", currentStep, phone)
	
	// Step 2: Find and activate the next pending step by EARLIEST trigger time
	// Use FOR UPDATE SKIP LOCKED to handle concurrent access from 3000 devices
	activateNextQuery := `
		WITH next_step AS (
			SELECT id, current_step, current_trigger, next_trigger_time
			FROM sequence_contacts
			WHERE sequence_id = $1 
				AND contact_phone = $2 
				AND status = 'pending'
				AND next_trigger_time <= NOW()
			ORDER BY next_trigger_time ASC  -- EARLIEST first, not by step number
			LIMIT 1
			FOR UPDATE SKIP LOCKED         -- Skip if another device is processing
		)
		UPDATE sequence_contacts sc
		SET status = 'active'
		FROM next_step ns
		WHERE sc.id = ns.id
		RETURNING ns.id, ns.current_step, ns.current_trigger, ns.next_trigger_time
	`
	
	var nextStepID string
	var nextStep int
	var nextStepTrigger string
	var nextTriggerTime time.Time
	
	err = tx.QueryRow(activateNextQuery, sequenceID, phone).Scan(
		&nextStepID, &nextStep, &nextStepTrigger, &nextTriggerTime)
	
	if err == sql.ErrNoRows {
		// No more pending steps ready - check if sequence is complete
		var pendingCount int
		tx.QueryRow(`
			SELECT COUNT(*) 
			FROM sequence_contacts 
			WHERE sequence_id = $1 AND contact_phone = $2 AND status = 'pending'
		`, sequenceID, phone).Scan(&pendingCount)
		
		if pendingCount == 0 {
			logrus.Infof("SEQUENCE COMPLETE: All steps finished for %s", phone)
			
			// Get the trigger to remove from lead
			var trigger string
			tx.QueryRow("SELECT current_trigger FROM sequence_contacts WHERE id = $1", contactID).Scan(&trigger)
			
			// Commit transaction before removing trigger
			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit transaction: %w", err)
			}
			
			// Remove trigger from lead
			s.removeCompletedTriggerFromLead(phone, trigger)
			return nil
		} else {
			logrus.Infof("WAITING: %d steps pending but not ready yet for %s", pendingCount, phone)
		}
	} else if err != nil {
		return fmt.Errorf("failed to activate next step: %w", err)
	} else {
		// Successfully activated next step
		logrus.Infof("ACTIVATED: Step %d for %s (trigger: %s) - next trigger at %v", 
			nextStep, phone, nextStepTrigger, nextTriggerTime.Format("15:04:05"))
		
		// Check how many steps remain
		var remainingCount int
		tx.QueryRow(`
			SELECT COUNT(*) 
			FROM sequence_contacts 
			WHERE sequence_id = $1 
			AND contact_phone = $2 
			AND status = 'pending'
		`, sequenceID, phone).Scan(&remainingCount)
		
		logrus.Infof("Progress: %s completed step %d -> activated step %d (with %d more pending)", 
			phone, currentStep, nextStep, remainingCount)
	}
	
	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}'''
    
    content = content[:start_pos] + new_func + content[end_pos:]
    print("Replaced function successfully")
else:
    print("Could not find updateContactProgress function")

# Write back with UTF-8 encoding
with open('src/usecase/sequence_trigger_processor.go', 'w', encoding='utf-8') as f:
    f.write(content)

print("Fixed sequence_trigger_processor.go")