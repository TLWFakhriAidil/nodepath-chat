package repository

import (
	"database/sql"
	"fmt"
	"time"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
)

// WhatsAppChat represents a chat/conversation
type WhatsAppChat struct {
	ID              int       `json:"id"`
	DeviceID        string    `json:"device_id"`
	ChatJID         string    `json:"chat_jid"`
	ChatName        string    `json:"chat_name"`
	IsGroup         bool      `json:"is_group"`
	IsMuted         bool      `json:"is_muted"`
	LastMessageText string    `json:"last_message_text"`
	LastMessageTime time.Time `json:"last_message_time"`
	UnreadCount     int       `json:"unread_count"`
	AvatarURL       string    `json:"avatar_url"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// WhatsAppMessage represents a message
type WhatsAppMessage struct {
	ID          int       `json:"id"`
	DeviceID    string    `json:"device_id"`
	ChatJID     string    `json:"chat_jid"`
	MessageID   string    `json:"message_id"`
	SenderJID   string    `json:"sender_jid"`
	SenderName  string    `json:"sender_name"`
	MessageText string    `json:"message_text"`
	MessageType string    `json:"message_type"`
	MediaURL    string    `json:"media_url"`
	IsSent      bool      `json:"is_sent"`
	IsRead      bool      `json:"is_read"`
	Timestamp   time.Time `json:"timestamp"`
	CreatedAt   time.Time `json:"created_at"`
}

// WhatsAppRepository handles WhatsApp data persistence
type WhatsAppRepository struct {
	db *sql.DB
}

// NewWhatsAppRepository creates a new WhatsApp repository
func NewWhatsAppRepository(db *sql.DB) *WhatsAppRepository {
	return &WhatsAppRepository{db: db}
}

// SaveOrUpdateChat saves or updates a chat
func (r *WhatsAppRepository) SaveOrUpdateChat(chat *WhatsAppChat) error {
	query := `
		INSERT INTO whatsapp_chats 
		(device_id, chat_jid, chat_name, is_group, is_muted, last_message_text, 
		 last_message_time, unread_count, avatar_url, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (device_id, chat_jid) 
		DO UPDATE SET 
			chat_name = EXCLUDED.chat_name,
			is_group = EXCLUDED.is_group,
			is_muted = EXCLUDED.is_muted,
			last_message_text = EXCLUDED.last_message_text,
			last_message_time = EXCLUDED.last_message_time,
			unread_count = EXCLUDED.unread_count,
			avatar_url = EXCLUDED.avatar_url,
			updated_at = EXCLUDED.updated_at`
	
	err := r.db.QueryRow(query, 
		chat.DeviceID, chat.ChatJID, chat.ChatName, chat.IsGroup, chat.IsMuted,
		chat.LastMessageText, chat.LastMessageTime, chat.UnreadCount, 
		chat.AvatarURL, time.Now()).Scan(&chat.ID)
	
	return err
}

// GetChats retrieves all chats for a device
func (r *WhatsAppRepository) GetChats(deviceID string) ([]WhatsAppChat, error) {
	query := `
		SELECT id, device_id, chat_jid, chat_name, is_group, is_muted, 
		       last_message_text, last_message_time, unread_count, avatar_url, 
		       created_at, updated_at
		from whatsapp_chats
		WHERE device_id = ?
		ORDER BY last_message_time DESC`
	
	rows, err := r.db.Query(query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var chats []WhatsAppChat
	for rows.Next() {
		var chat WhatsAppChat
		err := rows.Scan(&chat.ID, &chat.DeviceID, &chat.ChatJID, &chat.ChatName,
			&chat.IsGroup, &chat.IsMuted, &chat.LastMessageText, &chat.LastMessageTime,
			&chat.UnreadCount, &chat.AvatarURL, &chat.CreatedAt, &chat.UpdatedAt)
		if err != nil {
			return nil, err
		}
		chats = append(chats, chat)
	}
	
	return chats, nil
}

// SaveMessage saves a new message
func (r *WhatsAppRepository) SaveMessage(msg *WhatsAppMessage) error {
	query := `
		INSERT INTO whatsapp_messages(device_id, chat_jid, message_id, sender_jid, sender_name, message_text,
		 message_type, media_url, is_sent, is_read, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (device_id, message_id) DO NOTHING`
	
	err := r.db.QueryRow(query,
		msg.DeviceID, msg.ChatJID, msg.MessageID, msg.SenderJID, msg.SenderName,
		msg.MessageText, msg.MessageType, msg.MediaURL, msg.IsSent, msg.IsRead,
		msg.Timestamp).Scan(&msg.ID)
	
	if err == sql.ErrNoRows {
		// Message already exists, not an error
		return nil
	}
	
	return err
}

// GetMessages retrieves messages for a chat
func (r *WhatsAppRepository) GetMessages(deviceID, chatJID string, limit int) ([]WhatsAppMessage, error) {
	query := `
		SELECT id, device_id, chat_jid, message_id, sender_jid, sender_name,
		       message_text, message_type, media_url, is_sent, is_read, 
		       timestamp, created_at
		from whatsapp_messages
		WHERE device_id = ? AND chat_jid = ?
		ORDER BY timestamp DESC
		LIMIT ?`
	
	rows, err := r.db.Query(query, deviceID, chatJID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var messages []WhatsAppMessage
	for rows.Next() {
		var msg WhatsAppMessage
		err := rows.Scan(&msg.ID, &msg.DeviceID, &msg.ChatJID, &msg.MessageID,
			&msg.SenderJID, &msg.SenderName, &msg.MessageText, &msg.MessageType,
			&msg.MediaURL, &msg.IsSent, &msg.IsRead, &msg.Timestamp, &msg.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	
	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}
	
	return messages, nil
}

// GetChatByJID retrieves a single chat by JID
func (r *WhatsAppRepository) GetChatByJID(deviceID, chatJID string) (*WhatsAppChat, error) {
	query := `
		SELECT id, device_id, chat_jid, chat_name, is_group, is_muted, 
		       last_message_text, last_message_time, unread_count, avatar_url, 
		       created_at, updated_at
		from whatsapp_chats
		WHERE device_id = ? AND chat_jid = ?`
	
	var chat WhatsAppChat
	err := r.db.QueryRow(query, deviceID, chatJID).Scan(
		&chat.ID, &chat.DeviceID, &chat.ChatJID, &chat.ChatName,
		&chat.IsGroup, &chat.IsMuted, &chat.LastMessageText, &chat.LastMessageTime,
		&chat.UnreadCount, &chat.AvatarURL, &chat.CreatedAt, &chat.UpdatedAt)
	
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("chat not found")
	}
	
	return &chat, err
}

// ClearDeviceMessages clears all messages for a device
func (r *WhatsAppRepository) ClearDeviceMessages(deviceID string) error {
	query := `DELETE FROM whatsapp_messages WHERE device_id = ?`
	_, err := r.db.Exec(query, deviceID)
	if err != nil {
		return fmt.Errorf("failed to clear device messages: %w", err)
	}
	return nil
}

// ClearDeviceChats clears all chats for a device
func (r *WhatsAppRepository) ClearDeviceChats(deviceID string) error {
	query := `DELETE FROM whatsapp_chats WHERE device_id = ?`
	_, err := r.db.Exec(query, deviceID)
	if err != nil {
		return fmt.Errorf("failed to clear device chats: %w", err)
	}
	return nil
}

// WhatsApp repository singleton
var whatsappRepo *WhatsAppRepository

// GetWhatsAppRepository returns the WhatsApp repository instance
func GetWhatsAppRepository() *WhatsAppRepository {
	if whatsappRepo == nil {
		whatsappRepo = NewWhatsAppRepository(database.GetDB())
	}
	return whatsappRepo
}
