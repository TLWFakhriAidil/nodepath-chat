package database

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
	
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
)

var (
	db   *sql.DB
	once sync.Once
)

// GetDB returns the database connection
func GetDB() *sql.DB {
	once.Do(func() {
		var err error
		
		// Check for MySQL configuration first
		mysqlURI := os.Getenv("MYSQL_URI")
		if mysqlURI == "" {
			mysqlURI = os.Getenv("DB_URI")
		}
		
		if mysqlURI != "" && strings.Contains(mysqlURI, "mysql") {
			// Use MySQL for application data
			log.Println("Using MySQL for application data")
			
			// Convert mysql:// URL to DSN format if needed
			if strings.HasPrefix(mysqlURI, "mysql://") {
				mysqlURI = strings.TrimPrefix(mysqlURI, "mysql://")
				parts := strings.Split(mysqlURI, "@")
				if len(parts) == 2 {
					userPass := parts[0]
					hostDb := parts[1]
					mysqlURI = userPass + "@tcp(" + strings.Replace(hostDb, "/", ")/", 1) + "?parseTime=true&charset=utf8mb4"
				}
			}
			
			db, err = sql.Open("mysql", mysqlURI)
		} else {
			// Fallback to PostgreSQL
			log.Println("Using PostgreSQL for application data")
			db, err = sql.Open("postgres", config.DBURI)
		}
		
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		
		// Configure connection pool for 200+ users
		// Optimized for 3000 devices
		db.SetMaxOpenConns(500)     // Increased from 100
		db.SetMaxIdleConns(100)     // Increased from 10  
		db.SetConnMaxLifetime(5 * time.Minute)  // Reduced from 1 hour
		
		// Test connection
		if err := db.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}
		
		// Initialize schema - SKIP MIGRATIONS
		// Uncomment the lines below to enable migrations
		/*
		if err := InitializeSchema(); err != nil {
			log.Fatalf("Failed to initialize schema: %v", err)
		}
		*/
		log.Println("Skipping schema migrations as requested")
		
		log.Println("Database connection established")
	})
	
	return db
}
// InitializeSchema creates tables if they don't exist
func InitializeSchema() error {
	schema := `
	-- Create extension for UUID generation if not exists
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

	-- Create users table
	CREATE TABLE IF NOT EXISTS users (
		id UUID PRIMARY KEY DEFAULT UUID(),
		email VARCHAR(255) UNIQUE NOT NULL,
		full_name VARCHAR(255) NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		is_active BOOLEAN DEFAULT true,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		last_login TIMESTAMP
	);

	-- Create user_devices table
	CREATE TABLE IF NOT EXISTS user_devices (
		id UUID PRIMARY KEY DEFAULT UUID(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		device_name VARCHAR(255) NOT NULL,
		phone VARCHAR(50),
		jid VARCHAR(255),
		status VARCHAR(50) DEFAULT 'offline',
		last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(user_id, jid)
	);

	-- Create user_sessions table
	CREATE TABLE IF NOT EXISTS user_sessions (
		id UUID PRIMARY KEY DEFAULT UUID(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		token VARCHAR(255) UNIQUE NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create message_analytics table
	CREATE TABLE IF NOT EXISTS message_analytics (
		id UUID PRIMARY KEY DEFAULT UUID(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		device_id UUID REFERENCES user_devices(id) ON DELETE SET NULL,
		message_id VARCHAR(255) NOT NULL,
		jid VARCHAR(255) NOT NULL,
		content TEXT,
		is_from_me BOOLEAN DEFAULT false,
		status VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(message_id)
	);

	-- Create leads table
	CREATE TABLE IF NOT EXISTS leads (
		id SERIAL PRIMARY KEY,
		device_id UUID NOT NULL,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		phone VARCHAR(50) NOT NULL,
		niche VARCHAR(255),
		journey TEXT,
		status VARCHAR(50) DEFAULT 'new',
		last_interaction TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create campaigns table
	CREATE TABLE IF NOT EXISTS campaigns (
		id SERIAL PRIMARY KEY,
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		campaign_date DATE NOT NULL,
		title VARCHAR(255) NOT NULL,
		niche VARCHAR(255),
		message TEXT NOT NULL,
		image_url TEXT,
		time_schedule TEXT,
		min_delay_seconds INTEGER DEFAULT 10,
		max_delay_seconds INTEGER DEFAULT 30,
		status VARCHAR(50) DEFAULT 'scheduled',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Remove old unique constraint if it exists
	ALTER TABLE campaigns DROP CONSTRAINT IF EXISTS campaigns_user_id_campaign_date_key;
	
	-- Add missing columns to campaigns if they don't exist
	ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS device_id UUID;
	ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS niche VARCHAR(255);
	ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS image_url TEXT;
	ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS target_status VARCHAR(50) DEFAULT 'all';
	ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS time_schedule TEXT;
	ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'scheduled';
	ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10;
	ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30;
	
	-- Add min/max delay columns to user_devices
	ALTER TABLE user_devices ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 5;
	ALTER TABLE user_devices ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 15;

	-- Create whatsapp_chats table to store chat list
	CREATE TABLE IF NOT EXISTS whatsapp_chats (
		id SERIAL PRIMARY KEY,
		device_id VARCHAR(255) NOT NULL,
		chat_jid VARCHAR(255) NOT NULL,
		chat_name VARCHAR(255) NOT NULL,
		is_group BOOLEAN DEFAULT 0,
		is_muted BOOLEAN DEFAULT 0,
		last_message_text TEXT,
		last_message_time TIMESTAMP,
		unread_count INTEGER DEFAULT 0,
		avatar_url TEXT,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(device_id, chat_jid)
	);

	-- Create whatsapp_messages table to store message history
	CREATE TABLE IF NOT EXISTS whatsapp_messages (
		id SERIAL PRIMARY KEY,
		device_id VARCHAR(255) NOT NULL,
		chat_jid VARCHAR(255) NOT NULL,
		message_id VARCHAR(255) NOT NULL,
		sender_jid VARCHAR(255),
		sender_name VARCHAR(255),
		message_text TEXT,
		message_type VARCHAR(50),
		media_url TEXT,
		is_sent BOOLEAN DEFAULT 0,
		is_read BOOLEAN DEFAULT 0,
		timestamp TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(device_id, message_id)
	);
	
	-- Create sequences table
	CREATE TABLE IF NOT EXISTS sequences (
		id UUID PRIMARY KEY DEFAULT UUID(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		niche VARCHAR(255),
		time_schedule TEXT,
		min_delay_seconds INTEGER DEFAULT 10,
		max_delay_seconds INTEGER DEFAULT 30,
		status VARCHAR(50) DEFAULT 'draft',
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
	
	-- Create sequence_steps table
	CREATE TABLE IF NOT EXISTS sequence_steps (
		id UUID PRIMARY KEY DEFAULT UUID(),
		sequence_id UUID NOT NULL REFERENCES sequences(id) ON DELETE CASCADE,
		day_number INTEGER NOT NULL,
		content TEXT,
		image_url TEXT,
		time_schedule TEXT,
		min_delay_seconds INTEGER DEFAULT 5,
		max_delay_seconds INTEGER DEFAULT 15,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(sequence_id, day_number)
	);
	
	-- Create sequence_contacts table
	CREATE TABLE IF NOT EXISTS sequence_contacts (
		id UUID PRIMARY KEY DEFAULT UUID(),
		sequence_id UUID NOT NULL REFERENCES sequences(id) ON DELETE CASCADE,
		contact_phone VARCHAR(50) NOT NULL,
		contact_name VARCHAR(255),
		current_step INTEGER DEFAULT 0,
		status VARCHAR(50) DEFAULT 'active',
		completed_at TIMESTAMP,
		current_trigger VARCHAR(255),
		next_trigger_time TIMESTAMP,
		processing_device_id UUID,
		last_error TEXT,
		retry_count INTEGER DEFAULT 0,
		assigned_device_id UUID,
		processing_started_at TIMESTAMP,
		sequence_stepid UUID,
		user_id UUID,
		UNIQUE(sequence_id, contact_phone)
	);
	
	-- Create broadcast_messages table for tracking
	CREATE TABLE IF NOT EXISTS broadcast_messages (
		id UUID PRIMARY KEY DEFAULT UUID(),
		user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		device_id UUID NOT NULL REFERENCES user_devices(id) ON DELETE CASCADE,
		campaign_id INTEGER REFERENCES campaigns(id) ON DELETE SET NULL,
		sequence_id UUID REFERENCES sequences(id) ON DELETE SET NULL,
		recipient_phone VARCHAR(50) NOT NULL,
		message_type VARCHAR(50) NOT NULL,
		content TEXT,
		media_url TEXT,
		status VARCHAR(50) DEFAULT 'pending',
		error_message TEXT,
		scheduled_at TIMESTAMP,
		sent_at TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		group_id VARCHAR(255),
		group_order INTEGER
	);

	-- Create indexes
	CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	CREATE INDEX IF NOT EXISTS idx_user_devices_user_id ON user_devices(user_id);
	CREATE INDEX IF NOT EXISTS idx_user_sessions_token ON user_sessions(token);
	CREATE INDEX IF NOT EXISTS idx_message_analytics_user_id ON message_analytics(user_id);
	CREATE INDEX IF NOT EXISTS idx_message_analytics_created_at ON message_analytics(created_at);
	CREATE INDEX IF NOT EXISTS idx_leads_device_id ON leads(device_id);
	CREATE INDEX IF NOT EXISTS idx_leads_user_id ON leads(user_id);
	CREATE INDEX IF NOT EXISTS idx_campaigns_user_id ON campaigns(user_id);
	CREATE INDEX IF NOT EXISTS idx_campaigns_date ON campaigns(campaign_date);
	CREATE INDEX IF NOT EXISTS idx_whatsapp_chats_device_id ON whatsapp_chats(device_id);
	CREATE INDEX IF NOT EXISTS idx_whatsapp_chats_updated ON whatsapp_chats(updated_at DESC);
	CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_device_chat ON whatsapp_messages(device_id, chat_jid);
	CREATE INDEX IF NOT EXISTS idx_whatsapp_messages_timestamp ON whatsapp_messages(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_sequences_user_id ON sequences(user_id);
	CREATE INDEX IF NOT EXISTS idx_sequences_status ON sequences(status);
	CREATE INDEX IF NOT EXISTS idx_sequence_steps_sequence_id ON sequence_steps(sequence_id);
	CREATE INDEX IF NOT EXISTS idx_sequence_contacts_sequence_id ON sequence_contacts(sequence_id);
	CREATE INDEX IF NOT EXISTS idx_sequence_contacts_next_send ON sequence_contacts(next_trigger_time);
	CREATE INDEX IF NOT EXISTS idx_broadcast_messages_status ON broadcast_messages(status);
	CREATE INDEX IF NOT EXISTS idx_broadcast_messages_scheduled ON broadcast_messages(scheduled_at);
	`
	
	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}
	
	// SKIP ALTER SCHEMA - Database already has correct structure
	// Commented out on January 8, 2025 - Prevents conflicting ADD/DROP column operations
	/*
	// Add missing columns for sequences (simplified version compatibility)
	alterSchema := `
	-- Add missing columns to sequences table
	ALTER TABLE sequences ADD COLUMN IF NOT EXISTS device_id UUID;
	ALTER TABLE sequences ADD COLUMN IF NOT EXISTS total_days INTEGER DEFAULT 0;
	ALTER TABLE sequences ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;
	ALTER TABLE sequences ADD COLUMN IF NOT EXISTS time_schedule TEXT;
	
	-- Add missing columns to broadcast_messages if they don't exist
	ALTER TABLE broadcast_messages ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;
	ALTER TABLE sequences ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10;
	ALTER TABLE sequences ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30;
	
	-- Make device_id nullable since sequences use all user devices
	ALTER TABLE sequences ALTER COLUMN device_id DROP NOT NULL;
	
	-- Add `trigger` columns to sequences table
	ALTER TABLE sequences ADD COLUMN IF NOT EXISTS start_trigger VARCHAR(255);
	ALTER TABLE sequences ADD COLUMN IF NOT EXISTS end_trigger VARCHAR(255);

	-- Remove specific columns from sequence_steps table
	ALTER TABLE sequence_steps DROP COLUMN IF EXISTS day CASCADE;
	ALTER TABLE sequence_steps DROP COLUMN IF EXISTS send_time CASCADE;
	ALTER TABLE sequence_steps DROP COLUMN IF EXISTS updated_at CASCADE;

	-- Fix sequence_steps table by removing other timestamp columns
	ALTER TABLE sequence_steps DROP COLUMN IF EXISTS created_at CASCADE;
	ALTER TABLE sequence_steps DROP COLUMN IF EXISTS schedule_time CASCADE;

	-- Add missing columns to sequence_steps table
	ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS message_type VARCHAR(50) DEFAULT 'text';
	ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS media_url TEXT;
	ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS caption TEXT;
	ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS `trigger` VARCHAR(255) DEFAULT '';
	ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS next_trigger VARCHAR(255) DEFAULT '';
	ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger_delay_hours INTEGER DEFAULT 24;
	ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS is_entry_point BOOLEAN DEFAULT false;
	ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS min_delay_seconds INTEGER DEFAULT 10;
	ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS max_delay_seconds INTEGER DEFAULT 30;
	ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS delay_days INTEGER DEFAULT 0;

	-- Add missing columns to sequence_contacts table  
	ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS current_trigger VARCHAR(255) DEFAULT '';
	ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS next_trigger_time TIMESTAMP;
	ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS processing_device_id UUID;
	ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS assigned_device_id UUID;
	ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS last_error TEXT;
	ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS retry_count INTEGER DEFAULT 0;

	-- Remove unnecessary columns from sequence_contacts
	ALTER TABLE sequence_contacts DROP COLUMN IF EXISTS enrolled_at CASCADE;
	ALTER TABLE sequence_contacts DROP COLUMN IF EXISTS last_sent_at CASCADE;
	ALTER TABLE sequence_contacts DROP COLUMN IF EXISTS next_trigger_time CASCADE;
	ALTER TABLE sequence_contacts DROP COLUMN IF EXISTS current_day CASCADE;
	ALTER TABLE sequence_contacts DROP COLUMN IF EXISTS added_at CASCADE;
	ALTER TABLE sequence_contacts DROP COLUMN IF EXISTS last_message_at CASCADE;
	ALTER TABLE sequence_contacts DROP COLUMN IF EXISTS processing_started_at CASCADE;

	-- Add indexes for 3000 device optimization
	CREATE INDEX IF NOT EXISTS idx_sc_active_trigger ON sequence_contacts(status, next_trigger_time) 
	WHERE `status` = 'active' AND processing_device_id IS NULL;
	
	CREATE INDEX IF NOT EXISTS idx_sc_current_trigger ON sequence_contacts(current_trigger);
	
	CREATE INDEX IF NOT EXISTS idx_ss_sequence_trigger ON sequence_steps(sequence_id, `trigger`);
	
	CREATE INDEX IF NOT EXISTS idx_leads_phone_trigger ON leads(phone, `trigger`) WHERE `trigger` IS NOT NULL;
	
	-- Optimize sequences table for active sequences
	CREATE INDEX IF NOT EXISTS idx_sequences_active ON sequences(status) WHERE `status` = 'active';

	-- Create sequence_logs table if not exists
	CREATE TABLE IF NOT EXISTS sequence_logs (
		id UUID PRIMARY KEY DEFAULT UUID(),
		sequence_id UUID NOT NULL REFERENCES sequences(id) ON DELETE CASCADE,
		contact_id UUID NOT NULL,
		step_id UUID NOT NULL,
		day INTEGER NOT NULL,
		status VARCHAR(50) NOT NULL,
		message_id VARCHAR(255),
		error_message TEXT,
		sent_at TIMESTAMP NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	-- Create indexes for sequence_logs
	CREATE INDEX IF NOT EXISTS idx_sequence_logs_sequence_id ON sequence_logs(sequence_id);
	CREATE INDEX IF NOT EXISTS idx_sequence_logs_contact_id ON sequence_logs(contact_id);
	CREATE INDEX IF NOT EXISTS idx_sequence_logs_sent_at ON sequence_logs(sent_at);
	
	-- Add group columns to broadcast_messages
	ALTER TABLE broadcast_messages ADD COLUMN IF NOT EXISTS group_id VARCHAR(255);
	ALTER TABLE broadcast_messages ADD COLUMN IF NOT EXISTS group_order INTEGER;
	`
	
	_, err = db.Exec(alterSchema)
	if err != nil {
		log.Printf("Warning: Failed to add sequence columns: %v", err)
		// Don't fail initialization, just log the warning
	}
	*/
	
	log.Println("Skipping alter schema - database structure already correct")
	
	// Create default admin user if not exists
	var adminExists bool
	err = db.QueryRow("SELECT EXISTS(SELECT 1 from users WHERE email = 'admin@whatsapp.com')").Scan(&adminExists)
	if err != nil {
		return fmt.Errorf("failed to check admin user: %w", err)
	}
	
	if !adminExists {
		// Encode password with base64 for the default admin
		encodedPassword := base64.StdEncoding.EncodeToString([]byte("changeme123"))
		
		_, err = db.Exec(`
			INSERT INTO users (email, full_name, password_hash, is_active) 
			VALUES (?, ?, ?, ?)`,
			"admin@whatsapp.com", "Administrator", encodedPassword, true)
		if err != nil {
			return fmt.Errorf("failed to create admin user: %w", err)
		}
		log.Printf("Created default admin user: admin@whatsapp.com / changeme123 (encoded: %s)\n", encodedPassword)
	}
	
	// SKIP AUTO-MIGRATIONS - Database schema is already set up correctly
	// Commented out on January 8, 2025 - All tables and columns exist as needed
	/*
	// Run auto-migrations for time_schedule and other updates
	log.Println("Running database migrations...")
	migrations := GetMigrations() // Use the new migration system
	
	for _, migration := range migrations {
		log.Printf("Running migration: %s", migration.Name)
		_, err := db.Exec(migration.SQL)
		if err != nil {
			log.Printf("Warning: Migration '%s' failed (may already exist): %v", migration.Name, err)
		} else {
			log.Printf("✓ Migration '%s' completed successfully", migration.Name)
		}
	}
	
	log.Println("All migrations completed")
	*/
	
	log.Println("Skipping auto-migrations - database schema already configured")
	
	// Run team members migration
	if err := runTeamMembersMigration(db); err != nil {
		log.Printf("Warning: Team members migration failed: %v", err)
	}
	
	// Run cleanup for expired sessions
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		
		for range ticker.C {
			db.Exec("DELETE FROM user_sessions WHERE expires_at < CURRENT_TIMESTAMP")
		}
	}()
	
	return nil
}
// runTeamMembersMigration runs the migration for team members functionality
func runTeamMembersMigration(db *sql.DB) error {
	log.Println("Running team members migration...")
	
	// Check if team_members table exists
	var tableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT from information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'team_members'
		)
	`).Scan(&tableExists)
	
	if err != nil {
		return fmt.Errorf("failed to check if team_members table exists: %w", err)
	}
	
	if tableExists {
		log.Println("Team members tables already exist, skipping migration")
		return nil
	}
	
	// Run the migration
	migrationSQL := `
		-- Create team members table
		CREATE TABLE IF NOT EXISTS team_members (
			id UUID PRIMARY KEY DEFAULT UUID(),
			username VARCHAR(255) UNIQUE NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_by UUID,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			is_active BOOLEAN DEFAULT true
		);

		-- Create team sessions table
		CREATE TABLE IF NOT EXISTS team_sessions (
			id UUID PRIMARY KEY DEFAULT UUID(),
			team_member_id UUID REFERENCES team_members(id) ON DELETE CASCADE,
			token VARCHAR(255) UNIQUE NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);

		-- Create indexes
		CREATE INDEX IF NOT EXISTS idx_team_members_username ON team_members(username);
		CREATE INDEX IF NOT EXISTS idx_team_members_created_by ON team_members(created_by);
		CREATE INDEX IF NOT EXISTS idx_team_sessions_token ON team_sessions(token);
		CREATE INDEX IF NOT EXISTS idx_team_sessions_expires_at ON team_sessions(expires_at);

		-- Add database trigger to update updated_at timestamp
		-- PostgreSQL trigger removed for MySQL compatibility
		-- MySQL handles updated_at with ON UPDATE CURRENT_TIMESTAMP in column definition
	`
	
	_, err = db.Exec(migrationSQL)
	if err != nil {
		return fmt.Errorf("failed to run team members migration: %w", err)
	}
	
	log.Println("✓ Team members migration completed successfully")
	
	// Run cleanup for expired team sessions
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		
		for range ticker.C {
			db.Exec("DELETE FROM team_sessions WHERE expires_at < CURRENT_TIMESTAMP")
		}
	}()
	
	return nil
}
