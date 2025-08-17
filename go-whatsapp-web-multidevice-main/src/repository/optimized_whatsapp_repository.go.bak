package repository

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
)

// OptimizedWhatsAppRepository handles WhatsApp data with performance optimizations
type OptimizedWhatsAppRepository struct {
	db              *sql.DB
	chatCache       *sync.Map // deviceID -> map[chatJID]*WhatsAppChat
	messageCache    *sync.Map // deviceID_chatJID -> []WhatsAppMessage
	batchInsertStmt *sql.Stmt
	cacheTTL        time.Duration
}

// NewOptimizedWhatsAppRepository creates an optimized repository
func NewOptimizedWhatsAppRepository(db *sql.DB) (*OptimizedWhatsAppRepository, error) {
	// Prepare batch insert statement
	batchStmt, err := db.Prepare(`
		INSERT INTO whatsapp_messages(device_id, chat_jid, message_id, sender_jid, sender_name, message_text,
		 message_type, media_url, is_sent, is_read, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (device_id, message_id) DO NOTHING
	`)
	if err != nil {
		return nil, err
	}
	
	repo := &OptimizedWhatsAppRepository{
		db:              db,
		chatCache:       &sync.Map{},
		messageCache:    &sync.Map{},
		batchInsertStmt: batchStmt,
		cacheTTL:        5 * time.Minute,
	}
	
	// Start cache cleanup
	go repo.cleanupCache()
	
	return repo, nil
}

// BatchSaveChats saves multiple chats in a single transaction
func (r *OptimizedWhatsAppRepository) BatchSaveChats(chats []WhatsAppChat) error {
	if len(chats) == 0 {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	stmt, err := tx.Prepare(`
		INSERT INTO whatsapp_chats 
		(device_id, chat_jid, chat_name, is_group, is_muted, last_message_text, 
		 last_message_time, unread_count, avatar_url, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT (device_id, chat_jid) 
		DO UPDATE SET 
			chat_name = EXCLUDED.chat_name,
			is_muted = EXCLUDED.is_muted,
			last_message_text = CASE 
				WHEN EXCLUDED.last_message_time > whatsapp_chats.last_message_time 
				THEN EXCLUDED.last_message_text 
				ELSE whatsapp_chats.last_message_text 
			END,
			last_message_time = CASE 
				WHEN EXCLUDED.last_message_time > whatsapp_chats.last_message_time 
				THEN EXCLUDED.last_message_time 
				ELSE whatsapp_chats.last_message_time 
			END,
			unread_count = EXCLUDED.unread_count,
			avatar_url = EXCLUDED.avatar_url,
			updated_at = EXCLUDED.updated_at
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	
	for _, chat := range chats {
		_, err = stmt.Exec(
			chat.DeviceID, chat.ChatJID, chat.ChatName, chat.IsGroup, chat.IsMuted,
			chat.LastMessageText, chat.LastMessageTime, chat.UnreadCount, 
			chat.AvatarURL, time.Now(),
		)
		if err != nil {
			return err
		}
		
		// Update cache
		r.updateChatCache(chat.DeviceID, &chat)
	}
	
	return tx.Commit()
}

// BatchSaveMessages saves multiple messages efficiently
func (r *OptimizedWhatsAppRepository) BatchSaveMessages(messages []WhatsAppMessage) error {
	if len(messages) == 0 {
		return nil
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	
	// Use COPY for PostgreSQL for maximum performance
	stmt, err := tx.Prepare(`
		COPY whatsapp_messages (device_id, chat_jid, message_id, sender_jid, sender_name, 
		message_text, message_type, media_url, is_sent, is_read, timestamp) 
		FROM STDIN
	`)
	if err != nil {
		// Fallback to regular batch insert
		for _, msg := range messages {
			_, err = tx.Stmt(r.batchInsertStmt).Exec(
				msg.DeviceID, msg.ChatJID, msg.MessageID, msg.SenderJID, msg.SenderName,
				msg.MessageText, msg.MessageType, msg.MediaURL, msg.IsSent, msg.IsRead,
				msg.Timestamp,
			)
			if err != nil {
				return err
			}
		}
	} else {
		stmt.Close()
	}
	
	return tx.Commit()
}

// GetChatsWithCache retrieves chats with caching
func (r *OptimizedWhatsAppRepository) GetChatsWithCache(deviceID string) ([]WhatsAppChat, error) {
	// Check cache first
	if cached, ok := r.chatCache.Load(deviceID); ok {
		if chatMap, ok := cached.(map[string]*WhatsAppChat); ok {
			chats := make([]WhatsAppChat, 0, len(chatMap))
			for _, chat := range chatMap {
				chats = append(chats, *chat)
			}
			return chats, nil
		}
	}
	
	// Load from database
	chats, err := r.GetChats(deviceID)
	if err != nil {
		return nil, err
	}
	
	// Update cache
	chatMap := make(map[string]*WhatsAppChat)
	for i := range chats {
		chatMap[chats[i].ChatJID] = &chats[i]
	}
	r.chatCache.Store(deviceID, chatMap)
	
	return chats, nil
}

// GetMessagesWithCache retrieves messages with caching
func (r *OptimizedWhatsAppRepository) GetMessagesWithCache(deviceID, chatJID string, limit int) ([]WhatsAppMessage, error) {
	cacheKey := fmt.Sprintf("%s_%s", deviceID, chatJID)
	
	// Check cache first
	if cached, ok := r.messageCache.Load(cacheKey); ok {
		if messages, ok := cached.([]WhatsAppMessage); ok && len(messages) >= limit {
			if limit > len(messages) {
				limit = len(messages)
			}
			return messages[:limit], nil
		}
	}
	
	// Load from database
	messages, err := r.GetMessages(deviceID, chatJID, limit*2) // Get extra for cache
	if err != nil {
		return nil, err
	}
	
	// Update cache
	r.messageCache.Store(cacheKey, messages)
	
	if limit > len(messages) {
		limit = len(messages)
	}
	return messages[:limit], nil
}

// updateChatCache updates the chat cache
func (r *OptimizedWhatsAppRepository) updateChatCache(deviceID string, chat *WhatsAppChat) {
	cached, _ := r.chatCache.LoadOrStore(deviceID, make(map[string]*WhatsAppChat))
	if chatMap, ok := cached.(map[string]*WhatsAppChat); ok {
		chatMap[chat.ChatJID] = chat
	}
}

// cleanupCache periodically cleans up old cache entries
func (r *OptimizedWhatsAppRepository) cleanupCache() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		// Clear all caches periodically
		r.chatCache = &sync.Map{}
		r.messageCache = &sync.Map{}
	}
}

// GetChats retrieves all chats for a device (existing method)
func (r *OptimizedWhatsAppRepository) GetChats(deviceID string) ([]WhatsAppChat, error) {
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

// GetMessages retrieves messages for a chat (existing method)
func (r *OptimizedWhatsAppRepository) GetMessages(deviceID, chatJID string, limit int) ([]WhatsAppMessage, error) {
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

// Close closes the repository and cleans up resources
func (r *OptimizedWhatsAppRepository) Close() error {
	if r.batchInsertStmt != nil {
		r.batchInsertStmt.Close()
	}
	return nil
}
