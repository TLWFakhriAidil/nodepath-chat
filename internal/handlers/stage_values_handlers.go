package handlers

import (
	"database/sql"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// StageSetValue represents the stage value configuration
type StageSetValue struct {
	StageSetValueID int     `json:"stageSetValue_id"`
	IDDevice        string  `json:"id_device"`
	Stage           string  `json:"stage"`
	TypeInputData   string  `json:"type_inputData"`
	ColumnsData     string  `json:"columnsData"`
	InputHardCode   *string `json:"inputHardCode"` // Changed to *string to properly handle null
}

// GetAllStageValues gets all stage values for authenticated user's devices
func (h *Handlers) GetAllStageValues(c *fiber.Ctx) error {
	// Get user ID from auth context
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	// Check if database is available
	if h.db == nil {
		logrus.Warn("Database not available for stage values")
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Database service unavailable",
		})
	}

	// First, get all devices for this user
	deviceQuery := `
		SELECT id_device FROM device_setting_nodepath WHERE user_id = ?
	`
	deviceRows, err := h.db.Query(deviceQuery, userIDStr)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch user devices")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch user devices",
		})
	}
	defer deviceRows.Close()

	var deviceIDs []string
	for deviceRows.Next() {
		var deviceID string
		if err := deviceRows.Scan(&deviceID); err != nil {
			continue
		}
		deviceIDs = append(deviceIDs, deviceID)
	}

	if len(deviceIDs) == 0 {
		// Return empty array if user has no devices
		return c.JSON([]StageSetValue{})
	}

	// Build query for stage values
	query := `
		SELECT stageSetValue_id, id_device, stage, type_inputData, columnsData, inputHardCode
		FROM stageSetValue_nodepath
		WHERE id_device IN (`

	// Add placeholders for IN clause
	args := make([]interface{}, len(deviceIDs))
	for i, id := range deviceIDs {
		if i > 0 {
			query += ","
		}
		query += "?"
		args[i] = id
	}
	query += ") ORDER BY id_device, stage ASC"

	rows, err := h.db.Query(query, args...)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch stage values")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch stage values",
		})
	}
	defer rows.Close()

	var stageValues []StageSetValue
	for rows.Next() {
		var sv StageSetValue
		var inputHardCode sql.NullString
		err := rows.Scan(&sv.StageSetValueID, &sv.IDDevice, &sv.Stage, &sv.TypeInputData, &sv.ColumnsData, &inputHardCode)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan stage value")
			continue
		}
		// Convert sql.NullString to *string
		if inputHardCode.Valid {
			sv.InputHardCode = &inputHardCode.String
		} else {
			sv.InputHardCode = nil
		}
		stageValues = append(stageValues, sv)
	}

	// Return empty array if no values found
	if stageValues == nil {
		stageValues = []StageSetValue{}
	}

	return c.JSON(stageValues)
}

// CreateStageValue creates a new stage value configuration
func (h *Handlers) CreateStageValue(c *fiber.Ctx) error {
	// Get user ID from auth context
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	var req struct {
		IDDevice      string  `json:"id_device"`
		Stage         string  `json:"stage"` // Changed from int to string
		TypeInputData string  `json:"type_inputData"`
		ColumnsData   string  `json:"columnsData"`
		InputHardCode *string `json:"inputHardCode"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get device from context if available
	deviceID := c.Locals("selectedDevice")
	if deviceID != nil {
		req.IDDevice = deviceID.(string)
	}

	// If no device ID provided, get first device for user
	if req.IDDevice == "" {
		var firstDevice string
		err := h.db.QueryRow("SELECT id_device FROM device_setting_nodepath WHERE user_id = ? LIMIT 1", userIDStr).Scan(&firstDevice)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No device found for user",
			})
		}
		req.IDDevice = firstDevice
	}

	// Validate required fields
	if req.Stage == "" || req.TypeInputData == "" || req.ColumnsData == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
		})
	}

	// Verify user owns the device
	var count int
	err := h.db.QueryRow("SELECT COUNT(*) FROM device_setting_nodepath WHERE id_device = ? AND user_id = ?", req.IDDevice, userIDStr).Scan(&count)
	if err != nil || count == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to modify this device",
		})
	}

	// Check if database is available
	if h.db == nil {
		logrus.Warn("Database not available for stage values")
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Database service unavailable",
		})
	}

	// Create the table if it doesn't exist
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS stageSetValue_nodepath (
			stageSetValue_id INT AUTO_INCREMENT PRIMARY KEY,
			id_device VARCHAR(255),
			stage VARCHAR(255),
			type_inputData VARCHAR(255),
			columnsData VARCHAR(255),
			inputHardCode VARCHAR(255),
			INDEX idx_device (id_device),
			INDEX idx_stage (stage)
		)
	`
	_, err = h.db.Exec(createTableQuery)
	if err != nil {
		logrus.WithError(err).Error("Failed to create stage values table")
		// Continue anyway, table might already exist
	}

	// Try to alter existing table if stage column is INT
	alterTableQuery := `
		ALTER TABLE stageSetValue_nodepath 
		MODIFY COLUMN stage VARCHAR(255)
	`
	_, err = h.db.Exec(alterTableQuery)
	if err != nil {
		// This is expected if column is already VARCHAR or table doesn't exist
		logrus.Debug("Stage column might already be VARCHAR or table doesn't exist")
	}

	// Insert the new stage value
	insertQuery := `
		INSERT INTO stageSetValue_nodepath (id_device, stage, type_inputData, columnsData, inputHardCode)
		VALUES (?, ?, ?, ?, ?)
	`

	result, err := h.db.Exec(insertQuery, req.IDDevice, req.Stage, req.TypeInputData, req.ColumnsData, req.InputHardCode)
	if err != nil {
		logrus.WithError(err).Error("Failed to insert stage value")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create stage value",
		})
	}

	lastID, _ := result.LastInsertId()

	return c.JSON(fiber.Map{
		"message": "Stage value created successfully",
		"id":      lastID,
	})
}

