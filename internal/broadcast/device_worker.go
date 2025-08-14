package broadcast

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"

	"github.com/sirupsen/logrus"
)

// MessageSender interface to avoid circular imports
type MessageSender interface {
	SendMessage(phoneNumber, message string) error
	SendMediaMessage(phoneNumber, caption, mediaURL, mediaType string) error
}

// DeviceWorker handles message processing for a specific device with high performance
type DeviceWorker struct {
	deviceID       string
	messageQueue   chan models.BroadcastMessage
	ctx            context.Context
	cancel         context.CancelFunc
	processedCount int64
	failedCount    int64
	lastActivity   time.Time
	status         string
	mu             sync.RWMutex
	running        bool
	messageSender  MessageSender
	minDelay       time.Duration
	maxDelay       time.Duration
	repo           *repository.BroadcastRepository
}

// NewDeviceWorker creates a new device worker with optimized settings
func NewDeviceWorker(deviceID string, parentCtx context.Context, messageSender MessageSender, repo *repository.BroadcastRepository) *DeviceWorker {
	ctx, cancel := context.WithCancel(parentCtx)
	
	return &DeviceWorker{
		deviceID:      deviceID,
		messageQueue:  make(chan models.BroadcastMessage, 1000), // Large buffer for high throughput
		ctx:           ctx,
		cancel:        cancel,
		lastActivity:  time.Now(),
		status:        "idle",
		running:       false,
		messageSender: messageSender,
		minDelay:      5 * time.Second,  // Default delays
		maxDelay:      15 * time.Second,
		repo:          repo,
	}
}

// Start begins the worker's message processing loop
func (dw *DeviceWorker) Start() error {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	
	if dw.running {
		return fmt.Errorf("worker for device %s is already running", dw.deviceID)
	}
	
	dw.running = true
	dw.status = "starting"
	
	// Initialize WhatsApp service connection for this device
	// Note: This would need to be implemented based on your WhatsApp service architecture
	// dw.whatsappSvc = whatsapp.GetServiceForDevice(dw.deviceID)
	
	logrus.Debugf("Starting worker for device %s", dw.deviceID)
	
	// Start the main processing goroutine
	go dw.processMessages()
	
	dw.status = "running"
	return nil
}

// Stop gracefully shuts down the worker
func (dw *DeviceWorker) Stop() {
	dw.mu.Lock()
	defer dw.mu.Unlock()
	
	if !dw.running {
		return
	}
	
	logrus.Debugf("Stopping worker for device %s", dw.deviceID)
	dw.running = false
	dw.status = "stopping"
	dw.cancel()
	
	// Close the message queue
	close(dw.messageQueue)
	
	dw.status = "stopped"
}

// QueueMessage adds a message to the worker's queue with timeout
func (dw *DeviceWorker) QueueMessage(msg models.BroadcastMessage) error {
	select {
	case dw.messageQueue <- msg:
		// Update message status to queued
		dw.repo.UpdateMessageStatus(msg.ID, "queued", "")
		return nil
	case <-time.After(10 * time.Second): // Increased timeout for high load
		return fmt.Errorf("queue full for device %s, message timeout", dw.deviceID)
	case <-dw.ctx.Done():
		return fmt.Errorf("worker for device %s is shutting down", dw.deviceID)
	}
}

// processMessages is the main message processing loop
func (dw *DeviceWorker) processMessages() {
	defer func() {
		if r := recover(); r != nil {
			logrus.Errorf("Worker for device %s panicked: %v", dw.deviceID, r)
			dw.mu.Lock()
			dw.status = "crashed"
			dw.mu.Unlock()
		}
	}()
	
	logrus.Debugf("Worker for device %s started processing messages", dw.deviceID)
	
	for {
		select {
		case <-dw.ctx.Done():
			logrus.Debugf("Worker for device %s received shutdown signal", dw.deviceID)
			return
		case msg, ok := <-dw.messageQueue:
			if !ok {
				logrus.Debugf("Message queue closed for device %s", dw.deviceID)
				return
			}
			
			dw.processMessage(msg)
		}
	}
}

