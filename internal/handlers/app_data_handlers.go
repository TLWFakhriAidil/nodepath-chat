package handlers

import (
	"database/sql"
	"nodepath-chat/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

// AppDataHandlers handles consolidated application data operations
type AppDataHandlers struct {
	db *sql.DB
}

// NewAppDataHandlers creates a new instance of AppDataHandlers
func NewAppDataHandlers(db *sql.DB) *AppDataHandlers {
	return &AppDataHandlers{
		db: db,
	}
}

// AppDataResponse represents the consolidated response structure
type AppDataResponse struct {
	User        models.User `json:"user"`
	HasDevices  bool        `json:"has_devices"`
	DeviceCount int         `json:"device_count"`
	DeviceIDs   []string    `json:"device_ids"`
}

// GetAppData returns consolidated user profile and device status in a single optimized query
// This endpoint replaces multiple separate calls to /api/profile/ and /api/auth/device-status
func (adh *AppDataHandlers) GetAppData(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	// Check if database connection is available
	if adh.db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Database service is not available",
		})
	}

	userIDStr := userID.(string)

	// OPTIMIZED QUERY: Get user data and device count in a single LEFT JOIN query
	// This eliminates the need for separate queries to user_nodepath and device_setting_nodepath
	query := `
		SELECT 
			u.id, u.email, u.full_name, u.gmail, u.phone, u.status, u.expired, 
			u.is_active, u.created_at, u.updated_at, u.last_login,
			COUNT(CASE WHEN d.id_device IS NOT NULL AND d.id_device != '' THEN 1 END) as device_count,
			GROUP_CONCAT(CASE WHEN d.id_device IS NOT NULL AND d.id_device != '' THEN d.id_device END) as device_ids_concat
		FROM user_nodepath u
		LEFT JOIN device_setting_nodepath d ON u.id = d.user_id
		WHERE u.id = ?
		GROUP BY u.id, u.email, u.full_name, u.gmail, u.phone, u.status, u.expired, 
		         u.is_active, u.created_at, u.updated_at, u.last_login
	`

	var user models.User
	var deviceCount int
	var deviceIDsConcat sql.NullString

	err := adh.db.QueryRow(query, userIDStr).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Gmail, &user.Phone,
		&user.Status, &user.Expired, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
		&deviceCount, &deviceIDsConcat,
	)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	} else if err != nil {
		logrus.WithError(err).WithField("userID", userIDStr).Error("Failed to fetch consolidated app data")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch app data",
		})
	}

	// Parse device IDs from concatenated string
	var deviceIDs []string
	if deviceIDsConcat.Valid && deviceIDsConcat.String != "" {
		// Split the concatenated device IDs (GROUP_CONCAT result)
		deviceIDsStr := deviceIDsConcat.String
		if deviceIDsStr != "" {
			// Simple split by comma for GROUP_CONCAT result
			deviceIDs = splitString(deviceIDsStr, ",")
		}
	}

	// Construct response
	response := AppDataResponse{
		User:        user,
		HasDevices:  deviceCount > 0,
		DeviceCount: deviceCount,
		DeviceIDs:   deviceIDs,
	}

	logrus.WithFields(logrus.Fields{
		"userID":       userIDStr,
		"device_count": deviceCount,
		"has_devices":  deviceCount > 0,
		"query_type":   "optimized_join",
	}).Info("Consolidated app data fetched successfully")

	return c.JSON(fiber.Map{
		"success": true,
		"data":    response,
	})
}

// GetAppDataStatus returns just the status information (lighter version)
// Useful for health checks or quick status updates
func (adh *AppDataHandlers) GetAppDataStatus(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	userIDStr := userID.(string)

	// Lightweight query for just status and device count
	query := `
		SELECT 
			u.status, u.expired, u.is_active,
			COUNT(CASE WHEN d.id_device IS NOT NULL AND d.id_device != '' THEN 1 END) as device_count
		FROM user_nodepath u
		LEFT JOIN device_setting_nodepath d ON u.id = d.user_id
		WHERE u.id = ?
		GROUP BY u.status, u.expired, u.is_active
	`

	var status string
	var expired *string
	var isActive bool
	var deviceCount int

	err := adh.db.QueryRow(query, userIDStr).Scan(&status, &expired, &isActive, &deviceCount)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	} else if err != nil {
		logrus.WithError(err).WithField("userID", userIDStr).Error("Failed to fetch app status")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch app status",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"status":       status,
			"expired":      expired,
			"is_active":    isActive,
			"has_devices":  deviceCount > 0,
			"device_count": deviceCount,
		},
	})
}

// SetupAppDataRoutes configures the optimized app data routes
func (adh *AppDataHandlers) SetupAppDataRoutes(api fiber.Router) {
	app := api.Group("/app")

	// Consolidated data endpoint - replaces /api/profile/ + /api/auth/device-status
	app.Get("/data", adh.GetAppData)

	// Lightweight status endpoint for quick checks
	app.Get("/status", adh.GetAppDataStatus)
}

// splitString is a simple string splitter utility
func splitString(s, sep string) []string {
	if s == "" {
		return []string{}
	}

	result := []string{}
	start := 0

	for i := 0; i < len(s); i++ {
		if i == len(s)-1 || (i < len(s)-len(sep) && s[i:i+len(sep)] == sep) {
			end := i
			if i == len(s)-1 {
				end = i + 1
			}

			if start < end {
				result = append(result, s[start:end])
			}

			if i < len(s)-len(sep) {
				start = i + len(sep)
				i += len(sep) - 1
			}
		}
	}

	return result
}
