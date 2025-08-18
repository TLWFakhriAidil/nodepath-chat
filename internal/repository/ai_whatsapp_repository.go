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
	GetAIWhatsappByDevice(idDevice string) ([]models.AIWhatsapp, error)
	GetAIWhatsappByNiche(niche string) ([]models.AIWhatsapp, error)
	GetActiveAIConversations() ([]models.AIWhatsapp, error)
	GetConversationHistory(prospectNum string, limit int) ([]models.ConversationLog, error)
	GetConversationLogsByStage(stage string) ([]models.ConversationLog, error)
	GetAIWhatsappByProspectAndDevice(prospectNum, idDevice string) (*models.AIWhatsapp, error)

	// Update operations
	UpdateAIWhatsapp(ai *models.AIWhatsapp) error
	UpdateConversationStage(prospectNum string, stage string) error
	UpdateHumanTakeover(prospectNum string, human int) error
	UpdateConvCurrent(prospectNum string, convCurrent string) error
	UpdateConvLast(prospectNum string, convLast interface{}) error
	SaveConversationHistory(prospectNum, idDevice, userMessage, botResponse, stage string) error

	// Delete operations
	DeleteAIWhatsapp(id int) error
	DeleteConversationLogs(prospectNum string) error

	// Analytics operations
	GetConversationStats(idDevice string) (map[string]int, error)
	GetActiveConversationCount() (int, error)
	GetConversationsByDateRange(startDate, endDate time.Time) ([]models.AIWhatsapp, error)
	GetAnalyticsData(startDate, endDate time.Time, idDevice string) (map[string]interface{}, error)
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
// Saves NULL instead of empty string when there's no conversation data
func (r *aiWhatsappRepository) CreateAIWhatsapp(ai *models.AIWhatsapp) error {
	ai.CreatedAt = time.Now()
	ai.UpdatedAt = time.Now()

	// Determine conv_last value - use NULL if empty, otherwise marshal to JSON
	var convLastValue interface{}
	if ai.ConvLast == nil {
		convLastValue = nil // This will be stored as NULL in the database
	} else {
		// Check if it's empty JSON
		convLastJSON, err := json.Marshal(ai.ConvLast)
		if err != nil {
			return fmt.Errorf("failed to marshal conv_last: %w", err)
		}
		// Check if the marshaled result is empty JSON
		jsonStr := string(convLastJSON)
		if jsonStr == "null" || jsonStr == "{}" || jsonStr == "[]" || jsonStr == "\"\"" {
			convLastValue = nil
		} else {
			convLastValue = jsonStr
		}
	}

	query := `
		INSERT INTO ai_whatsapp_nodepath (
			id_prospect, id_device, prospect_num, stage, date_order, conv_last, 
			conv_current, human, niche, jam, intro, 
			catatan_staff, balas, data_image, conv_stage, 
			bot_balas, keywordiklan, marketer, update_today, 
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Handle ConvCurrent as sql.NullString
	var convCurrentValue interface{}
	if ai.ConvCurrent.Valid {
		convCurrentValue = ai.ConvCurrent.String
	} else {
		convCurrentValue = nil
	}

	_, err := r.db.Exec(query,
		ai.IDProspect, ai.IDDevice, ai.ProspectNum, ai.Stage, ai.DateOrder, convLastValue,
		convCurrentValue, ai.Human, ai.Niche, ai.Jam, ai.Intro,
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
			prospect_num, id_device, message, sender, stage, created_at
		) VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.Exec(query,
		log.ProspectNum, log.IDDevice, log.Message, log.Sender, log.Stage, log.CreatedAt,
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
		SELECT id_prospect, id_device, prospect_num, stage, date_order, conv_last, 
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

	var convCurrentSQL sql.NullString
	err := row.Scan(
		&ai.IDProspect, &ai.IDDevice, &ai.ProspectNum, &ai.Stage, &ai.DateOrder, &convLastJSON,
		&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
		&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
		&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
		&ai.CreatedAt, &ai.UpdatedAt,
	)

	ai.ConvCurrent = convCurrentSQL

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		logrus.WithError(err).Error("Failed to get AI WhatsApp conversation by prospect number")
		return nil, fmt.Errorf("failed to get AI WhatsApp conversation: %w", err)
	}

	// Handle conv_last data (both JSON and plain text formats)
	if convLastJSON.Valid && convLastJSON.String != "" {
		// Try to parse as JSON first (for backward compatibility)
		var testJSON interface{}
		if err := json.Unmarshal([]byte(convLastJSON.String), &testJSON); err == nil {
			// It's valid JSON, store as is
			ai.ConvLast = json.RawMessage(convLastJSON.String)
		} else {
			// It's plain text, store as is
			ai.ConvLast = json.RawMessage(convLastJSON.String)
		}
	}

	return ai, nil
}

// GetAIWhatsappByID retrieves AI WhatsApp conversation by ID
func (r *aiWhatsappRepository) GetAIWhatsappByID(id int) (*models.AIWhatsapp, error) {
	query := `
		SELECT id_prospect, id_device, prospect_num, stage, date_order, conv_last, 
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

	var convCurrentSQL sql.NullString
	err := row.Scan(
		&ai.IDProspect, &ai.IDDevice, &ai.ProspectNum, &ai.Stage, &ai.DateOrder, &convLastJSON,
		&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
		&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
		&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
		&ai.CreatedAt, &ai.UpdatedAt,
	)

	ai.ConvCurrent = convCurrentSQL

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		logrus.WithError(err).Error("Failed to get AI WhatsApp conversation by ID")
		return nil, fmt.Errorf("failed to get AI WhatsApp conversation: %w", err)
	}

	// Handle conv_last data (both JSON and plain text formats)
	if convLastJSON.Valid && convLastJSON.String != "" {
		// Try to parse as JSON first (for backward compatibility)
		var testJSON interface{}
		if err := json.Unmarshal([]byte(convLastJSON.String), &testJSON); err == nil {
			// It's valid JSON, store as is
			ai.ConvLast = json.RawMessage(convLastJSON.String)
		} else {
			// It's plain text, store as is
			ai.ConvLast = json.RawMessage(convLastJSON.String)
		}
	}

	return ai, nil
}

