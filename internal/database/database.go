package database

import (
	"database/sql"
	"fmt"
	"time"

	"nodepath-chat/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

// Initialize creates and returns a database connection using MYSQL_URI exclusively
func Initialize(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.GetDSN()
	if dsn == "" {
		return nil, fmt.Errorf("MYSQL_URI environment variable is required")
	}

	// Log which database URL is being used
	logrus.Info("Connecting to MySQL database using MYSQL_URI")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool for high concurrency (3000+ users)
	// Optimized settings for handling 3000+ concurrent users with real-time messaging
	db.SetMaxOpenConns(500)                 // Increased significantly for 3000+ concurrent users
	db.SetMaxIdleConns(100)                 // Higher idle connections to reduce connection overhead
	db.SetConnMaxLifetime(60 * time.Minute) // Longer lifetime to reduce connection churn
	db.SetConnMaxIdleTime(15 * time.Minute) // Balanced idle time for resource efficiency

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logrus.Info("Database connection established successfully")
	return db, nil
}

// RunMigrations runs all database migrations
func RunMigrations(db *sql.DB) error {
	logrus.Info("Running database migrations")

	// Test chat executions table removed from migrations
	migrations := []string{
		createFlowsTable,
		createDeviceSettingsTable,
		createUsersTable,
		createUserSessionsTable,
		createAIWhatsappTable,
		createWasapBotTable,
		createConversationLogTable,
		createOrdersTable,
		createAIWhatsappSessionTable,
		createWasapBotSessionTable,
	}

	for i, migration := range migrations {
		logrus.WithField("migration", i+1).Debug("Running migration")
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", i+1, err)
		}
	}

	// Remove deprecated columns
	if err := removeDeprecatedColumnsFromFlowsTable(db); err != nil {
		logrus.WithError(err).Warn("Some deprecated columns may not exist, continuing...")
	}

	// Add missing columns with error handling
	if err := addMissingColumnsToFlowsTable(db); err != nil {
		logrus.WithError(err).Warn("Some columns may already exist, continuing...")
	}

	if err := addMissingColumnsToDeviceSettingsTable(db); err != nil {
		logrus.WithError(err).Warn("Some device settings columns may already exist, continuing...")
	}

	// Update provider values from 'rvsb_wasap' to 'waha'
	if err := updateProviderRvsbWasapToWaha(db); err != nil {
		logrus.WithError(err).Warn("Failed to update provider values, continuing...")
	}

	// Update provider ENUM to include 'waha' (critical for WAHA provider support)
	if err := updateProviderEnum(db); err != nil {
		logrus.WithError(err).Warn("Failed to update provider ENUM, continuing...")
	}

	// Drop deprecated billing columns from orders_nodepath
	if err := dropDeprecatedBillingColumns(db); err != nil {
		logrus.WithError(err).Warn("Failed to drop deprecated billing columns, continuing...")
	}

	// Convert user_id column from INT to CHAR(36) in orders_nodepath
	if err := convertUserIDToChar36(db); err != nil {
		logrus.WithError(err).Warn("Failed to convert user_id column in orders_nodepath, continuing...")
	}

	// Convert user_id column from INT to CHAR(36) in device_setting_nodepath
	if err := convertDeviceUserIDToChar36(db); err != nil {
		logrus.WithError(err).Warn("Failed to convert user_id column in device_setting_nodepath, continuing...")
	}

	// Make device_id nullable for manual device creation
	if err := makeDeviceIDNullable(db); err != nil {
		logrus.WithError(err).Warn("Failed to make device_id nullable in device_setting_nodepath, continuing...")
	}

	logrus.Info("Database migrations completed successfully")
	return nil
}

