package database

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
)

// MigrateSequenceSteps fixes the sequence_steps table structure
func MigrateSequenceSteps(db *sql.DB) error {
	logrus.Info("Running sequence_steps migration...")
	
	// 1. Check if table exists
	var tableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'sequence_steps'
		)
	`).Scan(&tableExists)
	
	if err != nil {
		return fmt.Errorf("failed to check if table exists: %v", err)
	}
	
	if !tableExists {
		// Create the table from scratch
		logrus.Info("Creating sequence_steps table...")
		createQuery := `
		CREATE TABLE sequence_steps (
			id UUID PRIMARY KEY DEFAULT UUID(),
			sequence_id UUID NOT NULL,
			day_number INTEGER NOT NULL DEFAULT 1,
			message_type VARCHAR(50) DEFAULT 'text',
			content TEXT DEFAULT '',
			media_url VARCHAR(500) DEFAULT '',
			caption TEXT DEFAULT '',
			time_schedule VARCHAR(10) DEFAULT '10:00',
			trigger VARCHAR(255) DEFAULT '',
			next_trigger VARCHAR(255) DEFAULT '',
			trigger_delay_hours INTEGER DEFAULT 24,
			is_entry_point BOOLEAN DEFAULT false,
			min_delay_seconds INTEGER DEFAULT 10,
			max_delay_seconds INTEGER DEFAULT 30,
			delay_days INTEGER DEFAULT 0,
			FOREIGN KEY (sequence_id) REFERENCES sequences(id) ON DELETE CASCADE
		)`
		
		_, err = db.Exec(createQuery)
		if err != nil {
			return fmt.Errorf("failed to create table: %v", err)
		}
		logrus.Info("✓ Created sequence_steps table")
	} else {
		// Drop problematic columns if they exist
		logrus.Info("Removing timestamp columns...")
		dropColumns := []string{"send_time", "created_at", "updated_at", "day", "schedule_time"}
		
		for _, col := range dropColumns {
			query := fmt.Sprintf("ALTER TABLE sequence_steps DROP COLUMN IF EXISTS %s", col)
			_, err = db.Exec(query)
			if err != nil {
				logrus.Warnf("Warning dropping column %s: %v", col, err)
			} else {
				logrus.Infof("✓ Dropped column %s", col)
			}
		}
		
		// Add missing columns
		logrus.Info("Adding missing columns...")
		addColumns := []struct {
			name string
			definition string
		}{
			{"min_delay_seconds", "INTEGER DEFAULT 10"},
			{"max_delay_seconds", "INTEGER DEFAULT 30"},
			{"delay_days", "INTEGER DEFAULT 0"},
		}
		
		for _, col := range addColumns {
			query := fmt.Sprintf("ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS %s %s", col.name, col.definition)
			_, err = db.Exec(query)
			if err != nil {
				logrus.Warnf("Warning adding column %s: %v", col.name, err)
			} else {
				logrus.Infof("✓ Added column %s", col.name)
			}
		}
	}
	
	logrus.Info("✓ Sequence steps migration completed")
	return nil
}
