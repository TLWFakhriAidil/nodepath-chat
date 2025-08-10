package broadcast

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	waProto "go.mau.fi/whatsmeow/binary/proto"
	"google.golang.org/protobuf/proto"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/antipattern"
)

// DeviceWorker handles broadcasting for a single device
type DeviceWorker struct {
	deviceID        string
	client          *whatsmeow.Client
	minDelay        int
	maxDelay        int
	messageQueue    chan domainBroadcast.BroadcastMessage
	ctx             context.Context
	cancel          context.CancelFunc
	mu              sync.RWMutex
	status          string
	processedCount  int
	failedCount     int
	lastActivity    time.Time
	broadcastRepo   *repository.BroadcastRepository
	greetingProcessor *antipattern.GreetingProcessor
	messageRandomizer *antipattern.MessageRandomizer
}

// NewDeviceWorker creates a new device worker
func NewDeviceWorker(deviceID string, client *whatsmeow.Client, minDelay, maxDelay int) *DeviceWorker {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &DeviceWorker{
		deviceID:      deviceID,
		client:        client,
		minDelay:      minDelay,
		maxDelay:      maxDelay,
		messageQueue:  make(chan domainBroadcast.BroadcastMessage, config.WorkerQueueSize),
		ctx:           ctx,
		cancel:        cancel,
		status:        "idle",
		lastActivity:  time.Now(),
		broadcastRepo: repository.GetBroadcastRepository(),
		greetingProcessor: antipattern.NewGreetingProcessor(),
		messageRandomizer: antipattern.NewMessageRandomizer(),
	}
}

// Start starts the worker
func (dw *DeviceWorker) Start() {
	go dw.processMessages()
	// go dw.healthCheck() // DISABLED - Device Health Monitor is the single source of truth
}

// Stop stops the worker
func (dw *DeviceWorker) Stop() {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	
	if dw.cancel != nil {
		dw.cancel()
	}
	close(dw.messageQueue)
	dw.status = "stopped"
}

// QueueMessage adds a message to the worker's queue
func (dw *DeviceWorker) QueueMessage(msg domainBroadcast.BroadcastMessage) error {
	select {
	case dw.messageQueue <- msg:
		return nil
	case <-time.After(time.Second * 5):
		return fmt.Errorf("queue full for device %s", dw.deviceID)
	}
}

// GetStatus returns worker status
func (dw *DeviceWorker) GetStatus() domainBroadcast.WorkerStatus {
	dw.mu.RLock()
	defer dw.mu.RUnlock()
	
	return domainBroadcast.WorkerStatus{
		DeviceID:       dw.deviceID,
		Status:         dw.status,
		QueueSize:      len(dw.messageQueue),
		ProcessedCount: dw.processedCount,
		FailedCount:    dw.failedCount,
		LastActivity:   dw.lastActivity,
	}
}