// Migration SQL statements
const createFlowsTable = `
CREATE TABLE IF NOT EXISTS chatbot_flows_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT COLLATE utf8mb4_unicode_ci,
    niche TEXT COLLATE utf8mb4_unicode_ci,
    id_device VARCHAR(255),
    nodes JSON,
    edges JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// Test chat executions table schema removed

const createDeviceSettingsTable = `
CREATE TABLE IF NOT EXISTS device_setting_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    device_id VARCHAR(255), -- Changed to allow NULL for manual creation
    api_key_option ENUM('openai/gpt-5-chat', 'openai/gpt-5-mini', 'openai/chatgpt-4o-latest', 'openai/gpt-4.1', 'google/gemini-2.5-pro', 'google/gemini-pro-1.5') DEFAULT 'openai/gpt-4.1',
    webhook_id VARCHAR(500),
    provider ENUM('whacenter', 'wablas', 'waha') DEFAULT 'wablas',
    phone_number VARCHAR(20),
    api_key TEXT,
    id_device VARCHAR(255),
    id_erp VARCHAR(255),
    id_admin VARCHAR(255),
    user_id CHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// Create users table for authentication (matches existing schema)
const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id CHAR(36) NOT NULL PRIMARY KEY,
	email VARCHAR(255) NOT NULL,
	full_name VARCHAR(255) NOT NULL,
	password_hash VARCHAR(255) NOT NULL,
	is_active TINYINT(1) DEFAULT 1,
	created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP,
	last_login TIMESTAMP NULL DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// Create user_sessions table for authentication sessions (matches existing schema)
