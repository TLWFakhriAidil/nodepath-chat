package whatsapp

import (
	"database/sql"
	"fmt"
	"github.com/sirupsen/logrus"
)

// ClearAllWhatsAppSessions clears all WhatsApp session data from the database
// This is useful when you want to start fresh with all devices
func ClearAllWhatsAppSessions(db *sql.DB) error {
	// List of whatsmeow tables to clear
	tables := []string{
		"whatsmeow_app_state_mutation_macs",
		"whatsmeow_app_state_sync_keys",
		"whatsmeow_app_state_version", 
		"whatsmeow_chat_settings",
		"whatsmeow_contacts",
		"whatsmeow_disappearing_timers",
		"whatsmeow_group_participants",
		"whatsmeow_groups",
		"whatsmeow_history_syncs",
		"whatsmeow_media_backfill_requests",
		"whatsmeow_message_secrets",
		"whatsmeow_portal_backfill",
		"whatsmeow_portal_backfill_queue",
		"whatsmeow_portal_message",
		"whatsmeow_portal_message_part", 
		"whatsmeow_portal_reaction",
		"whatsmeow_portal",
		"whatsmeow_privacy_settings",
		"whatsmeow_sender_keys",
		"whatsmeow_sessions",
		"whatsmeow_pre_keys",
		"whatsmeow_identity_keys",
		"whatsmeow_device",
	}
	
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Try to truncate each table
	for _, table := range tables {
		// Check if table exists
		var exists bool
		err = tx.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = ?
			)
		`, table).Scan(&exists)
		
		if err != nil {
			logrus.Warnf("Error checking if table %s exists: %v", table, err)
			continue
		}
		
		if exists {
			_, err = tx.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
			if err != nil {
				logrus.Warnf("Error truncating %s: %v", table, err)
			} else {
				logrus.Infof("Truncated table %s", table)
			}
		}
	}
	
	return tx.Commit()
}

// GetWhatsAppSessionInfo returns information about WhatsApp sessions in the database
func GetWhatsAppSessionInfo(db *sql.DB) (map[string]int, error) {
	info := make(map[string]int)
	
	// Check device count
	var deviceCount int
	err := db.QueryRow("SELECT COUNT(*) FROM whatsmeow_device").Scan(&deviceCount)
	if err != nil {
		logrus.Warnf("Error counting devices: %v", err)
	} else {
		info["devices"] = deviceCount
	}
	
	// Check session count
	var sessionCount int
	err = db.QueryRow("SELECT COUNT(*) FROM whatsmeow_sessions").Scan(&sessionCount)
	if err != nil {
		logrus.Warnf("Error counting sessions: %v", err)
	} else {
		info["sessions"] = sessionCount
	}
	
	return info, nil
}
