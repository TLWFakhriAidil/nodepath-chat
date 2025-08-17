package whatsapp

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/websocket"
	"github.com/sirupsen/logrus"
)

// NotifyMessageUpdate sends a WebSocket notification when a new message is received
func NotifyMessageUpdate(deviceID string, chatJID string, message string) {
	// Send WebSocket broadcast for real-time UI update
	websocket.Broadcast <- websocket.BroadcastMessage{
		Code:    "NEW_MESSAGE",
		Message: "New message received",
		Result: map[string]interface{}{
			"deviceId": deviceID,
			"chatJid":  chatJID,
			"message":  message,
			"action":   "refresh_chat_list",
		},
	}
	
	logrus.Debugf("Sent WebSocket notification for new message in chat %s", chatJID)
}

// NotifyChatUpdate sends a WebSocket notification when chat list should be updated
func NotifyChatUpdate(deviceID string) {
	// Send WebSocket broadcast for chat list update
	websocket.Broadcast <- websocket.BroadcastMessage{
		Code:    "CHAT_UPDATE",
		Message: "Chat list updated",
		Result: map[string]interface{}{
			"deviceId": deviceID,
			"action":   "refresh_chat_list",
		},
	}
	
	logrus.Debugf("Sent WebSocket notification for chat update on device %s", deviceID)
}