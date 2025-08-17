// GetSequenceStepLeads gets lead details for a specific step in a sequence on a device
func (handler *App) GetSequenceStepLeads(c *fiber.Ctx) error {
	sequenceId := c.Params("id") // Already a string UUID

	deviceId := c.Params("deviceId")
	stepId := c.Params("stepId")
	status := c.Query("status", "all")

	// Get date filters from query params
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Get session from cookie
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token",
		})
	}

	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Invalid session",
		})
	}

	// Get leads from broadcast messages for this specific step
	db := database.GetDB()

	query := `
		SELECT bm.recipient_phone, bm.status, bm.sent_at, l.name, bm.error_message
		FROM broadcast_messages bm
		LEFT JOIN leads l ON l.phone = bm.recipient_phone AND l.user_id = bm.user_id
		WHERE bm.sequence_id = ?
		AND bm.device_id = ?
		AND bm.sequence_stepid = ?
		AND bm.user_id = ?
	`

	// Use sequence ID directly as string
	args := []interface{}{sequenceId, deviceId, stepId, session.UserID}

	log.Printf("GetSequenceStepLeads - Sequence: %s, Device: %s, Step: %s, Status: %s, DateRange: %s to %s",
		sequenceId, deviceId, stepId, status, startDate, endDate)

	// Add status filter if not "all"
	if status != "all" {
		if status == "success" {
			query += ` AND bm.status IN ('sent', 'delivered', 'success')`
		} else if status == "pending" {
			query += ` AND bm.status IN ('pending', 'queued')`
		} else if status == "failed" {
			query += ` AND bm.status IN ('failed', 'error')`
		}
	}
	
	// Add date filter if provided
	if startDate != "" && endDate != "" {
		query += ` AND DATE(bm.sent_at) BETWEEN ? AND ?`
		args = append(args, startDate, endDate)
	} else if startDate != "" {
		query += ` AND DATE(bm.sent_at) >= ?`
		args = append(args, startDate)
	} else if endDate != "" {
		query += ` AND DATE(bm.sent_at) <= ?`
		args = append(args, endDate)
	}
	
	query += ` ORDER BY bm.sent_at DESC`
	
	rows, err := db.Query(query, args...)
	if err != nil {
		log.Printf("Error executing sequence step lead details query: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get lead details",
		})
	}
	defer rows.Close()
	
	leads := []map[string]interface{}{}
	
	for rows.Next() {
		var phone, msgStatus string
		var sentAt sql.NullTime
		var name sql.NullString
		var errorMessage sql.NullString
		
		err := rows.Scan(&phone, &msgStatus, &sentAt, &name, &errorMessage)
		if err != nil {
			log.Printf("Error scanning lead row: %v", err)
			continue
		}
		
		leadName := "Unknown"
		if name.Valid && name.String != "" {
			leadName = name.String
		}
		
		lead := map[string]interface{}{
			"name":   leadName,
			"phone":  phone,
			"status": msgStatus,
		}
		
		// Add error message if exists
		if errorMessage.Valid && errorMessage.String != "" {
			lead["error_message"] = errorMessage.String
		} else {
			lead["error_message"] = "-"
		}
		
		if sentAt.Valid {
			lead["sent_at"] = sentAt.Time.Format("2006-01-02 03:04 PM")
		} else {
			lead["sent_at"] = "-"
		}
		
		leads = append(leads, lead)
	}
	
	log.Printf("GetSequenceStepLeads - Found %d leads for sequence %s, device %s, step %s, status %s",
		len(leads), sequenceId, deviceId, stepId, status)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Lead details retrieved successfully",
		Results: leads,
	})
}
