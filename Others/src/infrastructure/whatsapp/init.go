package whatsapp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	pkgError "github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/error"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/ui/websocket"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/appstate"
	"go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"google.golang.org/protobuf/proto"
)

// Type definitions
type ExtractedMedia struct {
	MediaPath string `json:"media_path"`
	MimeType  string `json:"mime_type"`
	Caption   string `json:"caption"`
}

type evtReaction struct {
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}

type evtMessage struct {
	ID            string `json:"id,omitempty"`
	Text          string `json:"text,omitempty"`
	RepliedId     string `json:"replied_id,omitempty"`
	QuotedMessage string `json:"quoted_message,omitempty"`
}

// Global variables
var (
	cli           *whatsmeow.Client
	log           waLog.Logger
	historySyncID int32
	startupTime   = time.Now().Unix()
)

// InitWaDB initializes the WhatsApp database connection
func InitWaDB(ctx context.Context) *sqlstore.Container {
	log = waLog.Stdout("Main", config.WhatsappLogLevel, true)
	dbLog := waLog.Stdout("Database", config.WhatsappLogLevel, true)

	storeContainer, err := initDatabase(ctx, dbLog)
	if err != nil {
		log.Errorf("Database initialization error: %v", err)
		panic(pkgError.InternalServerError(fmt.Sprintf("Database initialization error: %v", err)))
	}

	return storeContainer
}

// initDatabase creates and returns a database store container based on the configured URI
func initDatabase(ctx context.Context, dbLog waLog.Logger) (*sqlstore.Container, error) {
	if strings.HasPrefix(config.DBURI, "file:") {
		return sqlstore.New(ctx, "sqlite3", config.DBURI, dbLog)
	} else if strings.HasPrefix(config.DBURI, "postgres:") || strings.HasPrefix(config.DBURI, "postgresql:") {
		// Convert postgresql:// to postgres:// for the driver
		dbUri := config.DBURI
		if strings.HasPrefix(dbUri, "postgresql://") {
			dbUri = strings.Replace(dbUri, "postgresql://", "postgres://", 1)
		}
		return sqlstore.New(ctx, "postgres", dbUri, dbLog)
	}

	return nil, fmt.Errorf("unknown database type: %s. Currently only sqlite3(file:) and postgres/postgresql are supported", config.DBURI)
}

// InitWaCLI initializes the WhatsApp client
func InitWaCLI(ctx context.Context, storeContainer *sqlstore.Container) *whatsmeow.Client {
	device, err := storeContainer.GetFirstDevice(ctx)
	if err != nil {
		log.Errorf("Failed to get device: %v", err)
		panic(err)
	}

	if device == nil {
		log.Infof("No device found - devices will be created when users add them")
		// For multi-device support, return nil instead of panicking
		// Devices will be created dynamically when users add them
		return nil
	}

	// Configure device properties
	osName := fmt.Sprintf("%s %s", config.AppOs, config.AppVersion)
	store.DeviceProps.PlatformType = &config.AppPlatform
	store.DeviceProps.Os = &osName

	// Create and configure the client
	cli = whatsmeow.NewClient(device, waLog.Stdout("Client", config.WhatsappLogLevel, true))
	cli.EnableAutoReconnect = true
	cli.AutoTrustIdentity = true
	cli.AddEventHandler(func(rawEvt interface{}) {
		// Process all events asynchronously to prevent WebSocket blocking
		go handler(ctx, rawEvt)
	})
	
	// Set as global client for debugging
	SetGlobalClient(cli)

	return cli
}

// handler is the main event handler for WhatsApp events
func handler(ctx context.Context, rawEvt interface{}) {
	switch evt := rawEvt.(type) {
	case *events.DeleteForMe:
		handleDeleteForMe(ctx, evt)
	case *events.AppStateSyncComplete:
		handleAppStateSyncComplete(ctx, evt)
	case *events.PairSuccess:
		handlePairSuccess(ctx, evt)
	case *events.LoggedOut:
		handleLoggedOut(ctx)
	case *events.Connected, *events.PushNameSetting:
		handleConnectionEvents(ctx)
	case *events.StreamReplaced:
		handleStreamReplaced(ctx)
	case *events.Message:
		handleMessage(ctx, evt)
	case *events.Receipt:
		handleReceipt(ctx, evt)
	case *events.Presence:
		handlePresence(ctx, evt)
	case *events.HistorySync:
		handleHistorySync(ctx, evt)
	case *events.AppState:
		handleAppState(ctx, evt)
	}
}

