package whatsapp

import (
	"context"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
	"time"
)

// Context key for device ID
type contextKey string

const deviceIDKey contextKey = "deviceID"

// WithDeviceID adds device ID to context
func WithDeviceID(ctx context.Context, deviceID string) context.Context {
	return context.WithValue(ctx, deviceIDKey, deviceID)
}

// GetDeviceIDFromContext retrieves device ID from context
func GetDeviceIDFromContext(ctx context.Context) string {
	if deviceID, ok := ctx.Value(deviceIDKey).(string); ok {
		return deviceID
	}
	return ""
}

// StoreWhatsAppMessage stores a message in the whatsapp_messages table
func StoreWhatsAppMessage(deviceID, chatJID, messageID, senderJID, messageText, messageType string) {
	StoreWhatsAppMessageWithTimestamp(deviceID, chatJID, messageID, senderJID, messageText, messageType, time.Now().Unix())
}

// StoreWhatsAppMessageWithMedia stores a message with media URL
func StoreWhatsAppMessageWithMedia(deviceID, chatJID, messageID, senderJID, messageText, messageType, mediaURL string) {
	userRepo := repository.GetUserRepository()
	db := userRepo.DB()
	
	// If sender is empty, it's a sent message
	if senderJID == "" {
		// Get our JID from client
		cm := GetClientManager()
		client, err := cm.GetClient(deviceID)
		if err != nil {
			logrus.Warnf("Failed to get client for storing message: %v", err)
			return
		}
		if client.Store.ID != nil {
			senderJID = client.Store.ID.String()
		}
	}
	
	// Store in database with media URL
	query := `
		INSERT INTO whatsapp_messages(device_id, chat_jid, message_id, sender_jid, message_text, message_type, message_secrets, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			message_text = VALUES(message_text),
			message_secrets = VALUES(message_secrets),
			timestamp = VALUES(timestamp)
	`
	
	_, err := db.Exec(query, deviceID, chatJID, messageID, senderJID, messageText, messageType, mediaURL, time.Now().Unix())
	if err != nil {
		logrus.Errorf("Failed to store message with media: %v", err)
	} else {
		logrus.Debugf("Stored %s message with media URL: %s", messageType, mediaURL)
	}
}

// StoreWhatsAppMessageWithTimestamp stores a message with specific timestamp
func StoreWhatsAppMessageWithTimestamp(deviceID, chatJID, messageID, senderJID, messageText, messageType string, timestamp int64) {
	StoreWhatsAppMessageWithMediaAndTimestamp(deviceID, chatJID, messageID, senderJID, messageText, messageType, "", timestamp)
}

// StoreWhatsAppMessageWithMediaAndTimestamp stores a message with media URL and specific timestamp
func StoreWhatsAppMessageWithMediaAndTimestamp(deviceID, chatJID, messageID, senderJID, messageText, messageType, mediaURL string, timestamp int64) {
	userRepo := repository.GetUserRepository()
	db := userRepo.DB()
	
	// If sender is empty, it's a sent message
	if senderJID == "" {
		// Get our JID from client
		cm := GetClientManager()
		client, err := cm.GetClient(deviceID)
		if err != nil {
			logrus.Warnf("Failed to get client for storing message: %v", err)
			return
		}
		if client.Store.ID != nil {
			senderJID = client.Store.ID.String()
		}
	}
	
	// Insert message with media URL if provided
	query := `
		INSERT INTO whatsapp_messages(device_id, chat_jid, message_id, sender_jid, message_text, message_type, message_secrets, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			message_text = VALUES(message_text),
			message_type = VALUES(message_type),
			message_secrets = VALUES(message_secrets),
			timestamp = VALUES(timestamp)
	`
	
	_, err := db.Exec(query, deviceID, chatJID, messageID, senderJID, messageText, messageType, mediaURL, timestamp)
	if err != nil {
		logrus.Warnf("Failed to store message: %v", err)
	} else {
		logrus.Debugf("Stored %s message for chat %s (media: %s)", messageType, chatJID, mediaURL)
	}
}

// ImageRequest extension for WhatsApp Web
type ImageRequestExt struct {
	ImageB64   string
	ImageBytes []byte
}
