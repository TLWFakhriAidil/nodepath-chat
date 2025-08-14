package repository

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"sync"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/google/uuid"
)

// UserRepository handles user data persistence in PostgreSQL
type UserRepository struct {
	db *sql.DB
}

var (
	userRepo     *UserRepository
	userRepoOnce sync.Once
)

// GetUserRepository returns singleton instance of UserRepository
func GetUserRepository() *UserRepository {
	userRepoOnce.Do(func() {
		userRepo = &UserRepository{
			db: database.GetDB(),
		}
		
		// Create default admin user if not exists
		// The schema.sql already handles this, but we'll check anyway
		_, err := userRepo.GetUserByEmail("admin@whatsapp.com")
		if err != nil {
			userRepo.CreateUser("admin@whatsapp.com", "Administrator", "changeme123")
		}
	})
	return userRepo
}

// DB returns the database connection
func (r *UserRepository) DB() *sql.DB {
	return r.db
}

// CreateUser creates a new user in the database
func (r *UserRepository) CreateUser(email, fullName, password string) (*models.User, error) {
	// Check if user exists
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 from users WHERE email = ?)", email).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}
	
	// Encode password with base64 (for easy viewing)
	encodedPassword := base64.StdEncoding.EncodeToString([]byte(password))
	
	// Create user
	user := &models.User{
		ID:           uuid.New().String(),
		Email:        email,
		FullName:     fullName,
		PasswordHash: encodedPassword,
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	
	// Insert into database
	query := `
		INSERT INTO users (id, email, full_name, password_hash, is_active, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?), updated_at
	`
	err = r.db.QueryRow(query, user.ID, user.Email, user.FullName, user.PasswordHash, 
		user.IsActive, user.CreatedAt, user.UpdatedAt).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(email string) (*models.User, error) {
	// fmt.Printf("Debug GetUserByEmail: Looking for email: '%s'\n", email)
	user := &models.User{}
	var lastLogin sql.NullTime
	query := `
		SELECT id, email, full_name, password_hash, is_active, created_at, updated_at, last_login
		from users WHERE email = ?
	`
	err := r.db.QueryRow(query, email).Scan(
		&user.ID, &user.Email, &user.FullName, &user.PasswordHash,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &lastLogin,
	)
	if err == sql.ErrNoRows {
		// fmt.Printf("Debug GetUserByEmail: User not found for email: '%s'\n", email)
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		// fmt.Printf("Debug GetUserByEmail: Database error: %v\n", err)
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}
	
	// fmt.Printf("Debug GetUserByEmail: Found user - Email: '%s', ID: %s\n", user.Email, user.ID)
	return user, nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(id string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, email, full_name, password_hash, is_active, created_at, updated_at, last_login
		from users WHERE id = ?
	`
	var lastLogin sql.NullTime
	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Email, &user.FullName, &user.PasswordHash,
		&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &lastLogin,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	if lastLogin.Valid {
		user.LastLogin = lastLogin.Time
	}
	
	return user, nil
}

// ValidatePassword checks if password is correct
func (r *UserRepository) ValidatePassword(email, password string) (*models.User, error) {
	user, err := r.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	
	if !user.IsActive {
		return nil, fmt.Errorf("user account is disabled")
	}
	
	// Debug logging - temporarily enabled
	fmt.Printf("Debug: Validating password for email: %s\n", email)
	fmt.Printf("Debug: Password provided: %s\n", password)
	fmt.Printf("Debug: Encoded password from DB: %s\n", user.PasswordHash)
	
	// Decode the stored password
	decodedPassword, err := base64.StdEncoding.DecodeString(user.PasswordHash)
	if err != nil {
		fmt.Printf("Debug: Failed to decode password: %v\n", err)
		return nil, fmt.Errorf("invalid password format")
	}
	
	// Compare passwords
	if string(decodedPassword) != password {
		fmt.Printf("Debug: Password mismatch - Stored: '%s', Provided: '%s'\n", string(decodedPassword), password)
		return nil, fmt.Errorf("invalid password")
	}
	
	fmt.Printf("Debug: Password validation successful\n")
	
	// Update last login
	_, err = r.db.Exec("UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?", user.ID)
	if err != nil {
		// Log error but don't fail login
		fmt.Printf("Failed to update last login: %v\n", err)
	}
	
	return user, nil
}

// CreateSession creates a new user session
func (r *UserRepository) CreateSession(userID string) (*models.UserSession, error) {
	session := &models.UserSession{
		ID:        uuid.New().String(),
		UserID:    userID,
		Token:     uuid.New().String(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
	}
	
	query := `
		INSERT INTO user_sessions(id, user_id, token, expires_at, created_at)
		VALUES (?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, session.ID, session.UserID, session.Token, session.ExpiresAt, session.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}
	
	return session, nil
}

// GetSession retrieves a session by token
func (r *UserRepository) GetSession(token string) (*models.UserSession, error) {
	session := &models.UserSession{}
	query := `
		SELECT id, user_id, token, expires_at, created_at
		from user_sessions 
		WHERE token = ? AND expires_at > CURRENT_TIMESTAMP
	`
	err := r.db.QueryRow(query, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.ExpiresAt, &session.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	
	return session, nil
}

// AddUserDevice adds a device for a user
func (r *UserRepository) AddUserDevice(userID, deviceName string) (*models.UserDevice, error) {
	device := &models.UserDevice{
		ID:              uuid.New().String(),
		UserID:          userID,
		DeviceName:      deviceName,
		Status:          "offline",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		LastSeen:        time.Now(),
		MinDelaySeconds: 5,
		MaxDelaySeconds: 15,
	}
	
	query := `
		INSERT INTO user_devices(id, user_id, device_name, status, last_seen, created_at, updated_at, min_delay_seconds, max_delay_seconds)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, device.ID, device.UserID, device.DeviceName, 
		device.Status, device.LastSeen, device.CreatedAt, device.UpdatedAt, 
		device.MinDelaySeconds, device.MaxDelaySeconds)
	if err != nil {
		return nil, fmt.Errorf("failed to add device: %w", err)
	}
	
	return device, nil
}

// AddUserDeviceWithPhone adds a device for a user with phone number
func (r *UserRepository) AddUserDeviceWithPhone(userID, deviceName, phone string) (*models.UserDevice, error) {
	device := &models.UserDevice{
		ID:              uuid.New().String(),
		UserID:          userID,
		DeviceName:      deviceName,
		Phone:           phone,
		Status:          "offline",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		LastSeen:        time.Now(),
		MinDelaySeconds: 5,
		MaxDelaySeconds: 15,
	}
	
	query := `
		INSERT INTO user_devices(id, user_id, device_name, phone, status, last_seen, created_at, updated_at, min_delay_seconds, max_delay_seconds)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.Exec(query, device.ID, device.UserID, device.DeviceName, 
		device.Phone, device.Status, device.LastSeen, device.CreatedAt, device.UpdatedAt,
		device.MinDelaySeconds, device.MaxDelaySeconds)
	if err != nil {
		return nil, fmt.Errorf("failed to add device with phone: %w", err)
	}
	
	return device, nil
}

// GetUserDevices gets all devices for a user
func (r *UserRepository) GetUserDevices(userID string) ([]*models.UserDevice, error) {
	query := `
		SELECT id, user_id, device_name, phone, jid, status, last_seen, created_at, 
		       COALESCE(platform, '') as platform
		FROM user_devices 
		WHERE user_id = ?
		order BY created_at DESC
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}
	defer rows.Close()
	
	var devices []*models.UserDevice
	for rows.Next() {
		device := &models.UserDevice{}
		var phone, jid sql.NullString
		err := rows.Scan(&device.ID, &device.UserID, &device.DeviceName, 
			&phone, &jid, &device.Status, &device.LastSeen, &device.CreatedAt,
			&device.Platform)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}
		device.Phone = phone.String
		device.JID = jid.String
		
		// Platform devices always appear as online
		if device.Platform != "" {
			device.Status = "online"
		}
		
		devices = append(devices, device)
	}
	
	return devices, nil
}

// GetUserDevice gets a specific device for a user
func (r *UserRepository) GetUserDevice(userID, deviceID string) (*models.UserDevice, error) {
	query := `
		SELECT id, user_id, device_name, phone, jid, status, last_seen, created_at
		FROM user_devices 
		WHERE user_id = ? AND id = ?
	`
	device := &models.UserDevice{}
	var phone, jid sql.NullString
	
	err := r.db.QueryRow(query, userID, deviceID).Scan(
		&device.ID, &device.UserID, &device.DeviceName, 
		&phone, &jid, &device.Status, &device.LastSeen, &device.CreatedAt)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("device not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	
	device.Phone = phone.String
	device.JID = jid.String
	
	return device, nil
}

// GetDeviceByID gets a device by ID
func (r *UserRepository) GetDeviceByID(deviceID string) (*models.UserDevice, error) {
	var device models.UserDevice
	query := `
		SELECT id, user_id, device_name, COALESCE(phone, ''), COALESCE(jid, ''), status, 
		       COALESCE(min_delay_seconds, 5), COALESCE(max_delay_seconds, 15),
		       created_at, COALESCE(updated_at, created_at), last_seen, COALESCE(platform, '')
		FROM user_devices
		WHERE id = ?
	`
	
	err := r.db.QueryRow(query, deviceID).Scan(
		&device.ID, &device.UserID, &device.DeviceName, &device.Phone,
		&device.JID, &device.Status, &device.MinDelaySeconds, &device.MaxDelaySeconds,
		&device.CreatedAt, &device.UpdatedAt, &device.LastSeen, &device.Platform,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("device not found")
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	
	return &device, nil
}

// UpdateDeviceStatus updates device status
func (r *UserRepository) UpdateDeviceStatus(deviceID, status string, phone, jid string) error {
	query := `
		UPDATE user_devices SET status = ?, last_seen = CURRENT_TIMESTAMP, phone = ?, jid = ?
		WHERE id = ?
	`
	_, err := r.db.Exec(query, deviceID, status, phone, jid)
	if err != nil {
		return fmt.Errorf("failed to update device status: %w", err)
	}
	
	return nil
}

// DeleteDevice deletes a device
func (r *UserRepository) DeleteDevice(deviceID string) error {
	// Start a transaction to ensure data consistency
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()
	
	// First, delete all WhatsApp messages for this device
	result, err := tx.Exec("DELETE FROM whatsapp_messages WHERE device_id = ?", deviceID)
	if err != nil {
		log.Printf("Warning: failed to delete WhatsApp messages: %v", err)
		// Continue with deletion even if this fails (table might not exist)
	} else {
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("Deleted %d WhatsApp messages for device %s", rowsAffected, deviceID)
		}
	}
	
	// Delete all WhatsApp chats for this device
	result, err = tx.Exec("DELETE FROM whatsapp_chats WHERE device_id = ?", deviceID)
	if err != nil {
		log.Printf("Warning: failed to delete WhatsApp chats: %v", err)
		// Continue with deletion even if this fails (table might not exist)
	} else {
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			log.Printf("Deleted %d WhatsApp chats for device %s", rowsAffected, deviceID)
		}
	}
	
	// Delete all leads associated with this device
	result, err = tx.Exec("DELETE FROM leads WHERE device_id = ?", deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete leads for device: %w", err)
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Deleted %d leads for device %s", rowsAffected, deviceID)
	}
	
	// Also delete any broadcast messages for this device
	result, err = tx.Exec("DELETE FROM broadcast_messages WHERE device_id = ?", deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete broadcast messages for device: %w", err)
	}
	
	rowsAffected, _ = result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Deleted %d broadcast messages for device %s", rowsAffected, deviceID)
	}
	
	// Finally, delete the device itself
	_, err = tx.Exec("DELETE FROM user_devices WHERE id = ?", deviceID)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}
	
	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	log.Printf("Successfully deleted device %s and all associated data (including WhatsApp chats and messages)", deviceID)
	return nil
}

// GetDevice gets a specific device by ID
func (r *UserRepository) GetDevice(userID, deviceID string) (*models.UserDevice, error) {
	query := `
		SELECT id, user_id, device_name, COALESCE(phone, ''), COALESCE(jid, ''), status, created_at, last_seen
		FROM user_devices
		WHERE user_id = ? AND id = ?
	`
	
	var device models.UserDevice
	
	err := r.db.QueryRow(query, userID, deviceID).Scan(
		&device.ID, &device.UserID, &device.DeviceName,
		&device.Phone, &device.JID, &device.Status,
		&device.CreatedAt, &device.LastSeen,
	)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("device not found")
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	
	return &device, nil
}

// GetAllUsers returns all users (for admin)
func (r *UserRepository) GetAllUsers() ([]*models.User, error) {
	query := `
		SELECT id, email, full_name, is_active, created_at, updated_at, last_login
		from users 
		order BY created_at DESC
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()
	
	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var lastLogin sql.NullTime
		
		err := rows.Scan(
			&user.ID, &user.Email, &user.FullName, 
			&user.IsActive, &user.CreatedAt, &user.UpdatedAt, &lastLogin,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		
		if lastLogin.Valid {
			user.LastLogin = lastLogin.Time
		}
		
		users = append(users, user)
	}
	
	return users, nil
}

// CleanupExpiredSessions removes expired sessions
func (r *UserRepository) CleanupExpiredSessions() error {
	_, err := r.db.Exec("DELETE FROM user_sessions WHERE expires_at < CURRENT_TIMESTAMP")
	return err
}

// UpdateDevicePhone updates the phone number for a user's device
func (r *UserRepository) UpdateDevicePhone(userID, deviceID, phone string) error {
	query := `
		UPDATE user_devices 
		SET phone = ?, last_seen = CURRENT_TIMESTAMP
		WHERE user_id = ? AND id = ?
	`
	result, err := r.db.Exec(query, phone, userID, deviceID)
	if err != nil {
		return fmt.Errorf("failed to update device phone: %w", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	
	if rowsAffected == 0 {
		return fmt.Errorf("device not found or not owned by user")
	}
	
	return nil
}

// GetAllDevices retrieves all devices from the database
func (r *UserRepository) GetAllDevices() ([]*models.UserDevice, error) {
	query := `
		SELECT id, user_id, device_name, phone, status, last_seen, created_at, updated_at,
		       COALESCE(jid, '') as jid, COALESCE(platform, '') as platform
		FROM user_devices 
		order BY created_at DESC
	`
	
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all devices: %w", err)
	}
	defer rows.Close()
	
	var devices []*models.UserDevice
	for rows.Next() {
		device := &models.UserDevice{}
		err := rows.Scan(
			&device.ID,
			&device.UserID,
			&device.DeviceName,
			&device.Phone,
			&device.Status,
			&device.LastSeen,
			&device.CreatedAt,
			&device.UpdatedAt,
			&device.JID,
			&device.Platform,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device row: %w", err)
		}
		
		// Platform devices always appear as online
		if device.Platform != "" {
			device.Status = "online"
		}
		
		devices = append(devices, device)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating device rows: %w", err)
	}
	
	return devices, nil
}

// CreateDevice creates a new device with all fields
func (r *UserRepository) CreateDevice(device *models.UserDevice) error {
	// Ensure we have required fields
	if device.ID == "" {
		device.ID = uuid.New().String()
	}
	if device.Status == "" {
		device.Status = "offline"
	}
	if device.CreatedAt.IsZero() {
		device.CreatedAt = time.Now()
	}
	if device.UpdatedAt.IsZero() {
		device.UpdatedAt = time.Now()
	}
	if device.LastSeen.IsZero() {
		device.LastSeen = time.Now()
	}
	
	query := `
		INSERT INTO user_devices(
			id, user_id, device_name, phone, jid, status, 
			last_seen, created_at, updated_at, 
			min_delay_seconds, max_delay_seconds, platform
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	
	_, err := r.db.Exec(query, 
		device.ID, 
		device.UserID, 
		device.DeviceName,
		device.Phone,
		device.JID,
		device.Status,
		device.LastSeen,
		device.CreatedAt,
		device.UpdatedAt,
		device.MinDelaySeconds,
		device.MaxDelaySeconds,
		device.Platform,
	)
	
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}
	
	return nil
}


// GetDeviceByUserAndJID gets a device by user ID and JID combination
func (r *UserRepository) GetDeviceByUserAndJID(userID, jid string) (*models.UserDevice, error) {
	device := &models.UserDevice{}
	query := `
		SELECT id, user_id, device_name, phone, jid, status, last_seen, created_at, updated_at,
		       COALESCE(min_delay_seconds, 5) as min_delay_seconds,
		       COALESCE(max_delay_seconds, 15) as max_delay_seconds,
		       COALESCE(platform, '') as platform
		FROM user_devices
		WHERE user_id = ? AND jid = ?
		limit 1
	`
	
	err := r.db.QueryRow(query, userID, jid).Scan(
		&device.ID,
		&device.UserID,
		&device.DeviceName,
		&device.Phone,
		&device.JID,
		&device.Status,
		&device.LastSeen,
		&device.CreatedAt,
		&device.UpdatedAt,
		&device.MinDelaySeconds,
		&device.MaxDelaySeconds,
		&device.Platform,
	)
	
	if err != nil {
		return nil, err
	}
	
	return device, nil
}

// GetDeviceByUserAndName gets a device by user ID and device name combination
func (r *UserRepository) GetDeviceByUserAndName(userID, deviceName string) (*models.UserDevice, error) {
	device := &models.UserDevice{}
	query := `
		SELECT id, user_id, device_name, phone, jid, status, last_seen, created_at, updated_at,
		       COALESCE(min_delay_seconds, 5) as min_delay_seconds,
		       COALESCE(max_delay_seconds, 15) as max_delay_seconds,
		       COALESCE(platform, '') as platform
		FROM user_devices
		WHERE user_id = ? AND device_name = ?
		limit 1
	`
	
	err := r.db.QueryRow(query, userID, deviceName).Scan(
		&device.ID,
		&device.UserID,
		&device.DeviceName,
		&device.Phone,
		&device.JID,
		&device.Status,
		&device.LastSeen,
		&device.CreatedAt,
		&device.UpdatedAt,
		&device.MinDelaySeconds,
		&device.MaxDelaySeconds,
		&device.Platform,
	)
	
	if err != nil {
		return nil, err
	}
	
	return device, nil
}

// GetDB returns the database connection
func (r *UserRepository) GetDB() *sql.DB {
	return r.db
}
