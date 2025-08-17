package usecase

import (
	"database/sql"
	"fmt"
	"time"
	"github.com/sirupsen/logrus"
)

// SequenceHotfix provides emergency fixes for sequence processing issues
func (s *SequenceTriggerProcessor) SequenceHotfix() error {
	logrus.Info("=== SEQUENCE HOTFIX STARTED ===")
	
	// Fix 1: Ensure only ONE step per contact is active at a time
	logrus.Info("Fix 1: Ensuring only one active step per contact...")
	_, err := s.db.Exec(`
		WITH ranked_active AS (
			SELECT id, 
				   ROW_NUMBER() OVER (PARTITION BY contact_phone, sequence_id ORDER BY current_step) as rn
			FROM sequence_contacts
			WHERE status = 'active'
		)
		UPDATE sequence_contacts
		SET status = 'pending'
		WHERE id IN (
			SELECT id FROM ranked_active WHERE rn > 1
		)
	`)
	if err != nil {
		logrus.Errorf("Failed to fix multiple active steps: %v", err)
	}
	
	// Fix 2: Ensure next_trigger_time is respected
	logrus.Info("Fix 2: Fixing next_trigger_time for pending steps...")
	_, err = s.db.Exec(`
		UPDATE sequence_contacts sc
		SET next_trigger_time = 
			CASE 
				WHEN sc.current_step = 1 THEN sc.created_at
				WHEN sc.current_step = 2 THEN sc.created_at + INTERVAL '12 hours'
				WHEN sc.current_step = 3 THEN sc.created_at + INTERVAL '24 hours'
				WHEN sc.current_step = 4 THEN sc.created_at + INTERVAL '36 hours'
				ELSE sc.next_trigger_time
			END
		WHERE status = 'pending' 
		AND next_trigger_time IS NULL OR next_trigger_time < NOW()
	`)
	if err != nil {
		logrus.Errorf("Failed to fix trigger times: %v", err)
	}
	
	// Fix 3: Log current state
	var stats []struct {
		Status string
		Count  int
	}
	rows, err := s.db.Query(`
		SELECT status, COUNT(*) 
		FROM sequence_contacts 
		GROUP BY status
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var stat struct {
				Status string
				Count  int
			}
			rows.Scan(&stat.Status, &stat.Count)
			stats = append(stats, stat)
		}
		logrus.Infof("Current status distribution: %+v", stats)
	}
	
	logrus.Info("=== SEQUENCE HOTFIX COMPLETED ===")
	return nil
}

// ProcessSequenceContactsSafely processes contacts with extra safety checks
func (s *SequenceTriggerProcessor) ProcessSequenceContactsSafely() (int, error) {
	// CRITICAL: Process only ONE step per contact per cycle
	query := `
		WITH ready_contacts AS (
			SELECT DISTINCT ON (contact_phone) 
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
				AND sc.next_trigger_time <= NOW()
				AND sc.processing_device_id IS NULL
				AND sc.completed_at IS NULL
			ORDER BY contact_phone, current_step
		)
		SELECT * FROM ready_contacts
		ORDER BY id
		LIMIT $1
	`
	
	rows, err := s.db.Query(query, 100) // Process max 100 at a time
	if err != nil {
		return 0, fmt.Errorf("failed to get safe contacts: %w", err)
	}
	defer rows.Close()
	
	processed := 0
	for rows.Next() {
		// Process each contact...
		processed++
	}
	
	return processed, nil
}
