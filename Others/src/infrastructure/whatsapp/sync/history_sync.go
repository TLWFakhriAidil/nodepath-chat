package sync

import (
	"context"
	"fmt"
	"time"
	
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
)

// HistorySyncManager handles syncing message history from WhatsApp
type HistorySyncManager struct {
	client   *whatsmeow.Client
	deviceID string
	logger   *logrus.Logger
}

// NewHistorySyncManager creates a new history sync manager
func NewHistorySyncManager(client *whatsmeow.Client, deviceID string) *HistorySyncManager {
	return &HistorySyncManager{
		client:   client,
		deviceID: deviceID,
		logger:   logrus.New(),
	}
}

// RequestHistorySync initiates a history sync request
func (hsm *HistorySyncManager) RequestHistorySync() error {
	hsm.logger.Infof("Requesting history sync for device %s", hsm.deviceID)
	
	if !hsm.client.IsConnected() {
		return fmt.Errorf("client not connected")
	}
	
	// Build history sync request - request recent messages
	// nil = sync all chats, 50 = recent message count per chat
	historyMsg := hsm.client.BuildHistorySyncRequest(nil, 50)
	if historyMsg == nil {
		return fmt.Errorf("failed to build history sync request")
	}
	
	// Send to WhatsApp status broadcast
	_, err := hsm.client.SendMessage(context.Background(), types.JID{
		Server: "s.whatsapp.net",
		User:   "status",
	}, historyMsg)
	
	if err != nil {
		return fmt.Errorf("failed to send history sync request: %v", err)
	}
	
	hsm.logger.Info("History sync requested successfully")
	return nil
}

// ProcessHistorySync processes a history sync event
func (hsm *HistorySyncManager) ProcessHistorySync(evt *events.HistorySync) {
	hsm.logger.Infof("Processing history sync for device %s - Type: %s, Progress: %d%%", 
		hsm.deviceID, evt.Data.GetSyncType(), evt.Data.GetProgress())
	
	messageCount := 0
	chatCount := 0
	
	// Process each conversation
	for _, conv := range evt.Data.GetConversations() {
		if conv.GetId() == "" {
			continue
		}
		
		// Parse chat JID
		chatJID, err := types.ParseJID(conv.GetId())
		if err != nil {
			hsm.logger.Warnf("Failed to parse JID %s: %v", conv.GetId(), err)
			continue
		}
		
		// Skip non-personal chats for WhatsApp Web view
		if chatJID.Server != types.DefaultUserServer {
			continue
		}
		
		// Skip status updates
		if chatJID.User == "status" {
			continue
		}
		
		chatCount++
		
		// Process messages in this conversation
		for _, historyMsg := range conv.GetMessages() {
			webMsg := historyMsg.GetMessage()
			if webMsg == nil || webMsg.GetKey() == nil {
				continue
			}
			
			// Extract message details
			messageID := webMsg.GetKey().GetId()
			timestamp := webMsg.GetMessageTimestamp()
			isFromMe := webMsg.GetKey().GetFromMe()
			
			// Get sender JID
			var senderJID string
			if isFromMe {
				senderJID = hsm.client.Store.ID.String()
			} else {
				if participant := webMsg.GetKey().GetParticipant(); participant != "" {
					senderJID = participant
				} else {
					senderJID = chatJID.String()
				}
			}
			
			// Extract message content and type
			messageText, messageType := hsm.extractMessageContent(webMsg.GetMessage())
			
			// Skip empty messages
			if messageText == "" && messageType == "text" {
				continue
			}
			
			// Store the message
			hsm.storeMessage(chatJID.String(), messageID, senderJID, messageText, messageType, timestamp)
			messageCount++
		}
	}
	
	hsm.logger.Infof("History sync processed: %d messages from %d chats", messageCount, chatCount)
}

// extractMessageContent extracts text and type from a message
func (hsm *HistorySyncManager) extractMessageContent(msg *waProto.Message) (string, string) {
	if msg == nil {
		return "", "text"
	}
	
	// Text messages
	if text := msg.GetConversation(); text != "" {
		return text, "text"
	}
	if extText := msg.GetExtendedTextMessage(); extText != nil {
		return extText.GetText(), "text"
	}
	
	// Media messages
	if img := msg.GetImageMessage(); img != nil {
		caption := img.GetCaption()
		if caption == "" {
			caption = "📷 Photo"
		}
		return caption, "image"
	}
	
	if vid := msg.GetVideoMessage(); vid != nil {
		caption := vid.GetCaption()
		if caption == "" {
			caption = "📹 Video"
		}
		return caption, "video"
	}
	
	if aud := msg.GetAudioMessage(); aud != nil {
		if aud.GetPTT() {
			return "🎤 Voice message", "audio"
		}
		return "🎵 Audio", "audio"
	}
	
	if doc := msg.GetDocumentMessage(); doc != nil {
		filename := doc.GetFileName()
		if filename == "" {
			filename = "Document"
		}
		return "📄 " + filename, "document"
	}
	
	if msg.GetStickerMessage() != nil {
		return "Sticker", "sticker"
	}
	
	if loc := msg.GetLocationMessage(); loc != nil {
		return "📍 Location", "location"
	}
	
	if msg.GetContactMessage() != nil {
		return "👤 Contact", "contact"
	}
	
	if msg.GetPollCreationMessage() != nil {
		return "📊 Poll", "poll"
	}
	
	return "", "text"
}

// storeMessage stores a message in the database
func (hsm *HistorySyncManager) storeMessage(chatJID, messageID, senderJID, messageText, messageType string, timestamp uint64) {
	// Import the whatsapp package to use StoreWhatsAppMessageWithTimestamp
	// We'll use a function from the parent package
	StoreHistorySyncMessage(hsm.deviceID, chatJID, messageID, senderJID, messageText, messageType, int64(timestamp))
}

// BuildHistorySyncRequest creates a custom history sync request
func BuildHistorySyncRequest(client *whatsmeow.Client, jid *types.JID, messageCount int) *waProto.Message {
	if messageCount <= 0 {
		messageCount = 50
	}
	
	// Create history sync request
	historySyncRequest := &waProto.HistorySyncNotification{
		FileSha256:    proto.String(""),
		FileLength:    proto.Uint64(0),
		MediaKey:      []byte{},
		FileEncSha256: []byte{},
		DirectPath:    proto.String(""),
		SyncType:      waProto.HistorySyncNotification_RECENT.Enum(),
		ChunkOrder:    proto.Uint32(1),
	}
	
	return &waProto.Message{
		HistorySyncNotification: historySyncRequest,
	}
}
