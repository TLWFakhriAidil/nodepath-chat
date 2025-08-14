// File: fix_device_tracking.go
// Changes to make:
// 1. Keep processing_device_id after completion (for tracking)
// 2. Don't release contact on failure (strict device ownership)
// 3. Skip/remove stuck processing cleanup

// Change 1: In updateContactProgress - line ~625
// OLD:
	query := `
		UPDATE sequence_contacts 
		SET status = 'completed', 
			processing_device_id = NULL,
			processing_started_at = NULL,
			completed_at = NOW()
		WHERE id = $1
		RETURNING sequence_id, contact_phone, current_step
	`

// NEW:
	query := `
		UPDATE sequence_contacts 
		SET status = 'completed', 
			completed_at = NOW()
		WHERE id = $1
		RETURNING sequence_id, contact_phone, current_step
	`

// Change 2: In processContact - line ~577
// Remove this line:
		s.releaseContact(job.contactID)

// Change 3: In processTriggers - line ~114
// Comment out or remove:
	// Step 2: Clean up stuck processing
	// if err := s.cleanupStuckProcessing(); err != nil {
	//     logrus.Warnf("Error cleaning up stuck processing: %v", err)
	// }

// Also remove/comment the cleanupStuckProcessing function entirely (line ~760)