// GetAIWhatsappByDevice retrieves all AI WhatsApp conversations for a specific device
func (r *aiWhatsappRepository) GetAIWhatsappByDevice(idDevice string) ([]models.AIWhatsapp, error) {
	query := `
		SELECT id_prospect, id_device, prospect_num, stage, date_order, conv_last, 
		       conv_current, human, niche, jam, intro, 
		       catatan_staff, balas, data_image, conv_stage, 
		       bot_balas, keywordiklan, marketer, update_today, 
		       created_at, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE id_device = ?
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Query(query, idDevice)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp conversations by device")
		return nil, fmt.Errorf("failed to get AI WhatsApp conversations: %w", err)
	}
	defer rows.Close()

	var conversations []models.AIWhatsapp
	for rows.Next() {
		ai := models.AIWhatsapp{}
		var convLastJSON sql.NullString

		var convCurrentSQL sql.NullString
		err := rows.Scan(
			&ai.IDProspect, &ai.IDDevice, &ai.ProspectNum, &ai.Stage, &ai.DateOrder, &convLastJSON,
			&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
			&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
			&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
			&ai.CreatedAt, &ai.UpdatedAt,
		)

		ai.ConvCurrent = convCurrentSQL

		if err != nil {
			logrus.WithError(err).Error("Failed to scan AI WhatsApp conversation")
			continue
		}

		// Handle conv_last data (both JSON and plain text formats)
		if convLastJSON.Valid && convLastJSON.String != "" {
			// Try to parse as JSON first (for backward compatibility)
			var testJSON interface{}
			if err := json.Unmarshal([]byte(convLastJSON.String), &testJSON); err == nil {
				// It's valid JSON, store as is
				ai.ConvLast = json.RawMessage(convLastJSON.String)
			} else {
				// It's plain text, store as is
				ai.ConvLast = json.RawMessage(convLastJSON.String)
			}
		}

		conversations = append(conversations, ai)
	}

	return conversations, nil
}

// GetAnalyticsData retrieves analytics data from ai_whatsapp_nodepath table with date filtering
func (r *aiWhatsappRepository) GetAnalyticsData(startDate, endDate time.Time, idDevice string) (map[string]interface{}, error) {
	// Base query for filtering by date_order
	baseQuery := `
		SELECT 
			COUNT(*) as total_conversations,
			COUNT(CASE WHEN human = 0 THEN 1 END) as ai_active,
			COUNT(CASE WHEN human = 1 THEN 1 END) as human_takeover,
			COUNT(DISTINCT id_device) as unique_devices,
			COUNT(DISTINCT niche) as unique_niches,
			COUNT(CASE WHEN stage IS NOT NULL AND stage != '' THEN 1 END) as conversations_with_stage
		FROM ai_whatsapp_nodepath 
		WHERE date_order BETWEEN ? AND ?
	`

	// Add device filter if specified
	args := []interface{}{startDate, endDate}
	if idDevice != "" && idDevice != "all" {
		baseQuery += " AND id_device = ?"
		args = append(args, idDevice)
	}

	// Execute main analytics query
	var totalConversations, aiActive, humanTakeover, uniqueDevices, uniqueNiches, conversationsWithStage int
	err := r.db.QueryRow(baseQuery, args...).Scan(
		&totalConversations, &aiActive, &humanTakeover, &uniqueDevices, &uniqueNiches, &conversationsWithStage,
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to get analytics data")
		return nil, fmt.Errorf("failed to get analytics data: %w", err)
	}

	// Get daily breakdown
	dailyQuery := `
		SELECT 
			DATE(date_order) as date,
			COUNT(*) as conversations,
			COUNT(CASE WHEN human = 0 THEN 1 END) as ai_conversations,
			COUNT(CASE WHEN human = 1 THEN 1 END) as human_conversations
		FROM ai_whatsapp_nodepath 
		WHERE date_order BETWEEN ? AND ?
	`

	dailyArgs := []interface{}{startDate, endDate}
	if idDevice != "" && idDevice != "all" {
		dailyQuery += " AND id_device = ?"
		dailyArgs = append(dailyArgs, idDevice)
	}
	dailyQuery += " GROUP BY DATE(date_order) ORDER BY DATE(date_order)"

	rows, err := r.db.Query(dailyQuery, dailyArgs...)
	if err != nil {
		logrus.WithError(err).Error("Failed to get daily analytics data")
		return nil, fmt.Errorf("failed to get daily analytics data: %w", err)
	}
	defer rows.Close()

	var dailyData []map[string]interface{}
	for rows.Next() {
		var date string
		var conversations, aiConversations, humanConversations int
		err := rows.Scan(&date, &conversations, &aiConversations, &humanConversations)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan daily analytics data")
			continue
		}

		dailyData = append(dailyData, map[string]interface{}{
			"date":                date,
			"conversations":       conversations,
			"ai_conversations":    aiConversations,
			"human_conversations": humanConversations,
		})
	}

	// Get stage distribution
	stageQuery := `
		SELECT 
			stage,
			COUNT(*) as count
		FROM ai_whatsapp_nodepath 
		WHERE date_order BETWEEN ? AND ? AND stage IS NOT NULL AND stage != ''
	`

	stageArgs := []interface{}{startDate, endDate}
	if idDevice != "" && idDevice != "all" {
		stageQuery += " AND id_device = ?"
		stageArgs = append(stageArgs, idDevice)
	}
	stageQuery += " GROUP BY stage ORDER BY count DESC"

	stageRows, err := r.db.Query(stageQuery, stageArgs...)
	if err != nil {
		logrus.WithError(err).Error("Failed to get stage distribution data")
		return nil, fmt.Errorf("failed to get stage distribution data: %w", err)
	}
	defer stageRows.Close()

	var stageDistribution []map[string]interface{}
	for stageRows.Next() {
		var stage string
		var count int
		err := stageRows.Scan(&stage, &count)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan stage distribution data")
			continue
		}

		stageDistribution = append(stageDistribution, map[string]interface{}{
			"stage": stage,
			"count": count,
		})
	}

	// Return comprehensive analytics data
	return map[string]interface{}{
		"summary": map[string]interface{}{
			"total_conversations":       totalConversations,
			"ai_active":                 aiActive,
			"human_takeover":            humanTakeover,
			"unique_devices":            uniqueDevices,
			"unique_niches":             uniqueNiches,
			"conversations_with_stage":  conversationsWithStage,
			"ai_active_percentage":      float64(aiActive) / float64(totalConversations) * 100,
			"human_takeover_percentage": float64(humanTakeover) / float64(totalConversations) * 100,
		},
		"daily_data":         dailyData,
		"stage_distribution": stageDistribution,
		"date_range": map[string]interface{}{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
	}, nil
}

// GetAIWhatsappByNiche retrieves all AI WhatsApp conversations for a specific niche
func (r *aiWhatsappRepository) GetAIWhatsappByNiche(niche string) ([]models.AIWhatsapp, error) {
	query := `
		SELECT id_prospect, id_device, prospect_num, stage, date_order, conv_last, 
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

		var convCurrentSQL sql.NullString
		err := rows.Scan(
			&ai.IDProspect, &ai.IDDevice, &ai.ProspectNum, &ai.Stage, &ai.DateOrder, &convLastJSON,
			&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
			&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
			&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
			&ai.CreatedAt, &ai.UpdatedAt,
		)

		ai.ConvCurrent = convCurrentSQL

		if err != nil {
			logrus.WithError(err).Error("Failed to scan AI WhatsApp conversation")
			continue
		}

		// Handle conv_last data (both JSON and plain text formats)
		if convLastJSON.Valid && convLastJSON.String != "" {
			// Try to parse as JSON first (for backward compatibility)
			var testJSON interface{}
			if err := json.Unmarshal([]byte(convLastJSON.String), &testJSON); err == nil {
				// It's valid JSON, store as is
				ai.ConvLast = json.RawMessage(convLastJSON.String)
			} else {
				// It's plain text, store as is
				ai.ConvLast = json.RawMessage(convLastJSON.String)
			}
		}

		conversations = append(conversations, ai)
	}

	return conversations, nil
}