// processMessages processes messages from the queue
func (dw *DeviceWorker) processMessages() {
	for {
		select {
		case <-dw.ctx.Done():
			return
		case msg, ok := <-dw.messageQueue:
			if !ok {
				return
			}
			
			dw.mu.Lock()
			dw.status = "processing"
			dw.lastActivity = time.Now()
			dw.mu.Unlock()
			
			// CRITICAL: Double-check message wasn't already sent to prevent duplicates
			db := database.GetDB()
			var currentStatus string
			checkErr := db.QueryRow("SELECT status FROM broadcast_messages WHERE id = ?", msg.ID).Scan(&currentStatus)
			if checkErr == nil && currentStatus == "sent" {
				logrus.Warnf("Message %s already sent, skipping to prevent duplicate", msg.ID)
				dw.mu.Lock()
				dw.status = "idle"
				dw.mu.Unlock()
				continue
			}
			
			// Process the message
			err := dw.sendMessage(msg)
			
			dw.mu.Lock()
			if err != nil {
				dw.failedCount++
				logrus.Errorf("Failed to send message: %v", err)
				// Update broadcast status to failed
				if msg.ID != "" {
					// Direct update like skipped
					db := database.GetDB()
					_, updateErr := db.Exec(`UPDATE broadcast_messages SET STATUS = 'failed', error_message = ?, updated_at = NOW() WHERE id = ?`, err.Error(), msg.ID)
					if updateErr != nil {
						logrus.Errorf("Failed to update message status to failed: %v", updateErr)
					}
				}
			} else {
				dw.processedCount++
				// Update broadcast status to sent
				if msg.ID != "" {
					// Direct update like skipped
					db := database.GetDB()
					_, updateErr := db.Exec(`UPDATE broadcast_messages SET status = 'sent', sent_at = NOW(), updated_at = NOW() WHERE id = ? AND status IN ('processing', 'pending', 'queued')`, msg.ID)
					if updateErr != nil {
						logrus.Errorf("Failed to update message status to sent: %v", updateErr)
					}
					logrus.Infof("Message %s sent successfully to %s", msg.ID, msg.RecipientPhone)
					
					// Update campaign processing timestamp if this is a campaign message
					if msg.CampaignID != nil {
						_, updateErr = db.Exec(`
							UPDATE campaigns 
							SET updated_at = NOW()
							WHERE id = ?
						`, *msg.CampaignID)
						if updateErr != nil {
							logrus.Errorf("Failed to update campaign processing time: %v", updateErr)
						}
					}
				}
			}
			dw.status = "idle"
			dw.mu.Unlock()
			
			// Determine delay based on message's campaign/sequence settings
			var delay time.Duration
			
			// Use delays from the message (set by campaign/sequence)
			if msg.MinDelay > 0 || msg.MaxDelay > 0 {
				delay = getRandomDelayBetween(msg.MinDelay, msg.MaxDelay)
				logrus.Debugf("Using campaign/sequence delays: %d-%d seconds", msg.MinDelay, msg.MaxDelay)
			} else {
				// Fallback to device defaults if message doesn't have delays
				delay = dw.getRandomDelay()
				logrus.Debugf("Using device default delays: %d-%d seconds", dw.minDelay, dw.maxDelay)
			}
			
			// Log the message processing
			logrus.Infof("Processing message for %s - Type: %s, Has image: %v, Has text: %v, Delay: %v", 
				msg.RecipientPhone, msg.Type, msg.MediaURL != "", msg.Content != "", delay)
			
			time.Sleep(delay)
		}
	}
}

// sendMessage sends a message based on type
func (dw *DeviceWorker) sendMessage(msg domainBroadcast.BroadcastMessage) error {
	// Parse recipient JID
	recipient, err := whatsapp.ParseJID(msg.RecipientPhone)
	if err != nil {
		return fmt.Errorf("invalid recipient: %v", err)
	}
	
	switch msg.Type {
	case "text":
		return dw.sendTextMessage(recipient, msg)
	case "image":
		return dw.sendImageMessage(recipient, msg)
	case "video":
		return dw.sendVideoMessage(recipient, msg)
	case "document":
		return dw.sendDocumentMessage(recipient, msg)
	default:
		return fmt.Errorf("unsupported message type: %s", msg.Type)
	}
}