// Event handler functions

func handleDeleteForMe(_ context.Context, evt *events.DeleteForMe) {
	log.Infof("Deleted message %s for %s", evt.MessageID, evt.SenderJID.String())
}

func handleAppStateSyncComplete(_ context.Context, evt *events.AppStateSyncComplete) {
	if len(cli.Store.PushName) > 0 && evt.Name == appstate.WAPatchCriticalBlock {
		if err := cli.SendPresence(types.PresenceAvailable); err != nil {
			log.Warnf("Failed to send available presence: %v", err)
		} else {
			log.Infof("Marked self as available")
		}
	}
}

func handlePairSuccess(_ context.Context, evt *events.PairSuccess) {
	log.Infof("Pair success! ID: %s, BusinessName: %s, Platform: %s", 
		evt.ID.String(), evt.BusinessName, evt.Platform)
	
	// Log the phone number for debugging
	if evt.ID.User != "" {
		log.Infof("Phone number from pair success: %s", evt.ID.User)
		
		// Try to find and update the device immediately
		userRepo := repository.GetUserRepository()
		for userID, session := range connectionSessions {
			if session != nil && session.DeviceID != "" {
				log.Infof("Found session for user %s, updating device %s", userID, session.DeviceID)
				// Update with phone number from pair success
				err := userRepo.UpdateDeviceStatus(session.DeviceID, "connecting", evt.ID.User, evt.ID.String())
				if err != nil {
					log.Errorf("Failed to update device on pair success: %v", err)
				}
				break
			}
		}
	}
	
	// Update device status in database
	if userRepo := repository.GetUserRepository(); userRepo != nil {
		// TODO: Get current user and device ID from context
		// For now, just broadcast the success
		log.Infof("Device paired successfully, waiting for full connection...")
	}
	
	websocket.Broadcast <- websocket.BroadcastMessage{
		Code:    "LOGIN_SUCCESS",
		Message: fmt.Sprintf("Successfully paired with %s", evt.ID.String()),
	}
	
	// The device is paired but not fully connected yet
	// Wait for Connected event to update status to "online"
}

func handleLoggedOut(_ context.Context) {
	websocket.Broadcast <- websocket.BroadcastMessage{
		Code:   "LIST_DEVICES",
		Result: nil,
	}
}

func handleConnectionEvents(_ context.Context) {
	log.Infof("WhatsApp connection event received")
	
	// DISABLED: Global client handler causes multi-device issues
	// Each device should handle its own connection events through device-specific handlers
	// See app.go Login() function for device-specific handling
	return
}

func handleStreamReplaced(_ context.Context) {
	os.Exit(0)
}

func handleMessage(ctx context.Context, evt *events.Message) {
	// Log message metadata
	metaParts := buildMessageMetaParts(evt)
	log.Infof("Received message %s from %s (%s): %+v",
		evt.Info.ID,
		evt.Info.SourceString(),
		strings.Join(metaParts, ", "),
		evt.Message,
	)

	// Record the message with analytics
	message := ExtractMessageText(evt)
	utils.RecordMessage(evt.Info.ID, evt.Info.Sender.String(), message)
	
	log.Infof("WhatsappChatStorage enabled: %v", config.WhatsappChatStorage)
	
	// Save message and chat info like whatsapp-mcp-main does
	if config.WhatsappChatStorage {
		cm := GetClientManager()
		allClients := cm.GetAllClients()
		for deviceID, client := range allClients {
			// Skip if not the right client
			if client != cli {
				continue
			}
			
			// Store the message immediately for real-time updates
			if evt.Info.Chat.Server == types.DefaultUserServer && !evt.Info.IsGroup {
				// Update chat info first
				HandleMessageForChats(deviceID, client, evt)
				
				// Store the message
				HandleMessageForWebView(deviceID, evt)
				
				log.Debugf("Stored message in real-time for device %s in chat %s", deviceID, evt.Info.Chat.String())
			}
			
			// Also register for real-time sync
			EnableRealtimeSync(deviceID, client, evt)
			
			break
		}
	}

	// Handle image message if present
	handleImageMessage(ctx, evt)

	// Handle auto-reply if configured
	handleAutoReply(evt)

	// Forward to webhook if configured
	handleWebhookForward(ctx, evt)
}

