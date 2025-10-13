package handlers

import (
	"database/sql"
	"nodepath-chat/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// ProfileHandlers handles user profile operations
type ProfileHandlers struct {
	db *sql.DB
}

// NewProfileHandlers creates a new instance of ProfileHandlers
func NewProfileHandlers(db *sql.DB) *ProfileHandlers {
	return &ProfileHandlers{
		db: db,
	}
}

// GetProfile returns the current user's profile information
func (ph *ProfileHandlers) GetProfile(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	// Check if database connection is available
	if ph.db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Database service is not available",
		})
	}

	// Fetch user from database
	var user models.User
	err := ph.db.QueryRow(`
		SELECT id, email, full_name, gmail, phone, status, expired, is_active, created_at, updated_at, last_login 
		FROM user_nodepath 
		WHERE id = ?
	`, userID).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Gmail, &user.Phone,
		&user.Status, &user.Expired, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
	)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	} else if err != nil {
		logrus.WithError(err).Error("Failed to fetch user profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch profile",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

// UpdateProfile updates the current user's profile information
func (ph *ProfileHandlers) UpdateProfile(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	// Check if database connection is available
	if ph.db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Database service is not available",
		})
	}

	var req models.UserProfileUpdate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request format",
		})
	}

	// Start building update query
	updateFields := []string{}
	args := []interface{}{}

	if req.FullName != "" {
		updateFields = append(updateFields, "full_name = ?")
		args = append(args, req.FullName)
	}

	if req.Gmail != nil {
		updateFields = append(updateFields, "gmail = ?")
		args = append(args, req.Gmail)
	}

	if req.Phone != nil {
		updateFields = append(updateFields, "phone = ?")
		args = append(args, req.Phone)
	}

	// Handle password update if provided
	if req.Password != nil && req.NewPassword != nil {
		// Verify current password first
		var currentHashedPassword string
		err := ph.db.QueryRow("SELECT password FROM user_nodepath WHERE id = ?", userID).Scan(&currentHashedPassword)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to verify current password",
			})
		}

		// Check if current password is correct
		err = bcrypt.CompareHashAndPassword([]byte(currentHashedPassword), []byte(*req.Password))
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"error":   "Current password is incorrect",
			})
		}

		// Hash new password
		hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(*req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Failed to process new password",
			})
		}

		updateFields = append(updateFields, "password = ?")
		args = append(args, string(hashedNewPassword))
	}

	// If no fields to update
	if len(updateFields) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "No fields to update",
		})
	}

	// Add updated_at field
	updateFields = append(updateFields, "updated_at = NOW()")
	args = append(args, userID) // for WHERE clause

	// Execute update
	query := "UPDATE user_nodepath SET " + joinStrings(updateFields, ", ") + " WHERE id = ?"
	_, err := ph.db.Exec(query, args...)
	if err != nil {
		logrus.WithError(err).Error("Failed to update user profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to update profile",
		})
	}

	// Fetch updated user data
	var user models.User
	err = ph.db.QueryRow(`
		SELECT id, email, full_name, gmail, phone, status, expired, is_active, created_at, updated_at, last_login 
		FROM user_nodepath 
		WHERE id = ?
	`, userID).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Gmail, &user.Phone,
		&user.Status, &user.Expired, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
	)

	if err != nil {
		logrus.WithError(err).Error("Failed to fetch updated user profile")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Profile updated but failed to fetch updated data",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Profile updated successfully",
		"data":    user,
	})
}

// GetUserStatus returns the current user's status and expiry information
func (ph *ProfileHandlers) GetUserStatus(c *fiber.Ctx) error {
	// Get user ID from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	// Check if database connection is available
	if ph.db == nil {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Database service is not available",
		})
	}

	// Fetch user status and expiry
	var status string
	var expired *string
	err := ph.db.QueryRow(`
		SELECT status, expired 
		FROM user_nodepath 
		WHERE id = ?
	`, userID).Scan(&status, &expired)

	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"success": false,
			"error":   "User not found",
		})
	} else if err != nil {
		logrus.WithError(err).Error("Failed to fetch user status")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch user status",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"status":  status,
			"expired": expired,
		},
	})
}

// Helper function to join strings
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
