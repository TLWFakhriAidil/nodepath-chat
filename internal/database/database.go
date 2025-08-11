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
		createNodeManualTable,
		createAISettingsTable,
		createLeadsTable,
		createMediaFilesTable,
		addMissingColumns,
	}

	for i, migration := range migrations {
		logrus.WithField("migration", i+1).Debug("Running migration")
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration %d: %w", i+1, err)
		}
	}

	logrus.Info("Database migrations completed successfully")
	return nil
}

// Migration SQL statements
const createFlowsTable = `
CREATE TABLE IF NOT EXISTS chatbot_flows_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    flow_id VARCHAR(255) NOT NULL,
    node_id VARCHAR(255) NOT NULL,
    node_type VARCHAR(50) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT COLLATE utf8mb4_unicode_ci,
    system_prompt TEXT COLLATE utf8mb4_unicode_ci,
    instance VARCHAR(255),
    apiprovider VARCHAR(255),
    global_instance VARCHAR(255),
    global_open_router_key VARCHAR(500),
    mode VARCHAR(50) DEFAULT 'standard',
    nodes JSON,
    edges JSON,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_flow_id (flow_id),
    INDEX idx_node_type (node_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const createExecutionsTable = `
CREATE TABLE IF NOT EXISTS chatbot_executions_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    flow_id VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    staff_id VARCHAR(255),
    conv_last JSON,
    conv_current TEXT COLLATE utf8mb4_unicode_ci,
    current_node VARCHAR(255),
    variables JSON,
    status ENUM('active', 'completed', 'failed') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_flow_id (flow_id),
    INDEX idx_phone_number (phone_number),
    INDEX idx_staff_id (staff_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const createNodeManualTable = `
CREATE TABLE IF NOT EXISTS chatbot_node_manual (
    id VARCHAR(255) PRIMARY KEY,
    flow_id VARCHAR(255) NOT NULL,
    node_id VARCHAR(255) NOT NULL,
    message TEXT COLLATE utf8mb4_unicode_ci,
    media_type ENUM('text', 'image', 'audio', 'video') DEFAULT 'text',
    media_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_flow_id (flow_id),
    INDEX idx_node_id (node_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const createAISettingsTable = `
CREATE TABLE IF NOT EXISTS ai_setting_nodepath (
    id VARCHAR(255) PRIMARY KEY,
    flow_id VARCHAR(255) NOT NULL,
    node_id VARCHAR(255) NOT NULL,
    system_prompt TEXT COLLATE utf8mb4_unicode_ci,
    instance VARCHAR(255),
    apiprovider VARCHAR(255),
    model VARCHAR(100) DEFAULT 'openai/gpt-4.1',
    temperature DECIMAL(3,2) DEFAULT 0.7,
    max_tokens INT DEFAULT 1000,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_flow_id (flow_id),
    INDEX idx_node_id (node_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const createLeadsTable = `
CREATE TABLE IF NOT EXISTS chatbot_leads (
    id VARCHAR(255) PRIMARY KEY,
    phone_number VARCHAR(20) NOT NULL,
    staff_id VARCHAR(255) NOT NULL,
    name VARCHAR(255),
    email VARCHAR(255),
    status ENUM('new', 'contacted', 'qualified', 'converted') DEFAULT 'new',
    source VARCHAR(100) DEFAULT 'whatsapp',
    notes TEXT COLLATE utf8mb4_unicode_ci,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    UNIQUE KEY unique_phone_staff (phone_number, staff_id),
    INDEX idx_phone_number (phone_number),
    INDEX idx_staff_id (staff_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const createMediaFilesTable = `
CREATE TABLE IF NOT EXISTS media_files (
    id VARCHAR(255) PRIMARY KEY,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255),
    mime_type VARCHAR(100),
    size BIGINT,
    url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_filename (filename),
    INDEX idx_mime_type (mime_type)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const addMissingColumns = `
ALTER TABLE chatbot_flows_nodepath 
ADD COLUMN IF NOT EXISTS global_instance VARCHAR(255),
ADD COLUMN IF NOT EXISTS global_open_router_key VARCHAR(500),
ADD COLUMN IF NOT EXISTS mode VARCHAR(50) DEFAULT 'standard';
`