// GetActiveAIConversations retrieves all active AI conversations (human = 0)
func (r *aiWhatsappRepository) GetActiveAIConversations() ([]models.AIWhatsapp, error) {
	query := `
		SELECT id_prospect, id_device, prospect_num, stage, date_order, conv_last, 
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
		var convCurrentSQL sql.NullString

		err := rows.Scan(
			&ai.IDProspect, &ai.IDDevice, &ai.ProspectNum, &ai.Stage, &ai.DateOrder, &convLastJSON,
			&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
			&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
			&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
			&ai.CreatedAt, &ai.UpdatedAt,
		)

		ai.ConvCurrent = convCurrentSQL

		if err != nil {
			logrus.WithError(err).Error("Failed to scan AI WhatsApp conversation")
			continue
		}

		// Handle conv_last data (both JSON and plain text formats)
		if convLastJSON.Valid && convLastJSON.String != "" {
			// Try to parse as JSON first (for backward compatibility)
			var testJSON interface{}
			if err := json.Unmarshal([]byte(convLastJSON.String), &testJSON); err == nil {
				// It's valid JSON, store as is
				ai.ConvLast = json.RawMessage(convLastJSON.String)
			} else {
				// It's plain text, store as is
				ai.ConvLast = json.RawMessage(convLastJSON.String)
			}
		}

		conversations = append(conversations, ai)
	}

	return conversations, nil
}

// GetConversationHistory retrieves conversation history for a prospect
func (r *aiWhatsappRepository) GetConversationHistory(prospectNum string, limit int) ([]models.ConversationLog, error) {
	query := `
		SELECT id, prospect_num, id_device, message, sender, stage, created_at
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
			&log.ID, &log.ProspectNum, &log.IDDevice, &log.Message, 
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
		SELECT id, prospect_num, id_device, message, sender, stage, created_at
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
			&log.ID, &log.ProspectNum, &log.IDDevice, &log.Message, 
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
// Saves NULL instead of empty string when there's no conversation data
func (r *aiWhatsappRepository) UpdateAIWhatsapp(ai *models.AIWhatsapp) error {
	ai.UpdatedAt = time.Now()

	// Determine conv_last value - use NULL if empty, otherwise marshal to JSON
	var convLastValue interface{}
	if ai.ConvLast == nil {
		convLastValue = nil // This will be stored as NULL in the database
	} else {
		// Check if it's empty JSON
		convLastJSON, err := json.Marshal(ai.ConvLast)
		if err != nil {
			return fmt.Errorf("failed to marshal conv_last: %w", err)
		}
		// Check if the marshaled result is empty JSON
		jsonStr := string(convLastJSON)
		if jsonStr == "null" || jsonStr == "{}" || jsonStr == "[]" || jsonStr == "\"\"" {
			convLastValue = nil
		} else {
			convLastValue = jsonStr
		}
	}

	query := `
		UPDATE ai_whatsapp_nodepath SET 
			id_device = ?, stage = ?, date_order = ?, conv_last = ?, conv_current = ?, 
			human = ?, niche = ?, jam = ?, intro = ?, 
			catatan_staff = ?, balas = ?, data_image = ?, conv_stage = ?, 
			bot_balas = ?, keywordiklan = ?, marketer = ?, update_today = ?, 
			updated_at = ?
		WHERE id_prospect = ?
	`

	// Handle ConvCurrent as sql.NullString
	var convCurrentValue interface{}
	if ai.ConvCurrent.Valid {
		convCurrentValue = ai.ConvCurrent.String
	} else {
		convCurrentValue = nil
	}

	_, err := r.db.Exec(query,
		ai.IDDevice, ai.Stage, ai.DateOrder, convLastValue, convCurrentValue,
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

	// Handle empty string as NULL
	var convCurrentValue interface{}
	if convCurrent != "" {
		convCurrentValue = convCurrent
	} else {
		convCurrentValue = nil
	}

	_, err := r.db.Exec(query, convCurrentValue, time.Now(), prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to update conv_current")
		return fmt.Errorf("failed to update conv_current: %w", err)
	}

	return nil
}

// UpdateConvLast updates the last conversation JSON data
// Saves NULL instead of empty string when there's no conversation data
func (r *aiWhatsappRepository) UpdateConvLast(prospectNum string, convLast interface{}) error {
	// Determine conv_last value - use NULL if empty, otherwise marshal to JSON
	var convLastValue interface{}
	
	// Check if convLast is empty or nil
	if convLast == nil {
		convLastValue = nil // This will be stored as NULL in the database
	} else {
		// Check if it's an empty string, empty slice, or empty map
		switch v := convLast.(type) {
		case string:
			if v == "" {
				convLastValue = nil
			} else {
				convLastValue = v
			}
		case []interface{}:
			if len(v) == 0 {
				convLastValue = nil
			} else {
				// Convert to JSON string
				convLastJSON, err := json.Marshal(convLast)
				if err != nil {
					return fmt.Errorf("failed to marshal conv_last: %w", err)
				}
				convLastValue = string(convLastJSON)
			}
		case map[string]interface{}:
			if len(v) == 0 {
				convLastValue = nil
			} else {
				// Convert to JSON string
				convLastJSON, err := json.Marshal(convLast)
				if err != nil {
					return fmt.Errorf("failed to marshal conv_last: %w", err)
				}
				convLastValue = string(convLastJSON)
			}
		default:
			// Convert to JSON string for other types
			convLastJSON, err := json.Marshal(convLast)
			if err != nil {
				return fmt.Errorf("failed to marshal conv_last: %w", err)
			}
			// Check if the marshaled result is empty JSON
			jsonStr := string(convLastJSON)
			if jsonStr == "null" || jsonStr == "{}" || jsonStr == "[]" || jsonStr == "\"\"" {
				convLastValue = nil
			} else {
				convLastValue = jsonStr
			}
		}
	}

	query := `
		UPDATE ai_whatsapp_nodepath 
		SET conv_last = ?, updated_at = ?
		WHERE prospect_num = ?
	`

	_, err := r.db.Exec(query, convLastValue, time.Now(), prospectNum)
	if err != nil {
		logrus.WithError(err).Error("Failed to update conv_last")
		return fmt.Errorf("failed to update conv_last: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"conv_last_is_null": convLastValue == nil,
	}).Info("Conv_last updated successfully")

	return nil
}

// GetAIWhatsappByProspectAndDevice retrieves AI WhatsApp conversation by prospect number and device ID
func (r *aiWhatsappRepository) GetAIWhatsappByProspectAndDevice(prospectNum, idDevice string) (*models.AIWhatsapp, error) {
	query := `
		SELECT id_prospect, id_device, prospect_num, stage, date_order, conv_last, 
		       conv_current, human, niche, jam, intro, 
		       catatan_staff, balas, data_image, conv_stage, 
		       bot_balas, keywordiklan, marketer, update_today, 
		       created_at, updated_at
		FROM ai_whatsapp_nodepath 
		WHERE prospect_num = ? AND id_device = ?
	`

	row := r.db.QueryRow(query, prospectNum, idDevice)

	ai := &models.AIWhatsapp{}
	var convLastJSON sql.NullString
	var convCurrentSQL sql.NullString

	err := row.Scan(
		&ai.IDProspect, &ai.IDDevice, &ai.ProspectNum, &ai.Stage, &ai.DateOrder, &convLastJSON,
		&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
		&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
		&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
		&ai.CreatedAt, &ai.UpdatedAt,
	)

	ai.ConvCurrent = convCurrentSQL

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		logrus.WithError(err).Error("Failed to get AI WhatsApp conversation by prospect and device")
		return nil, fmt.Errorf("failed to get AI WhatsApp conversation: %w", err)
	}

	// Handle conv_last data (both JSON and plain text formats)
	if convLastJSON.Valid && convLastJSON.String != "" {
		// Try to parse as JSON first (for backward compatibility)
		var testJSON interface{}
		if err := json.Unmarshal([]byte(convLastJSON.String), &testJSON); err == nil {
			// It's valid JSON, store as is
			ai.ConvLast = json.RawMessage(convLastJSON.String)
		} else {
			// It's plain text, store as is
			ai.ConvLast = json.RawMessage(convLastJSON.String)
		}
	}

	return ai, nil
}

// SaveConversationHistory saves conversation history to conv_last field as plain text
// If record exists, it updates the conv_last field; otherwise, it creates a new record
// Saves NULL instead of empty string when there's no conversation data
func (r *aiWhatsappRepository) SaveConversationHistory(prospectNum, idDevice, userMessage, botResponse, stage string) error {
	// Check if record exists
	existingRecord, err := r.GetAIWhatsappByProspectAndDevice(prospectNum, idDevice)
	if err != nil {
		return fmt.Errorf("failed to check existing record: %w", err)
	}

	// Get existing conversation history as plain text
	var convHistory string
	if existingRecord != nil && existingRecord.ConvLast != nil {
		// Check if existing data is JSON format (for backward compatibility)
		var existingHistory interface{}
		if err := json.Unmarshal(existingRecord.ConvLast, &existingHistory); err == nil {
			// Convert JSON format to plain text format
			if historySlice, ok := existingHistory.([]interface{}); ok {
				for _, item := range historySlice {
					if itemMap, ok := item.(map[string]interface{}); ok {
						for k, v := range itemMap {
							if str, ok := v.(string); ok {
								if k == "user" {
									if convHistory != "" {
										convHistory += "\n"
									}
									convHistory += "USER:" + str
								} else if k == "bot" {
									if convHistory != "" {
										convHistory += "\n"
									}
									convHistory += "BOT:" + str
								}
							}
						}
					}
				}
			}
		} else {
			// Already in plain text format
			convHistory = string(existingRecord.ConvLast)
		}
	}

	// Add new conversation entries in plain text format
	if userMessage != "" {
		if convHistory != "" {
			convHistory += "\n"
		}
		convHistory += "USER:" + userMessage
	}
	if botResponse != "" {
		if convHistory != "" {
			convHistory += "\n"
		}
		convHistory += "BOT:" + botResponse
	}

	// Determine conv_last value - use NULL if empty, otherwise use the conversation history
	var convLastValue interface{}
	if convHistory == "" {
		convLastValue = nil // This will be stored as NULL in the database
	} else {
		convLastValue = convHistory
	}

	if existingRecord != nil {
		// Update existing record
		query := `
			UPDATE ai_whatsapp_nodepath 
			SET conv_last = ?, stage = ?, updated_at = ?
			WHERE prospect_num = ? AND id_device = ?
		`
		_, err = r.db.Exec(query, convLastValue, stage, time.Now(), prospectNum, idDevice)
		if err != nil {
			logrus.WithError(err).Error("Failed to update conversation history")
			return fmt.Errorf("failed to update conversation history: %w", err)
		}
		logrus.WithFields(logrus.Fields{
			"prospect_num": prospectNum,
			"id_device": idDevice,
		}).Info("Conversation history updated successfully")
	} else {
		// Create new record
		now := time.Now()
		query := `
			INSERT INTO ai_whatsapp_nodepath (
				id_device, prospect_num, stage, conv_last, human, 
				created_at, updated_at
			) VALUES (?, ?, ?, ?, ?, ?, ?)
		`
		_, err = r.db.Exec(query, idDevice, prospectNum, stage, convLastValue, 0, now, now)
		if err != nil {
			logrus.WithError(err).Error("Failed to create new conversation record")
			return fmt.Errorf("failed to create new conversation record: %w", err)
		}
		logrus.WithFields(logrus.Fields{
			"prospect_num": prospectNum,
			"id_device": idDevice,
		}).Info("New conversation record created successfully")
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

// GetConversationStats returns conversation statistics for a device
func (r *aiWhatsappRepository) GetConversationStats(idDevice string) (map[string]int, error) {
	stats := make(map[string]int)

	// Total conversations
	var total int
	query := `SELECT COUNT(*) FROM ai_whatsapp_nodepath WHERE id_device = ?`
	row := r.db.QueryRow(query, idDevice)
	err := row.Scan(&total)
	if err != nil {
		return nil, fmt.Errorf("failed to get total conversations: %w", err)
	}
	stats["total"] = total

	// Active AI conversations
	var activeAI int
	query = `SELECT COUNT(*) FROM ai_whatsapp_nodepath WHERE id_device = ? AND human = 0`
	row = r.db.QueryRow(query, idDevice)
	err = row.Scan(&activeAI)
	if err != nil {
		return nil, fmt.Errorf("failed to get active AI conversations: %w", err)
	}
	stats["active_ai"] = activeAI

	// Human takeover conversations
	var humanTakeover int
	query = `SELECT COUNT(*) FROM ai_whatsapp_nodepath WHERE id_device = ? AND human = 1`
	row = r.db.QueryRow(query, idDevice)
	err = row.Scan(&humanTakeover)
	if err != nil {
		return nil, fmt.Errorf("failed to get human takeover conversations: %w", err)
	}
	stats["human_takeover"] = humanTakeover

	// Today's conversations
	var today int
	query = `SELECT COUNT(*) FROM ai_whatsapp_nodepath WHERE id_device = ? AND DATE(created_at) = CURDATE()`
	row = r.db.QueryRow(query, idDevice)
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
		SELECT id_prospect, id_device, prospect_num, stage, date_order, conv_last, 
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
		var convCurrentSQL sql.NullString

		err := rows.Scan(
			&ai.IDProspect, &ai.IDDevice, &ai.ProspectNum, &ai.Stage, &ai.DateOrder, &convLastJSON,
			&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Jam, &ai.Intro,
			&ai.CatatanStaff, &ai.Balas, &ai.DataImage, &ai.ConvStage,
			&ai.BotBalas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
			&ai.CreatedAt, &ai.UpdatedAt,
		)

		ai.ConvCurrent = convCurrentSQL

		if err != nil {
			logrus.WithError(err).Error("Failed to scan AI WhatsApp conversation")
			continue
		}

		// Handle conv_last data (both JSON and plain text formats)
		if convLastJSON.Valid && convLastJSON.String != "" {
			// Try to parse as JSON first (for backward compatibility)
			var testJSON interface{}
			if err := json.Unmarshal([]byte(convLastJSON.String), &testJSON); err == nil {
				// It's valid JSON, store as is
				ai.ConvLast = json.RawMessage(convLastJSON.String)
			} else {
				// It's plain text, store as is
				ai.ConvLast = json.RawMessage(convLastJSON.String)
			}
		}

		conversations = append(conversations, ai)
	}

	return conversations, nil
}