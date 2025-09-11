package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/utils"

	"github.com/sirupsen/logrus"
)

// AIWhatsappRepository interface defines methods for AI WhatsApp conversation management
type AIWhatsappRepository interface {
	// Create operations
	CreateAIWhatsapp(ai *models.AIWhatsapp) error
	// CreateConversationLog removed - no longer using conversation_log_nodepath table

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
	UpdateFlowTrackingFields(prospectNum, idDevice string, flowID, currentNodeID, lastNodeID string, waitingForReply int, executionStatus, executionID string) error
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
	GetAnalyticsData(startDate, endDate time.Time, idDevice string, userID int) (map[string]interface{}, error)

	// Data table operations
	GetAllAIWhatsappData(limit, offset int, deviceFilter, stageFilter, search string, userID int) ([]models.AIWhatsapp, int, error)
	
	// Database access for transactions
	GetDB() *sql.DB
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

// GetDB returns the database connection for transaction handling
func (r *aiWhatsappRepository) GetDB() *sql.DB {
	return r.db
}

// CreateAIWhatsapp creates a new AI WhatsApp conversation record
// Saves NULL instead of empty string when there's no conversation data
// Includes all flow tracking fields to ensure data integrity
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
			id_device, prospect_num, stage, date_order, conv_last, 
			conv_current, human, niche, intro, 
			balas, keywordiklan, marketer, update_today, 
			current_node_id, waiting_for_reply, flow_id, last_node_id,
			flow_reference, execution_id, execution_status,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	// Handle ConvCurrent as sql.NullString
	var convCurrentValue interface{}
	if ai.ConvCurrent.Valid {
		convCurrentValue = ai.ConvCurrent.String
	} else {
		convCurrentValue = nil
	}

	// Handle flow tracking fields as sql.NullString and sql.NullInt32
	var currentNodeIDValue, flowIDValue, lastNodeIDValue interface{}
	var waitingForReplyValue interface{}
	var flowReferenceValue, executionIDValue, executionStatusValue interface{}
	
	if ai.CurrentNodeID.Valid {
		currentNodeIDValue = ai.CurrentNodeID.String
	} else {
		currentNodeIDValue = nil
	}
	
	if ai.FlowID.Valid {
		flowIDValue = ai.FlowID.String
	} else {
		flowIDValue = nil
	}
	
	if ai.LastNodeID.Valid {
		lastNodeIDValue = ai.LastNodeID.String
	} else {
		lastNodeIDValue = nil
	}
	
	if ai.WaitingForReply.Valid {
		waitingForReplyValue = ai.WaitingForReply.Int32
	} else {
		waitingForReplyValue = nil
	}
	
	if ai.FlowReference.Valid {
		flowReferenceValue = ai.FlowReference.String
	} else {
		flowReferenceValue = nil
	}
	
	if ai.ExecutionID.Valid {
		executionIDValue = ai.ExecutionID.String
	} else {
		executionIDValue = nil
	}
	
	if ai.ExecutionStatus.Valid {
		executionStatusValue = ai.ExecutionStatus.String
	} else {
		executionStatusValue = nil
	}

	_, err := r.db.Exec(query,
		ai.IDDevice, ai.ProspectNum, ai.Stage, ai.DateOrder, convLastValue,
		convCurrentValue, ai.Human, ai.Niche, ai.Intro,
		ai.Balas, ai.KeywordIklan, ai.Marketer, ai.UpdateToday,
		currentNodeIDValue, waitingForReplyValue, flowIDValue, lastNodeIDValue,
		flowReferenceValue, executionIDValue, executionStatusValue,
		ai.CreatedAt, ai.UpdatedAt,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to create AI WhatsApp conversation")
		return fmt.Errorf("failed to create AI WhatsApp conversation: %w", err)
	}

	logrus.WithField("prospect_num", ai.ProspectNum).Info("AI WhatsApp conversation created successfully")
	return nil
}

