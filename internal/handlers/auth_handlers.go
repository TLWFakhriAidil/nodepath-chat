package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
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

// NewAuthHandlers creates a new instance of AuthHandlers
func NewAuthHandlers(db *sql.DB) *AuthHandlers {
	return &AuthHandlers{
		db: db,
	}
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

	// Check if user already exists
	var existingUserID int
	err := ah.db.QueryRow("SELECT id FROM users WHERE email = ?", req.Email).Scan(&existingUserID)
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

	// Insert new user
	result, err := ah.db.Exec(
		"INSERT INTO users (email, full_name, password, created_at, updated_at) VALUES (?, ?, ?, NOW(), NOW())",
		req.Email, req.FullName, string(hashedPassword),
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to create user")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create user",
		})
	}

	// Get the created user ID
	userID, err := result.LastInsertId()
	if err != nil {
		logrus.WithError(err).Error("Failed to get user ID")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create user",
		})
	}

	// Fetch the created user
	var user models.User
	err = ah.db.QueryRow(
		"SELECT id, email, full_name, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Email, &user.FullName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch created user")
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
		logrus.WithError(err).Error("Failed to store session")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   "Failed to create session",
		})
	}

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

// Login handles user authentication
func (ah *AuthHandlers) Login(c *fiber.Ctx) error {
	logrus.Info("Processing user login request")

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

	// Fetch user from database
	var user models.User
	var hashedPassword string
	err := ah.db.QueryRow(
		"SELECT id, email, full_name, password, created_at, updated_at FROM users WHERE email = ?",
		req.Email,
	).Scan(&user.ID, &user.Email, &user.FullName, &hashedPassword, &user.CreatedAt, &user.UpdatedAt)
	if err == sql.ErrNoRows {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"success": false,
			"error":   "Invalid email or password",
		})
	} else if err != nil {
		logrus.WithError(err).Error("Failed to fetch user")
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



	logrus.WithField("user_id", user.ID).Info("User logged in successfully")

	return c.JSON(fiber.Map{
		"success": true,
		"message": "Login successful",
		"data": AuthResponse{
			User:  user,
			Token: token,
		},
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

	// Fetch user from database
	var user models.User
	err := ah.db.QueryRow(
		"SELECT id, email, full_name, created_at, updated_at FROM users WHERE id = ?",
		userID,
	).Scan(&user.ID, &user.Email, &user.FullName, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch current user")
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

// storeSession stores a session token with user ID in database
func (ah *AuthHandlers) storeSession(token string, userID int, ipAddress, userAgent string) error {
	// Set expiration time to 24 hours from now
	expiresAt := time.Now().Add(24 * time.Hour)
	
	_, err := ah.db.Exec(`
		INSERT INTO user_sessions (session_token, user_id, expires_at, ip_address, user_agent) 
		VALUES (?, ?, ?, ?, ?)
	`, token, userID, expiresAt, ipAddress, userAgent)
	
	return err
}

// getSession retrieves user ID from session token in database
func (ah *AuthHandlers) getSession(token string) (int, bool) {
	var userID int
	var expiresAt time.Time
	
	err := ah.db.QueryRow(`
		SELECT user_id, expires_at FROM user_sessions 
		WHERE session_token = ? AND is_active = TRUE
	`, token).Scan(&userID, &expiresAt)
	
	if err != nil {
		return 0, false
	}
	
	// Check if session has expired
	if time.Now().After(expiresAt) {
		// Mark session as inactive
		ah.db.Exec(`UPDATE user_sessions SET is_active = FALSE WHERE session_token = ?`, token)
		return 0, false
	}
	
	// Update last accessed time
	ah.db.Exec(`UPDATE user_sessions SET last_accessed = NOW() WHERE session_token = ?`, token)
	
	return userID, true
}

// removeSession removes a session token from database
func (ah *AuthHandlers) removeSession(token string) error {
	_, err := ah.db.Exec(`UPDATE user_sessions SET is_active = FALSE WHERE session_token = ?`, token)
	return err
}

// cleanupExpiredSessions removes expired sessions from database
func (ah *AuthHandlers) cleanupExpiredSessions() error {
	_, err := ah.db.Exec(`DELETE FROM user_sessions WHERE expires_at < NOW() OR is_active = FALSE`)
	return err
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

		// Set user ID in context
		c.Locals("user_id", userID)
		return c.Next()
	}
}

// SetupAuthRoutes configures authentication routes
func (ah *AuthHandlers) SetupAuthRoutes(api fiber.Router) {
	auth := api.Group("/auth")
	auth.Post("/register", ah.Register)
	auth.Post("/login", ah.Login)
	auth.Post("/logout", ah.Logout)
	auth.Get("/me", ah.AuthMiddleware(), ah.GetCurrentUser)
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