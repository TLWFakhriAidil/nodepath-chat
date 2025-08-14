package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type sequenceRepository struct {
	db *sql.DB
}

var sequenceRepo *sequenceRepository

// GetSequenceRepository returns sequence repository instance
func GetSequenceRepository() *sequenceRepository {
	if sequenceRepo == nil {
		sequenceRepo = &sequenceRepository{
			db: database.GetDB(),
		}
	}
	return sequenceRepo
}

// CreateSequence creates a new sequence
func (r *sequenceRepository) CreateSequence(sequence *models.Sequence) error {
	sequence.ID = uuid.New().String()
	sequence.CreatedAt = time.Now()
	sequence.UpdatedAt = time.Now()

	query := `
		INSERT INTO sequences(id, user_id, device_id, name, description, niche, status, ` + "`trigger`" + `, start_trigger, end_trigger, total_days, is_active, time_schedule, 
		                      min_delay_seconds, max_delay_seconds, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.Exec(query, sequence.ID, sequence.UserID, nil, // device_id is NULL - sequences use all user devices
		sequence.Name, sequence.Description, sequence.Niche, sequence.Status, 
		sequence.Trigger, sequence.StartTrigger, sequence.EndTrigger, sequence.TotalDays, 
		sequence.IsActive, sequence.TimeSchedule, sequence.MinDelaySeconds, sequence.MaxDelaySeconds,
		sequence.CreatedAt, sequence.UpdatedAt)
		
	return err
}

// GetSequences gets all sequences for a user
func (r *sequenceRepository) GetSequences(userID string) ([]models.Sequence, error) {
	query := `
		SELECT id, user_id, device_id, name, description, niche, status, 
		       COALESCE(start_trigger, '') AS start_trigger,
		       COALESCE(end_trigger, '') AS end_trigger,
		       total_days, is_active, 
		       COALESCE(schedule_time, '09:00') AS schedule_time, 
		       COALESCE(min_delay_seconds, 10) AS min_delay_seconds,
		       COALESCE(max_delay_seconds, 30) AS max_delay_seconds,
		       created_at, updated_at
		FROM sequences
		WHERE user_id = ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query, userID)
	if err != nil {
		logrus.Errorf("Failed to query sequences: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	var sequences []models.Sequence
	for rows.Next() {
		var seq models.Sequence
		err := rows.Scan(&seq.ID, &seq.UserID, &seq.DeviceID, &seq.Name, 
			&seq.Description, &seq.Niche, &seq.Status, &seq.StartTrigger, &seq.EndTrigger,
			&seq.TotalDays, &seq.IsActive, 
			&seq.TimeSchedule, &seq.MinDelaySeconds, &seq.MaxDelaySeconds,
			&seq.CreatedAt, &seq.UpdatedAt)
		if err != nil {
			logrus.Errorf("Failed to scan sequence row: %v", err)
			continue
		}
		sequences = append(sequences, seq)
	}
	
	logrus.Infof("Repository: Found %d sequences for user %s", len(sequences), userID)
	return sequences, nil
}
// GetSequenceByID gets sequence by ID
func (r *sequenceRepository) GetSequenceByID(sequenceID string) (*models.Sequence, error) {
	query := `

		SELECT id, user_id, device_id, name, description, niche, status, 
		       COALESCE(start_trigger, '') AS start_trigger,
		       COALESCE(end_trigger, '') AS end_trigger,
		       total_days, is_active, 
		       COALESCE(schedule_time, '09:00') AS schedule_time,
		       COALESCE(min_delay_seconds, 10) AS min_delay_seconds,
		       COALESCE(max_delay_seconds, 30) AS max_delay_seconds,
		       created_at, updated_at
		FROM sequences
		WHERE id = ?
	`
	
	var seq models.Sequence
	err := r.db.QueryRow(query, sequenceID).Scan(&seq.ID, &seq.UserID, &seq.DeviceID, 
		&seq.Name, &seq.Description, &seq.Niche, &seq.Status, &seq.StartTrigger, &seq.EndTrigger,
		&seq.TotalDays, &seq.IsActive, 
		&seq.TimeSchedule, &seq.MinDelaySeconds, &seq.MaxDelaySeconds,
		&seq.CreatedAt, &seq.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			logrus.Errorf("Sequence not found: %s", sequenceID)
		} else {
			logrus.Errorf("Failed to get sequence %s: %v", sequenceID, err)
		}
		return nil, err
	}
	
	return &seq, nil
}