// CreateConversationLog - REMOVED: No longer using conversation_log_nodepath table
// All conversation history is now stored in ai_whatsapp_nodepath.conv_last field
// func (r *aiWhatsappRepository) CreateConversationLog(log *models.ConversationLog) error {
// 	// REMOVED - no longer needed
// 	return nil
// }

// GetAIWhatsappByProspectNum retrieves AI WhatsApp conversation by prospect number
func (r *aiWhatsappRepository) GetAIWhatsappByProspectNum(prospectNum string) (*models.AIWhatsapp, error) {
	// Check if database connection is available
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	query := `
		SELECT id_prospect, id_device, prospect_num, stage, date_order, conv_last, 
		       conv_current, human, niche, intro, 
		       balas, keywordiklan, marketer, update_today, 
		       created_at, updated_at,
		       current_node_id, waiting_for_reply, flow_id, last_node_id, 
		       flow_reference, execution_status, execution_id
		FROM ai_whatsapp_nodepath 
		WHERE prospect_num = ?
	`

	row := r.db.QueryRow(query, prospectNum)

	ai := &models.AIWhatsapp{}
	var convLastJSON sql.NullString

	var convCurrentSQL sql.NullString
	err := row.Scan(
		&ai.IDProspect, &ai.IDDevice, &ai.ProspectNum, &ai.Stage, &ai.DateOrder, &convLastJSON,
		&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Intro,
		&ai.Balas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
		&ai.CreatedAt, &ai.UpdatedAt,
		&ai.CurrentNodeID, &ai.WaitingForReply, &ai.FlowID, &ai.LastNodeID,
		&ai.FlowReference, &ai.ExecutionStatus, &ai.ExecutionID,
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
			// It's plain text, convert to proper JSON string
			jsonBytes, err := json.Marshal(convLastJSON.String)
			if err != nil {
				logrus.WithError(err).WithField("conv_last", convLastJSON.String).Warn("Failed to marshal conv_last as JSON, setting to null")
				ai.ConvLast = json.RawMessage("null")
			} else {
				ai.ConvLast = json.RawMessage(jsonBytes)
			}
		}
	} else {
		// Set to null JSON if empty or invalid
		ai.ConvLast = json.RawMessage("null")
	}

	return ai, nil
}

// GetAIWhatsappByID retrieves AI WhatsApp conversation by ID
func (r *aiWhatsappRepository) GetAIWhatsappByID(id int) (*models.AIWhatsapp, error) {
	// Check if database connection is available
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

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
		&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Intro,
		&ai.Balas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
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
			// It's plain text, convert to proper JSON string
			jsonBytes, err := json.Marshal(convLastJSON.String)
			if err != nil {
				logrus.WithError(err).WithField("conv_last", convLastJSON.String).Warn("Failed to marshal conv_last as JSON, setting to null")
				ai.ConvLast = json.RawMessage("null")
			} else {
				ai.ConvLast = json.RawMessage(jsonBytes)
			}
		}
	} else {
		// Set to null JSON if empty or invalid
		ai.ConvLast = json.RawMessage("null")
	}

	return ai, nil
}

// GetAIWhatsappByDevice retrieves all AI WhatsApp conversations for a specific device
func (r *aiWhatsappRepository) GetAIWhatsappByDevice(idDevice string) ([]models.AIWhatsapp, error) {
	// Check if database connection is available
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

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
			&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Intro,
			&ai.Balas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
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
				// It's plain text, convert to proper JSON string
				jsonBytes, err := json.Marshal(convLastJSON.String)
				if err != nil {
					logrus.WithError(err).WithField("conv_last", convLastJSON.String).Warn("Failed to marshal conv_last as JSON, setting to null")
					ai.ConvLast = json.RawMessage("null")
				} else {
					ai.ConvLast = json.RawMessage(jsonBytes)
				}
			}
		} else {
			// Set to null JSON if empty or invalid
			ai.ConvLast = json.RawMessage("null")
		}

		conversations = append(conversations, ai)
	}

	return conversations, nil
}