const createUserSessionsTable = `
CREATE TABLE IF NOT EXISTS user_sessions (
	id CHAR(36) NOT NULL PRIMARY KEY,
	user_id CHAR(36) NOT NULL,
	token VARCHAR(255) NOT NULL,
	expires_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	created_at TIMESTAMP NULL DEFAULT CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// AI WhatsApp conversation table for managing AI-powered WhatsApp conversations
const createAIWhatsappTable = `
CREATE TABLE IF NOT EXISTS ai_whatsapp_nodepath (
    id_prospect INT(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    flow_reference VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Reference to chatbot flow being executed',
    execution_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Unique execution identifier',
    date_order DATETIME DEFAULT NULL,
    id_device VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    niche VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    prospect_name VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    prospect_num VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    intro VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    stage VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    conv_last TEXT COLLATE utf8mb4_unicode_ci,
    conv_current TEXT COLLATE utf8mb4_unicode_ci,
    execution_status ENUM('active','completed','failed') COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Flow execution status',
    flow_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ID of the current chatbot flow being executed',
    current_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Current node ID in the chatbot flow',
    last_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Previous node ID for flow tracking',
    waiting_for_reply TINYINT(1) DEFAULT 0 COMMENT '1 = waiting for user reply, 0 = not waiting',
    balas VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    human INT(11) DEFAULT 0,
    keywordiklan VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    marketer VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    update_today DATETIME DEFAULT NULL,
    UNIQUE KEY uniq_execution_id (execution_id),
    KEY idx_flow_id (flow_id),
    KEY idx_current_node_id (current_node_id),
    KEY idx_id_device (id_device),
    KEY idx_prospect_num (prospect_num)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// WasapBot table for WasapBot Exama flow process
const createWasapBotTable = `
CREATE TABLE IF NOT EXISTS wasapBot_nodepath (
  id_prospect       INT(11) NOT NULL AUTO_INCREMENT,
  flow_reference    VARCHAR(255) COLLATE latin1_swedish_ci DEFAULT NULL,
  execution_id      VARCHAR(255) COLLATE latin1_swedish_ci DEFAULT NULL,
  execution_status  ENUM('active','completed','failed') COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Flow execution status',
  flow_id           VARCHAR(255) COLLATE latin1_swedish_ci DEFAULT NULL COMMENT 'ID of the current chatbot flow being executed',
  current_node_id   VARCHAR(255) COLLATE latin1_swedish_ci DEFAULT NULL COMMENT 'Current node ID in the chatbot flow',
  last_node_id      VARCHAR(255) COLLATE latin1_swedish_ci DEFAULT NULL COMMENT 'Previous node ID for flow tracking',
  waiting_for_reply TINYINT(1) DEFAULT 0 COMMENT '1 = waiting for user reply, 0 = not waiting',
  marketer_id       VARCHAR(100) COLLATE latin1_swedish_ci DEFAULT NULL,
  prospect_num      VARCHAR(100) COLLATE latin1_swedish_ci DEFAULT NULL,
  niche             VARCHAR(300) COLLATE latin1_swedish_ci DEFAULT NULL,
  instance          VARCHAR(255) COLLATE latin1_swedish_ci DEFAULT NULL,
  peringkat_sekolah VARCHAR(100) COLLATE latin1_swedish_ci DEFAULT NULL,
  alamat            VARCHAR(100) COLLATE latin1_swedish_ci DEFAULT NULL,
  nama              VARCHAR(100) COLLATE latin1_swedish_ci DEFAULT NULL,
  pakej             VARCHAR(100) COLLATE latin1_swedish_ci DEFAULT NULL,
  no_fon            VARCHAR(20)  COLLATE latin1_swedish_ci DEFAULT NULL,
  cara_bayaran      VARCHAR(100) COLLATE latin1_swedish_ci DEFAULT NULL,
  tarikh_gaji       VARCHAR(20)  COLLATE latin1_swedish_ci DEFAULT NULL,
  stage             VARCHAR(200) COLLATE latin1_swedish_ci DEFAULT NULL,
  temp_stage        VARCHAR(200) COLLATE latin1_swedish_ci DEFAULT NULL,
  conv_start        VARCHAR(200) COLLATE latin1_swedish_ci DEFAULT NULL,
  conv_last         TEXT         COLLATE latin1_swedish_ci,
  date_start        VARCHAR(50)  COLLATE latin1_swedish_ci DEFAULT NULL,
  date_last         VARCHAR(50)  COLLATE latin1_swedish_ci DEFAULT NULL,
  status            VARCHAR(200) COLLATE latin1_swedish_ci DEFAULT 'Prospek',
  staff_cls         VARCHAR(200) COLLATE latin1_swedish_ci DEFAULT NULL,
  umur              VARCHAR(200) COLLATE latin1_swedish_ci DEFAULT NULL,
  kerja             VARCHAR(200) COLLATE latin1_swedish_ci DEFAULT NULL,
  sijil             VARCHAR(200) COLLATE latin1_swedish_ci DEFAULT NULL,
  user_input        TEXT         COLLATE latin1_swedish_ci,
  alasan            VARCHAR(200) COLLATE latin1_swedish_ci DEFAULT NULL,
  nota              VARCHAR(200) COLLATE latin1_swedish_ci DEFAULT NULL,
  PRIMARY KEY (id_prospect),
  INDEX idx_prospect_num (prospect_num),
  INDEX idx_flow_id (flow_id),
  INDEX idx_execution_id (execution_id),
  INDEX idx_instance (instance)
) ENGINE=InnoDB DEFAULT CHARSET=latin1 COLLATE=latin1_swedish_ci;
`

// Conversation log table for storing all AI conversation history
const createConversationLogTable = `
CREATE TABLE IF NOT EXISTS conversation_log_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    prospect_num VARCHAR(20) NOT NULL,
    sender ENUM('user', 'bot', 'staff') NOT NULL,
    message TEXT COLLATE utf8mb4_unicode_ci NOT NULL,
    message_type ENUM('text', 'image', 'document', 'audio', 'video') DEFAULT 'text',
    stage VARCHAR(255),
    ai_response JSON COMMENT 'Full AI response with stage and content',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_prospect_num (prospect_num),
    INDEX idx_sender (sender),
    INDEX idx_created_at (created_at),
    INDEX idx_stage (stage)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// Orders table for Billplz payment integration
const createOrdersTable = `
CREATE TABLE IF NOT EXISTS orders_nodepath (
    id INT(11) NOT NULL AUTO_INCREMENT PRIMARY KEY,
    amount DECIMAL(10,2) NOT NULL COMMENT 'Amount in RM',
    collection_id VARCHAR(255) COLLATE utf8mb4_unicode_ci,
    status ENUM('Pending', 'Processing', 'Success', 'Failed') DEFAULT 'Pending',
    bill_id VARCHAR(255) COLLATE utf8mb4_unicode_ci,
    url TEXT COLLATE utf8mb4_unicode_ci COMMENT 'Billplz payment URL',
    product VARCHAR(255) COLLATE utf8mb4_unicode_ci NOT NULL,
    method VARCHAR(50) COLLATE utf8mb4_unicode_ci DEFAULT 'billplz',
    user_id CHAR(36),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_bill_id (bill_id),
    INDEX idx_status (status),
    INDEX idx_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const createAIWhatsappSessionTable = `
CREATE TABLE IF NOT EXISTS ai_whatsapp_session_nodepath (
    id_sessionX INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    phone_number VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    locked_at TIMESTAMP NULL DEFAULT NULL,
    last_activity TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ` + "`timestamp`" + ` VARCHAR(255) NOT NULL,
    UNIQUE KEY uniq_ai_whatsapp_session (phone_number, device_id),
    KEY idx_ai_whatsapp_session_device (device_id),
    KEY idx_ai_session_locked (locked_at),
    KEY idx_ai_session_activity (last_activity)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const createWasapBotSessionTable = `
CREATE TABLE IF NOT EXISTS wasapBot_session_nodepath (
    id_sessionY INT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    id_prospect VARCHAR(255) NOT NULL,
    id_device VARCHAR(255) NOT NULL,
    ` + "`timestamp`" + ` VARCHAR(255) NOT NULL,
    UNIQUE KEY uniq_wasapbot_session (id_prospect, id_device),
    KEY idx_wasapbot_session_device (id_device)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// addMissingColumnsToFlowsTable adds missing columns to the flows table
func addMissingColumnsToFlowsTable(db *sql.DB) error {
	columns := []struct {
		name       string
		definition string
	}{
		{"niche", "TEXT COLLATE utf8mb4_unicode_ci"},
		{"id_device", "VARCHAR(255)"},
	}

	for _, col := range columns {
		// Check if column exists
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'chatbot_flows_nodepath' 
			AND COLUMN_NAME = ?
		`, col.name).Scan(&count)

		if err != nil {
			return fmt.Errorf("failed to check column %s: %w", col.name, err)
		}

		if count == 0 {
			// Column doesn't exist, add it
			query := fmt.Sprintf("ALTER TABLE chatbot_flows_nodepath ADD COLUMN %s %s", col.name, col.definition)
			if _, err := db.Exec(query); err != nil {
				return fmt.Errorf("failed to add column %s: %w", col.name, err)
			}
			logrus.WithField("column", col.name).Info("Added missing column")
		} else {
			logrus.WithField("column", col.name).Debug("Column already exists")
		}
	}
	return nil
}

// updateProviderRvsbWasapToWaha updates provider values from 'rvsb_wasap' to 'waha'
func updateProviderRvsbWasapToWaha(db *sql.DB) error {
	// Update existing records that have 'rvsb_wasap' provider to 'waha'
	result, err := db.Exec("UPDATE device_setting_nodepath SET provider = 'waha' WHERE provider = 'rvsb_wasap'")
	if err != nil {
		return fmt.Errorf("failed to update provider values: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		logrus.WithField("rows_updated", rowsAffected).Info("Updated provider values from 'rvsb_wasap' to 'waha'")
	} else {
		logrus.Debug("No records found with 'rvsb_wasap' provider to update")
	}

	return nil
}

// updateProviderEnum updates the provider ENUM to include 'waha' and remove 'rvsb_wasap'
func updateProviderEnum(db *sql.DB) error {
	logrus.Info("🔧 Checking provider ENUM constraint in device_setting_nodepath table")

	// Check if table exists first
	var tableExists int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'device_setting_nodepath'
	`).Scan(&tableExists)

	if err != nil {
		logrus.WithError(err).Warn("Failed to check if device_setting_nodepath table exists, skipping provider ENUM update")
		return nil
	}

	if tableExists == 0 {
		logrus.Info("device_setting_nodepath table doesn't exist yet, skipping provider ENUM update")
		return nil
	}

	// Check current ENUM values for provider column
	var columnType string
	err = db.QueryRow(`
		SELECT COLUMN_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'device_setting_nodepath' 
		AND COLUMN_NAME = 'provider'
	`).Scan(&columnType)

	if err != nil {
		logrus.WithError(err).Warn("Failed to check provider column type, attempting to alter anyway")
	} else {
		logrus.WithField("current_enum", columnType).Info("Current provider ENUM constraint")

		// If already has 'waha' and doesn't have 'rvsb_wasap', skip
		if contains(columnType, "'waha'") && !contains(columnType, "'rvsb_wasap'") {
			logrus.Info("✅ Provider ENUM already has 'waha' and doesn't have 'rvsb_wasap' - no update needed")
			return nil
		}
	}

	// Update the ENUM to replace 'rvsb_wasap' with 'waha'
	logrus.Info("🔧 Updating provider ENUM to include 'waha' and remove 'rvsb_wasap'")
	_, err = db.Exec("ALTER TABLE device_setting_nodepath MODIFY COLUMN provider ENUM('whacenter', 'wablas', 'waha') DEFAULT 'wablas'")
	if err != nil {
		logrus.WithError(err).Error("❌ Failed to update provider ENUM - this will cause WAHA provider issues")
		return fmt.Errorf("failed to update provider ENUM: %w", err)
	}

	logrus.Info("✅ Successfully updated provider ENUM to support WAHA provider")

	// Verify the change was applied
	err = db.QueryRow(`
		SELECT COLUMN_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'device_setting_nodepath' 
		AND COLUMN_NAME = 'provider'
	`).Scan(&columnType)

	if err == nil {
		logrus.WithField("updated_enum", columnType).Info("✅ Verified provider ENUM after migration")
	}

	return nil
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || (len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && s[1:len(substr)+1] == substr ||
			findInString(s, substr))))
}

// findInString is a simple string search helper
func findInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// removeDeprecatedColumnsFromFlowsTable removes deprecated columns from the flows table
func removeDeprecatedColumnsFromFlowsTable(db *sql.DB) error {
	columns := []string{
		"global_instance",
		"global_open_router_key",
	}

	for _, col := range columns {
		// Check if column exists
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'chatbot_flows_nodepath' 
			AND COLUMN_NAME = ?
		`, col).Scan(&count)

		if err != nil {
			return fmt.Errorf("failed to check column %s: %w", col, err)
		}

		if count > 0 {
			// Column exists, drop it
			query := fmt.Sprintf("ALTER TABLE chatbot_flows_nodepath DROP COLUMN %s", col)
			if _, err := db.Exec(query); err != nil {
				return fmt.Errorf("failed to drop column %s: %w", col, err)
			}
			logrus.WithField("column", col).Info("Removed deprecated column")
		} else {
			logrus.WithField("column", col).Debug("Deprecated column does not exist")
		}
	}
	return nil
}

// addMissingColumnsToDeviceSettingsTable adds missing columns to the device settings table
func addMissingColumnsToDeviceSettingsTable(db *sql.DB) error {
	// Define columns that should exist
	columns := []struct {
		name       string
		definition string
	}{
		{"phone_number", "VARCHAR(20)"},
		{"instance", "TEXT"},
		{"user_id", "INT"},
	}

	for _, col := range columns {
		// Check if column exists
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'device_setting_nodepath' 
			AND COLUMN_NAME = ?
		`, col.name).Scan(&count)

		if err != nil {
			return fmt.Errorf("failed to check column %s: %w", col.name, err)
		}

		if count == 0 {
			// Column doesn't exist, add it
			query := fmt.Sprintf("ALTER TABLE device_setting_nodepath ADD COLUMN %s %s", col.name, col.definition)
			if _, err := db.Exec(query); err != nil {
				return fmt.Errorf("failed to add column %s: %w", col.name, err)
			}
			logrus.WithField("column", col.name).Info("Added missing column to device_setting_nodepath")
		} else {
			logrus.WithField("column", col.name).Debug("Column already exists in device_setting_nodepath")
		}
	}

	return nil
}

// dropDeprecatedBillingColumns drops deprecated billing columns from orders_nodepath table
func dropDeprecatedBillingColumns(db *sql.DB) error {
	columns := []string{
		"customer_email",
		"customer_name",
		"billing_phone",
		"billing_address",
		"billing_city",
		"billing_state",
		"billing_postcode",
	}

	for _, col := range columns {
		// Check if column exists
		var count int
		err := db.QueryRow(`
			SELECT COUNT(*) 
			FROM INFORMATION_SCHEMA.COLUMNS 
			WHERE TABLE_SCHEMA = DATABASE() 
			AND TABLE_NAME = 'orders_nodepath' 
			AND COLUMN_NAME = ?
		`, col).Scan(&count)

		if err != nil {
			return fmt.Errorf("failed to check column %s: %w", col, err)
		}

		if count > 0 {
			// Column exists, drop it
			query := fmt.Sprintf("ALTER TABLE orders_nodepath DROP COLUMN %s", col)
			if _, err := db.Exec(query); err != nil {
				return fmt.Errorf("failed to drop column %s: %w", col, err)
			}
			logrus.WithField("column", col).Info("Dropped deprecated billing column from orders_nodepath")
		} else {
			logrus.WithField("column", col).Debug("Deprecated billing column does not exist in orders_nodepath")
		}
	}
	return nil
}

// convertUserIDToChar36 converts user_id column from INT to CHAR(36) in orders_nodepath table
func convertUserIDToChar36(db *sql.DB) error {
	// Check current data type of user_id column
	var dataType string
	err := db.QueryRow(`
		SELECT DATA_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'orders_nodepath' 
		AND COLUMN_NAME = 'user_id'
	`).Scan(&dataType)

	if err != nil {
		return fmt.Errorf("failed to check user_id column type: %w", err)
	}

	// If already CHAR, skip
	if dataType == "char" {
		logrus.Info("user_id column is already CHAR(36) in orders_nodepath")
		return nil
	}

	// Convert INT to CHAR(36) - first set existing data to NULL or empty since INT can't convert to UUID
	_, err = db.Exec("UPDATE orders_nodepath SET user_id = NULL WHERE user_id IS NOT NULL")
	if err != nil {
		logrus.WithError(err).Warn("Failed to clear user_id values before conversion")
	}

	// Alter column type
	_, err = db.Exec("ALTER TABLE orders_nodepath MODIFY COLUMN user_id CHAR(36)")
	if err != nil {
		return fmt.Errorf("failed to convert user_id to CHAR(36): %w", err)
	}

	logrus.Info("Successfully converted user_id column from INT to CHAR(36) in orders_nodepath")
	return nil
}

// convertDeviceUserIDToChar36 converts user_id column from INT to CHAR(36) in device_setting_nodepath table
func convertDeviceUserIDToChar36(db *sql.DB) error {
	// Check current data type of user_id column
	var dataType string
	err := db.QueryRow(`
		SELECT DATA_TYPE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'device_setting_nodepath' 
		AND COLUMN_NAME = 'user_id'
	`).Scan(&dataType)

	if err != nil {
		return fmt.Errorf("failed to check user_id column type: %w", err)
	}

	// If already CHAR, skip
	if dataType == "char" {
		logrus.Info("user_id column is already CHAR(36) in device_setting_nodepath")
		return nil
	}

	logrus.Warn("⚠️  CRITICAL: user_id column needs conversion from INT to CHAR(36)")
	logrus.Warn("⚠️  This will set all user_id values to NULL!")
	logrus.Warn("⚠️  You MUST run the fix script: scripts/fix_device_user_links.sql")
	logrus.Warn("⚠️  Or manually re-link devices to users after migration")

	// Convert INT to CHAR(36) - WARNING: This sets existing data to NULL since INT can't convert to UUID
	// Users will need to run fix script or manually re-link devices
	_, err = db.Exec("UPDATE device_setting_nodepath SET user_id = NULL WHERE user_id IS NOT NULL")
	if err != nil {
		logrus.WithError(err).Warn("Failed to clear user_id values before conversion in device_setting_nodepath")
	}

	// Alter column type
	_, err = db.Exec("ALTER TABLE device_setting_nodepath MODIFY COLUMN user_id CHAR(36)")
	if err != nil {
		return fmt.Errorf("failed to convert user_id to CHAR(36) in device_setting_nodepath: %w", err)
	}

	// Add index if it doesn't exist
	_, err = db.Exec("ALTER TABLE device_setting_nodepath ADD INDEX IF NOT EXISTS idx_user_id (user_id)")
	if err != nil {
		logrus.WithError(err).Warn("Failed to add index to user_id column")
	}

	logrus.Info("✅ Converted user_id column from INT to CHAR(36) in device_setting_nodepath")
	logrus.Warn("⚠️  ACTION REQUIRED: Run scripts/fix_device_user_links.sql to restore device-user links")
	return nil
}

// makeDeviceIDNullable makes device_id column nullable to allow manual device creation
func makeDeviceIDNullable(db *sql.DB) error {
	logrus.Info("🔧 Checking device_id column nullability in device_setting_nodepath table")

	// Check if table exists first
	var tableExists int
	err := db.QueryRow(`
		SELECT COUNT(*) 
		FROM INFORMATION_SCHEMA.TABLES 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'device_setting_nodepath'
	`).Scan(&tableExists)

	if err != nil {
		logrus.WithError(err).Warn("Failed to check if device_setting_nodepath table exists, skipping migration")
		return nil // Don't fail the entire migration for this
	}

	if tableExists == 0 {
		logrus.Info("device_setting_nodepath table doesn't exist yet, skipping device_id nullable migration")
		return nil
	}

	// Check if device_id column allows NULL
	var isNullable string
	err = db.QueryRow(`
		SELECT IS_NULLABLE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'device_setting_nodepath' 
		AND COLUMN_NAME = 'device_id'
	`).Scan(&isNullable)

	if err != nil {
		logrus.WithError(err).Warn("Failed to check device_id column nullability, attempting to alter anyway")
		// Continue and try to alter, might be a column that doesn't exist yet
	} else {
		// If already nullable, skip
		if isNullable == "YES" {
			logrus.Info("✅ device_id column is already nullable in device_setting_nodepath")
			return nil
		}
		logrus.WithField("is_nullable", isNullable).Info("device_id column current nullability status")
	}

	// Make column nullable - this is the critical fix
	logrus.Info("🔧 Altering device_id column to allow NULL values")
	_, err = db.Exec("ALTER TABLE device_setting_nodepath MODIFY COLUMN device_id VARCHAR(255) NULL")
	if err != nil {
		logrus.WithError(err).Error("❌ Failed to make device_id nullable - this will cause WAHA provider issues")
		return fmt.Errorf("failed to make device_id nullable: %w", err)
	}

	logrus.Info("✅ Successfully made device_id column nullable in device_setting_nodepath for manual device creation")

	// Verify the change was applied
	err = db.QueryRow(`
		SELECT IS_NULLABLE 
		FROM INFORMATION_SCHEMA.COLUMNS 
		WHERE TABLE_SCHEMA = DATABASE() 
		AND TABLE_NAME = 'device_setting_nodepath' 
		AND COLUMN_NAME = 'device_id'
	`).Scan(&isNullable)

	if err == nil {
		logrus.WithField("is_nullable_after", isNullable).Info("✅ Verified device_id column nullability after migration")
	}

	return nil
}
