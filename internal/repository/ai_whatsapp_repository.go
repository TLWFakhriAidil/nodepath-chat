package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"nodepath-chat/internal/models"

	"github.com/sirupsen/logrus"
)

// AIWhatsappRepository interface defines methods for AI WhatsApp conversation management
type AIWhatsappRepository interface {
	// Create operations
	CreateAIWhatsapp(ai *models.AIWhatsapp) error
	CreateConversationLog(log *models.ConversationLog) error

	// Read operations
	GetAIWhatsappByProspectNum(prospectNum string) (*models.AIWhatsapp, error)
	GetAIWhatsappByID(id int) (*models.AIWhatsapp, error)
	GetAIWhatsappByStaff(idStaff string) ([]models.AIWhatsapp, error)
	GetAIWhatsappByNiche(niche string) ([]models.AIWhatsapp, error)
	GetActiveAIConversations() ([]models.AIWhatsapp, error)
	GetConversationHistory(prospectNum string, limit int) ([]models.ConversationLog, error)
	GetConversationLogsByStage(stage string) ([]models.ConversationLog, error)

	// Update operations
	UpdateAIWhatsapp(ai *models.AIWhatsapp) error
	UpdateConversationStage(prospectNum string, stage string) error
	UpdateHumanTakeover(prospectNum string, human int) error
	UpdateConvCurrent(prospectNum string, convCurrent string) error
	UpdateConvLast(prospectNum string, convLast interface{}) error

	// Delete operations
	DeleteAIWhatsapp(id int) error
	DeleteConversationLogs(prospectNum string) error

	// Analytics operations
	GetConversationStats(idStaff string) (map[string]int, error)
	GetActiveConversationCount() (int, error)
	GetConversationsByDateRange(startDate, endDate time.Time) ([]models.AIWhatsapp, error)
}

// aiWhatsappRepository implements AIWhatsappRepository interface
type aiWhatsappRepository struct {
	db *sql.DB
}

// NewAIWhatsappRepository creates a new instance of AIWhatsappRepository
func NewAIWhatsappRepository(db *sql.DB) AIWhatsappRepository {
	return &aiWhatsappRepository{
		db: db,
	}
}

// CreateAIWhatsapp creates a new AI WhatsApp conversation record
func (r *aiWhatsappRepository) CreateAIWhatsapp(ai *models.AIWhatsapp) error {
	ai.CreatedAt = time.Now()
	ai.UpdatedAt = time.Now()

	// Convert conv_last to JSON string if it's not nil
	var convLastJSON []byte
	var err error
	if ai.ConvLast != nil {
		convLastJSON, err = json.Marshal(ai.ConvLast)
		if err != nil {
			return fmt.Errorf("failed to marshal conv_last: %w", err)
		}
	}

	query := `
		INSERT INTO ai_whatsapp_nodepath (
			id_prospect, id_staff, prospect_num, stage, conv_last, 
			conv_current, human, niche, jam, intro, 
			catatan_staff, balas, data_image, conv_stage, 
			bot_balas, keywordiklan, marketer, update_today, 
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = r.db.Exec(query,
		ai.IDProspect, ai.IDStaff, ai.ProspectNum, ai.Stage, string(convLastJSON),
		ai.ConvCurrent, ai.Human, ai.Niche, ai.Jam, ai.Intro,
		ai.CatatanStaff, ai.Balas, ai.DataImage, ai.ConvStage,
		ai.BotBalas, ai.KeywordIklan, ai.Marketer, ai.UpdateToday,
		ai.CreatedAt, ai.UpdatedAt,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to create AI WhatsApp conversation")
		return fmt.Errorf("failed to create AI WhatsApp conversation: %w", err)
	}

	logrus.WithField("prospect_num", ai.ProspectNum).Info("AI WhatsApp conversation created successfully")
	return nil
}

// CreateConversationLog creates a new conversation log entry
func (r *aiWhatsappRepository) CreateConversationLog(log *models.ConversationLog) error {
	log.CreatedAt = time.Now()

	query := `
		INSERT INTO conversation_log_nodepath (
			prospect_num, id_staff, message, sender, stage, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		log.ProspectNum, log.IDStaff, log.Message, log.Sender, log.Stage, log.CreatedAt,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to create conversation log")
		return fmt.Errorf("failed to create conversation log: %w", err)
	}

	return nil
}