// GetAllAIWhatsappData retrieves all AI WhatsApp conversation records with pagination and filtering
func (r *aiWhatsappRepository) GetAllAIWhatsappData(limit, offset int, deviceFilter, stageFilter, search string, userID int) ([]models.AIWhatsapp, int, error) {
	// Build base query with JOIN to filter by user
	baseQuery := `
		SELECT a.id_prospect, a.id_device, a.prospect_num, a.stage, a.date_order, a.conv_last, 
		       a.conv_current, a.human, a.niche, a.jam, a.intro, 
		       a.catatan_staff, a.balas, a.data_image, a.conv_stage, 
		       a.bot_balas, a.keywordiklan, a.marketer, a.update_today, 
		       a.created_at, a.updated_at
		FROM ai_whatsapp_nodepath a
		JOIN device_setting_nodepath d ON a.id_device = d.id_device
		WHERE d.user_id = ?
	`

	countQuery := `SELECT COUNT(*) FROM ai_whatsapp_nodepath a JOIN device_setting_nodepath d ON a.id_device = d.id_device WHERE d.user_id = ?`

	// Build additional WHERE conditions
	var conditions []string
	var args []interface{}
	var countArgs []interface{}

	// Start with userID for both queries
	args = append(args, userID)
	countArgs = append(countArgs, userID)

	// Add device filter
	if deviceFilter != "" {
		conditions = append(conditions, "a.id_device = ?")
		args = append(args, deviceFilter)
		countArgs = append(countArgs, deviceFilter)
	}

	// Add stage filter
	if stageFilter != "" {
		conditions = append(conditions, "a.stage = ?")
		args = append(args, stageFilter)
		countArgs = append(countArgs, stageFilter)
	}

	// Add search filter (searches in prospect_num, niche, stage, and marketer)
	if search != "" {
		conditions = append(conditions, "(a.prospect_num LIKE ? OR a.niche LIKE ? OR a.stage LIKE ? OR a.marketer LIKE ?)")
		searchTerm := "%" + search + "%"
		args = append(args, searchTerm, searchTerm, searchTerm, searchTerm)
		countArgs = append(countArgs, searchTerm, searchTerm, searchTerm, searchTerm)
	}

	// Add additional WHERE conditions if they exist
	if len(conditions) > 0 {
		whereClause := " AND " + fmt.Sprintf("%s", conditions[0])
		for i := 1; i < len(conditions); i++ {
			whereClause += " AND " + conditions[i]
		}
		baseQuery += whereClause
		countQuery += whereClause
	}

	// Add ORDER BY and LIMIT for main query
	baseQuery += " ORDER BY updated_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Get total count first
	var total int
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Failed to get total count for AI WhatsApp data")
		return nil, 0, fmt.Errorf("failed to get total count: %w", err)
	}

	// Execute main query
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		logrus.WithError(err).Error("Failed to get AI WhatsApp data")
		return nil, 0, fmt.Errorf("failed to get AI WhatsApp data: %w", err)
	}
	defer rows.Close()

	var conversations []models.AIWhatsapp
	for rows.Next() {
		ai := models.AIWhatsapp{}
		var convLastJSON sql.NullString
		var convCurrentSQL sql.NullString

		err := rows.Scan(
			&ai.IDProspect, &ai.IDDevice, &ai.ProspectNum, &ai.Stage, &ai.DateOrder, &convLastJSON,
			&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Intro,
			&ai.Balas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
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
				// It's plain text, convert to proper JSON string
				jsonBytes, err := json.Marshal(convLastJSON.String)
				if err != nil {
					logrus.WithError(err).WithField("conv_last", convLastJSON.String).Warn("Failed to marshal conv_last as JSON, setting to null")
					ai.ConvLast = json.RawMessage("null")
				} else {
					ai.ConvLast = json.RawMessage(jsonBytes)
				}
			}
		} else {
			// Set to null JSON if empty or invalid
			ai.ConvLast = json.RawMessage("null")
		}

		conversations = append(conversations, ai)
	}

	if err = rows.Err(); err != nil {
		logrus.WithError(err).Error("Error iterating over AI WhatsApp data rows")
		return nil, 0, fmt.Errorf("error iterating over rows: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"total_records": total,
		"returned_records": len(conversations),
		"limit": limit,
		"offset": offset,
	}).Info("Successfully retrieved AI WhatsApp data")

	return conversations, total, nil
}