// processMessage processes a single message with anti-spam delays
func (dw *DeviceWorker) processMessage(msg models.BroadcastMessage) {
	start := time.Now()
	// Update status to processing
	dw.repo.UpdateMessageStatus(msg.ID, "processing", "")
	
	dw.mu.Lock()
	dw.status = "processing"
	dw.lastActivity = time.Now()
	dw.mu.Unlock()
	
	logrus.Debugf("Processing message %s for %s on device %s", msg.ID, msg.RecipientPhone, dw.deviceID)
	
	// Apply anti-spam delay based on message settings or device defaults
	minDelay := dw.minDelay
	maxDelay := dw.maxDelay
	
	if msg.MinDelay > 0 {
		minDelay = time.Duration(msg.MinDelay) * time.Second
	}
	if msg.MaxDelay > 0 {
		maxDelay = time.Duration(msg.MaxDelay) * time.Second
	}
	
	// Random delay between min and max for human-like behavior
	delay := minDelay
	if maxDelay > minDelay {
		delayRange := maxDelay - minDelay
		randomDelay := time.Duration(rand.Int63n(int64(delayRange)))
		delay = minDelay + randomDelay
	}
	
	// Apply the delay before sending
	if delay > 0 {
		logrus.Debugf("Applying %v delay before sending message %s", delay, msg.ID)
		select {
		case <-time.After(delay):
			// Continue with sending
		case <-dw.ctx.Done():
			// Worker is shutting down
			dw.repo.UpdateMessageStatus(msg.ID, "failed", "Worker shutdown during delay")
			return
		}
	}
	
	// Send the message
	err := dw.sendMessage(msg)
	if err != nil {
		logrus.Errorf("Failed to send message %s: %v", msg.ID, err)
		dw.repo.UpdateMessageStatus(msg.ID, "failed", err.Error())
		atomic.AddInt64(&dw.failedCount, 1)
	} else {
		logrus.Debugf("Successfully sent message %s to %s", msg.ID, msg.RecipientPhone)
		dw.repo.UpdateMessageStatus(msg.ID, "sent", "")
		atomic.AddInt64(&dw.processedCount, 1)
	}
	
	dw.mu.Lock()
	dw.status = "idle"
	dw.lastActivity = time.Now()
	dw.mu.Unlock()
	
	processingTime := time.Since(start)
	logrus.Debugf("Message %s processed in %v", msg.ID, processingTime)
}

// sendMessage sends the actual WhatsApp message
func (dw *DeviceWorker) sendMessage(msg models.BroadcastMessage) error {
	if dw.messageSender == nil {
		return fmt.Errorf("message sender not configured for device %s", dw.deviceID)
	}
	
	logrus.Debugf("Sending %s message to %s: %s", msg.Type, msg.RecipientPhone, msg.Content)
	
	// Send message based on type
	switch msg.Type {
	case "text":
		return dw.messageSender.SendMessage(msg.RecipientPhone, msg.Content)
	case "media":
		return dw.messageSender.SendMediaMessage(msg.RecipientPhone, msg.Content, msg.MediaURL, msg.Type)
	default:
		return dw.messageSender.SendMessage(msg.RecipientPhone, msg.Content)
	}
}

// IsHealthy checks if the worker is healthy and responsive
func (dw *DeviceWorker) IsHealthy() bool {
	dw.mu.RLock()
	defer dw.mu.RUnlock()
	
	if !dw.running {
		return false
	}
	
	// Check if worker has been inactive for too long
	if time.Since(dw.lastActivity) > 5*time.Minute {
		return false
	}
	
	// Check if status indicates a problem
	if dw.status == "crashed" || dw.status == "error" {
		return false
	}
	
	return true
}

// GetStatus returns the current status of the worker
func (dw *DeviceWorker) GetStatus() models.WorkerStatus {
	dw.mu.RLock()
	defer dw.mu.RUnlock()
	
	return models.WorkerStatus{
		DeviceID:       dw.deviceID,
		Status:         dw.status,
		QueueSize:      len(dw.messageQueue),
		ProcessedCount: atomic.LoadInt64(&dw.processedCount),
		FailedCount:    atomic.LoadInt64(&dw.failedCount),
		LastActivity:   dw.lastActivity,
	}
}

// GetQueueSize returns the current queue size
func (dw *DeviceWorker) GetQueueSize() int {
	return len(dw.messageQueue)
}

// GetProcessedCount returns the total number of processed messages
func (dw *DeviceWorker) GetProcessedCount() int64 {
	return atomic.LoadInt64(&dw.processedCount)
}

// GetFailedCount returns the total number of failed messages
func (dw *DeviceWorker) GetFailedCount() int64 {
	return atomic.LoadInt64(&dw.failedCount)
}