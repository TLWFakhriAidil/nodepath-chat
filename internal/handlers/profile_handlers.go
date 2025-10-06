package handlers

import (
	"database/sql"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"nodepath-chat/internal/models"
)

type ProfileHandlers struct {
	db *sql.DB
}

func NewProfileHandlers(db *sql.DB) *ProfileHandlers {
	return &ProfileHandlers{db: db}
}

// GetProfile retrieves the current user's profile
func (h *ProfileHandlers) GetProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// First, try the query with new columns (after migration)
	query := `
		SELECT id, email, full_name, 
		       COALESCE(gmail, '') as gmail, 
		       COALESCE(phone, '') as phone, 
		       COALESCE(status, 'active') as status, 
		       expired, is_active, created_at, updated_at, last_login
		FROM user_nodepath 
		WHERE id = ?
	`

	var user models.User
	err := h.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Email, &user.FullName, &user.Gmail, &user.Phone,
		&user.Status, &user.Expired, &user.IsActive, &user.CreatedAt,
		&user.UpdatedAt, &user.LastLogin,
	)
	
	if err != nil {
		// If columns don't exist, fall back to basic query
		fallbackQuery := `
			SELECT id, email, full_name, is_active, created_at, updated_at, last_login
			FROM user_nodepath 
			WHERE id = ?
		`
		
		err = h.db.QueryRow(fallbackQuery, userID).Scan(
			&user.ID, &user.Email, &user.FullName, &user.IsActive, 
			&user.CreatedAt, &user.UpdatedAt, &user.LastLogin,
		)
		
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error": "User not found",
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to retrieve profile",
				"details": err.Error(),
			})
		}
		
		// Set default values for missing fields
		emptyString := ""
		user.Gmail = &emptyString
		user.Phone = &emptyString
		user.Status = "active"
		user.Expired = nil
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": user,
	})
}

// UpdateProfile updates the current user's profile
func (h *ProfileHandlers) UpdateProfile(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	var updateData models.UserProfileUpdate
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
			"details": err.Error(),
		})
	}

	// Start building the update query
	updateQuery := "UPDATE user_nodepath SET full_name = ?, gmail = ?, phone = ?"
	args := []interface{}{updateData.FullName, updateData.Gmail, updateData.Phone}

	// Handle password update if provided
	if updateData.Password != nil && updateData.NewPassword != nil {
		// Verify current password first
		var currentHashedPassword string
		err := h.db.QueryRow("SELECT password FROM user_nodepath WHERE id = ?", userID).Scan(&currentHashedPassword)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to verify current password",
			})
		}

		// Check if current password is correct
		if err := bcrypt.CompareHashAndPassword([]byte(currentHashedPassword), []byte(*updateData.Password)); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Current password is incorrect",
			})
		}

		// Hash new password
		hashedNewPassword, err := bcrypt.GenerateFromPassword([]byte(*updateData.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to hash new password",
			})
		}

		updateQuery += ", password = ?"
		args = append(args, string(hashedNewPassword))
	}

	updateQuery += ", updated_at = NOW() WHERE id = ?"
	args = append(args, userID)

	// Execute the update
	_, err := h.db.Exec(updateQuery, args...)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update profile",
			"details": err.Error(),
		})
	}

	// Return updated user data
	return h.GetProfile(c)
}

// GetUserStatus returns the current user's status for system indicator
func (h *ProfileHandlers) GetUserStatus(c *fiber.Ctx) error {
	userID, ok := c.Locals("userID").(int)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Authentication required",
		})
	}

	// Try to get status, fallback if column doesn't exist
	query := `SELECT COALESCE(status, 'active') as status FROM user_nodepath WHERE id = ?`
	
	var status string
	err := h.db.QueryRow(query, userID).Scan(&status)
	if err != nil {
		// If status column doesn't exist, assume user is active
		if err.Error() == "no such column: status" {
			status = "active"
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to get user status",
				"details": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"success": true,
		"status": status,
		"system_online": status == "active",
	})
}