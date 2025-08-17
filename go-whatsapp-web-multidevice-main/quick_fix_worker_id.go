// Quick fix to add to GetPendingMessagesAndLock after line 540
// This will help debug what's happening:

		result, err := tx.Exec(updateQuery, args...)
		if err != nil {
			// This is the problem - it continues even if UPDATE fails!
			logrus.Errorf("Failed to update message status: %v", err)
			// Add this to make it fail properly:
			return nil, fmt.Errorf("failed to update processing_worker_id: %w", err)
		}
		
		// Add this to verify the update worked:
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected == 0 {
			logrus.Errorf("WARNING: No rows were updated! Messages might already be processed")
			return nil, fmt.Errorf("no rows updated - messages may already be claimed")
		}
		
		logrus.Infof("Updated %d messages with processing_worker_id = %s", rowsAffected, workerID)
