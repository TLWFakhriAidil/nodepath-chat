package database

import (
	"database/sql"
	"fmt"
	"time"

	"nodepath-chat/internal/config"

	_ "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
)

// Initialize creates and returns a database connection
func Initialize(cfg *config.Config) (*sql.DB, error) {
	dsn := cfg.GetDSN()
	logrus.WithField("host", cfg.MySQLHost).Info("Connecting to MySQL database")

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

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

	migrations := []string{
		createFlowsTable,
		createExecutionsTable,
		createDeviceSettingsTable,
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

const createExecutionsTable = `
CREATE TABLE IF NOT EXISTS chatbot_executions_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    flow_reference VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    staff_id VARCHAR(255),
    conv_last JSON,
    conv_current TEXT COLLATE utf8mb4_unicode_ci,
    current_node VARCHAR(255),
    variables JSON,
    status ENUM('active', 'completed', 'failed') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_flow_reference (flow_reference),
    INDEX idx_phone_number (phone_number),
    INDEX idx_staff_id (staff_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

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