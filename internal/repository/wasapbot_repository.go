package repository

import (
	"database/sql"
	"fmt"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/utils"

	"github.com/sirupsen/logrus"
)

// WasapBotRepository interface for wasapBot_nodepath operations
type WasapBotRepository interface {
	GetByProspectAndDevice(prospectNum, instance string) (*models.WasapBot, error)
	GetActiveExecution(prospectNum, instance string) (*models.WasapBot, error)
	GetByExecutionID(executionID string) (*models.WasapBot, error)
	Create(wasapBot *models.WasapBot) error
	Update(wasapBot *models.WasapBot) error
	UpdateExecutionStatus(executionID, status string) error
	UpdateCurrentNode(executionID, nodeID string) error
	UpdateWaitingStatus(executionID string, waitingValue int) error
	SaveConversationHistory(prospectNum, instance, userMessage, botResponse, stage, nama string) error
	GetAllWasapBotData(limit, offset int, deviceFilter, stageFilter, statusFilter, search string, userID int) ([]map[string]interface{}, int, error)
	GetWasapBotStats(deviceFilter string, userID int) (map[string]interface{}, error)
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

// GetByProspectAndDevice retrieves a wasapBot record by prospect number and instance
func (r *wasapBotRepository) GetByProspectAndDevice(prospectNum, instance string) (*models.WasapBot, error) {
	query := `
		SELECT id_prospect, flow_reference, execution_id, execution_status, flow_id,
		       current_node_id, last_node_id, waiting_for_reply, id_device,
		       prospect_num, niche, instance, peringkat_sekolah, alamat, nama,
		       pakej, no_fon, cara_bayaran, tarikh_gaji, stage, temp_stage,
		       conv_start, conv_last, date_start, date_last, status, staff_cls,
		       umur, kerja, sijil, user_input, alasan, nota
		FROM wasapBot_nodepath
		WHERE prospect_num = ? AND instance = ?
		LIMIT 1
	`

	var wb models.WasapBot
	err := r.db.QueryRow(query, prospectNum, instance).Scan(
		&wb.IDProspect, &wb.FlowReference, &wb.ExecutionID, &wb.ExecutionStatus,
		&wb.FlowID, &wb.CurrentNodeID, &wb.LastNodeID, &wb.WaitingForReply,
		&wb.IDDevice, &wb.ProspectNum, &wb.Niche, &wb.Instance,
		&wb.PeringkatSekolah, &wb.Alamat, &wb.Nama, &wb.Pakej,
		&wb.NoFon, &wb.CaraBayaran, &wb.TarikhGaji, &wb.Stage,
		&wb.TempStage, &wb.ConvStart, &wb.ConvLast, &wb.DateStart,
		&wb.DateLast, &wb.Status, &wb.StaffCls, &wb.Umur,
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

// GetActiveExecution retrieves an active execution for a prospect and instance
func (r *wasapBotRepository) GetActiveExecution(prospectNum, instance string) (*models.WasapBot, error) {
	query := `
		SELECT id_prospect, flow_reference, execution_id, execution_status, flow_id,
		       current_node_id, last_node_id, waiting_for_reply, id_device,
		       prospect_num, niche, instance, peringkat_sekolah, alamat, nama,
		       pakej, no_fon, cara_bayaran, tarikh_gaji, stage, temp_stage,
		       conv_start, conv_last, date_start, date_last, status, staff_cls,
		       umur, kerja, sijil, user_input, alasan, nota
		FROM wasapBot_nodepath
		WHERE prospect_num = ? AND instance = ? AND execution_status = 'active'
		LIMIT 1
	`

	var wb models.WasapBot
	err := r.db.QueryRow(query, prospectNum, instance).Scan(
		&wb.IDProspect, &wb.FlowReference, &wb.ExecutionID, &wb.ExecutionStatus,
		&wb.FlowID, &wb.CurrentNodeID, &wb.LastNodeID, &wb.WaitingForReply,
		&wb.IDDevice, &wb.ProspectNum, &wb.Niche, &wb.Instance,
		&wb.PeringkatSekolah, &wb.Alamat, &wb.Nama, &wb.Pakej,
		&wb.NoFon, &wb.CaraBayaran, &wb.TarikhGaji, &wb.Stage,
		&wb.TempStage, &wb.ConvStart, &wb.ConvLast, &wb.DateStart,
		&wb.DateLast, &wb.Status, &wb.StaffCls, &wb.Umur,
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
		       prospect_num, niche, instance, peringkat_sekolah, alamat, nama,
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
		&wb.IDDevice, &wb.ProspectNum, &wb.Niche, &wb.Instance,
		&wb.PeringkatSekolah, &wb.Alamat, &wb.Nama, &wb.Pakej,
		&wb.NoFon, &wb.CaraBayaran, &wb.TarikhGaji, &wb.Stage,
		&wb.TempStage, &wb.ConvStart, &wb.ConvLast, &wb.DateStart,
		&wb.DateLast, &wb.Status, &wb.StaffCls, &wb.Umur,
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
			prospect_num, niche, instance, peringkat_sekolah, alamat, nama,
			pakej, no_fon, cara_bayaran, tarikh_gaji, stage, temp_stage,
			conv_start, conv_last, date_start, date_last, status, staff_cls,
			umur, kerja, sijil, user_input, alasan, nota
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.Exec(query,
		wasapBot.FlowReference, wasapBot.ExecutionID, wasapBot.ExecutionStatus,
		wasapBot.FlowID, wasapBot.CurrentNodeID, wasapBot.LastNodeID,
		wasapBot.WaitingForReply, wasapBot.IDDevice, wasapBot.ProspectNum,
		wasapBot.Niche, wasapBot.Instance, wasapBot.PeringkatSekolah,
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
			prospect_num = ?, niche = ?, instance = ?, peringkat_sekolah = ?, alamat = ?,
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
		wasapBot.Niche, wasapBot.Instance, wasapBot.PeringkatSekolah,
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
func (r *wasapBotRepository) SaveConversationHistory(prospectNum, instance, userMessage, botResponse, stage, nama string) error {
	return utils.WithTransaction(r.db, func(tx *sql.Tx) error {
		// Check if record exists
		var existingID *int
		var existingConvLast sql.NullString
		checkQuery := `
			SELECT id_prospect, conv_last 
			FROM wasapBot_nodepath 
			WHERE prospect_num = ? AND instance = ?
			FOR UPDATE
		`
		err := tx.QueryRow(checkQuery, prospectNum, instance).Scan(&existingID, &existingConvLast)
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
				WHERE prospect_num = ? AND instance = ?
			`
			_, err = tx.Exec(updateQuery, convLastValue, stage, nama, now, prospectNum, instance)
			if err != nil {
				return fmt.Errorf("failed to update conversation history: %w", err)
			}
			logrus.WithFields(logrus.Fields{
				"prospect_num": prospectNum,
				"instance": instance,
			}).Info("WasapBot conversation history updated successfully")
		} else {
			// Create new record
			insertQuery := `
				INSERT INTO wasapBot_nodepath (
					prospect_num, instance, stage, conv_last, nama, 
					date_start, date_last, status, waiting_for_reply
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`
			_, err = tx.Exec(insertQuery, prospectNum, instance, stage, convLastValue, nama, 
				now, now, "Prospek", 0)
			if err != nil {
				return fmt.Errorf("failed to create new conversation record: %w", err)
			}
			logrus.WithFields(logrus.Fields{
				"prospect_num": prospectNum,
				"instance": instance,
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
			"execution_id":   executionID,
			"waiting_value": waitingValue,
		}).Error("Failed to update waiting status in wasapBot")
		return fmt.Errorf("failed to update waiting status: %w", err)
	}
	
	return nil
}


// GetAllWasapBotData retrieves all WasapBot data with filters
func (r *wasapBotRepository) GetAllWasapBotData(limit, offset int, deviceFilter, stageFilter, statusFilter, search string, userID int) ([]map[string]interface{}, int, error) {
	// Log incoming parameters
	logrus.WithFields(logrus.Fields{
		"limit": limit,
		"offset": offset,
		"deviceFilter": deviceFilter,
		"stageFilter": stageFilter,
		"statusFilter": statusFilter,
		"search": search,
		"userID": userID,
	}).Info("GetAllWasapBotData called")
	
	// Build query with filters
	query := `
		SELECT id_prospect, prospect_num, nama, no_fon, peringkat_sekolah, 
		       pakej, stage, status, date_last, instance
		FROM wasapBot_nodepath
		WHERE 1=1
	`
	
	countQuery := `SELECT COUNT(*) FROM wasapBot_nodepath WHERE 1=1`
	args := []interface{}{}
	countArgs := []interface{}{}
	
	// Apply filters
	if deviceFilter != "" && deviceFilter != "all" {
		query += " AND instance = ?"
		countQuery += " AND instance = ?"
		args = append(args, deviceFilter)
		countArgs = append(countArgs, deviceFilter)
		logrus.WithField("device_filter_applied", deviceFilter).Info("Applying device filter")
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
		"count_args": countArgs,
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
		"data_args": args,
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
			idProspect int
			prospectNum string
			nama sql.NullString
			noFon sql.NullString
			school sql.NullString
			pakej sql.NullString
			stage sql.NullString
			status sql.NullString
			dateLast sql.NullTime
			instance string
		)
		
		err := rows.Scan(
			&idProspect,
			&prospectNum,
			&nama,
			&noFon,
			&school,
			&pakej,
			&stage,
			&status,
			&dateLast,
			&instance,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan wasapBot row")
			continue
		}
		
		rowCount++
		
		// Convert to plain map for JSON
		record := map[string]interface{}{
			"id": idProspect,
			"name": "",
			"phone": prospectNum,
			"school": "",
			"package": "",
			"stage": "",
			"status": "",
			"payment": "",
			"lastUpdated": "",
			"instance": instance,
		}
		
		// Handle null values properly
		if nama.Valid {
			record["name"] = nama.String
		}
		if noFon.Valid {
			record["phone"] = noFon.String
		} else {
			record["phone"] = prospectNum // Use prospect_num if no_fon is null
		}
		if school.Valid {
			record["school"] = school.String
		}
		if pakej.Valid {
			record["package"] = pakej.String
		}
		if stage.Valid {
			record["stage"] = stage.String
		}
		if status.Valid {
			record["status"] = status.String
		}
		if dateLast.Valid {
			record["lastUpdated"] = dateLast.Time.Format("2006-01-02 15:04:05")
		}
		
		results = append(results, record)
		
		logrus.WithFields(logrus.Fields{
			"row_id": idProspect,
			"instance": instance,
			"prospect_num": prospectNum,
		}).Debug("Added record to results")
	}
	
	logrus.WithFields(logrus.Fields{
		"rows_scanned": rowCount,
		"results_count": len(results),
	}).Info("Query completed")
	
	return results, total, nil
}

// GetWasapBotStats retrieves WasapBot statistics
func (r *wasapBotRepository) GetWasapBotStats(deviceFilter string, userID int) (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"totalProspects": 0,
		"activeExecutions": 0,
		"completedExecutions": 0,
		"uniqueSchools": 0,
		"uniquePackages": 0,
		"totalWithPhone": 0,
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