// GetAnalyticsData retrieves analytics data from ai_whatsapp_nodepath table with date filtering
func (r *aiWhatsappRepository) GetAnalyticsData(startDate, endDate time.Time, idDevice string, userID int) (map[string]interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"startDate": startDate.Format("2006-01-02"),
		"endDate": endDate.Format("2006-01-02"),
		"idDevice": idDevice,
		"userID": userID,
	}).Info("GetAnalyticsData called")
	
	// First, let's get the user's devices
	var userDevices []string
	deviceQuery := `SELECT id_device FROM device_setting_nodepath WHERE user_id = ?`
	rows, err := r.db.Query(deviceQuery, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var device string
			if err := rows.Scan(&device); err == nil {
				userDevices = append(userDevices, device)
			}
		}
		logrus.WithFields(logrus.Fields{
			"userID": userID,
			"devices": userDevices,
		}).Info("User devices found")
	}
	
	// If no devices found for user, return empty data
	if len(userDevices) == 0 {
		logrus.WithField("userID", userID).Warn("No devices found for user")
		return map[string]interface{}{
			"summary": map[string]interface{}{
				"total_conversations":       0,
				"ai_active":                 0,
				"human_takeover":            0,
				"unique_devices":            0,
				"unique_niches":             0,
				"conversations_with_stage":  0,
				"ai_active_percentage":      0.0,
				"human_takeover_percentage": 0.0,
			},
			"daily_data":         []map[string]interface{}{},
			"stage_distribution": []map[string]interface{}{},
			"date_range": map[string]interface{}{
				"start_date": startDate.Format("2006-01-02"),
				"end_date":   endDate.Format("2006-01-02"),
			},
		}, nil
	}
	
	// Build query with IN clause for user's devices
	placeholders := make([]string, len(userDevices))
	args := []interface{}{}
	for i, device := range userDevices {
		placeholders[i] = "?"
		args = append(args, device)
	}
	args = append(args, startDate, endDate)
	
	// Base query using IN clause instead of JOIN
	baseQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_conversations,
			COUNT(CASE WHEN human = 0 THEN 1 END) as ai_active,
			COUNT(CASE WHEN human = 1 THEN 1 END) as human_takeover,
			COUNT(DISTINCT id_device) as unique_devices,
			COUNT(DISTINCT niche) as unique_niches,
			COUNT(CASE WHEN stage IS NOT NULL AND stage != '' THEN 1 END) as conversations_with_stage
		FROM ai_whatsapp_nodepath
		WHERE id_device IN (%s) AND date_order BETWEEN ? AND ?
	`, strings.Join(placeholders, ","))

	// Add specific device filter if specified
	if idDevice != "" && idDevice != "all" {
		baseQuery += " AND id_device = ?"
		args = append(args, idDevice)
		logrus.WithField("deviceFilter", idDevice).Info("Adding specific device filter to analytics query")
	}

	// Execute main analytics query
	logrus.WithField("query", baseQuery).WithField("args", args).Info("Executing analytics query")
	
	var totalConversations, aiActive, humanTakeover, uniqueDevices, uniqueNiches, conversationsWithStage int
	err = r.db.QueryRow(baseQuery, args...).Scan(
		&totalConversations, &aiActive, &humanTakeover, &uniqueDevices, &uniqueNiches, &conversationsWithStage,
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to get analytics data")
		// Return empty data instead of error
		return map[string]interface{}{
			"summary": map[string]interface{}{
				"total_conversations":       0,
				"ai_active":                 0,
				"human_takeover":            0,
				"unique_devices":            0,
				"unique_niches":             0,
				"conversations_with_stage":  0,
				"ai_active_percentage":      0.0,
				"human_takeover_percentage": 0.0,
			},
			"daily_data":         []map[string]interface{}{},
			"stage_distribution": []map[string]interface{}{},
			"date_range": map[string]interface{}{
				"start_date": startDate.Format("2006-01-02"),
				"end_date":   endDate.Format("2006-01-02"),
			},
		}, nil
	}
	
	logrus.WithFields(logrus.Fields{
		"totalConversations": totalConversations,
		"aiActive": aiActive,
		"humanTakeover": humanTakeover,
		"uniqueDevices": uniqueDevices,
		"uniqueNiches": uniqueNiches,
		"conversationsWithStage": conversationsWithStage,
	}).Info("Analytics query results")

	// Get daily breakdown
	dailyQuery := fmt.Sprintf(`
		SELECT 
			DATE(date_order) as date,
			COUNT(*) as conversations,
			COUNT(CASE WHEN human = 0 THEN 1 END) as ai_conversations,
			COUNT(CASE WHEN human = 1 THEN 1 END) as human_conversations
		FROM ai_whatsapp_nodepath
		WHERE id_device IN (%s) AND date_order BETWEEN ? AND ?
	`, strings.Join(placeholders, ","))

	// Reset args for daily query
	dailyArgs := []interface{}{}
	for _, device := range userDevices {
		dailyArgs = append(dailyArgs, device)
	}
	dailyArgs = append(dailyArgs, startDate, endDate)
	
	if idDevice != "" && idDevice != "all" {
		dailyQuery += " AND id_device = ?"
		dailyArgs = append(dailyArgs, idDevice)
	}
	dailyQuery += " GROUP BY DATE(date_order) ORDER BY DATE(date_order)"

	dailyRows, err := r.db.Query(dailyQuery, dailyArgs...)
	var dailyData []map[string]interface{}
	if err != nil {
		logrus.WithError(err).Error("Failed to get daily analytics data")
		// Don't return error, just use empty daily data
		dailyData = []map[string]interface{}{}
	} else {
		defer dailyRows.Close()
		
		for dailyRows.Next() {
			var date string
			var conversations, aiConversations, humanConversations int
			err := dailyRows.Scan(&date, &conversations, &aiConversations, &humanConversations)
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
	}

	// Get stage distribution
	stageQuery := fmt.Sprintf(`
		SELECT 
			stage,
			COUNT(*) as count
		FROM ai_whatsapp_nodepath
		WHERE id_device IN (%s) AND date_order BETWEEN ? AND ? AND stage IS NOT NULL AND stage != ''
	`, strings.Join(placeholders, ","))

	// Reset args for stage query
	stageArgs := []interface{}{}
	for _, device := range userDevices {
		stageArgs = append(stageArgs, device)
	}
	stageArgs = append(stageArgs, startDate, endDate)
	
	if idDevice != "" && idDevice != "all" {
		stageQuery += " AND id_device = ?"
		stageArgs = append(stageArgs, idDevice)
	}
	stageQuery += " GROUP BY stage ORDER BY count DESC"

	stageRows, err := r.db.Query(stageQuery, stageArgs...)
	var stageDistribution []map[string]interface{}
	if err != nil {
		logrus.WithError(err).Error("Failed to get stage distribution data")
		// Don't return error, just use empty stage data
		stageDistribution = []map[string]interface{}{}
	} else {
		defer stageRows.Close()
		
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
	}

	// Calculate percentages safely (avoid division by zero)
	var aiActivePercentage, humanTakeoverPercentage float64
	if totalConversations > 0 {
		aiActivePercentage = float64(aiActive) / float64(totalConversations) * 100
		humanTakeoverPercentage = float64(humanTakeover) / float64(totalConversations) * 100
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
			"ai_active_percentage":      aiActivePercentage,
			"human_takeover_percentage": humanTakeoverPercentage,
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
			&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Intro,
			&ai.Balas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
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
				// It's plain text, convert to proper JSON string
				jsonBytes, err := json.Marshal(convLastJSON.String)
				if err != nil {
					logrus.WithError(err).WithField("conv_last", convLastJSON.String).Warn("Failed to marshal conv_last as JSON, setting to null")
					ai.ConvLast = json.RawMessage("null")
				} else {
					ai.ConvLast = json.RawMessage(jsonBytes)
				}
			}
		} else {
			// Set to null JSON if empty or invalid
			ai.ConvLast = json.RawMessage("null")
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
			&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Intro,
			&ai.Balas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
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
				// It's plain text, convert to proper JSON string
				jsonBytes, err := json.Marshal(convLastJSON.String)
				if err != nil {
					logrus.WithError(err).WithField("conv_last", convLastJSON.String).Warn("Failed to marshal conv_last as JSON, setting to null")
					ai.ConvLast = json.RawMessage("null")
				} else {
					ai.ConvLast = json.RawMessage(jsonBytes)
				}
			}
		} else {
			// Set to null JSON if empty or invalid
			ai.ConvLast = json.RawMessage("null")
		}

		conversations = append(conversations, ai)
	}

	return conversations, nil
}

// GetConversationHistory retrieves conversation history for a prospect
func (r *aiWhatsappRepository) GetConversationHistory(prospectNum string, limit int) ([]models.ConversationLog, error) {
	query := `
		SELECT id, prospect_num, message, sender, stage, created_at
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
			&log.ID, &log.ProspectNum, &log.Message, 
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
// WARNING: This function overwrites ALL fields. Use UpdateFlowTrackingFields for flow-specific updates
// to preserve conversation history and other important data
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
			human = ?, niche = ?, intro = ?, 
			balas = ?, keywordiklan = ?, marketer = ?, update_today = ?, 
			current_node_id = ?, waiting_for_reply = ?, flow_id = ?, last_node_id = ?,
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

	// Handle flow tracking fields as sql.NullString and sql.NullInt32
	var currentNodeIDValue, flowIDValue, lastNodeIDValue interface{}
	var waitingForReplyValue interface{}
	
	if ai.CurrentNodeID.Valid {
		currentNodeIDValue = ai.CurrentNodeID.String
	} else {
		currentNodeIDValue = nil
	}
	
	if ai.FlowID.Valid {
		flowIDValue = ai.FlowID.String
	} else {
		flowIDValue = nil
	}
	
	if ai.LastNodeID.Valid {
		lastNodeIDValue = ai.LastNodeID.String
	} else {
		lastNodeIDValue = nil
	}
	
	if ai.WaitingForReply.Valid {
		waitingForReplyValue = ai.WaitingForReply.Int32
	} else {
		waitingForReplyValue = nil
	}

	_, err := r.db.Exec(query,
		ai.IDDevice, ai.Stage, ai.DateOrder, convLastValue, convCurrentValue,
		ai.Human, ai.Niche, ai.Intro,
		ai.Balas, ai.KeywordIklan, ai.Marketer, ai.UpdateToday,
		currentNodeIDValue, waitingForReplyValue, flowIDValue, lastNodeIDValue,
		ai.UpdatedAt, ai.IDProspect,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to update AI WhatsApp conversation")
		return fmt.Errorf("failed to update AI WhatsApp conversation: %w", err)
	}

	logrus.WithField("id_prospect", ai.IDProspect).Info("AI WhatsApp conversation updated successfully")
	return nil
}