// UpdateStageValue updates an existing stage value configuration
func (h *Handlers) UpdateStageValue(c *fiber.Ctx) error {
	// Get user ID from auth context
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Stage value ID is required",
		})
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid stage value ID",
		})
	}

	var req struct {
		Stage         string  `json:"stage"` // Changed from int to string
		TypeInputData string  `json:"type_inputData"`
		ColumnsData   string  `json:"columnsData"`
		InputHardCode *string `json:"inputHardCode"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Check if database is available
	if h.db == nil {
		logrus.Warn("Database not available for stage values")
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Database service unavailable",
		})
	}

	// First check if the stage value exists and get its device ID
	var deviceID string
	err = h.db.QueryRow("SELECT id_device FROM stageSetValue_nodepath WHERE stageSetValue_id = ?", id).Scan(&deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Stage value not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch stage value",
		})
	}

	// Verify user owns the device
	var count int
	err = h.db.QueryRow("SELECT COUNT(*) FROM device_setting_nodepath WHERE id_device = ? AND user_id = ?", deviceID, userIDStr).Scan(&count)
	if err != nil || count == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to modify this stage value",
		})
	}

	// Update the stage value
	updateQuery := `
		UPDATE stageSetValue_nodepath 
		SET stage = ?, type_inputData = ?, columnsData = ?, inputHardCode = ?
		WHERE stageSetValue_id = ?
	`
	result, err := h.db.Exec(updateQuery, req.Stage, req.TypeInputData, req.ColumnsData, req.InputHardCode, id)
	if err != nil {
		logrus.WithError(err).Error("Failed to update stage value")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update stage value",
		})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Stage value not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Stage value updated successfully",
	})
}

// DeleteStageValue deletes a stage value configuration
func (h *Handlers) DeleteStageValue(c *fiber.Ctx) error {
	// Get user ID from auth context
	userIDStr := c.Locals("user_id")
	if userIDStr == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not authenticated",
		})
	}

	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Stage value ID is required",
		})
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid stage value ID",
		})
	}

	// Check if database is available
	if h.db == nil {
		logrus.Warn("Database not available for stage values")
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Database service unavailable",
		})
	}

	// First check if the stage value exists and get its device ID
	var deviceID string
	err = h.db.QueryRow("SELECT id_device FROM stageSetValue_nodepath WHERE stageSetValue_id = ?", id).Scan(&deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "Stage value not found",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch stage value",
		})
	}

	// Verify user owns the device
	var count int
	err = h.db.QueryRow("SELECT COUNT(*) FROM device_setting_nodepath WHERE id_device = ? AND user_id = ?", deviceID, userIDStr).Scan(&count)
	if err != nil || count == 0 {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "You don't have permission to delete this stage value",
		})
	}

	deleteQuery := `DELETE FROM stageSetValue_nodepath WHERE stageSetValue_id = ?`
	result, err := h.db.Exec(deleteQuery, id)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete stage value")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete stage value",
		})
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Stage value not found",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Stage value deleted successfully",
	})
}
