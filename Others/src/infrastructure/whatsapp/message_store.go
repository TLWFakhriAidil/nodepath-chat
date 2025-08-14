package whatsapp

import (
	"context"
	"path/filepath"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
)

// HandleMessageForWebView stores messages for WhatsApp Web view
// Only stores personal chat messages (no groups)
func HandleMessageForWebView(deviceID string, evt *events.Message) {
	// Skip group messages - we only want personal chats
	if evt.Info.IsGroup || evt.Info.Chat.Server != types.DefaultUserServer {
		logrus.Debugf("Skipping group/non-personal message from %s", evt.Info.Chat.String())
		return
	}
	
	// Skip status updates
	if evt.Info.Chat.User == "status" {
		return
	}
	
	// Skip messages sent by us (they're already stored when sending)
	if evt.Info.IsFromMe {
		logrus.Debugf("Skipping own message in chat %s", evt.Info.Chat.String())
		return
	}
	
	logrus.Infof("=== Received message in chat %s from %s ===", evt.Info.Chat.String(), evt.Info.Sender.String())
	
	// Extract message text and handle images
	messageText := extractMessageText(evt)
	messageType := "text"
	mediaURL := ""
	
	// Check for different message types
	if imageMsg := evt.Message.GetImageMessage(); imageMsg != nil {
		messageType = "image"
		if caption := imageMsg.GetCaption(); caption != "" {
			messageText = caption
		}
		
		// Download and store image
		if media, err := ExtractMedia(context.Background(), config.PathStorages, imageMsg); err == nil {
			// Convert to web-accessible URL - use just the filename
			filename := filepath.Base(media.MediaPath)
			mediaURL = "/media/" + filename
			logrus.Infof("Downloaded image to: %s, URL: %s", media.MediaPath, mediaURL)
		} else {
			logrus.Errorf("Failed to download image: %v", err)
		}
	} else if evt.Message.GetVideoMessage() != nil {
		messageType = "video"
		if caption := evt.Message.GetVideoMessage().GetCaption(); caption != "" {
			messageText = caption
		}
	} else if evt.Message.GetAudioMessage() != nil {
		messageType = "audio"
	} else if evt.Message.GetDocumentMessage() != nil {
		messageType = "document"
		if fileName := evt.Message.GetDocumentMessage().GetFileName(); fileName != "" {
			messageText = "📄 " + fileName
		}
	}
	
	// Store message with media URL
	StoreWhatsAppMessageWithMedia(
		deviceID, 
		evt.Info.Chat.String(), 
		evt.Info.ID, 
		evt.Info.Sender.String(), 
		messageText, 
		messageType,
		mediaURL,
	)
	
	// Send WebSocket notification for real-time update
	NotifyMessageUpdate(deviceID, evt.Info.Chat.String(), messageText)
	
	logrus.Debugf("Stored %s message from %s in chat %s", messageType, evt.Info.Sender.String(), evt.Info.Chat.String())
}

// HandleHistorySyncForWebView processes history sync to get recent messages
func HandleHistorySyncForWebView(deviceID string, evt *events.HistorySync) {
	logrus.Infof("Processing history sync for device %s - Type: %s", deviceID, evt.Data.GetSyncType())
	
	cm := GetClientManager()
	client, err := cm.GetClient(deviceID)
	if err != nil {
		logrus.Errorf("Failed to get client for history sync: %v", err)
		return
	}
	
	messageCount := 0
	
	// Process conversations from history sync
	for _, conv := range evt.Data.GetConversations() {
		// Skip if no ID
		if conv.GetId() == "" {
			continue
		}
		
		// Parse chat JID
		chatJID, err := types.ParseJID(conv.GetId())
		if err != nil {
			continue
		}
		
		// Skip groups and broadcasts for WhatsApp Web
		if chatJID.Server != types.DefaultUserServer {
			continue
		}
		
		// Process messages in this conversation
		for _, historyMsg := range conv.GetMessages() {
			webMsg := historyMsg.GetMessage()
			if webMsg == nil {
				continue
			}
			
			// Parse the web message to get a proper Message event
			parsedMsg, err := client.ParseWebMessage(chatJID, webMsg)
			if err != nil {
				logrus.Debugf("Failed to parse history message: %v", err)
				continue
			}
			
			// Extract message details
			messageID := webMsg.GetKey().GetId()
			senderJID := webMsg.GetKey().GetFromMe()
			timestamp := webMsg.GetMessageTimestamp()
			
			// Get sender JID string
			var senderStr string
			if senderJID {
				senderStr = client.Store.ID.String()
			} else {
				if webMsg.GetParticipant() != "" {
					senderStr = webMsg.GetParticipant()
				} else {
					senderStr = chatJID.String()
				}
			}
			
			// Extract message text
			messageText := extractMessageFromParsed(parsedMsg)
			messageType := getMessageType(parsedMsg.Message)
			
			// Store in database with proper timestamp
			StoreHistoryMessage(
				deviceID,
				chatJID.String(),
				messageID,
				senderStr,
				messageText,
				messageType,
				int64(timestamp),
			)
			
			messageCount++
		}
	}
	
	logrus.Infof("Processed %d messages from history sync for device %s", messageCount, deviceID)
}