func buildMessageMetaParts(evt *events.Message) []string {
	metaParts := []string{
		fmt.Sprintf("pushname: %s", evt.Info.PushName),
		fmt.Sprintf("timestamp: %s", evt.Info.Timestamp),
	}
	if evt.Info.Type != "" {
		metaParts = append(metaParts, fmt.Sprintf("type: %s", evt.Info.Type))
	}
	if evt.Info.Category != "" {
		metaParts = append(metaParts, fmt.Sprintf("category: %s", evt.Info.Category))
	}
	if evt.IsViewOnce {
		metaParts = append(metaParts, "view once")
	}
	return metaParts
}

func handleImageMessage(ctx context.Context, evt *events.Message) {
	if img := evt.Message.GetImageMessage(); img != nil {
		if path, err := ExtractMedia(ctx, config.PathStorages, img); err != nil {
			log.Errorf("Failed to download image: %v", err)
		} else {
			log.Infof("Image downloaded to %s", path)
		}
	}
}

func handleAutoReply(evt *events.Message) {
	if config.WhatsappAutoReplyMessage != "" &&
		!isGroupJid(evt.Info.Chat.String()) &&
		!evt.Info.IsIncomingBroadcast() &&
		evt.Message.GetExtendedTextMessage() != nil &&
		evt.Message.GetExtendedTextMessage().GetText() != "" {
		_, _ = cli.SendMessage(
			context.Background(),
			FormatJID(evt.Info.Sender.String()),
			&waE2E.Message{Conversation: proto.String(config.WhatsappAutoReplyMessage)},
		)
	}
}

func handleWebhookForward(ctx context.Context, evt *events.Message) {
	if len(config.WhatsappWebhook) > 0 &&
		!strings.Contains(evt.Info.SourceString(), "broadcast") &&
		!isFromMySelf(evt.Info.SourceString()) {
		go func(evt *events.Message) {
			if err := forwardToWebhook(ctx, evt); err != nil {
				logrus.Error("Failed forward to webhook: ", err)
			}
		}(evt)
	}
}

func handleReceipt(ctx context.Context, evt *events.Receipt) {
	if evt.Type == types.ReceiptTypeRead || evt.Type == types.ReceiptTypeReadSelf {
		log.Infof("%v was read by %s at %s", evt.MessageIDs, evt.SourceString(), evt.Timestamp)
		// Update message status to "read" in database
		if analyticsRepo := ctx.Value("analyticsRepo"); analyticsRepo != nil {
			if repo, ok := analyticsRepo.(*repository.MessageAnalyticsRepository); ok {
				for _, msgID := range evt.MessageIDs {
					repo.UpdateMessageStatus(msgID, "read")
				}
			}
		}
	} else if evt.Type == types.ReceiptTypeDelivered {
		log.Infof("%s was delivered to %s at %s", evt.MessageIDs[0], evt.SourceString(), evt.Timestamp)
		// Update message status to "delivered" in database
		if analyticsRepo := ctx.Value("analyticsRepo"); analyticsRepo != nil {
			if repo, ok := analyticsRepo.(*repository.MessageAnalyticsRepository); ok {
				for _, msgID := range evt.MessageIDs {
					repo.UpdateMessageStatus(msgID, "delivered")
				}
			}
		}
	}
}

func handlePresence(_ context.Context, evt *events.Presence) {
	if evt.Unavailable {
		if evt.LastSeen.IsZero() {
			log.Infof("%s is now offline", evt.From)
		} else {
			log.Infof("%s is now offline (last seen: %s)", evt.From, evt.LastSeen)
		}
	} else {
		log.Infof("%s is now online", evt.From)
	}
}

