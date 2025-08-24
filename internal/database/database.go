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
	db.SetMaxOpenConns(500)  // Increased significantly for 3000+ concurrent users
	db.SetMaxIdleConns(100)  // Higher idle connections to reduce connection overhead
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
		createAIWhatsappTable,
		createConversationLogTable,
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
    provider ENUM('whacenter', 'wablas', 'rvsb_wasap') DEFAULT 'wablas',
    phone_number VARCHAR(20),
    api_key TEXT,
    id_device VARCHAR(255),
    id_erp VARCHAR(255),
    id_admin VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

// AI WhatsApp conversation table for managing AI-powered WhatsApp conversations
const createAIWhatsappTable = `
CREATE TABLE IF NOT EXISTS ai_whatsapp_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    id_prospect INT NOT NULL,
    id_device VARCHAR(255) NOT NULL,
    prospect_num VARCHAR(20) NOT NULL,
    stage VARCHAR(255) DEFAULT 'initial',
    date_order TIMESTAMP NULL,
    conv_last TEXT COLLATE utf8mb4_unicode_ci,
    conv_current TEXT COLLATE utf8mb4_unicode_ci,
    jam VARCHAR(255) DEFAULT NULL COMMENT 'Timestamp field for conversation tracking',
    intro VARCHAR(255) DEFAULT NULL COMMENT 'Introduction field for conversation flow',
    human TINYINT(1) DEFAULT 0 COMMENT '0=AI active, 1=human takeover',
    catatan_staff TEXT COLLATE utf8mb4_unicode_ci,
    balas TINYINT(1) DEFAULT 1 COMMENT '1=bot replies, 0=no bot reply',
    data_image TEXT COLLATE utf8mb4_unicode_ci,
    conv_stage VARCHAR(255),
    niche VARCHAR(255),
    bot_balas TIMESTAMP NULL,
    keywordiklan VARCHAR(255),
    marketer VARCHAR(255),
    update_today TIMESTAMP NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_prospect_num (prospect_num),
    INDEX idx_id_device (id_device),
    INDEX idx_stage (stage),
    INDEX idx_human (human),
    UNIQUE KEY unique_prospect_device (prospect_num, id_device)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
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



// addMissingColumnsToFlowsTable adds missing columns to the flows table
func addMissingColumnsToFlowsTable(db *sql.DB) error {
	columns := []struct {
		name string
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