// UpdateSequence updates a sequence
func (r *sequenceRepository) UpdateSequence(sequence *models.Sequence) error {
	sequence.UpdatedAt = time.Now()
	
	query := `

		UPDATE sequences 
		SET name = ?, description = ?, niche = ?, status = ?, 
		    start_trigger = ?, end_trigger = ?, total_days = ?, 
		    is_active = ?, schedule_time = ?, min_delay_seconds = ?, 
		    max_delay_seconds = ?, updated_at = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query, sequence.Name, sequence.Description, sequence.Niche, 
		sequence.Status, sequence.StartTrigger, sequence.EndTrigger, sequence.TotalDays, 
		sequence.IsActive, sequence.TimeSchedule, sequence.MinDelaySeconds, 
		sequence.MaxDelaySeconds, sequence.UpdatedAt, sequence.ID)
		
	return err
}

// DeleteSequence deletes a sequence
func (r *sequenceRepository) DeleteSequence(sequenceID string) error {
	_, err := r.db.Exec("DELETE FROM sequences WHERE id = ?", sequenceID)
	return err
}

// DeleteSequenceSteps deletes all steps for a sequence
func (r *sequenceRepository) DeleteSequenceSteps(sequenceID string) error {
	_, err := r.db.Exec("DELETE FROM sequence_steps WHERE sequence_id = ?", sequenceID)
	return err
}

// CreateSequenceStep creates a step in sequence
func (r *sequenceRepository) CreateSequenceStep(step *models.SequenceStep) error {
	step.ID = uuid.New().String()

	query := `
		INSERT INTO sequence_steps(
			id, sequence_id, day_number, message_type, content, 
			media_url, caption, delay_days, time_schedule, ` + "`trigger`" + `,
			next_trigger, trigger_delay_hours, is_entry_point,
			min_delay_seconds, max_delay_seconds
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	// Use DayNumber
	dayNumber := step.DayNumber
	if dayNumber == 0 {
		dayNumber = 1
	}
	
	// Default values
	if step.TimeSchedule == "" {
		step.TimeSchedule = "10:00"
	}
	if step.MessageType == "" {
		step.MessageType = "text"
	}
	if step.TriggerDelayHours == 0 {
		step.TriggerDelayHours = 24
	}
	if step.MinDelaySeconds == 0 {
		step.MinDelaySeconds = 10
	}
	if step.MaxDelaySeconds == 0 {
		step.MaxDelaySeconds = 30
	}
	
	// Default DelayDays if not set
	delayDays := step.DelayDays
	if delayDays == 0 {
		delayDays = 1
	}
	
	_, err := r.db.Exec(query, 
		step.ID, step.SequenceID, dayNumber, step.MessageType, step.Content,
		step.MediaURL, step.Caption, delayDays, step.TimeSchedule, step.Trigger,
		step.NextTrigger, step.TriggerDelayHours, step.IsEntryPoint,
		step.MinDelaySeconds, step.MaxDelaySeconds)
		
	if err != nil {
		logrus.Errorf("Failed to create sequence step: %v", err)
		logrus.Errorf("Step details - ID: %s, SequenceID: %s, DayNumber: %d, MessageType: %s", 
			step.ID, step.SequenceID, dayNumber, step.MessageType)
		logrus.Errorf("Content length: %d, MediaURL: %s, Caption: %s", 
			len(step.Content), step.MediaURL, step.Caption)
	}
	
	return err
}
// GetSequenceSteps gets all steps for a sequence
func (r *sequenceRepository) GetSequenceSteps(sequenceID string) ([]models.SequenceStep, error) {
	logrus.Infof("Getting steps for sequence: %s", sequenceID)
	
	// First, let's check if there are any steps at all
	var count int
	countQuery := `SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = ?`
	err := r.db.QueryRow(countQuery, sequenceID).Scan(&count)
	if err != nil {
		logrus.Errorf("Error counting sequence steps: %v", err)
		return nil, fmt.Errorf("failed to count steps: %v", err)
	} else {
		logrus.Infof("Total steps in database for sequence %s: %d", sequenceID, count)
	}
	
	// If no steps found, return empty slice instead of nil
	if count == 0 {
		logrus.Infof("No steps found for sequence %s, returning empty array", sequenceID)
		return []models.SequenceStep{}, nil
	}
	
	query := `
		SELECT 
			id, sequence_id, 
			COALESCE(day_number, 1) AS day_number,
			COALESCE(` + "`trigger`" + `, '') as ` + "`trigger`" + `, 
			COALESCE(next_trigger, '') as next_trigger,
			COALESCE(trigger_delay_hours, 24) as trigger_delay_hours,
			COALESCE(is_entry_point, false) as is_entry_point,
			COALESCE(message_type, 'text') as message_type, 
			COALESCE(content, '') as content, 
			COALESCE(media_url, '') as media_url, 
			COALESCE(caption, '') as caption, 
			COALESCE(time_schedule, '') as time_schedule,
			COALESCE(min_delay_seconds, 10) as min_delay_seconds,
			COALESCE(max_delay_seconds, 30) as max_delay_seconds,
			COALESCE(delay_days, 0) as delay_days
		FROM sequence_steps
		WHERE sequence_id = ?
		ORDER BY day_number ASC
	`
	
	rows, err := r.db.Query(query, sequenceID)
	if err != nil {
		logrus.Errorf("Error querying sequence steps: %v", err)
		return nil, err
	}
	defer rows.Close()
	
	var steps []models.SequenceStep
	for rows.Next() {
		var step models.SequenceStep
		err := rows.Scan(&step.ID, &step.SequenceID, &step.DayNumber, 
			&step.Trigger, &step.NextTrigger, &step.TriggerDelayHours, &step.IsEntryPoint,
			&step.MessageType, &step.Content, &step.MediaURL, &step.Caption, 
			&step.TimeSchedule, &step.MinDelaySeconds, &step.MaxDelaySeconds, &step.DelayDays)
		if err != nil {
			logrus.Errorf("Error scanning sequence step: %v", err)
			logrus.Errorf("Failed on sequence_id: %s, error details: %+v", sequenceID, err)
			// Try to get column info for debugging
			cols, _ := rows.Columns()
			logrus.Errorf("Column names: %v", cols)
			// Don't skip, return the error to understand what's wrong
			return nil, fmt.Errorf("failed to scan sequence step: %v", err)
		}
		steps = append(steps, step)
		logrus.Debugf("Successfully scanned step: day=%d, trigger=%s", step.DayNumber, step.Trigger)
	}
	
	if err = rows.Err(); err != nil {
		logrus.Errorf("Error iterating rows: %v", err)
	}
	
	logrus.Infof("Found %d steps for sequence %s", len(steps), sequenceID)
	return steps, nil
}

// AddContactToSequence adds a contact to sequence
func (r *sequenceRepository) AddContactToSequence(contact *models.SequenceContact) error {
	contact.ID = uuid.New().String()
	contact.CompletedAt = &[]time.Time{time.Now()}[0]
	contact.CurrentStep = 0
	contact.Status = "active"

	query := `

		INSERT INTO sequence_contacts(id, sequence_id, contact_phone, contact_name, current_step, status, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (sequence_id, contact_phone) DO NOTHING
	`
	
	_, err := r.db.Exec(query, contact.ID, contact.SequenceID, contact.ContactPhone,
		contact.ContactName, contact.CurrentStep, contact.Status, contact.CompletedAt)
		
	return err
}
// GetSequenceContacts gets all contacts in a sequence
func (r *sequenceRepository) GetSequenceContacts(sequenceID string) ([]models.SequenceContact, error) {
	query := `

		SELECT id, sequence_id, contact_phone, contact_name, current_step, status, 
			   completed_at
		FROM sequence_contacts
		WHERE sequence_id = ?
		ORDER BY completed_at DESC
	`
	
	rows, err := r.db.Query(query, sequenceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var contacts []models.SequenceContact
	for rows.Next() {
		var contact models.SequenceContact
		err := rows.Scan(&contact.ID, &contact.SequenceID, &contact.ContactPhone,
			&contact.ContactName, &contact.CurrentStep, &contact.Status,
			&contact.CompletedAt)
		if err != nil {
			continue
		}
		contacts = append(contacts, contact)
	}
	
	return contacts, nil
}

// GetActiveSequenceContacts gets contacts ready for next message
func (r *sequenceRepository) GetActiveSequenceContacts(currentTime time.Time) ([]models.SequenceContact, error) {
	query := `

		SELECT sc.id, sc.sequence_id, sc.contact_phone, sc.contact_name, 
			   sc.current_step, sc.status, sc.completed_at
		FROM sequence_contacts sc
		JOIN sequences s ON s.id = sc.sequence_id
		WHERE sc.status = 'active' 
		AND s.is_active = true
	`
	
	rows, err := r.db.Query(query, currentTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var contacts []models.SequenceContact
	for rows.Next() {
		var contact models.SequenceContact
		err := rows.Scan(&contact.ID, &contact.SequenceID, &contact.ContactPhone,
			&contact.ContactName, &contact.CurrentStep, &contact.Status,
			&contact.CompletedAt)
		if err != nil {
			continue
		}
		contacts = append(contacts, contact)
	}
	
	return contacts, nil
}
// UpdateContactProgress updates contact's progress in sequence
func (r *sequenceRepository) UpdateContactProgress(contactID string, currentStep int, status string) error {
	query := `

		UPDATE sequence_contacts 
		SET current_step = ?, status = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query, currentStep, status, contactID)
	return err
}

// MarkContactCompleted marks contact as completed
func (r *sequenceRepository) MarkContactCompleted(contactID string) error {
	now := time.Now()
	query := `

		UPDATE sequence_contacts SET status = 'completed', completed_at = ?
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query, now, contactID)
	return err
}

// CreateSequenceLog logs a sent message
func (r *sequenceRepository) CreateSequenceLog(log *models.SequenceLog) error {
	log.ID = uuid.New().String()
	log.SentAt = time.Now()

	query := `

		INSERT INTO sequence_logs(id, sequence_id, contact_id, step_id, day, status, message_id, error_message, sent_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.Exec(query, log.ID, log.SequenceID, log.ContactID, log.StepID,
		log.Day, log.Status, log.MessageID, log.ErrorMessage, log.SentAt)
		
	return err
}

// GetSequenceStats gets statistics for a sequence
func (r *sequenceRepository) GetSequenceStats(sequenceID string) (map[string]int, error) {
	stats := make(map[string]int)
	
	// Get contact counts by status
	query := `

		SELECT status, COUNT(*) AS count
		FROM sequence_contacts
		WHERE sequence_id = ?
		GROUP BY status
	`
	
	rows, err := r.db.Query(query, sequenceID)
	if err != nil {
		return stats, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err == nil {
			stats[status] = count
		}
	}
	
	// Get total messages sent
	var messageCount int
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM sequence_logs WHERE sequence_id = ?
	`, sequenceID).Scan(&messageCount)
	
	if err == nil {
		stats["messages_sent"] = messageCount
	}
	
	return stats, nil
}

// GetActiveSequencesWithNiche gets all active sequences that have a niche
func (r *sequenceRepository) GetActiveSequencesWithNiche() ([]models.Sequence, error) {
	query := `

		SELECT id, user_id, device_id, name, description, niche, total_days, is_active, created_at, updated_at
		FROM sequences
		WHERE is_active = true
		AND niche IS NOT NULL
		AND niche != ''
		ORDER BY created_at DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var sequences []models.Sequence
	for rows.Next() {
		var seq models.Sequence
		err := rows.Scan(&seq.ID, &seq.UserID, &seq.DeviceID, &seq.Name, 
			&seq.Description, &seq.Niche, &seq.TotalDays, &seq.IsActive, 
			&seq.CreatedAt, &seq.UpdatedAt)
		if err != nil {
			continue
		}
		sequences = append(sequences, seq)
	}
	
	return sequences, nil
}

// UpdateSequenceStep updates an existing sequence step
func (r *sequenceRepository) UpdateSequenceStep(step *models.SequenceStep) error {
	query := `
		UPDATE sequence_steps SET
			message_type = ?,
			content = ?,
			media_url = ?,
			caption = ?,
			time_schedule = ?,
			` + "`trigger`" + ` = ?,
			next_trigger = ?,
			trigger_delay_hours = ?,
			is_entry_point = ?,
			min_delay_seconds = ?,
			max_delay_seconds = ?,
			updated_at = NOW()
		WHERE id = ?
	`
	
	_, err := r.db.Exec(query,
		step.MessageType, step.Content, step.MediaURL, step.Caption,
		step.TimeSchedule, step.Trigger, step.NextTrigger,
		step.TriggerDelayHours, step.IsEntryPoint,
		step.MinDelaySeconds, step.MaxDelaySeconds,
		step.ID)
		
	if err != nil {
		logrus.Errorf("Failed to update sequence step %s: %v", step.ID, err)
	}
	
	return err
}

// DeleteSequenceStep deletes a single sequence step by ID
func (r *sequenceRepository) DeleteSequenceStep(stepID string) error {
	_, err := r.db.Exec("DELETE FROM sequence_steps WHERE id = ?", stepID)
	if err != nil {
		logrus.Errorf("Failed to delete sequence step %s: %v", stepID, err)
	}
	return err
}
