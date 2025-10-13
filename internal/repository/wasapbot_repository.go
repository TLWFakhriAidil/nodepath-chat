package repository

import (
	"database/sql"
	"fmt"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/utils"

	mysql "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

// WasapBotRepository interface for wasapBot_nodepath operations
type WasapBotRepository interface {
	GetByProspectAndDevice(prospectNum, deviceID string) (*models.WasapBot, error)
	GetActiveExecution(prospectNum, deviceID string) (*models.WasapBot, error)
	GetByExecutionID(executionID string) (*models.WasapBot, error)
	Create(wasapBot *models.WasapBot) error
	Update(wasapBot *models.WasapBot) error
	UpdateExecutionStatus(executionID, status string) error
	UpdateCurrentNode(executionID, nodeID string) error
	UpdateWaitingStatus(executionID string, waitingValue int) error
	SaveConversationHistory(prospectNum, deviceID, userMessage, botResponse, stage, nama string) error
	GetAllWasapBotData(limit, offset int, deviceFilter, stageFilter, statusFilter, search string, userID string) ([]map[string]interface{}, int, error)
	GetAllWasapBotDataWithDates(limit, offset int, deviceFilter, stageFilter, statusFilter, search, dateFrom, dateTo string, userID string) ([]map[string]interface{}, int, error)
	GetWasapBotStats(deviceFilter string, userID string) (map[string]interface{}, error)
	GetWasapBotStatsWithDates(deviceFilter, dateFrom, dateTo string, userID string) (map[string]interface{}, error)
	Delete(idProspect int) error

	// Session locking operations
	TryAcquireSession(prospectNum, deviceID string) (bool, error)
	ReleaseSession(prospectNum, deviceID string) error
}

type wasapBotRepository struct {
	db *sql.DB
}

// NewWasapBotRepository creates a new wasapBot repository
func NewWasapBotRepository(db *sql.DB) WasapBotRepository {
	return &wasapBotRepository{
		db: db,
	}
}

// GetByProspectAndDevice retrieves a wasapBot record by prospect number and device ID
func (r *wasapBotRepository) GetByProspectAndDevice(prospectNum, deviceID string) (*models.WasapBot, error) {
	query := `
		SELECT id_prospect, flow_reference, execution_id, execution_status, flow_id,
		       current_node_id, last_node_id, waiting_for_reply, id_device,
		       prospect_num, niche, peringkat_sekolah, alamat, nama,
		       pakej, no_fon, cara_bayaran, tarikh_gaji, stage, temp_stage,
		       conv_start, conv_last, date_start, date_last, status, staff_cls,
		       umur, kerja, sijil, user_input, alasan, nota
		FROM wasapBot_nodepath
		WHERE prospect_num = ? AND id_device = ?
		LIMIT 1
	`

	var wb models.WasapBot
	err := r.db.QueryRow(query, prospectNum, deviceID).Scan(
		&wb.IDProspect, &wb.FlowReference, &wb.ExecutionID, &wb.ExecutionStatus,
		&wb.FlowID, &wb.CurrentNodeID, &wb.LastNodeID, &wb.WaitingForReply,
		&wb.IDDevice, &wb.ProspectNum, &wb.Niche, &wb.PeringkatSekolah,
		&wb.Alamat, &wb.Nama, &wb.Pakej, &wb.NoFon, &wb.CaraBayaran,
		&wb.TarikhGaji, &wb.Stage, &wb.TempStage, &wb.ConvStart, &wb.ConvLast,
		&wb.DateStart, &wb.DateLast, &wb.Status, &wb.StaffCls, &wb.Umur,
		&wb.Kerja, &wb.Sijil, &wb.UserInput, &wb.Alasan, &wb.Nota,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wasapBot by prospect and device: %w", err)
	}

	return &wb, nil
}

// GetActiveExecution retrieves an active execution for a prospect and device ID
func (r *wasapBotRepository) GetActiveExecution(prospectNum, deviceID string) (*models.WasapBot, error) {
	query := `
		SELECT id_prospect, flow_reference, execution_id, execution_status, flow_id,
		       current_node_id, last_node_id, waiting_for_reply, id_device,
		       prospect_num, niche, peringkat_sekolah, alamat, nama,
		       pakej, no_fon, cara_bayaran, tarikh_gaji, stage, temp_stage,
		       conv_start, conv_last, date_start, date_last, status, staff_cls,
		       umur, kerja, sijil, user_input, alasan, nota
		FROM wasapBot_nodepath
		WHERE prospect_num = ? AND id_device = ? AND execution_status = 'active'
		LIMIT 1
	`

	var wb models.WasapBot
	err := r.db.QueryRow(query, prospectNum, deviceID).Scan(
		&wb.IDProspect, &wb.FlowReference, &wb.ExecutionID, &wb.ExecutionStatus,
		&wb.FlowID, &wb.CurrentNodeID, &wb.LastNodeID, &wb.WaitingForReply,
		&wb.IDDevice, &wb.ProspectNum, &wb.Niche, &wb.PeringkatSekolah,
		&wb.Alamat, &wb.Nama, &wb.Pakej, &wb.NoFon, &wb.CaraBayaran,
		&wb.TarikhGaji, &wb.Stage, &wb.TempStage, &wb.ConvStart, &wb.ConvLast,
		&wb.DateStart, &wb.DateLast, &wb.Status, &wb.StaffCls, &wb.Umur,
		&wb.Kerja, &wb.Sijil, &wb.UserInput, &wb.Alasan, &wb.Nota,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active execution: %w", err)
	}

	return &wb, nil
}

// GetByExecutionID retrieves a wasapBot record by execution ID
func (r *wasapBotRepository) GetByExecutionID(executionID string) (*models.WasapBot, error) {
	query := `
		SELECT id_prospect, flow_reference, execution_id, execution_status, flow_id,
		       current_node_id, last_node_id, waiting_for_reply, id_device,
		       prospect_num, niche, peringkat_sekolah, alamat, nama,
		       pakej, no_fon, cara_bayaran, tarikh_gaji, stage, temp_stage,
		       conv_start, conv_last, date_start, date_last, status, staff_cls,
		       umur, kerja, sijil, user_input, alasan, nota
		FROM wasapBot_nodepath
		WHERE execution_id = ?
		LIMIT 1
	`

	var wb models.WasapBot
	err := r.db.QueryRow(query, executionID).Scan(
		&wb.IDProspect, &wb.FlowReference, &wb.ExecutionID, &wb.ExecutionStatus,
		&wb.FlowID, &wb.CurrentNodeID, &wb.LastNodeID, &wb.WaitingForReply,
		&wb.IDDevice, &wb.ProspectNum, &wb.Niche, &wb.PeringkatSekolah,
		&wb.Alamat, &wb.Nama, &wb.Pakej, &wb.NoFon, &wb.CaraBayaran,
		&wb.TarikhGaji, &wb.Stage, &wb.TempStage, &wb.ConvStart, &wb.ConvLast,
		&wb.DateStart, &wb.DateLast, &wb.Status, &wb.StaffCls, &wb.Umur,
		&wb.Kerja, &wb.Sijil, &wb.UserInput, &wb.Alasan, &wb.Nota,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get wasapBot by execution ID: %w", err)
	}

	return &wb, nil
}

// Create creates a new wasapBot record
func (r *wasapBotRepository) Create(wasapBot *models.WasapBot) error {
	query := `
		INSERT INTO wasapBot_nodepath (
			flow_reference, execution_id, execution_status, flow_id,
			current_node_id, last_node_id, waiting_for_reply, id_device,
			prospect_num, niche, peringkat_sekolah, alamat, nama,
			pakej, no_fon, cara_bayaran, tarikh_gaji, stage, temp_stage,
			conv_start, conv_last, date_start, date_last, status, staff_cls,
			umur, kerja, sijil, user_input, alasan, nota
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		wasapBot.FlowReference, wasapBot.ExecutionID, wasapBot.ExecutionStatus,
		wasapBot.FlowID, wasapBot.CurrentNodeID, wasapBot.LastNodeID,
		wasapBot.WaitingForReply, wasapBot.IDDevice, wasapBot.ProspectNum,
		wasapBot.Niche, wasapBot.PeringkatSekolah,
		wasapBot.Alamat, wasapBot.Nama, wasapBot.Pakej, wasapBot.NoFon,
		wasapBot.CaraBayaran, wasapBot.TarikhGaji, wasapBot.Stage,
		wasapBot.TempStage, wasapBot.ConvStart, wasapBot.ConvLast,
		wasapBot.DateStart, wasapBot.DateLast, wasapBot.Status,
		wasapBot.StaffCls, wasapBot.Umur, wasapBot.Kerja, wasapBot.Sijil,
		wasapBot.UserInput, wasapBot.Alasan, wasapBot.Nota,
	)

	if err != nil {
		return fmt.Errorf("failed to create wasapBot record: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	wasapBot.IDProspect = int(id)
	return nil
}

// Update updates an existing wasapBot record
func (r *wasapBotRepository) Update(wasapBot *models.WasapBot) error {
	query := `
		UPDATE wasapBot_nodepath SET
			flow_reference = ?, execution_id = ?, execution_status = ?, flow_id = ?,
			current_node_id = ?, last_node_id = ?, waiting_for_reply = ?, id_device = ?,
			prospect_num = ?, niche = ?, peringkat_sekolah = ?, alamat = ?,
			nama = ?, pakej = ?, no_fon = ?, cara_bayaran = ?, tarikh_gaji = ?,
			stage = ?, temp_stage = ?, conv_start = ?, conv_last = ?, date_start = ?,
			date_last = ?, status = ?, staff_cls = ?, umur = ?, kerja = ?, sijil = ?,
			user_input = ?, alasan = ?, nota = ?
		WHERE id_prospect = ?
	`

	_, err := r.db.Exec(query,
		wasapBot.FlowReference, wasapBot.ExecutionID, wasapBot.ExecutionStatus,
		wasapBot.FlowID, wasapBot.CurrentNodeID, wasapBot.LastNodeID,
		wasapBot.WaitingForReply, wasapBot.IDDevice, wasapBot.ProspectNum,
		wasapBot.Niche, wasapBot.PeringkatSekolah,
		wasapBot.Alamat, wasapBot.Nama, wasapBot.Pakej, wasapBot.NoFon,
		wasapBot.CaraBayaran, wasapBot.TarikhGaji, wasapBot.Stage,
		wasapBot.TempStage, wasapBot.ConvStart, wasapBot.ConvLast,
		wasapBot.DateStart, wasapBot.DateLast, wasapBot.Status,
		wasapBot.StaffCls, wasapBot.Umur, wasapBot.Kerja, wasapBot.Sijil,
		wasapBot.UserInput, wasapBot.Alasan, wasapBot.Nota,
		wasapBot.IDProspect,
	)

	if err != nil {
		return fmt.Errorf("failed to update wasapBot record: %w", err)
	}

	return nil
}

// UpdateExecutionStatus updates the execution status
func (r *wasapBotRepository) UpdateExecutionStatus(executionID, status string) error {
	query := `UPDATE wasapBot_nodepath SET execution_status = ? WHERE execution_id = ?`
	_, err := r.db.Exec(query, status, executionID)
	if err != nil {
		return fmt.Errorf("failed to update execution status: %w", err)
	}
	return nil
}

// UpdateCurrentNode updates the current node ID
func (r *wasapBotRepository) UpdateCurrentNode(executionID, nodeID string) error {
	query := `UPDATE wasapBot_nodepath SET current_node_id = ? WHERE execution_id = ?`
	_, err := r.db.Exec(query, nodeID, executionID)
	if err != nil {
		return fmt.Errorf("failed to update current node: %w", err)
	}
	return nil
}

// SaveConversationHistory saves conversation history to conv_last field
func (r *wasapBotRepository) SaveConversationHistory(prospectNum, deviceID, userMessage, botResponse, stage, nama string) error {
	return utils.WithTransaction(r.db, func(tx *sql.Tx) error {
		// Check if record exists
		var existingID *int
		var existingConvLast sql.NullString
		checkQuery := `
			SELECT id_prospect, conv_last 
			FROM wasapBot_nodepath 
			WHERE prospect_num = ? AND id_device = ?
			FOR UPDATE
		`
		err := tx.QueryRow(checkQuery, prospectNum, deviceID).Scan(&existingID, &existingConvLast)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check existing record: %w", err)
		}

		// Build conversation history
		var convHistory string
		if existingID != nil && existingConvLast.Valid {
			convHistory = existingConvLast.String
		}

		// Add new conversation entries
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

		// Determine conv_last value
		var convLastValue interface{}
		if convHistory == "" {
			convLastValue = nil
		} else {
			convLastValue = convHistory
		}

		now := time.Now().Format("2006-01-02 15:04:05")

		if existingID != nil {
			// Update existing record
			updateQuery := `
				UPDATE wasapBot_nodepath 
				SET conv_last = ?, stage = ?, nama = ?, date_last = ?
				WHERE prospect_num = ? AND id_device = ?
			`
			_, err = tx.Exec(updateQuery, convLastValue, stage, nama, now, prospectNum, deviceID)
			if err != nil {
				return fmt.Errorf("failed to update conversation history: %w", err)
			}
			logrus.WithFields(logrus.Fields{
				"prospect_num": prospectNum,
			}).Info("WasapBot conversation history updated successfully")
		} else {
			// Create new record
			insertQuery := `
				INSERT INTO wasapBot_nodepath (
					prospect_num, id_device, stage, conv_last, nama, 
					date_start, date_last, status, waiting_for_reply
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`
			_, err = tx.Exec(insertQuery, prospectNum, deviceID, stage, convLastValue, nama,
				now, now, "Prospek", 0)
			if err != nil {
				return fmt.Errorf("failed to create new conversation record: %w", err)
			}
			logrus.WithFields(logrus.Fields{
				"prospect_num": prospectNum,
				"id_device":    deviceID,
			}).Info("New WasapBot conversation record created successfully")
		}

		return nil
	})
}

// UpdateWaitingStatus updates the waiting status for an execution
func (r *wasapBotRepository) UpdateWaitingStatus(executionID string, waitingValue int) error {
	query := `
		UPDATE wasapBot_nodepath 
		SET waiting_for_reply = ?, date_last = NOW()
		WHERE execution_id = ?
	`

	_, err := r.db.Exec(query, waitingValue, executionID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"execution_id":  executionID,
			"waiting_value": waitingValue,
		}).Error("Failed to update waiting status in wasapBot")
		return fmt.Errorf("failed to update waiting status: %w", err)
	}

	return nil
}

// TryAcquireSession attempts to create a session lock for a prospect/device pair
// Returns true when the lock was acquired, false if a lock already exists
func (r *wasapBotRepository) TryAcquireSession(prospectNum, deviceID string) (bool, error) {
	if r.db == nil {
		return false, fmt.Errorf("database connection is not available")
	}

	const query = `INSERT INTO wasapBot_session_nodepath (id_prospect, id_device, ` + "`timestamp`" + `) VALUES (?, ?, ?)`
	_, err := r.db.Exec(query, prospectNum, deviceID, time.Now().Format(time.RFC3339Nano))
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			if mysqlErr.Number == 1062 {
				return false, nil
			}
		}
		return false, fmt.Errorf("failed to acquire WasapBot session lock: %w", err)
	}

	return true, nil
}

// ReleaseSession removes the session lock for a prospect/device pair
func (r *wasapBotRepository) ReleaseSession(prospectNum, deviceID string) error {
	if r.db == nil {
		return fmt.Errorf("database connection is not available")
	}

	const query = `DELETE FROM wasapBot_session_nodepath WHERE id_prospect = ? AND id_device = ?`
	if _, err := r.db.Exec(query, prospectNum, deviceID); err != nil {
		return fmt.Errorf("failed to release WasapBot session lock: %w", err)
	}

	return nil
}

// GetAllWasapBotData retrieves all WasapBot data with filters
func (r *wasapBotRepository) GetAllWasapBotData(limit, offset int, deviceFilter, stageFilter, statusFilter, search string, userID string) ([]map[string]interface{}, int, error) {
	// Log incoming parameters
	logrus.WithFields(logrus.Fields{
		"limit":        limit,
		"offset":       offset,
		"deviceFilter": deviceFilter,
		"stageFilter":  stageFilter,
		"statusFilter": statusFilter,
		"search":       search,
		"userID":       userID,
	}).Info("GetAllWasapBotData called")

	// Build query with filters - select all needed columns
	query := `
		SELECT id_prospect, prospect_num, nama, stage, date_last, id_device,
		       niche, status, alamat, pakej, cara_bayaran, tarikh_gaji, current_node_id, no_fon
		FROM wasapBot_nodepath
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM wasapBot_nodepath WHERE 1=1`
	args := []interface{}{}
	countArgs := []interface{}{}

	// Apply filters
	if deviceFilter != "" && deviceFilter != "all" {
		// Handle multiple device IDs
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			placeholders := utils.GeneratePlaceholders(len(devices))
			query += " AND id_device IN (" + placeholders + ")"
			countQuery += " AND id_device IN (" + placeholders + ")"
			for _, device := range devices {
				args = append(args, device)
				countArgs = append(countArgs, device)
			}
			logrus.WithField("device_filter_applied", devices).Info("Applying device filter for multiple devices")
		}
	}

	if stageFilter != "" && stageFilter != "all" {
		query += " AND stage = ?"
		countQuery += " AND stage = ?"
		args = append(args, stageFilter)
		countArgs = append(countArgs, stageFilter)
	}

	if statusFilter != "" && statusFilter != "all" {
		query += " AND status = ?"
		countQuery += " AND status = ?"
		args = append(args, statusFilter)
		countArgs = append(countArgs, statusFilter)
	}

	if search != "" {
		query += " AND (prospect_num LIKE ? OR nama LIKE ? OR no_fon LIKE ? OR peringkat_sekolah LIKE ?)"
		countQuery += " AND (prospect_num LIKE ? OR nama LIKE ? OR no_fon LIKE ? OR peringkat_sekolah LIKE ?)"
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam, searchParam, searchParam)
		countArgs = append(countArgs, searchParam, searchParam, searchParam, searchParam)
	}

	// Log the final query
	logrus.WithFields(logrus.Fields{
		"count_query": countQuery,
		"count_args":  countArgs,
	}).Debug("Executing count query")

	// Get total count
	var total int
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Failed to get count")
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	logrus.WithField("total_count", total).Info("Total records found")

	// Add ORDER BY and pagination
	query += " ORDER BY date_last DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Log the data query
	logrus.WithFields(logrus.Fields{
		"data_query": query,
		"data_args":  args,
	}).Debug("Executing data query")

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		logrus.WithError(err).Error("Failed to query wasapBot data")
		return nil, 0, fmt.Errorf("failed to query wasapBot data: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	rowCount := 0
	for rows.Next() {
		var (
			idProspect    int
			prospectNum   sql.NullString
			nama          sql.NullString
			stage         sql.NullString
			dateLast      sql.NullString
			deviceID      sql.NullString
			niche         sql.NullString
			status        sql.NullString
			alamat        sql.NullString
			pakej         sql.NullString
			caraBayaran   sql.NullString
			tarikhGaji    sql.NullString
			currentNodeID sql.NullString
			noFon         sql.NullString
		)

		err := rows.Scan(
			&idProspect,
			&prospectNum,
			&nama,
			&stage,
			&dateLast,
			&deviceID,
			&niche,
			&status,
			&alamat,
			&pakej,
			&caraBayaran,
			&tarikhGaji,
			&currentNodeID,
			&noFon,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan wasapBot row")
			continue
		}

		rowCount++

		// Convert to plain map for JSON - match frontend expectations
		record := map[string]interface{}{
			"id_prospect":     idProspect,
			"id_device":       utils.GetStringValue(deviceID),
			"nama":            utils.GetStringValue(nama),
			"prospect_num":    utils.GetStringValue(prospectNum),
			"niche":           utils.GetStringValue(niche),
			"status":          utils.GetStringValue(status),
			"stage":           utils.GetStringValue(stage),
			"alamat":          utils.GetStringValue(alamat),
			"pakej":           utils.GetStringValue(pakej),
			"cara_bayaran":    utils.GetStringValue(caraBayaran),
			"tarikh_gaji":     utils.GetStringValue(tarikhGaji),
			"current_node_id": utils.GetStringValue(currentNodeID),
			"no_fon":          utils.GetStringValue(noFon),
			"date_last":       utils.GetStringValue(dateLast),
		}

		results = append(results, record)

		logrus.WithFields(logrus.Fields{
			"row_id":       idProspect,
			"device":       utils.GetStringValue(deviceID),
			"prospect_num": utils.GetStringValue(prospectNum),
		}).Debug("Added record to results")
	}

	logrus.WithFields(logrus.Fields{
		"rows_scanned":  rowCount,
		"results_count": len(results),
	}).Info("Query completed")

	return results, total, nil
}

// GetWasapBotStats retrieves WasapBot statistics
func (r *wasapBotRepository) GetWasapBotStats(deviceFilter string, userID string) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"totalProspects":      0,
		"activeExecutions":    0,
		"completedExecutions": 0,
		"uniqueSchools":       0,
		"uniquePackages":      0,
		"totalWithPhone":      0,
	}

	baseWhere := "1=1"
	args := []interface{}{}

	if deviceFilter != "" && deviceFilter != "all" {
		baseWhere += " AND instance = ?"
		args = append(args, deviceFilter)
	}

	// Total prospects
	var totalProspects int
	err := r.db.QueryRow("SELECT COUNT(DISTINCT prospect_num) FROM wasapBot_nodepath WHERE "+baseWhere, args...).Scan(&totalProspects)
	if err == nil {
		stats["totalProspects"] = totalProspects
	}

	// Active executions
	var activeExecutions int
	err = r.db.QueryRow("SELECT COUNT(*) FROM wasapBot_nodepath WHERE "+baseWhere+" AND execution_status = 'active'", args...).Scan(&activeExecutions)
	if err == nil {
		stats["activeExecutions"] = activeExecutions
	}

	// Completed executions
	var completedExecutions int
	err = r.db.QueryRow("SELECT COUNT(*) FROM wasapBot_nodepath WHERE "+baseWhere+" AND status = 'Customer'", args...).Scan(&completedExecutions)
	if err == nil {
		stats["completedExecutions"] = completedExecutions
	}

	// Unique schools
	var uniqueSchools int
	err = r.db.QueryRow("SELECT COUNT(DISTINCT peringkat_sekolah) FROM wasapBot_nodepath WHERE "+baseWhere+" AND peringkat_sekolah IS NOT NULL AND peringkat_sekolah != ''", args...).Scan(&uniqueSchools)
	if err == nil {
		stats["uniqueSchools"] = uniqueSchools
	}

	// Unique packages
	var uniquePackages int
	err = r.db.QueryRow("SELECT COUNT(DISTINCT pakej) FROM wasapBot_nodepath WHERE "+baseWhere+" AND pakej IS NOT NULL AND pakej != ''", args...).Scan(&uniquePackages)
	if err == nil {
		stats["uniquePackages"] = uniquePackages
	}

	// Total with phone
	var totalWithPhone int
	err = r.db.QueryRow("SELECT COUNT(*) FROM wasapBot_nodepath WHERE "+baseWhere+" AND no_fon IS NOT NULL AND no_fon != ''", args...).Scan(&totalWithPhone)
	if err == nil {
		stats["totalWithPhone"] = totalWithPhone
	}

	return stats, nil
}

// Delete deletes a WasapBot record by ID
func (r *wasapBotRepository) Delete(idProspect int) error {
	query := `DELETE FROM wasapBot_nodepath WHERE id_prospect = ?`

	result, err := r.db.Exec(query, idProspect)
	if err != nil {
		logrus.WithError(err).WithField("id_prospect", idProspect).Error("Failed to delete WasapBot record")
		return fmt.Errorf("failed to delete WasapBot record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("no record found with id_prospect: %d", idProspect)
	}

	logrus.WithField("id_prospect", idProspect).Info("WasapBot record deleted successfully")
	return nil
}

// GetAllWasapBotDataWithDates retrieves all WasapBot data with filters including date range
func (r *wasapBotRepository) GetAllWasapBotDataWithDates(limit, offset int, deviceFilter, stageFilter, statusFilter, search, dateFrom, dateTo string, userID string) ([]map[string]interface{}, int, error) {
	// Log incoming parameters
	logrus.WithFields(logrus.Fields{
		"limit":        limit,
		"offset":       offset,
		"deviceFilter": deviceFilter,
		"stageFilter":  stageFilter,
		"statusFilter": statusFilter,
		"search":       search,
		"dateFrom":     dateFrom,
		"dateTo":       dateTo,
		"userID":       userID,
	}).Info("GetAllWasapBotDataWithDates called")

	// Build query with filters - select all needed columns including date_start for display
	query := `
		SELECT id_prospect, prospect_num, nama, stage, date_last, date_start, id_device,
		       niche, status, alamat, pakej, cara_bayaran, tarikh_gaji, current_node_id, no_fon
		FROM wasapBot_nodepath
		WHERE 1=1
	`

	countQuery := `SELECT COUNT(*) FROM wasapBot_nodepath WHERE 1=1`
	args := []interface{}{}
	countArgs := []interface{}{}

	// Apply date filters using DATE() function to ignore time
	if dateFrom != "" {
		query += " AND DATE(date_start) >= ?"
		countQuery += " AND DATE(date_start) >= ?"
		args = append(args, dateFrom)
		countArgs = append(countArgs, dateFrom)
		logrus.WithField("date_from_applied", dateFrom).Info("Applying date from filter")
	}

	if dateTo != "" {
		query += " AND DATE(date_start) <= ?"
		countQuery += " AND DATE(date_start) <= ?"
		args = append(args, dateTo)
		countArgs = append(countArgs, dateTo)
		logrus.WithField("date_to_applied", dateTo).Info("Applying date to filter")
	}

	// Apply other filters
	if deviceFilter != "" && deviceFilter != "all" {
		// Handle multiple device IDs
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			placeholders := utils.GeneratePlaceholders(len(devices))
			query += " AND id_device IN (" + placeholders + ")"
			countQuery += " AND id_device IN (" + placeholders + ")"
			for _, device := range devices {
				args = append(args, device)
				countArgs = append(countArgs, device)
			}
			logrus.WithField("device_filter_applied", devices).Info("Applying device filter for multiple devices")
		}
	}

	if stageFilter != "" && stageFilter != "all" {
		if stageFilter == "No Stage" {
			query += " AND (stage IS NULL OR stage = '')"
			countQuery += " AND (stage IS NULL OR stage = '')"
		} else {
			query += " AND stage = ?"
			countQuery += " AND stage = ?"
			args = append(args, stageFilter)
			countArgs = append(countArgs, stageFilter)
		}
	}

	if statusFilter != "" && statusFilter != "all" {
		query += " AND status = ?"
		countQuery += " AND status = ?"
		args = append(args, statusFilter)
		countArgs = append(countArgs, statusFilter)
	}

	if search != "" {
		query += " AND (prospect_num LIKE ? OR nama LIKE ? OR no_fon LIKE ? OR peringkat_sekolah LIKE ? OR alamat LIKE ?)"
		countQuery += " AND (prospect_num LIKE ? OR nama LIKE ? OR no_fon LIKE ? OR peringkat_sekolah LIKE ? OR alamat LIKE ?)"
		searchParam := "%" + search + "%"
		args = append(args, searchParam, searchParam, searchParam, searchParam, searchParam)
		countArgs = append(countArgs, searchParam, searchParam, searchParam, searchParam, searchParam)
	}

	// Log the final query
	logrus.WithFields(logrus.Fields{
		"count_query": countQuery,
		"count_args":  countArgs,
	}).Debug("Executing count query")

	// Get total count
	var total int
	err := r.db.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Failed to get count")
		return nil, 0, fmt.Errorf("failed to get count: %w", err)
	}

	logrus.WithField("total_count", total).Info("Total records found")

	// Add ORDER BY and pagination
	query += " ORDER BY date_last DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	// Log the data query
	logrus.WithFields(logrus.Fields{
		"data_query": query,
		"data_args":  args,
	}).Debug("Executing data query")

	// Execute query
	rows, err := r.db.Query(query, args...)
	if err != nil {
		logrus.WithError(err).Error("Failed to query wasapBot data")
		return nil, 0, fmt.Errorf("failed to query wasapBot data: %w", err)
	}
	defer rows.Close()

	var results []map[string]interface{}
	rowCount := 0
	for rows.Next() {
		var (
			idProspect    int
			prospectNum   sql.NullString
			nama          sql.NullString
			stage         sql.NullString
			dateLast      sql.NullString
			dateStart     sql.NullString
			deviceID      sql.NullString
			niche         sql.NullString
			status        sql.NullString
			alamat        sql.NullString
			pakej         sql.NullString
			caraBayaran   sql.NullString
			tarikhGaji    sql.NullString
			currentNodeID sql.NullString
			noFon         sql.NullString
		)

		err := rows.Scan(&idProspect, &prospectNum, &nama, &stage, &dateLast, &dateStart,
			&deviceID, &niche, &status, &alamat, &pakej, &caraBayaran, &tarikhGaji,
			&currentNodeID, &noFon)

		if err != nil {
			logrus.WithError(err).Error("Failed to scan row")
			continue
		}

		rowCount++

		record := map[string]interface{}{
			"id_prospect":     idProspect,
			"prospect_num":    utils.GetStringValue(prospectNum),
			"nama":            utils.GetStringValue(nama),
			"stage":           utils.GetStringValue(stage),
			"date_last":       utils.GetStringValue(dateLast),
			"date_start":      utils.GetStringValue(dateStart),
			"id_device":       utils.GetStringValue(deviceID),
			"niche":           utils.GetStringValue(niche),
			"status":          utils.GetStringValue(status),
			"alamat":          utils.GetStringValue(alamat),
			"pakej":           utils.GetStringValue(pakej),
			"cara_bayaran":    utils.GetStringValue(caraBayaran),
			"tarikh_gaji":     utils.GetStringValue(tarikhGaji),
			"current_node_id": utils.GetStringValue(currentNodeID),
			"no_fon":          utils.GetStringValue(noFon),
		}

		results = append(results, record)

		logrus.WithFields(logrus.Fields{
			"row_id":       idProspect,
			"device":       utils.GetStringValue(deviceID),
			"prospect_num": utils.GetStringValue(prospectNum),
		}).Debug("Added record to results")
	}

	logrus.WithFields(logrus.Fields{
		"rows_scanned":  rowCount,
		"results_count": len(results),
	}).Info("Query completed")

	return results, total, nil
}

// GetWasapBotStatsWithDates retrieves WasapBot statistics with date filtering
func (r *wasapBotRepository) GetWasapBotStatsWithDates(deviceFilter, dateFrom, dateTo string, userID string) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"totalProspects":      0,
		"activeExecutions":    0,
		"completedExecutions": 0,
		"uniqueSchools":       0,
		"uniquePackages":      0,
		"totalWithPhone":      0,
		"stageBreakdown":      make(map[string]int),
	}

	baseWhere := "1=1"
	args := []interface{}{}

	// Apply date filters
	if dateFrom != "" {
		baseWhere += " AND DATE(date_start) >= ?"
		args = append(args, dateFrom)
	}

	if dateTo != "" {
		baseWhere += " AND DATE(date_start) <= ?"
		args = append(args, dateTo)
	}

	if deviceFilter != "" && deviceFilter != "all" {
		// Handle multiple device IDs
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			placeholders := utils.GeneratePlaceholders(len(devices))
			baseWhere += " AND id_device IN (" + placeholders + ")"
			for _, device := range devices {
				args = append(args, device)
			}
		}
	}

	// Total prospects
	var totalProspects int
	query := "SELECT COUNT(DISTINCT prospect_num) FROM wasapBot_nodepath WHERE " + baseWhere
	err := r.db.QueryRow(query, args...).Scan(&totalProspects)
	if err == nil {
		stats["totalProspects"] = totalProspects
	}

	// Active executions (current_node_id is NOT 'end')
	var activeExecutions int
	query = "SELECT COUNT(*) FROM wasapBot_nodepath WHERE " + baseWhere + " AND (current_node_id IS NULL OR current_node_id != 'end')"
	err = r.db.QueryRow(query, args...).Scan(&activeExecutions)
	if err == nil {
		stats["activeExecutions"] = activeExecutions
	} else {
		logrus.WithError(err).Error("Failed to get active executions count")
	}

	// Completed executions (current_node_id is 'end')
	var completedExecutions int
	query = "SELECT COUNT(*) FROM wasapBot_nodepath WHERE " + baseWhere + " AND current_node_id = 'end'"
	err = r.db.QueryRow(query, args...).Scan(&completedExecutions)
	if err == nil {
		stats["completedExecutions"] = completedExecutions
	} else {
		logrus.WithError(err).Error("Failed to get completed executions count")
	}

	// Unique schools
	var uniqueSchools int
	query = "SELECT COUNT(DISTINCT peringkat_sekolah) FROM wasapBot_nodepath WHERE " + baseWhere + " AND peringkat_sekolah IS NOT NULL AND peringkat_sekolah != ''"
	err = r.db.QueryRow(query, args...).Scan(&uniqueSchools)
	if err == nil {
		stats["uniqueSchools"] = uniqueSchools
	}

	// Unique packages
	var uniquePackages int
	query = "SELECT COUNT(DISTINCT pakej) FROM wasapBot_nodepath WHERE " + baseWhere + " AND pakej IS NOT NULL AND pakej != ''"
	err = r.db.QueryRow(query, args...).Scan(&uniquePackages)
	if err == nil {
		stats["uniquePackages"] = uniquePackages
	}

	// Total with phone
	var totalWithPhone int
	query = "SELECT COUNT(*) FROM wasapBot_nodepath WHERE " + baseWhere + " AND no_fon IS NOT NULL AND no_fon != ''"
	err = r.db.QueryRow(query, args...).Scan(&totalWithPhone)
	if err == nil {
		stats["totalWithPhone"] = totalWithPhone
	}

	// Get stage breakdown
	stageQuery := `
		SELECT 
			CASE 
				WHEN stage IS NULL OR stage = '' THEN 'No Stage' 
				ELSE stage 
			END as stage_name, 
			COUNT(*) as count 
		FROM wasapBot_nodepath 
		WHERE ` + baseWhere + ` 
		GROUP BY stage_name
	`

	rows, err := r.db.Query(stageQuery, args...)
	if err == nil {
		defer rows.Close()
		stageBreakdown := make(map[string]int)
		for rows.Next() {
			var stageName string
			var count int
			if err := rows.Scan(&stageName, &count); err == nil {
				stageBreakdown[stageName] = count
			}
		}
		stats["stageBreakdown"] = stageBreakdown
	}

	// Get daily data (prospects per day) - Fix date filtering issue
	// Build separate query for daily data to ensure we get results
	dailyBaseWhere := "1=1"
	dailyArgs := []interface{}{}

	// Apply date filters for daily query
	if dateFrom != "" {
		dailyBaseWhere += " AND DATE(date_start) >= ?"
		dailyArgs = append(dailyArgs, dateFrom)
	}

	if dateTo != "" {
		dailyBaseWhere += " AND DATE(date_start) <= ?"
		dailyArgs = append(dailyArgs, dateTo)
	}

	if deviceFilter != "" && deviceFilter != "all" {
		// Handle multiple device IDs for daily query
		devices := utils.SplitAndTrim(deviceFilter, ",")
		if len(devices) > 0 {
			placeholders := utils.GeneratePlaceholders(len(devices))
			dailyBaseWhere += " AND id_device IN (" + placeholders + ")"
			for _, device := range devices {
				dailyArgs = append(dailyArgs, device)
			}
		}
	}

	dailyQuery := `
		SELECT 
			DATE_FORMAT(DATE(date_start), '%Y-%m-%d') as date,
			COUNT(DISTINCT prospect_num) as prospects
		FROM wasapBot_nodepath 
		WHERE ` + dailyBaseWhere + `
		  AND date_start IS NOT NULL 
		  AND date_start != ''
		  AND DATE(date_start) IS NOT NULL
		GROUP BY DATE_FORMAT(DATE(date_start), '%Y-%m-%d')
		ORDER BY DATE_FORMAT(DATE(date_start), '%Y-%m-%d')
	`

	logrus.WithFields(logrus.Fields{
		"dailyQuery":     dailyQuery,
		"dailyArgs":      dailyArgs,
		"dateFrom":       dateFrom,
		"dateTo":         dateTo,
		"deviceFilter":   deviceFilter,
		"dailyBaseWhere": dailyBaseWhere,
	}).Info("Executing WasapBot daily_data query with separate args")

	dailyRows, err := r.db.Query(dailyQuery, dailyArgs...)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"dailyQuery": dailyQuery,
			"dailyArgs":  dailyArgs,
		}).Error("Failed to query WasapBot daily_data")
		stats["daily_data"] = []map[string]interface{}{} // Return empty array on error
	} else {
		defer dailyRows.Close()
		dailyData := []map[string]interface{}{}
		rowCount := 0
		for dailyRows.Next() {
			var date string
			var prospects int
			if err := dailyRows.Scan(&date, &prospects); err == nil {
				rowCount++
				// Clean date format - ensure only YYYY-MM-DD format without any timestamp
				cleanDate := date
				if len(cleanDate) > 10 {
					cleanDate = cleanDate[:10] // Take only YYYY-MM-DD part
				}

				dailyData = append(dailyData, map[string]interface{}{
					"date":      cleanDate, // Simple date format: YYYY-MM-DD
					"prospects": prospects,
				})
				logrus.WithFields(logrus.Fields{
					"date":      date,
					"prospects": prospects,
				}).Debug("WasapBot daily data row processed")
			} else {
				logrus.WithError(err).Error("Failed to scan WasapBot daily row")
			}
		}
		stats["daily_data"] = dailyData
		logrus.WithFields(logrus.Fields{
			"daily_data_count": len(dailyData),
			"rows_processed":   rowCount,
		}).Info("WasapBot daily_data query completed")

		// Clean completion - debug logging removed as issue is resolved
	}

	return stats, nil
}
