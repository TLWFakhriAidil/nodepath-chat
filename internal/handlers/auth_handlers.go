package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"time"

	"nodepath-chat/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// AuthHandlers handles user authentication operations
type AuthHandlers struct {
	db *sql.DB
}

// autoMigrate creates or updates the user_nodepath and user_sessions_nodepath tables
func (ah *AuthHandlers) autoMigrate() error {
	// Create user_nodepath table if not exists
	createUserTable := `
		CREATE TABLE IF NOT EXISTS user_nodepath (
			id CHAR(36) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			full_name VARCHAR(255) NOT NULL,
			password VARCHAR(255) NOT NULL,
			gmail VARCHAR(255) DEFAULT NULL,
			phone VARCHAR(20) DEFAULT NULL,
			is_active TINYINT(1) DEFAULT 1,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
			last_login TIMESTAMP NULL,
			status VARCHAR(255) DEFAULT 'Trial',
			expired VARCHAR(255) NULL
		)
	`

	if _, err := ah.db.Exec(createUserTable); err != nil {
		logrus.WithError(err).Error("Failed to create user_nodepath table")
		return err
	}

	// Create user_sessions_nodepath table if not exists
	createSessionTable := `
		CREATE TABLE IF NOT EXISTS user_sessions_nodepath (
			id CHAR(36) PRIMARY KEY,
			user_id CHAR(36) NOT NULL,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			INDEX idx_token (token),
			INDEX idx_user_id (user_id),
			INDEX idx_expires_at (expires_at)
		)
	`

	if _, err := ah.db.Exec(createSessionTable); err != nil {
		logrus.WithError(err).Error("Failed to create user_sessions_nodepath table")
		return err
	}

	// Check and add missing columns to user_nodepath
	columns := []struct {
		name       string
		definition string
	}{
		{"status", "ALTER TABLE user_nodepath ADD COLUMN status VARCHAR(255) DEFAULT 'Trial'"},
		{"expired", "ALTER TABLE user_nodepath ADD COLUMN expired VARCHAR(255) NULL"},
		{"gmail", "ALTER TABLE user_nodepath ADD COLUMN gmail VARCHAR(255) DEFAULT NULL"},
		{"phone", "ALTER TABLE user_nodepath ADD COLUMN phone VARCHAR(20) DEFAULT NULL"},
	}

	for _, col := range columns {
		var count int
		err := ah.db.QueryRow(`
			SELECT COUNT(*) 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_NAME = 'user_nodepath' 
			AND COLUMN_NAME = ?
		`, col.name).Scan(&count)

		if err != nil {
			logrus.WithError(err).Errorf("Failed to check column %s in user_nodepath", col.name)
			continue
		}

		if count == 0 {
			if _, err := ah.db.Exec(col.definition); err != nil {
				logrus.WithError(err).Errorf("Failed to add column %s to user_nodepath", col.name)
			} else {
				logrus.Infof("Added column %s to user_nodepath", col.name)
			}
		}
	}

	logrus.Info("Auth tables migration completed successfully")
	return nil
}

