package whatsapp

import (
	"database/sql"
	"fmt"
	"strings"
	
	"github.com/sirupsen/logrus"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
)

// ClearWhatsAppSessionData clears all WhatsApp session data for a device
func ClearWhatsAppSessionData(deviceID string) error {
	// Validate device ID format (should be UUID)
	if strings.HasPrefix(deviceID, "device_") {
		logrus.Warnf("Invalid device ID format: %s (expected UUID)", deviceID)
		return fmt.Errorf("invalid device ID format")
	}
	
	// Get repository to access DB
	userRepo := repository.GetUserRepository()
	db := userRepo.DB()
	
	// First, get the JID and phone from user_devices
	var jid sql.NullString
	var phone sql.NullString
	err := db.QueryRow("SELECT jid, phone from user_devices WHERE id = ?", deviceID).Scan(&jid, &phone)
	if err != nil {
		if err == sql.ErrNoRows {
			logrus.Warnf("Device %s not found in database", deviceID)
			return nil
		}
		logrus.Warnf("Failed to get device info for %s: %v", deviceID, err)
		return nil
	}
	
	// If no JID found, we can't clean WhatsApp sessions
	if !jid.Valid || jid.String == "" {
		logrus.Infof("No JID found for device %s, skipping WhatsApp session cleanup", deviceID)
		return nil
	}
	
	logrus.Infof("Clearing WhatsApp session for device %s with JID %s", deviceID, jid.String)
	
	// Disable foreign key checks temporarily for cleanup
	_, err = db.Exec("SET session_replication_role = 'replica'")
	if err != nil {
		logrus.Warnf("Failed to disable FK checks: %v", err)
	}
	defer func() {
		_, err = db.Exec("SET session_replication_role = 'origin'")
		if err != nil {
			logrus.Warnf("Failed to re-enable FK checks: %v", err)
		}
	}()
	
	// Clear all whatsmeow tables that might contain this JID
	// Using a more robust approach with different column names
	cleanupQueries := []struct {
		table   string
		columns []string
	}{
		// Tables with 'jid' column
		{"whatsmeow_device", []string{"jid"}},
		{"whatsmeow_identity_keys", []string{"their_id"}},
		{"whatsmeow_sessions", []string{"their_id"}},
		{"whatsmeow_sender_keys", []string{"chat_id", "sender_id"}},
		{"whatsmeow_app_state_version", []string{"jid"}},
		{"whatsmeow_app_state_mutation_macs", []string{"jid"}},
		{"whatsmeow_app_state_sync_keys", []string{"jid"}},
		{"whatsmeow_contacts", []string{"jid", "our_jid"}},
		{"whatsmeow_chat_settings", []string{"jid"}},
		{"whatsmeow_disappearing_timers", []string{"jid"}},
		{"whatsmeow_groups", []string{"jid"}},
		{"whatsmeow_group_participants", []string{"group_jid", "jid"}},
		{"whatsmeow_history_syncs", []string{"device_jid"}},
		{"whatsmeow_message_secrets", []string{"chat_jid", "sender_id"}},
		{"whatsmeow_privacy_tokens", []string{"our_jid"}},
		{"whatsmeow_pre_keys", []string{"jid"}},
		// Portal tables
		{"whatsmeow_portal", []string{"jid", "receiver"}},
		{"whatsmeow_portal_message", []string{"portal_jid"}},
		{"whatsmeow_portal_message_part", []string{"message_id"}},
		{"whatsmeow_portal_reaction", []string{"portal_jid"}},
		{"whatsmeow_portal_backfill", []string{"portal_jid"}},
		{"whatsmeow_portal_backfill_queue", []string{"portal_jid"}},
		{"whatsmeow_media_backfill_requests", []string{"user_jid", "portal_jid"}},
	}
	
	// Execute cleanup for each table
	successCount := 0
	totalTables := 0
	
	for _, cleanup := range cleanupQueries {
		// Check if table exists
		var exists bool
		err = db.QueryRow(`
			SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_schema = 'public' 
				AND table_name = ?
			)
		`, cleanup.table).Scan(&exists)
		
		if !exists {
			continue
		}
		
		totalTables++
		
		// Build WHERE clause for multiple columns
		conditions := []string{}
		params := []interface{}{}
		paramIdx := 1
		
		for _, col := range cleanup.columns {
			conditions = append(conditions, fmt.Sprintf("%s = $%d", col, paramIdx))
			params = append(params, jid.String)
			paramIdx++
		}
		
		query := fmt.Sprintf("DELETE FROM %s WHERE %s", cleanup.table, strings.Join(conditions, " OR "))
		
		_, err = db.Exec(query, params...)
		if err != nil {
			logrus.Debugf("Failed to clean %s: %v", cleanup.table, err)
		} else {
			successCount++
		}
	}
	
	logrus.Infof("Successfully cleared %d/%d WhatsApp tables for device %s", successCount, totalTables, deviceID)
	
	return nil
}
