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
	UpdateProspectName(prospectNum, idDevice, prospectName string) error
	UpdateHumanTakeover(prospectNum string, human int) error
	UpdateHumanStatus(idProspect string, human int) error
	UpdateConvCurrent(prospectNum string, convCurrent string) error
	UpdateConvLast(prospectNum string, convLast interface{}) error
	UpdateWaitingStatus(executionID string, waitingValue int32) error
	SaveConversationHistory(prospectNum, idDevice, userMessage, botResponse, stage, prospectName string) error

	// Delete operations
	DeleteAIWhatsapp(id int) error
	DeleteConversationLogs(prospectNum string) error

	// Analytics operations
	GetConversationStats(idDevice string) (map[string]int, error)
	GetActiveConversationCount() (int, error)
	GetConversationsByDateRange(startDate, endDate time.Time) ([]models.AIWhatsapp, error)
	GetAnalyticsData(startDate, endDate time.Time, idDevice string, userID string) (map[string]interface{}, error)

	// Data table operations
	GetAllAIWhatsappData(limit, offset int, deviceFilter, stageFilter, search string, userID string, startDate, endDate *time.Time) ([]models.AIWhatsapp, int, error)

	// Database access for transactions
	GetDB() *sql.DB

	// Session locking operations
	TryAcquireSession(prospectNum, idDevice string) (bool, error)
	ReleaseSession(prospectNum, idDevice string) error
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

	// Handle ConvLast as sql.NullString
	var convLastValue interface{}
	if ai.ConvLast.Valid {
		convLastValue = ai.ConvLast.String
	} else {
		convLastValue = nil
	}

	query := `
		INSERT INTO ai_whatsapp_nodepath (
			id_device, prospect_num, prospect_name, stage, date_order, conv_last, 
			conv_current, human, niche, intro, 
			balas, keywordiklan, marketer, update_today, 
			current_node_id, waiting_for_reply, flow_id, last_node_id,
			flow_reference, execution_id, execution_status,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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

	// Handle ProspectName as sql.NullString - Default to "Sis" if empty
	var prospectNameValue interface{}
	if ai.ProspectName.Valid && ai.ProspectName.String != "" {
		prospectNameValue = ai.ProspectName.String
	} else {
		prospectNameValue = "Sis" // Default value
	}

	// Handle Stage as sql.NullString - MUST be NULL not empty string
	var stageValue interface{}
	if ai.Stage.Valid && ai.Stage.String != "" {
		stageValue = ai.Stage.String
	} else {
		stageValue = nil
	}

	// Handle Intro properly - should be NULL if empty, not empty string
	var introValue interface{}
	if ai.Intro.Valid && ai.Intro.String != "" {
		introValue = ai.Intro.String
	} else {
		introValue = nil
	}

	// Handle other nullable fields - MUST be NULL not empty string
	var balasValue interface{}
	if ai.Balas.Valid && ai.Balas.String != "" {
		balasValue = ai.Balas.String
	} else {
		balasValue = nil
	}

	var keywordIklanValue interface{}
	if ai.KeywordIklan.Valid && ai.KeywordIklan.String != "" {
		keywordIklanValue = ai.KeywordIklan.String
	} else {
		keywordIklanValue = nil
	}

	var marketerValue interface{}
	if ai.Marketer.Valid && ai.Marketer.String != "" {
		marketerValue = ai.Marketer.String
	} else {
		marketerValue = nil
	}

	_, err := r.db.Exec(query,
		ai.IDDevice, ai.ProspectNum, prospectNameValue, stageValue, ai.DateOrder, convLastValue,
		convCurrentValue, ai.Human, ai.Niche, introValue,
		balasValue, keywordIklanValue, marketerValue, ai.UpdateToday,
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

	// Handle conv_last data - store as sql.NullString
	ai.ConvLast = convLastJSON

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
		       conv_current, human, niche, intro, 
		       balas, keywordiklan, marketer, update_today, 
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

	// Handle conv_last data - store as sql.NullString
	ai.ConvLast = convLastJSON

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
		       conv_current, human, niche, intro, 
		       balas, keywordiklan, marketer, update_today, 
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

		// Handle conv_last data - store as sql.NullString
		ai.ConvLast = convLastJSON

		conversations = append(conversations, ai)
	}

	return conversations, nil
}

// UpdateProspectName updates the prospect_name field in ai_whatsapp_nodepath
func (r *aiWhatsappRepository) UpdateProspectName(prospectNum, idDevice, prospectName string) error {
	// Check if database connection is available
	if r.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	// Handle prospect name - default to "Sis" if empty, NULL if still empty somehow
	var nameValue interface{}
	if prospectName == "" {
		nameValue = "Sis" // Default value when no name provided
	} else {
		nameValue = prospectName
	}

	query := `UPDATE ai_whatsapp_nodepath SET prospect_name = ?, updated_at = ? WHERE prospect_num = ? AND id_device = ?`
	now := time.Now()

	result, err := r.db.Exec(query, nameValue, now, prospectNum, idDevice)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"prospect_num":  prospectNum,
			"id_device":     idDevice,
			"prospect_name": prospectName,
		}).Error("Failed to update prospect_name")
		return fmt.Errorf("failed to update prospect_name: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		logrus.WithFields(logrus.Fields{
			"prospect_num":  prospectNum,
			"id_device":     idDevice,
			"prospect_name": prospectName,
		}).Info("Prospect name updated successfully")
	}

	return nil
}

// GetAllAIWhatsappData retrieves all AI WhatsApp conversation records with pagination and filtering
func (r *aiWhatsappRepository) GetAllAIWhatsappData(limit, offset int, deviceFilter, stageFilter, search string, userID string, startDate, endDate *time.Time) ([]models.AIWhatsapp, int, error) {
	// Build base query with JOIN to filter by user
	baseQuery := `
		SELECT a.id_prospect, a.id_device, a.prospect_num, a.prospect_name, a.stage, a.date_order, a.conv_last, 
		       a.conv_current, a.human, a.niche, a.intro, 
		       a.balas, a.keywordiklan, a.marketer, a.update_today, 
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

	// Add device filter (supports single device or comma-separated multiple devices)
	if deviceFilter != "" {
		// Handle comma-separated device IDs
		if strings.Contains(deviceFilter, ",") {
			deviceIDs := strings.Split(deviceFilter, ",")
			placeholders := make([]string, len(deviceIDs))
			for i, deviceID := range deviceIDs {
				placeholders[i] = "?"
				args = append(args, strings.TrimSpace(deviceID))
				countArgs = append(countArgs, strings.TrimSpace(deviceID))
			}
			conditions = append(conditions, fmt.Sprintf("a.id_device IN (%s)", strings.Join(placeholders, ",")))
			logrus.WithFields(logrus.Fields{
				"deviceFilter": deviceFilter,
				"deviceIDs":    deviceIDs,
			}).Info("Added multiple device filter to AI WhatsApp data query")
		} else {
			// Single device filter
			conditions = append(conditions, "a.id_device = ?")
			args = append(args, deviceFilter)
			countArgs = append(countArgs, deviceFilter)
		}
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

	// Add date range filter
	if startDate != nil && endDate != nil {
		conditions = append(conditions, "a.created_at BETWEEN ? AND ?")
		args = append(args, *startDate, *endDate)
		countArgs = append(countArgs, *startDate, *endDate)
		logrus.WithFields(logrus.Fields{
			"startDate": startDate.Format("2006-01-02"),
			"endDate":   endDate.Format("2006-01-02"),
		}).Info("Added date range filter to AI WhatsApp data query")
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
		// If there's no data or query fails, return empty result instead of error
		logrus.WithError(err).Warn("No data found or failed to get total count for AI WhatsApp data - returning empty result")
		return []models.AIWhatsapp{}, 0, nil
	}

	// If total is 0, return empty result early
	if total == 0 {
		logrus.Info("No AI WhatsApp data found for given filters - returning empty result")
		return []models.AIWhatsapp{}, 0, nil
	}

	// Execute main query
	rows, err := r.db.Query(baseQuery, args...)
	if err != nil {
		// If query fails but we have a count, something is wrong with the query
		logrus.WithError(err).Warn("Failed to execute main query for AI WhatsApp data - returning empty result")
		return []models.AIWhatsapp{}, 0, nil
	}
	defer rows.Close()

	var conversations []models.AIWhatsapp
	for rows.Next() {
		ai := models.AIWhatsapp{}
		var convLastJSON sql.NullString
		var convCurrentSQL sql.NullString

		err := rows.Scan(
			&ai.IDProspect, &ai.IDDevice, &ai.ProspectNum, &ai.ProspectName, &ai.Stage, &ai.DateOrder, &convLastJSON,
			&convCurrentSQL, &ai.Human, &ai.Niche, &ai.Intro,
			&ai.Balas, &ai.KeywordIklan, &ai.Marketer, &ai.UpdateToday,
			&ai.CreatedAt, &ai.UpdatedAt,
		)

		ai.ConvCurrent = convCurrentSQL

		if err != nil {
			logrus.WithError(err).Error("Failed to scan AI WhatsApp conversation")
			continue
		}

		// Handle conv_last data - store as sql.NullString
		ai.ConvLast = convLastJSON

		conversations = append(conversations, ai)
	}

	if err = rows.Err(); err != nil {
		logrus.WithError(err).Error("Error iterating over AI WhatsApp data rows")
		return nil, 0, fmt.Errorf("error iterating over rows: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"total_records":    total,
		"returned_records": len(conversations),
		"limit":            limit,
		"offset":           offset,
	}).Info("Successfully retrieved AI WhatsApp data")

	return conversations, total, nil
}