// extractMessageFromParsed extracts text from a parsed message
func extractMessageFromParsed(msg *events.Message) string {
	if msg.Message == nil {
		return ""
	}
	
	// Try different message types
	if text := msg.Message.GetConversation(); text != "" {
		return text
	}
	if extText := msg.Message.GetExtendedTextMessage(); extText != nil {
		return extText.GetText()
	}
	if imageMsg := msg.Message.GetImageMessage(); imageMsg != nil {
		return imageMsg.GetCaption()
	}
	if videoMsg := msg.Message.GetVideoMessage(); videoMsg != nil {
		return videoMsg.GetCaption()
	}
	if docMsg := msg.Message.GetDocumentMessage(); docMsg != nil {
		return "📄 " + docMsg.GetFileName()
	}
	
	return ""
}

// getMessageType determines the type of message
func getMessageType(msg *waProto.Message) string {
	if msg.GetImageMessage() != nil {
		return "image"
	}
	if msg.GetVideoMessage() != nil {
		return "video"
	}
	if msg.GetAudioMessage() != nil {
		return "audio"
	}
	if msg.GetDocumentMessage() != nil {
		return "document"
	}
	if msg.GetStickerMessage() != nil {
		return "sticker"
	}
	return "text"
}

// extractMessageText extracts text from various message types
func extractMessageText(evt *events.Message) string {
	if evt.Message == nil {
		return ""
	}
	
	// Regular text message
	if text := evt.Message.GetConversation(); text != "" {
		return text
	}
	
	// Extended text message
	if extText := evt.Message.GetExtendedTextMessage(); extText != nil {
		return extText.GetText()
	}
	
	// Image caption
	if img := evt.Message.GetImageMessage(); img != nil && img.GetCaption() != "" {
		return img.GetCaption()
	}
	
	// Video caption
	if vid := evt.Message.GetVideoMessage(); vid != nil && vid.GetCaption() != "" {
		return vid.GetCaption()
	}
	
	// Document filename
	if doc := evt.Message.GetDocumentMessage(); doc != nil {
		return "📄 " + doc.GetFileName()
	}
	
	// Audio/voice message
	if evt.Message.GetAudioMessage() != nil {
		return "🎵 Voice message"
	}
	
	// Sticker
	if evt.Message.GetStickerMessage() != nil {
		return "Sticker"
	}
	
	// Location
	if loc := evt.Message.GetLocationMessage(); loc != nil {
		return "📍 Location"
	}
	
	// Contact
	if evt.Message.GetContactMessage() != nil {
		return "👤 Contact"
	}
	
	// Poll
	if evt.Message.GetPollCreationMessage() != nil {
		return "📊 Poll"
	}
	
	return ""
}

// StoreHistoryMessage stores a message from history sync with specific timestamp
func StoreHistoryMessage(deviceID, chatJID, messageID, senderJID, messageText, messageType string, timestamp int64) {
	StoreWhatsAppMessageWithMediaAndTimestamp(deviceID, chatJID, messageID, senderJID, messageText, messageType, "", timestamp)
}