// GetAIWhatsappByProspectNum retrieves AI WhatsApp conversation by prospect number
func (r *aiWhatsappRepository) GetAIWhatsappByProspectNum(prospectNum string) (*models.AIWhatsapp, error) {
	query := `
		SELECT id_prospect, id_staff, prospect_num, stage, conv_last, 
		       conv_current, human, niche, jam, intro, 
		       catatan_staff, balas, data_image, conv_stage, 
		       bot_balas, keywordiklan, marketer, update_today, 
		       created_at, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE prospect_num = ?
	`

	row := r.db.QueryRow(query, prospectNum)

	ai := &models.AIWhatsapp{}
	var convLastJSON sql.NullString

	err := row.Scan(
		&ai.IDProspect, &ai.IDStaff, &ai.ProspectNum, &ai.Stage, &convLastJSON,
		&ai.ConvCurrent, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
		&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
		&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
		&ai.CreatedAt, &ai.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		logrus.WithError(err).Error("Failed to get AI WhatsApp conversation by prospect number")
		return nil, fmt.Errorf("failed to get AI WhatsApp conversation: %w", err)
	}

	// Parse conv_last JSON if it exists
	if convLastJSON.Valid && convLastJSON.String != "" {
		if err := json.Unmarshal([]byte(convLastJSON.String), &ai.ConvLast); err != nil {
			logrus.WithError(err).Warn("Failed to unmarshal conv_last JSON")
		}
	}

	return ai, nil
}

// GetAIWhatsappByID retrieves AI WhatsApp conversation by ID
func (r *aiWhatsappRepository) GetAIWhatsappByID(id int) (*models.AIWhatsapp, error) {
	query := `
		SELECT id_prospect, id_staff, prospect_num, stage, conv_last, 
		       conv_current, human, niche, jam, intro, 
		       catatan_staff, balas, data_image, conv_stage, 
		       bot_balas, keywordiklan, marketer, update_today, 
		       created_at, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE id_prospect = ?
	`

	row := r.db.QueryRow(query, id)

	ai := &models.AIWhatsapp{}
	var convLastJSON sql.NullString

	err := row.Scan(
		&ai.IDProspect, &ai.IDStaff, &ai.ProspectNum, &ai.Stage, &convLastJSON,
		&ai.ConvCurrent, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
		&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
		&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
		&ai.CreatedAt, &ai.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		logrus.WithError(err).Error("Failed to get AI WhatsApp conversation by ID")
		return nil, fmt.Errorf("failed to get AI WhatsApp conversation: %w", err)
	}

	// Parse conv_last JSON if it exists
	if convLastJSON.Valid && convLastJSON.String != "" {
		if err := json.Unmarshal([]byte(convLastJSON.String), &ai.ConvLast); err != nil {
			logrus.WithError(err).Warn("Failed to unmarshal conv_last JSON")
		}
	}

	return ai, nil
}

// GetAIWhatsappByStaff retrieves all AI WhatsApp conversations for a specific staff member
func (r *aiWhatsappRepository) GetAIWhatsappByStaff(idStaff string) ([]models.AIWhatsapp, error) {
	query := `
		SELECT id_prospect, id_staff, prospect_num, stage, conv_last, 
		       conv_current, human, niche, jam, intro, 
		       catatan_staff, balas, data_image, conv_stage, 
		       bot_balas, keywordiklan, marketer, update_today, 
		       created_at, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE id_staff = ?
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Query(query, idStaff)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp conversations by staff")
		return nil, fmt.Errorf("failed to get AI WhatsApp conversations: %w", err)
	}
	defer rows.Close()

	var conversations []models.AIWhatsapp
	for rows.Next() {
		ai := models.AIWhatsapp{}
		var convLastJSON sql.NullString

		err := rows.Scan(
			&ai.IDProspect, &ai.IDStaff, &ai.ProspectNum, &ai.Stage, &convLastJSON,
			&ai.ConvCurrent, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
			&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
			&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
			&ai.CreatedAt, &ai.UpdatedAt,
		)

		if err != nil {
			logrus.WithError(err).Error("Failed to scan AI WhatsApp conversation")
			continue
		}

		// Parse conv_last JSON if it exists
		if convLastJSON.Valid && convLastJSON.String != "" {
			if err := json.Unmarshal([]byte(convLastJSON.String), &ai.ConvLast); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal conv_last JSON")
			}
		}

		conversations = append(conversations, ai)
	}

	return conversations, nil
}

// GetAIWhatsappByNiche retrieves all AI WhatsApp conversations for a specific niche
func (r *aiWhatsappRepository) GetAIWhatsappByNiche(niche string) ([]models.AIWhatsapp, error) {
	query := `
		SELECT id_prospect, id_staff, prospect_num, stage, conv_last, 
		       conv_current, human, niche, jam, intro, 
		       catatan_staff, balas, data_image, conv_stage, 
		       bot_balas, keywordiklan, marketer, update_today, 
		       created_at, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE niche = ?
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Query(query, niche)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp conversations by niche")
		return nil, fmt.Errorf("failed to get AI WhatsApp conversations: %w", err)
	}
	defer rows.Close()

	var conversations []models.AIWhatsapp
	for rows.Next() {
		ai := models.AIWhatsapp{}
		var convLastJSON sql.NullString

		err := rows.Scan(
			&ai.IDProspect, &ai.IDStaff, &ai.ProspectNum, &ai.Stage, &convLastJSON,
			&ai.ConvCurrent, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
			&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
			&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
			&ai.CreatedAt, &ai.UpdatedAt,
		)

		if err != nil {
			logrus.WithError(err).Error("Failed to scan AI WhatsApp conversation")
			continue
		}

		// Parse conv_last JSON if it exists
		if convLastJSON.Valid && convLastJSON.String != "" {
			if err := json.Unmarshal([]byte(convLastJSON.String), &ai.ConvLast); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal conv_last JSON")
			}
		}

		conversations = append(conversations, ai)
	}

	return conversations, nil
}

