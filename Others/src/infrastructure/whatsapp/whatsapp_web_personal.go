package whatsapp

import (
	"database/sql"
	"fmt"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
)

// GetWhatsAppWebChats gets recent chats from database (like whatsapp-mcp-main)
func GetWhatsAppWebChats(deviceID string) ([]map[string]interface{}, error) {
	logrus.Infof("=== GetWhatsAppWebChats called for device: %s ===", deviceID)
	
	// Get chats from database (exactly like whatsapp-mcp-main)
	chats, err := GetChatsFromDatabase(deviceID)
	if err != nil {
		logrus.Errorf("Failed to get chats from database: %v", err)
		return []map[string]interface{}{}, nil
	}
	
	logrus.Infof("Found %d chats from database", len(chats))
	return chats, nil
}

// GetStoredMessagesFromDB gets messages from database without requiring client connection
func GetStoredMessagesFromDB(deviceID, chatJID string, limit int) ([]map[string]interface{}, error) {
	logrus.Infof("=== GetStoredMessagesFromDB called for device: %s, chat: %s ===", deviceID, chatJID)
	
	userRepo := repository.GetUserRepository()
	db := userRepo.DB()
	
	// Ensure table exists
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS whatsapp_messages (
			id SERIAL PRIMARY KEY,
			device_id TEXT NOT NULL,
			chat_jid TEXT NOT NULL,
			message_id TEXT NOT NULL,
			sender_jid TEXT NOT NULL,
			message_text TEXT,
			message_type TEXT DEFAULT 'text',
			timestamp BIGINT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(device_id, message_id)
		);
		
		CREATE INDEX IF NOT EXISTS idx_device_chat_time ON whatsapp_messages(device_id, chat_jid, timestamp DESC);
	`
	
	db.Exec(createTableQuery)
	
	// Query messages
	query := `
		SELECT message_id,
			sender_jid,
			message_text,
			message_type,
			message_secrets,
			timestamp
		FROM whatsapp_messages
		WHERE device_id = ? AND chat_jid = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`
	
	rows, err := db.Query(query, deviceID, chatJID, limit)
	if err != nil {
		logrus.Errorf("Failed to query messages: %v", err)
		return []map[string]interface{}{}, nil
	}
	defer rows.Close()
	
	var messages []map[string]interface{}
	
	// Try to determine our JID from stored messages or device info
	ourJID := ""
	
	// First, try to get device JID from user_devices table
	var deviceJID sql.NullString
	deviceQuery := `SELECT jid FROM user_devices WHERE id = ?`
	db.QueryRow(deviceQuery, deviceID).Scan(&deviceJID)
	if deviceJID.Valid && deviceJID.String != "" {
		ourJID = deviceJID.String
		logrus.Infof("Got device JID from database: %s", ourJID)
	}
	
	for rows.Next() {
		var messageID, senderJID, messageType string
		var messageText, messageSecrets sql.NullString
		var timestamp int64
		
		err := rows.Scan(&messageID, &senderJID, &messageText, &messageType, &messageSecrets, &timestamp)
		if err != nil {
			continue
		}
		
		// Determine if sent or received
		// If we have ourJID, use it. Otherwise, assume messages from the chat JID are received
		sent := false
		if ourJID != "" {
			sent = senderJID == ourJID
		} else {
			// Heuristic: if sender is not the chat JID, it's probably sent by us
			sent = senderJID != chatJID
		}
		
		// Format time using Malaysia timezone
		malaysiaTime := GetMalaysiaTime(timestamp)
		timeStr := malaysiaTime.Format("15:04")
		
		text := ""
		if messageText.Valid {
			text = messageText.String
		}
		
		message := map[string]interface{}{
			"id":        messageID,
			"text":      text,
			"type":      messageType,
			"sent":      sent,
			"time":      timeStr,
			"timestamp": timestamp,
		}
		
		// Add image URL if it's an image message
		if messageType == "image" && messageSecrets.Valid && messageSecrets.String != "" {
			message["image"] = messageSecrets.String
		}
		
		messages = append(messages, message)
	}
	
	// Reverse to show oldest first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	logrus.Infof("=== Found %d stored messages for chat %s ===", len(messages), chatJID)
	
	return messages, nil
}

// GetWhatsAppWebMessages gets messages for a specific chat
func GetWhatsAppWebMessages(deviceID, chatJID string, limit int) ([]map[string]interface{}, error) {
	logrus.Infof("=== GetWhatsAppWebMessages called for device: %s, chat: %s ===", deviceID, chatJID)
	
	cm := GetClientManager()
	client, err := cm.GetClient(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not connected: %v", err)
	}
	
	if client.Store.ID == nil {
		return nil, fmt.Errorf("device not logged in")
	}
	
	ourJID := client.Store.ID.String()
	
	// Get messages from database
	userRepo := repository.GetUserRepository()
	db := userRepo.DB()
	
	// Ensure table exists
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS whatsapp_messages (
			id SERIAL PRIMARY KEY,
			device_id TEXT NOT NULL,
			chat_jid TEXT NOT NULL,
			message_id TEXT NOT NULL,
			sender_jid TEXT NOT NULL,
			message_text TEXT,
			message_type TEXT DEFAULT 'text',
			timestamp BIGINT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(device_id, message_id)
		);
		
		CREATE INDEX IF NOT EXISTS idx_device_chat_time ON whatsapp_messages(device_id, chat_jid, timestamp DESC);
	`
	
	db.Exec(createTableQuery)
	
	// Query messages
	query := `
		SELECT message_id,
			sender_jid,
			message_text,
			message_type,
			message_secrets,
			timestamp
		FROM whatsapp_messages
		WHERE device_id = ? AND chat_jid = ?
		ORDER BY timestamp DESC
		LIMIT ?
	`
	
	rows, err := db.Query(query, deviceID, chatJID, limit)
	if err != nil {
		logrus.Errorf("Failed to query messages: %v", err)
		return []map[string]interface{}{}, nil
	}
	defer rows.Close()
	
	var messages []map[string]interface{}
	
	for rows.Next() {
		var messageID, senderJID, messageType string
		var messageText, messageSecrets sql.NullString
		var timestamp int64
		
		err := rows.Scan(&messageID, &senderJID, &messageText, &messageType, &messageSecrets, &timestamp)
		if err != nil {
			continue
		}
		
		// Determine if sent or received
		sent := senderJID == ourJID
		
		// Format time using Malaysia timezone
		malaysiaTime := GetMalaysiaTime(timestamp)
		timeStr := malaysiaTime.Format("15:04")
		
		text := ""
		if messageText.Valid {
			text = messageText.String
		}
		
		message := map[string]interface{}{
			"id":        messageID,
			"text":      text,
			"type":      messageType,
			"sent":      sent,
			"time":      timeStr,
			"timestamp": timestamp,
		}
		
		// Add image URL if it's an image message
		if messageType == "image" && messageSecrets.Valid && messageSecrets.String != "" {
			message["image"] = messageSecrets.String
		}
		
		messages = append(messages, message)
	}
	
	// Reverse to show oldest first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	logrus.Infof("=== Found %d messages for chat %s ===", len(messages), chatJID)
	
	// If no messages found, create a welcome message
	if len(messages) == 0 {
		messages = append(messages, map[string]interface{}{
			"id":        "welcome",
			"text":      "Send a message to start the conversation",
			"type":      "text",
			"sent":      false,
			"time":      time.Now().Format("15:04"),
			"timestamp": time.Now().Unix(),
		})
	}
	
	return messages, nil
}

