package services

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"sync"
	"time"

	"nodepath-chat/internal/broadcast"
	"nodepath-chat/internal/models"
	"nodepath-chat/internal/repository"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// BroadcastService is the main service that orchestrates the broadcast system
type BroadcastService struct {
	mu             sync.RWMutex
	manager        *broadcast.BroadcastManager
	queueCleaner   *QueueCleaner
	metrics        *MetricsService
	repo           *repository.BroadcastRepository
	running        bool
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewBroadcastService creates a new broadcast service
func NewBroadcastService(db *sql.DB, messageSender broadcast.MessageSender) *BroadcastService {
	ctx, cancel := context.WithCancel(context.Background())
	
	repo := repository.NewBroadcastRepository(db)
	manager := broadcast.NewBroadcastManager(ctx, repo)
	if messageSender != nil {
		manager.SetMessageSender(messageSender)
	}
	
	return &BroadcastService{
		manager:      manager,
		queueCleaner: NewQueueCleaner(),
		metrics:      NewMetricsService(),
		repo:         repo,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start initializes and starts the broadcast service
func (bs *BroadcastService) Start() error {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	
	if bs.running {
		return fmt.Errorf("broadcast service is already running")
	}
	
	logrus.Info("Starting broadcast service")
	
	// Start the broadcast manager
	err := bs.manager.Start()
	if err != nil {
		return fmt.Errorf("failed to start broadcast manager: %v", err)
	}
	
	// Start the queue cleaner
	bs.queueCleaner.Start()
	
	// Start periodic metrics reporting
	bs.metrics.StartPeriodicReporting(10 * time.Minute)
	
	bs.running = true
	logrus.Info("Broadcast service started successfully")
	
	return nil
}

// Stop gracefully shuts down the broadcast service
func (bs *BroadcastService) Stop() {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	
	if !bs.running {
		return
	}
	
	logrus.Info("Stopping broadcast service")
	
	// Stop all components
	bs.manager.Stop()
	bs.queueCleaner.Stop()
	bs.cancel()
	
	bs.running = false
	logrus.Info("Broadcast service stopped")
}

// QueueCampaignMessage queues a campaign message for broadcast
func (bs *BroadcastService) QueueCampaignMessage(userID, deviceID, campaignID, recipientPhone, messageType, content, mediaURL string, minDelay, maxDelay int) (string, error) {
	// Convert campaignID string to int
	campaignIDInt, err := strconv.Atoi(campaignID)
	if err != nil {
		return "", fmt.Errorf("invalid campaign ID: %v", err)
	}
	
	message := models.BroadcastMessage{
		ID:             uuid.New().String(),
		UserID:         userID,
		DeviceID:       deviceID,
		CampaignID:     &campaignIDInt,
		RecipientPhone: recipientPhone,
		Type:           messageType,
		Content:        content,
		MediaURL:       mediaURL,
		Status:         "pending",
		ScheduledAt:    time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		MinDelay:       minDelay,
		MaxDelay:       maxDelay,
	}
	
	err = bs.repo.QueueMessage(message)
	if err != nil {
		return "", fmt.Errorf("failed to queue campaign message: %v", err)
	}
	
	// Record metrics
	bs.metrics.RecordMessageQueued(deviceID)
	
	logrus.Debugf("Queued campaign message %s for %s on device %s", message.ID, recipientPhone, deviceID)
	return message.ID, nil
}

// QueueSequenceMessage queues a sequence message for broadcast
func (bs *BroadcastService) QueueSequenceMessage(userID, deviceID, sequenceID, sequenceStepID, recipientPhone, messageType, content, mediaURL string, minDelay, maxDelay int, stepDelay time.Duration) (string, error) {
	message := models.BroadcastMessage{
		ID:             uuid.New().String(),
		UserID:         userID,
		DeviceID:       deviceID,
		SequenceID:     &sequenceID,
		SequenceStepID: &sequenceStepID,
		RecipientPhone: recipientPhone,
		Type:           messageType,
		Content:        content,
		MediaURL:       mediaURL,
		Status:         "pending",
		ScheduledAt:    time.Now().Add(stepDelay), // Apply sequence step delay
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		MinDelay:       minDelay,
		MaxDelay:       maxDelay,
	}
	
	err := bs.repo.QueueMessage(message)
	if err != nil {
		return "", fmt.Errorf("failed to queue sequence message: %v", err)
	}
	
	// Record metrics
	bs.metrics.RecordMessageQueued(deviceID)
	
	logrus.Debugf("Queued sequence message %s for %s on device %s (step %s)", message.ID, recipientPhone, deviceID, sequenceStepID)
	return message.ID, nil
}

// QueueBulkMessages queues multiple messages in a single transaction for better performance
func (bs *BroadcastService) QueueBulkMessages(messages []models.BroadcastMessage) ([]string, error) {
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages to queue")
	}
	
	// Assign IDs and timestamps
	messageIDs := make([]string, len(messages))
	now := time.Now()
	
	for i := range messages {
		messages[i].ID = uuid.New().String()
		messages[i].Status = "pending"
		messages[i].CreatedAt = now
		messages[i].UpdatedAt = now
		if messages[i].ScheduledAt.IsZero() {
			messages[i].ScheduledAt = now
		}
		messageIDs[i] = messages[i].ID
	}
	
	// Queue all messages in a single transaction
	err := bs.repo.QueueBulkMessages(messages)
	if err != nil {
		return nil, fmt.Errorf("failed to queue bulk messages: %v", err)
	}
	
	// Record metrics for each device
	deviceCounts := make(map[string]int)
	for _, msg := range messages {
		deviceCounts[msg.DeviceID]++
	}
	
	for deviceID, count := range deviceCounts {
		for i := 0; i < count; i++ {
			bs.metrics.RecordMessageQueued(deviceID)
		}
	}
	
	logrus.Infof("Queued %d bulk messages across %d devices", len(messages), len(deviceCounts))
	return messageIDs, nil
}

// GetMessageStatus returns the current status of a message
func (bs *BroadcastService) GetMessageStatus(messageID string) (string, error) {
	status, err := bs.repo.GetMessageStatus(messageID)
	if err != nil {
		return "", fmt.Errorf("failed to get message status: %v", err)
	}
	return status, nil
}

// GetQueueStats returns current queue statistics
func (bs *BroadcastService) GetQueueStats() (map[string]int, error) {
	return bs.queueCleaner.GetQueueStats()
}

// GetMetrics returns current system metrics
func (bs *BroadcastService) GetMetrics() *BroadcastMetrics {
	return bs.metrics.GetMetrics()
}

// GetDeviceMetrics returns metrics for a specific device
func (bs *BroadcastService) GetDeviceMetrics(deviceID string) *DeviceMetrics {
	return bs.metrics.GetDeviceMetrics(deviceID)
}

// GetHealthStatus returns the overall health status
func (bs *BroadcastService) GetHealthStatus() string {
	return bs.metrics.GetHealthStatus()
}

// ForceCleanup manually triggers a queue cleanup
func (bs *BroadcastService) ForceCleanup() {
	bs.queueCleaner.ForceCleanup()
}

// ResetMetrics resets all performance metrics
func (bs *BroadcastService) ResetMetrics() {
	bs.metrics.ResetMetrics()
}

// GetWorkerStatus returns the status of all workers
func (bs *BroadcastService) GetWorkerStatus() []models.WorkerStatus {
	return bs.manager.GetWorkerStatus()
}

// GetActiveDevices returns a list of devices with active workers
func (bs *BroadcastService) GetActiveDevices() []string {
	return bs.manager.GetActiveDevices()
}

// PauseDevice pauses message processing for a specific device
func (bs *BroadcastService) PauseDevice(deviceID string) error {
	return bs.manager.PauseDevice(deviceID)
}

// ResumeDevice resumes message processing for a specific device
func (bs *BroadcastService) ResumeDevice(deviceID string) error {
	return bs.manager.ResumeDevice(deviceID)
}

// RestartDevice restarts the worker for a specific device
func (bs *BroadcastService) RestartDevice(deviceID string) error {
	return bs.manager.RestartDevice(deviceID)
}

// UpdateDeviceSettings updates processing settings for a device
func (bs *BroadcastService) UpdateDeviceSettings(deviceID string, minDelay, maxDelay time.Duration) error {
	return bs.manager.UpdateDeviceSettings(deviceID, minDelay, maxDelay)
}

// GetPendingMessageCount returns the number of pending messages for a device
func (bs *BroadcastService) GetPendingMessageCount(deviceID string) (int, error) {
	count, err := bs.repo.GetPendingMessageCount(deviceID)
	if err != nil {
		return 0, fmt.Errorf("failed to get pending message count: %v", err)
	}
	return count, nil
}

// GetTotalPendingMessageCount returns the total number of pending messages across all devices
func (bs *BroadcastService) GetTotalPendingMessageCount() (int, error) {
	count, err := bs.repo.GetTotalPendingMessageCount()
	if err != nil {
		return 0, fmt.Errorf("failed to get total pending message count: %v", err)
	}
	return count, nil
}

// CancelMessage cancels a pending message
func (bs *BroadcastService) CancelMessage(messageID string) error {
	err := bs.repo.UpdateMessageStatus(messageID, "cancelled", "Cancelled by user")
	if err != nil {
		return fmt.Errorf("failed to cancel message: %v", err)
	}
	
	logrus.Infof("Message %s cancelled", messageID)
	return nil
}

// CancelCampaignMessages cancels all pending messages for a campaign
func (bs *BroadcastService) CancelCampaignMessages(campaignID string) (int, error) {
	// Convert campaignID string to int
	campaignIDInt, err := strconv.Atoi(campaignID)
	if err != nil {
		return 0, fmt.Errorf("invalid campaign ID: %v", err)
	}
	
	count, err := bs.repo.CancelCampaignMessages(campaignIDInt)
	if err != nil {
		return 0, fmt.Errorf("failed to cancel campaign messages: %v", err)
	}
	
	logrus.Infof("Cancelled %d messages for campaign %s", count, campaignID)
	return count, nil
}

// CancelSequenceMessages cancels all pending messages for a sequence
func (bs *BroadcastService) CancelSequenceMessages(sequenceID string) (int, error) {
	count, err := bs.repo.CancelSequenceMessages(sequenceID)
	if err != nil {
		return 0, fmt.Errorf("failed to cancel sequence messages: %v", err)
	}
	
	logrus.Infof("Cancelled %d messages for sequence %s", count, sequenceID)
	return count, nil
}

// IsRunning returns whether the broadcast service is currently running
func (bs *BroadcastService) IsRunning() bool {
	bs.mu.RLock()
	defer bs.mu.RUnlock()
	return bs.running
}