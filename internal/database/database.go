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
    device_id VARCHAR(255) NOT NULL,
    api_key_option ENUM('openai/gpt-5-chat', 'openai/gpt-5-mini', 'openai/chatgpt-4o-latest', 'openai/gpt-4.1', 'google/gemini-2.5-pro', 'google/gemini-pro-1.5') DEFAULT 'openai/gpt-4.1',
    webhook_id VARCHAR(500),
    provider ENUM('whacenter', 'wablas', 'waha') DEFAULT 'wablas',
    phone_number VARCHAR(20),
    api_key TEXT,
    id_device VARCHAR(255),
    id_erp VARCHAR(255),
    id_admin VARCHAR(255),
    user_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
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
    user_id INT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_bill_id (bill_id),
    INDEX idx_status (status),
    INDEX idx_user_id (user_id)
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
