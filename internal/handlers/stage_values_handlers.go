package handlers

import (
	"database/sql"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// StageSetValue represents the stage value configuration
type StageSetValue struct {
	StageSetValueID int            `json:"stageSetValue_id"`
	IDDevice        string         `json:"id_device"`
	Stage           int            `json:"stage"`
	TypeInputData   string         `json:"type_inputData"`
	ColumnsData     string         `json:"columnsData"`
	InputHardCode   sql.NullString `json:"inputHardCode"`
}

// GetStageValuesByDevice gets all stage values for a specific device
func (h *Handlers) GetStageValuesByDevice(c *fiber.Ctx) error {
	deviceID := c.Params("deviceId")
	if deviceID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Device ID is required",
		})
	}

	// Check if database is available
	if h.db == nil {
		logrus.Warn("Database not available for stage values")
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "Database service unavailable",
		})
	}

	query := `
		SELECT stageSetValue_id, id_device, stage, type_inputData, columnsData, inputHardCode
		FROM stageSetValue_nodepath
		WHERE id_device = ?
		ORDER BY stage ASC
	`

	rows, err := h.db.Query(query, deviceID)
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
		err := rows.Scan(&sv.StageSetValueID, &sv.IDDevice, &sv.Stage, &sv.TypeInputData, &sv.ColumnsData, &sv.InputHardCode)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan stage value")
			continue
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
	var req struct {
		IDDevice      string  `json:"id_device"`
		Stage         int     `json:"stage"`
		TypeInputData string  `json:"type_inputData"`
		ColumnsData   string  `json:"columnsData"`
		InputHardCode *string `json:"inputHardCode"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate required fields
	if req.IDDevice == "" || req.Stage == 0 || req.TypeInputData == "" || req.ColumnsData == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing required fields",
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
			stage INT,
			type_inputData VARCHAR(255),
			columnsData VARCHAR(255),
			inputHardCode VARCHAR(255),
			INDEX idx_device (id_device)
		)
	`
	_, err := h.db.Exec(createTableQuery)
	if err != nil {
		logrus.WithError(err).Error("Failed to create stage values table")
		// Continue anyway, table might already exist
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

// DeleteStageValue deletes a stage value configuration
func (h *Handlers) DeleteStageValue(c *fiber.Ctx) error {
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