// GetActiveAIConversations retrieves all active AI conversations (human = 0)
func (r *aiWhatsappRepository) GetActiveAIConversations() ([]models.AIWhatsapp, error) {
	query := `
		SELECT id_prospect, id_staff, prospect_num, stage, conv_last, 
		       conv_current, human, niche, jam, intro, 
		       catatan_staff, balas, data_image, conv_stage, 
		       bot_balas, keywordiklan, marketer, update_today, 
		       created_at, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE human = 0
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		logrus.WithError(err).Error("Failed to get active AI conversations")
		return nil, fmt.Errorf("failed to get active AI conversations: %w", err)
	}
	defer rows.Close()

	var conversations []models.AIWhatsapp
	for rows.Next() {
		ai := models.AIWhatsapp{}
		var convLastJSON sql.NullString

		err := rows.Scan(
			&ai.IDProspect, &ai.IDStaff, &ai.ProspectNum, &ai.Stage, &convLastJSON,
			&ai.ConvCurrent, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
			&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
			&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
			&ai.CreatedAt, &ai.UpdatedAt,
		)

		if err != nil {
			logrus.WithError(err).Error("Failed to scan AI WhatsApp conversation")
			continue
		}

		// Parse conv_last JSON if it exists
		if convLastJSON.Valid && convLastJSON.String != "" {
			if err := json.Unmarshal([]byte(convLastJSON.String), &ai.ConvLast); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal conv_last JSON")
			}
		}

		conversations = append(conversations, ai)
	}

	return conversations, nil
}

// GetConversationHistory retrieves conversation history for a prospect
func (r *aiWhatsappRepository) GetConversationHistory(prospectNum string, limit int) ([]models.ConversationLog, error) {
	query := `
		SELECT id, prospect_num, id_staff, message, sender, stage, created_at
		FROM conversation_log_nodepath 
		WHERE prospect_num = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := r.db.Query(query, prospectNum, limit)
	if err != nil {
		logrus.WithError(err).Error("Failed to get conversation history")
		return nil, fmt.Errorf("failed to get conversation history: %w", err)
	}
	defer rows.Close()

	var logs []models.ConversationLog
	for rows.Next() {
		log := models.ConversationLog{}
		err := rows.Scan(
			&log.ID, &log.ProspectNum, &log.IDStaff, &log.Message, 
			&log.Sender, &log.Stage, &log.CreatedAt,
		)

		if err != nil {
			logrus.WithError(err).Error("Failed to scan conversation log")
			continue
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetConversationLogsByStage retrieves conversation logs by stage
func (r *aiWhatsappRepository) GetConversationLogsByStage(stage string) ([]models.ConversationLog, error) {
	query := `
		SELECT id, prospect_num, id_staff, message, sender, stage, created_at
		FROM conversation_log_nodepath 
		WHERE stage = ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, stage)
	if err != nil {
		logrus.WithError(err).Error("Failed to get conversation logs by stage")
		return nil, fmt.Errorf("failed to get conversation logs: %w", err)
	}
	defer rows.Close()

	var logs []models.ConversationLog
	for rows.Next() {
		log := models.ConversationLog{}
		err := rows.Scan(
			&log.ID, &log.ProspectNum, &log.IDStaff, &log.Message, 
			&log.Sender, &log.Stage, &log.CreatedAt,
		)

		if err != nil {
			logrus.WithError(err).Error("Failed to scan conversation log")
			continue
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// UpdateAIWhatsapp updates an existing AI WhatsApp conversation
func (r *aiWhatsappRepository) UpdateAIWhatsapp(ai *models.AIWhatsapp) error {
	ai.UpdatedAt = time.Now()

	// Convert conv_last to JSON string if it's not nil
	var convLastJSON []byte
	var err error
	if ai.ConvLast != nil {
		convLastJSON, err = json.Marshal(ai.ConvLast)
		if err != nil {
			return fmt.Errorf("failed to marshal conv_last: %w", err)
		}
	}

	query := `
		UPDATE ai_whatsapp_nodepath SET 
			id_staff = ?, stage = ?, conv_last = ?, conv_current = ?, 
			human = ?, niche = ?, jam = ?, intro = ?, 
			catatan_staff = ?, balas = ?, data_image = ?, conv_stage = ?, 
			bot_balas = ?, keywordiklan = ?, marketer = ?, update_today = ?, 
			updated_at = ?
		WHERE id_prospect = ?
	`

	_, err = r.db.Exec(query,
		ai.IDStaff, ai.Stage, string(convLastJSON), ai.ConvCurrent,
		ai.Human, ai.Niche, ai.Jam, ai.Intro,
		ai.CatatanStaff, ai.Balas, ai.DataImage, ai.ConvStage,
		ai.BotBalas, ai.KeywordIklan, ai.Marketer, ai.UpdateToday,
		ai.UpdatedAt, ai.IDProspect,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to update AI WhatsApp conversation")
		return fmt.Errorf("failed to update AI WhatsApp conversation: %w", err)
	}

	logrus.WithField("id_prospect", ai.IDProspect).Info("AI WhatsApp conversation updated successfully")
	return nil
}

// UpdateConversationStage updates the conversation stage for a prospect
func (r *aiWhatsappRepository) UpdateConversationStage(prospectNum string, stage string) error {
	query := `
		UPDATE ai_whatsapp_nodepath 
		SET stage = ?, updated_at = ?
		WHERE prospect_num = ?
	`

	_, err := r.db.Exec(query, stage, time.Now(), prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to update conversation stage")
		return fmt.Errorf("failed to update conversation stage: %w", err)
	}

	return nil
}

// UpdateHumanTakeover updates the human takeover status
func (r *aiWhatsappRepository) UpdateHumanTakeover(prospectNum string, human int) error {
	query := `
		UPDATE ai_whatsapp_nodepath 
		SET human = ?, updated_at = ?
		WHERE prospect_num = ?
	`

	_, err := r.db.Exec(query, human, time.Now(), prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to update human takeover status")
		return fmt.Errorf("failed to update human takeover status: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"human":       human,
	}).Info("Human takeover status updated")
	return nil
}

// UpdateConvCurrent updates the current conversation text
func (r *aiWhatsappRepository) UpdateConvCurrent(prospectNum string, convCurrent string) error {
	query := `
		UPDATE ai_whatsapp_nodepath 
		SET conv_current = ?, updated_at = ?
		WHERE prospect_num = ?
	`

	_, err := r.db.Exec(query, convCurrent, time.Now(), prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to update conv_current")
		return fmt.Errorf("failed to update conv_current: %w", err)
	}

	return nil
}

// UpdateConvLast updates the last conversation JSON data
func (r *aiWhatsappRepository) UpdateConvLast(prospectNum string, convLast interface{}) error {
	// Convert conv_last to JSON string
	convLastJSON, err := json.Marshal(convLast)
	if err != nil {
		return fmt.Errorf("failed to marshal conv_last: %w", err)
	}

	query := `
		UPDATE ai_whatsapp_nodepath 
		SET conv_last = ?, updated_at = ?
		WHERE prospect_num = ?
	`

	_, err = r.db.Exec(query, string(convLastJSON), time.Now(), prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to update conv_last")
		return fmt.Errorf("failed to update conv_last: %w", err)
	}

	return nil
}

// DeleteAIWhatsapp deletes an AI WhatsApp conversation by ID
func (r *aiWhatsappRepository) DeleteAIWhatsapp(id int) error {
	query := `DELETE FROM ai_whatsapp_nodepath WHERE id_prospect = ?`

	_, err := r.db.Exec(query, id)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete AI WhatsApp conversation")
		return fmt.Errorf("failed to delete AI WhatsApp conversation: %w", err)
	}

	logrus.WithField("id_prospect", id).Info("AI WhatsApp conversation deleted successfully")
	return nil
}

// DeleteConversationLogs deletes all conversation logs for a prospect
func (r *aiWhatsappRepository) DeleteConversationLogs(prospectNum string) error {
	query := `DELETE FROM conversation_log_nodepath WHERE prospect_num = ?`

	_, err := r.db.Exec(query, prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete conversation logs")
		return fmt.Errorf("failed to delete conversation logs: %w", err)
	}

	logrus.WithField("prospect_num", prospectNum).Info("Conversation logs deleted successfully")
	return nil
}

// GetConversationStats returns conversation statistics for a staff member
func (r *aiWhatsappRepository) GetConversationStats(idStaff string) (map[string]int, error) {
	stats := make(map[string]int)

	// Total conversations
	var total int
	query := `SELECT COUNT(*) FROM ai_whatsapp_nodepath WHERE id_staff = ?`
	row := r.db.QueryRow(query, idStaff)
	err := row.Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total conversations: %w", err)
	}
	stats["total"] = total

	// Active AI conversations
	var activeAI int
	query = `SELECT COUNT(*) FROM ai_whatsapp_nodepath WHERE id_staff = ? AND human = 0`
	row = r.db.QueryRow(query, idStaff)
	err = row.Scan(&activeAI)
	if err != nil {
		return nil, fmt.Errorf("failed to get active AI conversations: %w", err)
	}
	stats["active_ai"] = activeAI

	// Human takeover conversations
	var humanTakeover int
	query = `SELECT COUNT(*) FROM ai_whatsapp_nodepath WHERE id_staff = ? AND human = 1`
	row = r.db.QueryRow(query, idStaff)
	err = row.Scan(&humanTakeover)
	if err != nil {
		return nil, fmt.Errorf("failed to get human takeover conversations: %w", err)
	}
	stats["human_takeover"] = humanTakeover

	// Today's conversations
	var today int
	query = `SELECT COUNT(*) FROM ai_whatsapp_nodepath WHERE id_staff = ? AND DATE(created_at) = CURDATE()`
	row = r.db.QueryRow(query, idStaff)
	err = row.Scan(&today)
	if err != nil {
		return nil, fmt.Errorf("failed to get today's conversations: %w", err)
	}
	stats["today"] = today

	return stats, nil
}

// GetActiveConversationCount returns the total number of active AI conversations
func (r *aiWhatsappRepository) GetActiveConversationCount() (int, error) {
	query := `SELECT COUNT(*) FROM ai_whatsapp_nodepath WHERE human = 0`

	var count int
	row := r.db.QueryRow(query)
	err := row.Scan(&count)
	if err != nil {
		logrus.WithError(err).Error("Failed to get active conversation count")
		return 0, fmt.Errorf("failed to get active conversation count: %w", err)
	}

	return count, nil
}

// GetConversationsByDateRange retrieves conversations within a date range
func (r *aiWhatsappRepository) GetConversationsByDateRange(startDate, endDate time.Time) ([]models.AIWhatsapp, error) {
	query := `
		SELECT id_prospect, id_staff, prospect_num, stage, conv_last, 
		       conv_current, human, niche, jam, intro, 
		       catatan_staff, balas, data_image, conv_stage, 
		       bot_balas, keywordiklan, marketer, update_today, 
		       created_at, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE created_at BETWEEN ? AND ?
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(query, startDate, endDate)
	if err != nil {
		logrus.WithError(err).Error("Failed to get conversations by date range")
		return nil, fmt.Errorf("failed to get conversations by date range: %w", err)
	}
	defer rows.Close()

	var conversations []models.AIWhatsapp
	for rows.Next() {
		ai := models.AIWhatsapp{}
		var convLastJSON sql.NullString

		err := rows.Scan(
			&ai.IDProspect, &ai.IDStaff, &ai.ProspectNum, &ai.Stage, &convLastJSON,
			&ai.ConvCurrent, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
			&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
			&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
			&ai.CreatedAt, &ai.UpdatedAt,
		)

		if err != nil {
			logrus.WithError(err).Error("Failed to scan AI WhatsApp conversation")
			continue
		}

		// Parse conv_last JSON if it exists
		if convLastJSON.Valid && convLastJSON.String != "" {
			if err := json.Unmarshal([]byte(convLastJSON.String), &ai.ConvLast); err != nil {
				logrus.WithError(err).Warn("Failed to unmarshal conv_last JSON")
			}
		}

		conversations = append(conversations, ai)
	}

	return conversations, nil
}