// sendTextMessage sends text message
func (dw *DeviceWorker) sendTextMessage(recipient types.JID, msg domainBroadcast.BroadcastMessage) error {
	// Check if this is a platform device
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByID(dw.deviceID)
	if err != nil {
		return fmt.Errorf("failed to get device info: %v", err)
	}
	
	// If platform device, just send raw content (platform sender will handle greeting/anti-spam)
	if device != nil && device.Platform != "" {
		logrus.Debugf("Platform device %s detected, skipping greeting/anti-spam", device.Platform)
		message := &waProto.Message{
			Conversation: proto.String(msg.Content),
		}
		_, err := dw.client.SendMessage(context.Background(), recipient, message)
		return err
	}
	
	// For regular WhatsApp Web devices, apply greeting and anti-spam
	// STEP 1: Apply randomization to CONTENT ONLY
	randomizedContent := dw.messageRandomizer.RandomizeMessage(msg.Content)
	
	// STEP 2: Add greeting to the randomized content
	finalContent := dw.greetingProcessor.PrepareMessageWithGreeting(
		randomizedContent,
		msg.RecipientName,
		dw.deviceID,
		msg.RecipientPhone,
	)
	
	// Use simple Conversation message instead of ExtendedTextMessage for better compatibility
	message := &waProto.Message{
		Conversation: proto.String(finalContent),
	}
	
	_, err = dw.client.SendMessage(context.Background(), recipient, message)
	return err
}

// sendImageMessage sends image message with caption
func (dw *DeviceWorker) sendImageMessage(recipient types.JID, msg domainBroadcast.BroadcastMessage) error {
	// Download image from URL
	imageData, err := downloadMedia(msg.MediaURL)
	if err != nil {
		return fmt.Errorf("failed to download image: %v", err)
	}
	
	// Upload to WhatsApp servers
	uploaded, err := dw.client.Upload(context.Background(), imageData, whatsmeow.MediaImage)
	if err != nil {
		return fmt.Errorf("failed to upload image: %v", err)
	}
	
	// Process caption with spintax (same as text messages)
	processedCaption := ""
	if msg.Content != "" {
		// Check if this is a platform device
		userRepo := repository.GetUserRepository()
		device, err := userRepo.GetDeviceByID(dw.deviceID)
		if err == nil && device != nil && device.Platform != "" {
			// Platform device - just use raw content
			processedCaption = msg.Content
		} else {
			// Regular WhatsApp device - apply greeting and anti-spam
			// STEP 1: Apply randomization to CAPTION ONLY
			randomizedCaption := dw.messageRandomizer.RandomizeMessage(msg.Content)
			
			// STEP 2: Add greeting to the randomized caption
			processedCaption = dw.greetingProcessor.PrepareMessageWithGreeting(
				randomizedCaption,
				msg.RecipientName,
				dw.deviceID,
				msg.RecipientPhone,
			)
		}
	}
	
	// Create image message
	message := &waProto.Message{
		ImageMessage: &waProto.ImageMessage{
			Caption:       &processedCaption,
			URL:           &uploaded.URL,
			DirectPath:    &uploaded.DirectPath,
			MediaKey:      uploaded.MediaKey,
			FileEncSHA256: uploaded.FileEncSHA256,
			FileSHA256:    uploaded.FileSHA256,
			FileLength:    &uploaded.FileLength,
			Mimetype:      proto.String("image/jpeg"),
		},
	}
	
	_, err = dw.client.SendMessage(context.Background(), recipient, message)
	return err
}

// sendVideoMessage sends video message
func (dw *DeviceWorker) sendVideoMessage(recipient types.JID, msg domainBroadcast.BroadcastMessage) error {
	// TODO: Implement video sending
	return fmt.Errorf("video messages not yet implemented")
}

// sendDocumentMessage sends document message
func (dw *DeviceWorker) sendDocumentMessage(recipient types.JID, msg domainBroadcast.BroadcastMessage) error {
	// TODO: Implement document sending
	return fmt.Errorf("document messages not yet implemented")
}

// getRandomDelay returns a random delay between min and max
func (dw *DeviceWorker) getRandomDelay() time.Duration {
	if dw.minDelay == dw.maxDelay {
		return time.Duration(dw.minDelay) * time.Second
	}
	
	delay := rand.Intn(dw.maxDelay-dw.minDelay) + dw.minDelay
	return time.Duration(delay) * time.Second
}