func handleHistorySync(_ context.Context, evt *events.HistorySync) {
	log.Infof("=== HISTORY SYNC RECEIVED! Type: %s, Progress: %d%% ===", 
		evt.Data.GetSyncType(), evt.Data.GetProgress())
	
	// Process history sync for WhatsApp Web if enabled
	if config.WhatsappChatStorage {
		// Get all connected clients
		cm := GetClientManager()
		allClients := cm.GetAllClients()
		
		// Find the device ID for this history sync
		for deviceID, client := range allClients {
			if client == cli {
				log.Infof("Processing history sync for device %s", deviceID)
				
				// Create tables if needed
				CreateChatTable()
				
				// Process each conversation
				conversationCount := 0
				messageCount := 0
				
				for _, conv := range evt.Data.GetConversations() {
					if conv.GetId() == "" {
						continue
					}
					
					// Parse chat JID
					chatJID, err := types.ParseJID(conv.GetId())
					if err != nil {
						continue
					}
					
					// Skip non-personal chats
					if chatJID.Server != types.DefaultUserServer {
						continue
					}
					
					conversationCount++
					
					// Get chat name
					chatName := GetChatName(client, chatJID, conv.GetId())
					
					// Get last message time
					var lastMessageTime time.Time
					if len(conv.GetMessages()) > 0 {
						firstMsg := conv.GetMessages()[0]
						if firstMsg != nil && firstMsg.GetMessage() != nil {
							timestamp := firstMsg.GetMessage().GetMessageTimestamp()
							if timestamp > 0 {
								lastMessageTime = time.Unix(int64(timestamp), 0)
							}
						}
					}
					
					if lastMessageTime.IsZero() {
						lastMessageTime = time.Now()
					}
					
					// Store chat
					err = StoreChat(deviceID, conv.GetId(), chatName, lastMessageTime)
					if err != nil {
						log.Errorf("Failed to store chat %s: %v", conv.GetId(), err)
					} else {
						log.Debugf("Stored chat: %s (%s)", chatName, conv.GetId())
					}
					
					// Process messages
					for _, historyMsg := range conv.GetMessages() {
						webMsg := historyMsg.GetMessage()
						if webMsg == nil || webMsg.GetKey() == nil {
							continue
						}
						
						messageID := webMsg.GetKey().GetId()
						timestamp := webMsg.GetMessageTimestamp()
						isFromMe := webMsg.GetKey().GetFromMe()
						
						var senderJID string
						if isFromMe {
							senderJID = client.Store.ID.String()
						} else {
							senderJID = chatJID.String()
						}
						
						// Extract message content
						messageText := ""
						messageType := "text"
						
						if msg := webMsg.GetMessage(); msg != nil {
							if text := msg.GetConversation(); text != "" {
								messageText = text
							} else if extText := msg.GetExtendedTextMessage(); extText != nil {
								messageText = extText.GetText()
							} else if img := msg.GetImageMessage(); img != nil {
								messageText = img.GetCaption()
								if messageText == "" {
									messageText = "📷 Photo"
								}
								messageType = "image"
							} else if vid := msg.GetVideoMessage(); vid != nil {
								messageText = vid.GetCaption()
								if messageText == "" {
									messageText = "📹 Video"
								}
								messageType = "video"
							} else if aud := msg.GetAudioMessage(); aud != nil {
								if aud.GetPTT() {
									messageText = "🎤 Voice message"
								} else {
									messageText = "🎵 Audio"
								}
								messageType = "audio"
							} else if doc := msg.GetDocumentMessage(); doc != nil {
								messageText = "📄 " + doc.GetFileName()
								messageType = "document"
							}
						}
						
						if messageText != "" {
							StoreWhatsAppMessageWithTimestamp(deviceID, chatJID.String(), messageID, senderJID, messageText, messageType, int64(timestamp))
							messageCount++
						}
					}
				}
				
				log.Infof("=== History sync complete: %d conversations, %d messages stored ===", conversationCount, messageCount)
				break
			}
		}
	}
	
	// Also save to file as before
	id := atomic.AddInt32(&historySyncID, 1)
	fileName := fmt.Sprintf("%s/history-%d-%s-%d-%s.json",
		config.PathStorages,
		startupTime,
		cli.Store.ID.String(),
		id,
		evt.Data.SyncType.String(),
	)

	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Errorf("Failed to open file to write history sync: %v", err)
		return
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	enc.SetIndent("", "  ")
	if err = enc.Encode(evt.Data); err != nil {
		log.Errorf("Failed to write history sync: %v", err)
		return
	}

	log.Infof("Wrote history sync to %s", fileName)
}

func handleAppState(ctx context.Context, evt *events.AppState) {
	log.Debugf("App state event: %+v / %+v", evt.Index, evt.SyncActionValue)
}