// NewAuthHandlers creates a new instance of AuthHandlers
func NewAuthHandlers(db *sql.DB) *AuthHandlers {
	ah := &AuthHandlers{
		db: db,
	}
	// Run auto-migration for user_nodepath and user_sessions_nodepath tables
	if db != nil {
		if err := ah.autoMigrate(); err != nil {
			logrus.WithError(err).Error("Failed to auto-migrate auth tables")
		}
	}
	return ah
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email"`
	FullName string `json:"full_name" validate:"required,min=2"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User  models.User `json:"user"`
	Token string      `json:"token"`
}

// Register handles user registration
func (ah *AuthHandlers) Register(c *fiber.Ctx) error {
	logrus.Info("Processing user registration request")

	// Check if database connection is available
	if ah.db == nil {
		logrus.Error("Database connection is not available")
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"success": false,
			"error":   "Database service is not available",
		})
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse registration request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request format",
		})
	}

	// Validate required fields
	if req.Email == "" || req.FullName == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Email, full name, and password are required",
		})
	}

	// Check if user already exists in user_nodepath table
	var existingUserID string
	err := ah.db.QueryRow("SELECT id FROM user_nodepath WHERE email = ?", req.Email).Scan(&existingUserID)
	if err == nil {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success": false,
			"error":   "User with this email already exists",
		})
	} else if err != sql.ErrNoRows {
		logrus.WithError(err).Error("Failed to check existing user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Internal server error",
		})
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash password")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to process password",
		})
	}

	// Generate UUID for user
	userID := generateUUID()

	// Calculate expired date (current date + 7 days)
	expiredDate := time.Now().Add(7 * 24 * time.Hour).Format("2006-01-02 15:04:05")

	// Insert new user into user_nodepath table with status='Trial' and expired=now+7days
	_, err = ah.db.Exec(
		`INSERT INTO user_nodepath 
		(id, email, full_name, password, is_active, created_at, updated_at, status, expired) 
		VALUES (?, ?, ?, ?, 1, NOW(), NOW(), 'Trial', ?)`,
		userID, req.Email, req.FullName, string(hashedPassword), expiredDate,
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to create user in user_nodepath")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create user",
		})
	}

	// Fetch the created user from user_nodepath
	var user models.User
	err = ah.db.QueryRow(
		"SELECT id, email, full_name, is_active, created_at, updated_at, last_login FROM user_nodepath WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Email, &user.FullName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch created user from user_nodepath")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create user",
		})
	}

	// Generate session token
	token, err := generateSessionToken()
	if err != nil {
		logrus.WithError(err).Error("Failed to generate session token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create session",
		})
	}

	// Store session in database with client information
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	err = ah.storeSession(token, user.ID, ipAddress, userAgent)
	if err != nil {
		logrus.WithError(err).Error("Failed to store session in user_sessions_nodepath")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create session",
		})
	}

	// Set session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour), // 24 hours
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
	})

	logrus.WithField("user_id", user.ID).Info("User registered successfully")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"success": true,
		"message": "User registered successfully",
		"data": AuthResponse{
			User:  user,
			Token: token,
		},
	})
}

// Login handles user authentication with fallback for database unavailability
func (ah *AuthHandlers) Login(c *fiber.Ctx) error {
	logrus.Info("Processing user login request")

	// Check if database connection is available
	if ah.db == nil {
		logrus.Warn("Database connection is not available, using fallback authentication")
		return ah.loginWithFallback(c)
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse login request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request format",
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Email and password are required",
		})
	}

	// Fetch user from user_nodepath table
	var user models.User
	var hashedPassword string
	err := ah.db.QueryRow(
		"SELECT id, email, full_name, password, is_active, created_at, updated_at, last_login FROM user_nodepath WHERE email = ? AND is_active = 1",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.FullName, &hashedPassword, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid email or password",
		})
	} else if err != nil {
		logrus.WithError(err).Error("Failed to fetch user from user_nodepath")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Internal server error",
		})
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password))
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid email or password",
		})
	}

	// Update last_login timestamp in user_nodepath
	_, err = ah.db.Exec("UPDATE user_nodepath SET last_login = NOW() WHERE id = ?", user.ID)
	if err != nil {
		logrus.WithError(err).Error("Failed to update last_login in user_nodepath")
		// Don't fail the login for this error, just log it
	}

	// Generate session token
	token, err := generateSessionToken()
	if err != nil {
		logrus.WithError(err).Error("Failed to generate session token")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create session",
		})
	}

	// Set session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  time.Now().Add(24 * time.Hour), // 24 hours
		HTTPOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: "Lax",
	})

	// Store session in database with client information
	ipAddress := c.IP()
	userAgent := c.Get("User-Agent")
	err = ah.storeSession(token, user.ID, ipAddress, userAgent)
	if err != nil {
		logrus.WithError(err).Error("Failed to store session in user_sessions_nodepath")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create session",
		})
	}

	logrus.WithField("user_id", user.ID).Info("User logged in successfully")

	// Check if user has devices for redirect logic
	deviceCount, deviceIDs, err := ah.CheckUserDevices(user.ID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", user.ID).Error("Failed to check user devices after login")
		// Don't fail login, just log the error
		deviceCount = 0
	}

	// Determine redirect URL based on device ownership
	redirectURL := "/device-settings" // Default to device settings if no devices
	if deviceCount > 0 {
		redirectURL = "/dashboard" // Redirect to dashboard if user has devices
	}

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login successful",
		"data": AuthResponse{
			User:  user,
			Token: token,
		},
		"redirect_url": redirectURL,
		"has_devices":  deviceCount > 0,
		"device_count": deviceCount,
		"device_ids":   deviceIDs,
	})
}

// Logout handles user logout
func (ah *AuthHandlers) Logout(c *fiber.Ctx) error {
	logrus.Info("Processing user logout request")

	// Get session token from cookie
	token := c.Cookies("session_token")
	if token != "" {
		// Remove session from database
		err := ah.removeSession(token)
		if err != nil {
			logrus.WithError(err).Error("Failed to remove session")
		}
	}

	// Clear session cookie
	c.Cookie(&fiber.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // Expire immediately
		HTTPOnly: true,
		Secure:   false,
		SameSite: "Lax",
	})

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Logout successful",
	})
}

// GetCurrentUser returns the current authenticated user
func (ah *AuthHandlers) GetCurrentUser(c *fiber.Ctx) error {
	// Get user from context (set by auth middleware)
	userID := c.Locals("user_id")
	if userID == nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Not authenticated",
		})
	}

	// Fetch user from user_nodepath table
	var user models.User
	err := ah.db.QueryRow(
		"SELECT id, email, full_name, is_active, created_at, updated_at, last_login FROM user_nodepath WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Email, &user.FullName, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.LastLogin)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch current user from user_nodepath")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to fetch user data",
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    user,
	})
}

// Simple in-memory session store (use Redis or database in production)
// generateSessionToken generates a random session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// generateUUID generates a simple UUID-like string
func generateUUID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// storeSession stores a session token with user ID in user_sessions_nodepath table
func (ah *AuthHandlers) storeSession(token string, userID string, ipAddress, userAgent string) error {
	// Set expiration time to 24 hours from now
	expiresAt := time.Now().Add(24 * time.Hour)
	// Generate UUID for session ID
	sessionID := generateUUID()
	_, err := ah.db.Exec(`
		INSERT INTO user_sessions_nodepath (id, user_id, token, expires_at, created_at) 
		VALUES (?, ?, ?, ?, NOW())
	`, sessionID, userID, token, expiresAt)

	return err
}

// getSession retrieves user ID from session token in user_sessions_nodepath table
func (ah *AuthHandlers) getSession(token string) (string, bool) {
	var userID string
	var expiresAt time.Time

	err := ah.db.QueryRow(`
		SELECT user_id, expires_at FROM user_sessions_nodepath 
		WHERE token = ? AND expires_at > NOW()
	`, token).Scan(&userID, &expiresAt)

	if err != nil {
		return "", false
	}

	return userID, true
}

// removeSession removes a session token from user_sessions_nodepath table
func (ah *AuthHandlers) removeSession(token string) error {
	_, err := ah.db.Exec(`DELETE FROM user_sessions_nodepath WHERE token = ?`, token)
	return err
}

// cleanupExpiredSessions removes expired sessions from user_sessions_nodepath table
func (ah *AuthHandlers) cleanupExpiredSessions() error {
	_, err := ah.db.Exec(`DELETE FROM user_sessions_nodepath WHERE expires_at < NOW()`)
	return err
}

// convertUUIDToInt converts a UUID string to a consistent integer using hash function
func convertUUIDToInt(uuid string) int {
	h := fnv.New32a()
	h.Write([]byte(uuid))
	return int(h.Sum32())
}

// AuthMiddleware validates session tokens and sets user context
func (ah *AuthHandlers) AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get session token from cookie
		token := c.Cookies("session_token")
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authentication required",
			})
		}

		// Validate session using database
		userID, exists := ah.getSession(token)
		if !exists {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Invalid or expired session",
			})
		}

		// Set user ID in context as string UUID (CHAR(36) format)
		c.Locals("user_id", userID) // Primary key for all handlers
		return c.Next()
	}
}

// DeviceRequiredMiddleware checks if user has at least one device_setting_nodepath record
func (ah *AuthHandlers) DeviceRequiredMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user ID from context (should be set by AuthMiddleware)
		userIDStr, ok := c.Locals("user_id").(string)
		if !ok || userIDStr == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"success": false,
				"error":   "Authentication required",
			})
		}

		// Check if user has any device settings
		var count int
		err := ah.db.QueryRow(`
			SELECT COUNT(*) FROM device_setting_nodepath 
			WHERE user_id = ?
		`, userIDStr).Scan(&count)
		if err != nil {
			logrus.WithError(err).WithField("userID", userIDStr).Error("Failed to check user device count")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Internal server error",
			})
		}

		// If user has no devices, return device required error
		if count == 0 {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"success": false,
				"error":   "device_required",
				"message": "Please add a device first to access this feature",
			})
		}

		// User has devices, continue to next handler
		return c.Next()
	}
}

// CheckUserDevices returns device count and device IDs for a user
func (ah *AuthHandlers) CheckUserDevices(userID string) (int, []string, error) {
	// Check if database connection is available
	if ah.db == nil {
		return 0, nil, fmt.Errorf("database connection is not available")
	}

	var deviceIDs []string
	var count int

	// Get device count and IDs
	rows, err := ah.db.Query(`
		SELECT id_device FROM device_setting_nodepath 
		WHERE user_id = ? AND id_device IS NOT NULL AND id_device != ''
	`, userID)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to query user devices: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var deviceID string
		if err := rows.Scan(&deviceID); err != nil {
			return 0, nil, fmt.Errorf("failed to scan device ID: %w", err)
		}
		deviceIDs = append(deviceIDs, deviceID)
		count++
	}

	if err = rows.Err(); err != nil {
		return 0, nil, fmt.Errorf("error iterating device rows: %w", err)
	}

	return count, deviceIDs, nil
}

// GetDeviceStatus returns the device status for the authenticated user
func (ah *AuthHandlers) GetDeviceStatus(c *fiber.Ctx) error {
	// Get user ID from context
	userIDStr, ok := c.Locals("user_id").(string)
	if !ok || userIDStr == "" {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Authentication required",
		})
	}

	// Check user devices
	count, deviceIDs, err := ah.CheckUserDevices(userIDStr)
	if err != nil {
		logrus.WithError(err).WithField("userID", userIDStr).Error("Failed to check user devices")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Internal server error",
		})
	}

	return c.JSON(fiber.Map{
		"success":      true,
		"has_devices":  count > 0,
		"device_count": count,
		"device_ids":   deviceIDs,
	})
}

// SetupAuthRoutes configures authentication routes
func (ah *AuthHandlers) SetupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")
	auth.Post("/register", ah.Register)
	auth.Post("/login", ah.Login)
	auth.Post("/logout", ah.Logout)
	auth.Get("/me", ah.AuthMiddleware(), ah.GetCurrentUser)

	// Device check endpoint
	auth.Get("/device-status", ah.AuthMiddleware(), ah.GetDeviceStatus)
}

// SetupTemplateRoutes configures template serving routes
func (ah *AuthHandlers) SetupTemplateRoutes(app *fiber.App) {
	// Serve login page
	app.Get("/login", func(c *fiber.Ctx) error {
		return c.SendFile("./web/templates/login.html")
	})

	// Serve register page
	app.Get("/register", func(c *fiber.Ctx) error {
		return c.SendFile("./web/templates/register.html")
	})
}

// loginWithFallback provides temporary authentication when database is unavailable
func (ah *AuthHandlers) loginWithFallback(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse login request")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid request format",
		})
	}

	// Validate required fields
	if req.Email == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "Email and password are required",
		})
	}

	// Fallback credentials for development when database is unavailable
	// In production, this should be removed or use environment variables
	fallbackCredentials := map[string]string{
		"admin@nodepath.com": "admin123",
		"test@nodepath.com":  "test123",
		"demo@nodepath.com":  "demo123",
	}

	// Check fallback credentials
	if expectedPassword, exists := fallbackCredentials[req.Email]; exists && expectedPassword == req.Password {
		// Create a temporary user object
		user := models.User{
			ID:        generateUUID(),
			Email:     req.Email,
			FullName:  "Fallback User",
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		// Generate session token
		token, err := generateSessionToken()
		if err != nil {
			logrus.WithError(err).Error("Failed to generate session token")
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"error":   "Internal server error",
			})
		}

		logrus.WithFields(logrus.Fields{
			"user_id": user.ID,
			"email":   user.Email,
		}).Info("User logged in successfully with fallback authentication")

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"message": "Login successful (fallback mode - database unavailable)",
			"data": AuthResponse{
				User:  user,
				Token: token,
			},
		})
	}

	// Invalid credentials
	return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
		"success": false,
		"error":   "Invalid email or password (fallback mode)",
	})
}