// UpdateFlowTrackingFields updates only flow tracking fields without overwriting conversation history
// This function preserves conv_last, niche, intro and other important data
func (r *aiWhatsappRepository) UpdateFlowTrackingFields(prospectNum, idDevice string, flowID, currentNodeID, lastNodeID string, waitingForReply int, executionStatus, executionID string) error {
	query := `
		UPDATE ai_whatsapp_nodepath SET 
			flow_id = ?, flow_reference = ?, current_node_id = ?, last_node_id = ?, waiting_for_reply = ?,
			execution_status = ?, execution_id = ?, updated_at = ?
		WHERE prospect_num = ? AND id_device = ?
	`

	// Handle flow tracking fields as sql.NullString and sql.NullInt32
	var currentNodeIDValue, flowIDValue, lastNodeIDValue interface{}
	var waitingForReplyValue interface{}
	var executionStatusValue, executionIDValue interface{}
	
	if currentNodeID != "" {
		currentNodeIDValue = currentNodeID
	} else {
		currentNodeIDValue = nil
	}
	
	if flowID != "" {
		flowIDValue = flowID
	} else {
		flowIDValue = nil
	}
	
	if lastNodeID != "" {
		lastNodeIDValue = lastNodeID
	} else {
		lastNodeIDValue = nil
	}
	
	waitingForReplyValue = waitingForReply
	
	if executionStatus != "" {
		executionStatusValue = executionStatus
	} else {
		executionStatusValue = nil
	}
	
	if executionID != "" {
		executionIDValue = executionID
	} else {
		executionIDValue = nil
	}

	// Debug logging before update
	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"id_device": idDevice,
		"flow_id_input": flowID,
		"flow_id_value": flowIDValue,
		"current_node_id": currentNodeID,
		"execution_status": executionStatus,
		"execution_id": executionID,
	}).Info("DEBUG: About to update flow tracking fields")

	result, err := r.db.Exec(query,
		flowIDValue, flowIDValue, currentNodeIDValue, lastNodeIDValue, waitingForReplyValue,
		executionStatusValue, executionIDValue, time.Now(),
		prospectNum, idDevice,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to update flow tracking fields")
		return fmt.Errorf("failed to update flow tracking fields: %w", err)
	}

	// Check how many rows were affected
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logrus.WithError(err).Warn("Could not get rows affected count")
	}

	logrus.WithFields(logrus.Fields{
		"prospect_num": prospectNum,
		"id_device": idDevice,
		"flow_id": flowID,
		"current_node_id": currentNodeID,
		"rows_affected": rowsAffected,
	}).Info("Flow tracking fields updated successfully")
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
	// Check if database connection is available
	if r.db == nil {
		return nil, fmt.Errorf("database connection is not available")
	}

	query := `
		SELECT id_prospect, id_device, prospect_num, stage, date_order, conv_last, 
		       conv_current, human, niche, intro, 
		       balas, keywordiklan, marketer, update_today, 
		       created_at, updated_at,
		       current_node_id, waiting_for_reply, flow_id, last_node_id, 
		       flow_reference, execution_status, execution_id
		FROM ai_whatsapp_nodepath 
		WHERE prospect_num = ? AND id_device = ?
	`

	row := r.db.QueryRow(query, prospectNum, idDevice)

	ai := &models.AIWhatsapp{}
	var convLastJSON sql.NullString
	var convCurrentSQL sql.NullString

	err := row.Scan(
		&ai.IDProspect, &ai.IDDevice, &ai.ProspectNum, &ai.Stage, &ai.DateOrder, &convLastJSON,
		&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Intro,
		&ai.Balas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
		&ai.CreatedAt, &ai.UpdatedAt,
		&ai.CurrentNodeID, &ai.WaitingForReply, &ai.FlowID, &ai.LastNodeID,
		&ai.FlowReference, &ai.ExecutionStatus, &ai.ExecutionID,
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
			// It's plain text, convert to proper JSON string
			jsonBytes, err := json.Marshal(convLastJSON.String)
			if err != nil {
				logrus.WithError(err).WithField("conv_last", convLastJSON.String).Warn("Failed to marshal conv_last as JSON, setting to null")
				ai.ConvLast = json.RawMessage("null")
			} else {
				ai.ConvLast = json.RawMessage(jsonBytes)
			}
		}
	} else {
		// Set to null JSON if empty or invalid
		ai.ConvLast = json.RawMessage("null")
	}

	return ai, nil
}