// healthCheck monitors worker health - DISABLED (using Device Health Monitor instead)
/*
func (dw *DeviceWorker) healthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-dw.ctx.Done():
			return
		case <-ticker.C:
			dw.mu.RLock()
			lastActivity := dw.lastActivity
			status := dw.status
			dw.mu.RUnlock()
			
			// Check if worker is stuck
			if status == "processing" && time.Since(lastActivity) > 10*time.Minute {
				logrus.Warnf("Worker %s appears stuck, restarting...", dw.deviceID)
				// TODO: Implement restart logic
			}
		}
	}
}
*/

// SendMessage sends a message (implements direct sending without queue)
func (dw *DeviceWorker) SendMessage(msg domainBroadcast.BroadcastMessage) error {
	return dw.sendMessage(msg)
}

// Run starts the worker processes
func (dw *DeviceWorker) Run() {
	dw.Start()
}

// IsHealthy checks if the worker is healthy
func (dw *DeviceWorker) IsHealthy() bool {
	dw.mu.RLock()
	defer dw.mu.RUnlock()
	
	// Check if stopped
	if dw.status == "stopped" {
		return false
	}
	
	// Check if stuck processing
	if dw.status == "processing" && time.Since(dw.lastActivity) > 10*time.Minute {
		return false
	}
	
	// Check device status from DATABASE instead of WhatsApp
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByID(dw.deviceID)
	if err != nil || device == nil {
		return false
	}
	
	// Check if device is online in database
	if device.Status != "online" {
		return false
	}
	
	// Check queue health - if queue is full for too long
	if len(dw.messageQueue) > 900 && time.Since(dw.lastActivity) > 1*time.Minute {
		logrus.Warnf("Worker %s queue is nearly full (%d/1000) and inactive", dw.deviceID, len(dw.messageQueue))
		return false
	}
	
	return true
}

// downloadMedia downloads media from URL or data URL
func downloadMedia(url string) ([]byte, error) {
	// Check if it's a data URL
	if strings.HasPrefix(url, "data:") {
		// Parse data URL
		parts := strings.SplitN(url, ",", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid data URL format")
		}
		
		// Decode base64
		data, err := base64.StdEncoding.DecodeString(parts[1])
		if err != nil {
			return nil, fmt.Errorf("failed to decode base64: %v", err)
		}
		
		return data, nil
	}
	
	// Regular HTTP/HTTPS URL
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad status: %s", resp.Status)
	}
	
	return io.ReadAll(resp.Body)
}
// getRandomDelayBetween returns a random delay between min and max seconds
func getRandomDelayBetween(minDelay, maxDelay int) time.Duration {
	if minDelay <= 0 && maxDelay <= 0 {
		// Default delays if not set
		minDelay = 10
		maxDelay = 30
	}
	
	if minDelay == maxDelay || maxDelay <= minDelay {
		return time.Duration(minDelay) * time.Second
	}
	
	delay := rand.Intn(maxDelay-minDelay) + minDelay
	return time.Duration(delay) * time.Second
}

// RestartWorker restarts a worker with a new client
func (dw *DeviceWorker) RestartWorker() error {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	
	logrus.Infof("Restarting worker for device %s", dw.deviceID)
	
	// Get fresh client from ClientManager
	cm := whatsapp.GetClientManager()
	newClient, err := cm.GetClient(dw.deviceID)
	if err != nil {
		// Try to reconnect the device
		dhm := whatsapp.GetDeviceHealthMonitor(nil)
		if err := dhm.ManualReconnectDevice(dw.deviceID); err != nil {
			return fmt.Errorf("failed to reconnect device: %v", err)
		}
		
		// Try again to get client
		newClient, err = cm.GetClient(dw.deviceID)
		if err != nil {
			return fmt.Errorf("failed to get client after reconnection: %v", err)
		}
	}
	
	// Update client
	dw.client = newClient
	dw.status = "idle"
	dw.lastActivity = time.Now()
	
	logrus.Infof("Worker for device %s restarted successfully", dw.deviceID)
	return nil
}