// formatMessageTime formats timestamp to readable time
func formatMessageTime(timestamp int64) string {
	if timestamp == 0 {
		return ""
	}
	
	t := time.Unix(timestamp, 0)
	now := time.Now()
	
	// Today
	if t.Day() == now.Day() && t.Month() == now.Month() && t.Year() == now.Year() {
		return t.Format("15:04")
	}
	
	// Yesterday
	yesterday := now.AddDate(0, 0, -1)
	if t.Day() == yesterday.Day() && t.Month() == yesterday.Month() && t.Year() == yesterday.Year() {
		return "Yesterday"
	}
	
	// This week
	if now.Sub(t) < 7*24*time.Hour {
		return t.Format("Monday")
	}
	
	// Older
	return t.Format("Jan 2")
}

// RefreshWhatsAppChats - just return nil, no manual sync needed
func RefreshWhatsAppChats(deviceID string) error {
	logrus.Infof("=== RefreshWhatsAppChats called for device: %s ===", deviceID)
	logrus.Info("Sync happens automatically when WhatsApp sends events")
	return nil
}

// SyncWhatsAppHistory - removed manual sync, WhatsApp sends history automatically
func SyncWhatsAppHistory(deviceID string) error {
	logrus.Infof("=== SyncWhatsAppHistory called for device: %s ===", deviceID)
	logrus.Info("History sync happens automatically when WhatsApp sends events")
	return nil
}