// GetAnalyticsData retrieves analytics data from ai_whatsapp_nodepath table with date filtering
func (r *aiWhatsappRepository) GetAnalyticsData(startDate, endDate time.Time, idDevice string, userID string) (map[string]interface{}, error) {
	logrus.WithFields(logrus.Fields{
		"startDate": startDate.Format("2006-01-02"),
		"endDate":   endDate.Format("2006-01-02"),
		"idDevice":  idDevice,
		"userID":    userID,
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
			"userID":  userID,
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

	// Build query with IN clause for user's devices - using DATE() for Y-m-d filtering only
	placeholders := make([]string, len(userDevices))
	args := []interface{}{}
	for i, device := range userDevices {
		placeholders[i] = "?"
		args = append(args, device)
	}
	// Add date parameters for DATE() filtering (Y-m-d format only)
	args = append(args, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	// Base query using IN clause instead of JOIN - using DATE() for precise Y-m-d filtering
	baseQuery := fmt.Sprintf(`
		SELECT 
			COUNT(*) as total_conversations,
			COUNT(CASE WHEN human = 0 THEN 1 END) as ai_active,
			COUNT(CASE WHEN human = 1 THEN 1 END) as human_takeover,
			COUNT(DISTINCT id_device) as unique_devices,
			COUNT(DISTINCT niche) as unique_niches,
			COUNT(CASE WHEN stage IS NOT NULL AND stage != '' THEN 1 END) as conversations_with_stage
		FROM ai_whatsapp_nodepath
		WHERE id_device IN (%s) AND DATE(created_at) >= ? AND DATE(created_at) <= ?
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
		"totalConversations":     totalConversations,
		"aiActive":               aiActive,
		"humanTakeover":          humanTakeover,
		"uniqueDevices":          uniqueDevices,
		"uniqueNiches":           uniqueNiches,
		"conversationsWithStage": conversationsWithStage,
	}).Info("Analytics query results")

	// Get daily breakdown - using DATE() for Y-m-d filtering only
	dailyQuery := fmt.Sprintf(`
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as conversations,
			COUNT(CASE WHEN human = 0 THEN 1 END) as ai_conversations,
			COUNT(CASE WHEN human = 1 THEN 1 END) as human_conversations
		FROM ai_whatsapp_nodepath
		WHERE id_device IN (%s) AND DATE(created_at) >= ? AND DATE(created_at) <= ?
	`, strings.Join(placeholders, ","))

	// Reset args for daily query - using Y-m-d format only
	dailyArgs := []interface{}{}
	for _, device := range userDevices {
		dailyArgs = append(dailyArgs, device)
	}
	dailyArgs = append(dailyArgs, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	if idDevice != "" && idDevice != "all" {
		dailyQuery += " AND id_device = ?"
		dailyArgs = append(dailyArgs, idDevice)
	}
	dailyQuery += " GROUP BY DATE(created_at) ORDER BY DATE(created_at)"

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

			// Clean date format - ensure only YYYY-MM-DD format without any timestamp
			cleanDate := date
			if len(cleanDate) > 10 {
				cleanDate = cleanDate[:10] // Take only YYYY-MM-DD part
			}

			dailyData = append(dailyData, map[string]interface{}{
				"date":                cleanDate, // Simple date format: YYYY-MM-DD
				"conversations":       conversations,
				"ai_conversations":    aiConversations,
				"human_conversations": humanConversations,
			})
		}
	}

	// Get stage distribution including NULL stages - using DATE() for Y-m-d filtering only
	stageQuery := fmt.Sprintf(`
		SELECT 
			CASE 
				WHEN stage IS NULL OR stage = '' THEN 'Welcome Message' 
				ELSE stage 
			END as stage_name,
			COUNT(*) as count
		FROM ai_whatsapp_nodepath
		WHERE id_device IN (%s) AND DATE(created_at) >= ? AND DATE(created_at) <= ?
	`, strings.Join(placeholders, ","))

	// Reset args for stage query - using Y-m-d format only
	stageArgs := []interface{}{}
	for _, device := range userDevices {
		stageArgs = append(stageArgs, device)
	}
	stageArgs = append(stageArgs, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

	if idDevice != "" && idDevice != "all" {
		stageQuery += " AND id_device = ?"
		stageArgs = append(stageArgs, idDevice)
	}
	stageQuery += " GROUP BY stage_name ORDER BY count DESC"

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

		// Handle conv_last data - store as sql.NullString
		ai.ConvLast = convLastJSON

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
			// Store conv_last as sql.NullString
			ai.ConvLast = convLastJSON
		} else {
			// Set to empty sql.NullString if invalid
			ai.ConvLast = sql.NullString{Valid: false}
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

	// Handle conv_last as sql.NullString
	var convLastValue interface{}
	if ai.ConvLast.Valid {
		convLastValue = ai.ConvLast.String
	} else {
		convLastValue = nil
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
		"prospect_num":     prospectNum,
		"id_device":        idDevice,
		"flow_id_input":    flowID,
		"flow_id_value":    flowIDValue,
		"current_node_id":  currentNodeID,
		"execution_status": executionStatus,
		"execution_id":     executionID,
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
		"prospect_num":    prospectNum,
		"id_device":       idDevice,
		"flow_id":         flowID,
		"current_node_id": currentNodeID,
		"rows_affected":   rowsAffected,
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

// UpdateWaitingStatus updates the waiting_for_reply status for an execution
func (r *aiWhatsappRepository) UpdateWaitingStatus(executionID string, waitingValue int32) error {
	// Check if database connection is available
	if r.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	query := `UPDATE ai_whatsapp_nodepath SET waiting_for_reply = ?, updated_at = ? WHERE execution_id = ?`
	now := time.Now()

	result, err := r.db.Exec(query, waitingValue, now, executionID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"execution_id":  executionID,
			"waiting_value": waitingValue,
		}).Error("Failed to update waiting_for_reply status")
		return fmt.Errorf("failed to update waiting_for_reply status: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		logrus.WithFields(logrus.Fields{
			"execution_id":  executionID,
			"waiting_value": waitingValue,
		}).Info("Waiting status updated successfully")
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
		"human":        human,
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
		"prospect_num":      prospectNum,
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
		// Store conv_last as sql.NullString
		ai.ConvLast = convLastJSON
	} else {
		// Set to empty sql.NullString if invalid
		ai.ConvLast = sql.NullString{Valid: false}
	}

	return ai, nil
}

// SaveConversationHistory saves conversation history to conv_last field as plain text
// If record exists, it updates the conv_last field; otherwise, it creates a new record
// Saves NULL instead of empty string when there's no conversation data
// Uses database transactions to ensure data consistency
// Now includes prospect_name parameter to ensure names are always updated
func (r *aiWhatsappRepository) SaveConversationHistory(prospectNum, idDevice, userMessage, botResponse, stage, prospectName string) error {
	// CRITICAL: Handle stage - MUST be NULL if empty string for Chatbot AI
	var stageValue interface{}
	if stage != "" {
		stageValue = stage
	} else {
		stageValue = nil // ALWAYS NULL for empty stage - no exceptions
	}

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
			// IMPORTANT: Only update conv_last, stage, and updated_at
			// DO NOT update prospect_name, prospect_num, or human - these are set only on creation
			updateQuery := `
				UPDATE ai_whatsapp_nodepath 
				SET conv_last = ?, stage = ?, updated_at = ?
				WHERE prospect_num = ? AND id_device = ?
			`
			_, err = tx.Exec(updateQuery, convLastValue, stageValue, now, prospectNum, idDevice)
			if err != nil {
				return fmt.Errorf("failed to update conversation history: %w", err)
			}
			logrus.WithFields(logrus.Fields{
				"prospect_num":    prospectNum,
				"id_device":       idDevice,
				"updating_fields": "conv_last, stage, updated_at ONLY",
				"NOT_updating":    "prospect_name, human, prospect_num",
			}).Info("Conversation history updated successfully")
		} else {
			// Create new record within transaction
			// Only set prospect_name, prospect_num, and human when creating NEW records
			// Default prospect name to "Sis" if empty for new records ONLY
			if prospectName == "" {
				prospectName = "Sis"
			}

			logrus.WithFields(logrus.Fields{
				"prospect_num":  prospectNum,
				"id_device":     idDevice,
				"prospect_name": prospectName,
				"creating_new":  true,
			}).Info("Creating new conversation record with initial prospect data")

			insertQuery := `
				INSERT INTO ai_whatsapp_nodepath (
					id_device, prospect_num, stage, conv_last, prospect_name, human, 
					created_at, updated_at
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`
			_, err = tx.Exec(insertQuery, idDevice, prospectNum, stageValue, convLastValue, prospectName, 0, now, now)
			if err != nil {
				return fmt.Errorf("failed to create new conversation record: %w", err)
			}
			logrus.WithFields(logrus.Fields{
				"prospect_num": prospectNum,
				"id_device":    idDevice,
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

		// Store conv_last as sql.NullString
		ai.ConvLast = convLastJSON

		conversations = append(conversations, ai)
	}

	return conversations, nil
}

// UpdateHumanStatus updates the human status for a specific conversation by ID
func (r *aiWhatsappRepository) UpdateHumanStatus(idProspect string, human int) error {
	query := `
		UPDATE ai_whatsapp_nodepath 
		SET human = ?, updated_at = NOW() 
		WHERE id_prospect = ?
	`

	_, err := r.db.Exec(query, human, idProspect)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"id_prospect": idProspect,
			"human":       human,
		}).Error("Failed to update human status")
		return fmt.Errorf("failed to update human status: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"id_prospect": idProspect,
		"human":       human,
	}).Info("Successfully updated human status")

	return nil
}

// TryAcquireSession attempts to acquire a session lock for the given phone number and device
// Returns true if lock acquired, false if already locked
// Uses SELECT FOR UPDATE to create a true blocking lock that prevents concurrent processing
func (r *aiWhatsappRepository) TryAcquireSession(phoneNumber, deviceID string) (bool, error) {
	currentTimestamp := time.Now().Format("2006-01-02 15:04:05")
	lockTimeout := 30 // seconds - max time to hold the lock

	// Start a transaction for proper locking
	tx, err := r.db.Begin()
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Error("🔒 SESSION LOCK: ❌ Failed to start transaction")
		return false, err
	}

	// Set lock wait timeout to prevent indefinite blocking
	_, err = tx.Exec("SET SESSION innodb_lock_wait_timeout = 2")
	if err != nil {
		tx.Rollback()
		logrus.WithError(err).Error("🔒 SESSION LOCK: ❌ Failed to set lock timeout")
		return false, err
	}

	// Try to get existing lock with SELECT FOR UPDATE (blocks if another transaction holds it)
	var existingTimestamp string
	var lockedAt time.Time
	checkQuery := `
		SELECT timestamp, STR_TO_DATE(timestamp, '%Y-%m-%d %H:%i:%s') as locked_at
		FROM ai_whatsapp_session_nodepath 
		WHERE id_prospect = ? AND id_device = ?
		FOR UPDATE
	`

	err = tx.QueryRow(checkQuery, phoneNumber, deviceID).Scan(&existingTimestamp, &lockedAt)

	if err == sql.ErrNoRows {
		// No existing lock - create one
		insertQuery := `
			INSERT INTO ai_whatsapp_session_nodepath (id_prospect, id_device, timestamp)
			VALUES (?, ?, ?)
		`
		_, err = tx.Exec(insertQuery, phoneNumber, deviceID, currentTimestamp)
		if err != nil {
			tx.Rollback()
			logrus.WithError(err).WithFields(logrus.Fields{
				"phone_number": phoneNumber,
				"device_id":    deviceID,
			}).Error("🔒 SESSION LOCK: ❌ Failed to insert new lock")
			return false, err
		}

		// Commit the transaction - lock is now held until ReleaseSession is called
		err = tx.Commit()
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"phone_number": phoneNumber,
				"device_id":    deviceID,
			}).Error("🔒 SESSION LOCK: ❌ Failed to commit lock transaction")
			return false, err
		}

		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
			"timestamp":    currentTimestamp,
		}).Info("🔒 SESSION LOCK: ✅ Acquired successfully (NEW)")

		return true, nil
	} else if err != nil {
		tx.Rollback()
		// Lock wait timeout exceeded - another process is holding the lock
		if strings.Contains(err.Error(), "Lock wait timeout") {
			logrus.WithFields(logrus.Fields{
				"phone_number": phoneNumber,
				"device_id":    deviceID,
			}).Warn("🔒 SESSION LOCK: ⏸️ Already locked by another process - BLOCKING DUPLICATE")
			return false, nil // Not an error, just locked
		}

		logrus.WithError(err).WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Error("🔒 SESSION LOCK: ❌ Failed to check existing lock")
		return false, err
	}

	// Lock exists - check if it's stale (older than lockTimeout seconds)
	lockAge := time.Since(lockedAt).Seconds()
	if lockAge > float64(lockTimeout) {
		// Stale lock - update it and take over
		updateQuery := `
			UPDATE ai_whatsapp_session_nodepath 
			SET timestamp = ?
			WHERE id_prospect = ? AND id_device = ?
		`
		_, err = tx.Exec(updateQuery, currentTimestamp, phoneNumber, deviceID)
		if err != nil {
			tx.Rollback()
			logrus.WithError(err).WithFields(logrus.Fields{
				"phone_number": phoneNumber,
				"device_id":    deviceID,
			}).Error("🔒 SESSION LOCK: ❌ Failed to update stale lock")
			return false, err
		}

		err = tx.Commit()
		if err != nil {
			tx.Rollback()
			logrus.WithError(err).WithFields(logrus.Fields{
				"phone_number": phoneNumber,
				"device_id":    deviceID,
			}).Error("🔒 SESSION LOCK: ❌ Failed to commit stale lock update")
			return false, err
		}

		logrus.WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
			"lock_age_sec": lockAge,
		}).Warn("🔒 SESSION LOCK: ✅ Acquired by taking over STALE lock")

		return true, nil
	}

	// Active lock exists and is not stale - reject this request
	tx.Rollback()
	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
		"lock_age_sec": lockAge,
	}).Warn("🔒 SESSION LOCK: ⏸️ Already locked (active) - BLOCKING DUPLICATE")

	return false, nil
}

// ReleaseSession releases the session lock for the given phone number and device
func (r *aiWhatsappRepository) ReleaseSession(phoneNumber, deviceID string) error {
	// Use actual database columns: id_prospect, id_device, timestamp
	// Delete the lock record to properly clean up after processing
	query := `
		DELETE FROM ai_whatsapp_session_nodepath 
		WHERE id_prospect = ? AND id_device = ?
	`

	_, err := r.db.Exec(query, phoneNumber, deviceID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"phone_number": phoneNumber,
			"device_id":    deviceID,
		}).Error("🔒 SESSION LOCK: ❌ Failed to release session lock")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"device_id":    deviceID,
	}).Info("🔒 SESSION LOCK: ✅ Released successfully")

	return nil
}