// SaveConversationHistory saves conversation history to conv_last field as plain text
// If record exists, it updates the conv_last field; otherwise, it creates a new record
// Saves NULL instead of empty string when there's no conversation data
// Uses database transactions to ensure data consistency
func (r *aiWhatsappRepository) SaveConversationHistory(prospectNum, idDevice, userMessage, botResponse, stage string) error {
	return utils.WithTransaction(r.db, func(tx *sql.Tx) error {
		// Check if record exists within transaction
		var existingID *int
		var existingConvLast []byte
		checkQuery := `
			SELECT id_prospect, conv_last 
			FROM ai_whatsapp_nodepath 
			WHERE prospect_num = ? AND id_device = ?
			FOR UPDATE
		`
		err := tx.QueryRow(checkQuery, prospectNum, idDevice).Scan(&existingID, &existingConvLast)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check existing record: %w", err)
		}

		// Get existing conversation history as plain text
		var convHistory string
		if existingID != nil && existingConvLast != nil {
			// Check if existing data is JSON format (for backward compatibility)
			var existingHistory interface{}
			if err := json.Unmarshal(existingConvLast, &existingHistory); err == nil {
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
				convHistory = string(existingConvLast)
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

		now := time.Now()
		if existingID != nil {
			// Update existing record within transaction
			updateQuery := `
				UPDATE ai_whatsapp_nodepath 
				SET conv_last = ?, stage = ?, updated_at = ?
				WHERE prospect_num = ? AND id_device = ?
			`
			_, err = tx.Exec(updateQuery, convLastValue, stage, now, prospectNum, idDevice)
			if err != nil {
				return fmt.Errorf("failed to update conversation history: %w", err)
			}
			logrus.WithFields(logrus.Fields{
				"prospect_num": prospectNum,
				"id_device": idDevice,
			}).Info("Conversation history updated successfully")
		} else {
			// Create new record within transaction
			insertQuery := `
				INSERT INTO ai_whatsapp_nodepath (
					id_device, prospect_num, stage, conv_last, human, 
					created_at, updated_at
				) VALUES (?, ?, ?, ?, ?, ?, ?)
			`
			_, err = tx.Exec(insertQuery, idDevice, prospectNum, stage, convLastValue, 0, now, now)
			if err != nil {
				return fmt.Errorf("failed to create new conversation record: %w", err)
			}
			logrus.WithFields(logrus.Fields{
				"prospect_num": prospectNum,
				"id_device": idDevice,
			}).Info("New conversation record created successfully")
		}

		return nil
	})
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
			&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Intro,
			&ai.Balas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
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
				// It's plain text, convert to proper JSON string
				jsonBytes, err := json.Marshal(convLastJSON.String)
				if err != nil {
					logrus.WithError(err).WithField("conv_last", convLastJSON.String).Warn("Failed to marshal conv_last as JSON, setting to null")
					ai.ConvLast = json.RawMessage("null")
				} else {
					ai.ConvLast = json.RawMessage(jsonBytes)
				}
			}
		} else {
			// Set to null JSON if empty or invalid
			ai.ConvLast = json.RawMessage("null")
		}

		conversations = append(conversations, ai)
	}

	return conversations, nil
}