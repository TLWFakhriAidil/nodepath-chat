package broadcast

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
)

// OptimizedBroadcastManager handles broadcasting for 3,000+ devices
type OptimizedBroadcastManager struct {
	workers       map[string]*DeviceWorker
	workersMutex  sync.RWMutex
	activeWorkers int32
	maxWorkers    int32
	
	// Metrics
	totalProcessed  int64
	totalFailed     int64
	totalPending    int64
	
	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// DeviceWorker represents a worker for a specific device
type DeviceWorker struct {
	DeviceID        string
	Device          *whatsmeow.Client
	Queue           chan *BroadcastMessage
	Status          string
	LastActivity    time.Time
	ProcessedCount  int64
	FailedCount     int64
	MinDelay        int
	MaxDelay        int
	
	// Rate limiting
	messagesSentMinute int
	messagesSentHour   int
	messagesSentDay    int
	lastResetMinute    time.Time
	lastResetHour      time.Time
	lastResetDay       time.Time
	
	// Control
	ctx    context.Context
	cancel context.CancelFunc
	mutex  sync.Mutex
}

// BroadcastMessage represents a message to be broadcast
type BroadcastMessage struct {
	ID           string
	DeviceID     string
	RecipientJID string
	Message      string
	ImageURL     string
	CampaignID   *int
	SequenceID   *string
	GroupID      string
	GroupOrder   int
	RetryCount   int
	CreatedAt    time.Time
	// Delay settings from campaign/sequence
	MinDelay     int
	MaxDelay     int
}

// NewOptimizedBroadcastManager creates a new optimized broadcast manager
func NewOptimizedBroadcastManager() *OptimizedBroadcastManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &OptimizedBroadcastManager{
		workers:    make(map[string]*DeviceWorker),
		maxWorkers: int32(config.MaxConcurrentWorkers),
		ctx:        ctx,
		cancel:     cancel,
	}
	
	// Start health check routine
	go manager.healthCheckRoutine()
	
	// Start metrics collection
	go manager.metricsRoutine()
	
	return manager
}

// CreateOrGetWorker creates a new worker or returns existing one
func (m *OptimizedBroadcastManager) CreateOrGetWorker(deviceID string, device *whatsmeow.Client) (*DeviceWorker, error) {
	m.workersMutex.Lock()
	defer m.workersMutex.Unlock()
	
	// Check if worker already exists
	if worker, exists := m.workers[deviceID]; exists {
		return worker, nil
	}
	
	// Check if we can create more workers
	currentWorkers := atomic.LoadInt32(&m.activeWorkers)
	if currentWorkers >= m.maxWorkers {
		return nil, fmt.Errorf("max workers limit reached: %d", m.maxWorkers)
	}
	
	// Create new worker
	workerCtx, workerCancel := context.WithCancel(m.ctx)
	worker := &DeviceWorker{
		DeviceID:        deviceID,
		Device:          device,
		Queue:           make(chan *BroadcastMessage, config.WorkerQueueSize),
		Status:          "active",
		LastActivity:    time.Now(),
		ctx:             workerCtx,
		cancel:          workerCancel,
		lastResetMinute: time.Now(),
		lastResetHour:   time.Now(),
		lastResetDay:    time.Now(),
	}
	
	// Start worker routine
	m.wg.Add(1)
	go m.workerRoutine(worker)
	
	// Store worker
	m.workers[deviceID] = worker
	atomic.AddInt32(&m.activeWorkers, 1)
	
	logrus.Infof("Created worker for device %s (total workers: %d)", deviceID, currentWorkers+1)
	return worker, nil
}

// workerRoutine processes messages for a specific device
func (m *OptimizedBroadcastManager) workerRoutine(worker *DeviceWorker) {
	defer m.wg.Done()
	defer func() {
		atomic.AddInt32(&m.activeWorkers, -1)
		worker.Status = "stopped"
		logrus.Infof("Worker for device %s stopped", worker.DeviceID)
	}()
	
	idleTimer := time.NewTimer(time.Duration(config.WorkerIdleTimeoutMin) * time.Minute)
	defer idleTimer.Stop()
	
	for {
		select {
		case <-worker.ctx.Done():
			return
			
		case msg := <-worker.Queue:
			// Reset idle timer
			idleTimer.Reset(time.Duration(config.WorkerIdleTimeoutMin) * time.Minute)
			
			// Check rate limits
			if !worker.checkRateLimits() {
				// Re-queue message
				select {
				case worker.Queue <- msg:
				default:
					logrus.Warnf("Failed to re-queue message for device %s", worker.DeviceID)
				}
				time.Sleep(time.Minute) // Wait before retrying
				continue
			}
			
			// Process message with proper delays
			m.processMessageWithDelay(worker, msg)
			
		case <-idleTimer.C:
			// Worker has been idle for too long
			logrus.Infof("Worker for device %s idle timeout", worker.DeviceID)
			return
		}
	}
}

// processMessageWithDelay sends a message with proper delay handling
func (m *OptimizedBroadcastManager) processMessageWithDelay(worker *DeviceWorker, msg *BroadcastMessage) {
	worker.mutex.Lock()
	worker.LastActivity = time.Now()
	worker.mutex.Unlock()
	
	// Update metrics
	atomic.AddInt64(&m.totalPending, -1)
	
	// Prepare recipient JID
	recipientJID, err := whatsmeow.ParseJID(msg.RecipientJID)
	if err != nil {
		logrus.Errorf("Invalid JID %s: %v", msg.RecipientJID, err)
		m.handleFailedMessage(worker, msg, err)
		return
	}
	
	// Determine if this is a two-part message (image + text)
	hasImage := msg.ImageURL != ""
	hasText := msg.Message != ""
	
	var sendErr error
	
	if hasImage && hasText {
		// TWO-PART MESSAGE: Send image first, then text
		logrus.Debugf("Sending two-part message to %s (image + text)", recipientJID.String())
		
		// 1. Send image without caption
		sendErr = m.sendImageMessage(worker.Device, recipientJID, msg.ImageURL, "")
		if sendErr != nil {
			m.handleFailedMessage(worker, msg, sendErr)
			return
		}
		
		// 2. Wait 3 seconds between image and text
		time.Sleep(3 * time.Second)
		
		// 3. Send text message
		sendErr = m.sendTextMessage(worker.Device, recipientJID, msg.Message)
		if sendErr != nil {
			m.handleFailedMessage(worker, msg, sendErr)
			return
		}
	} else if hasImage {
		// Image only
		sendErr = m.sendImageMessage(worker.Device, recipientJID, msg.ImageURL, "")
	} else if hasText {
		// Text only
		sendErr = m.sendTextMessage(worker.Device, recipientJID, msg.Message)
	} else {
		logrus.Warnf("Message has no content: %s", msg.ID)
		return
	}
	
	if sendErr != nil {
		m.handleFailedMessage(worker, msg, sendErr)
	} else {
		m.handleSuccessMessage(worker, msg)
	}
	
	// Apply random delay between different leads
	// Use message-specific delays if provided, otherwise use worker defaults
	minDelay := msg.MinDelay
	maxDelay := msg.MaxDelay
	if minDelay == 0 || maxDelay == 0 {
		minDelay = config.DefaultMinDelaySeconds
		maxDelay = config.DefaultMaxDelaySeconds
	}
	
	// Calculate random delay
	delay := minDelay
	if maxDelay > minDelay {
		delay = minDelay + rand.Intn(maxDelay-minDelay+1)
	}
	
	logrus.Debugf("Waiting %d seconds before next message (min: %d, max: %d)", delay, minDelay, maxDelay)
	time.Sleep(time.Duration(delay) * time.Second)
}

// sendTextMessage sends a text message
func (m *OptimizedBroadcastManager) sendTextMessage(device *whatsmeow.Client, recipient whatsmeow.JID, message string) error {
	// TODO: Implement actual WhatsApp sending
	// For now, just log
	logrus.Infof("Sending text to %s: %s", recipient.String(), message)
	
	// Simulate sending
	time.Sleep(100 * time.Millisecond)
	
	return nil
}

// sendImageMessage sends an image message
func (m *OptimizedBroadcastManager) sendImageMessage(device *whatsmeow.Client, recipient whatsmeow.JID, imageURL, caption string) error {
	// TODO: Implement actual WhatsApp image sending
	// For now, just log
	if caption != "" {
		logrus.Infof("Sending image with caption to %s: %s (caption: %s)", recipient.String(), imageURL, caption)
	} else {
		logrus.Infof("Sending image to %s: %s", recipient.String(), imageURL)
	}
	
	// Simulate sending
	time.Sleep(200 * time.Millisecond)
	
	return nil
}

// handleSuccessMessage handles successful message send
func (m *OptimizedBroadcastManager) handleSuccessMessage(worker *DeviceWorker, msg *BroadcastMessage) {
	atomic.AddInt64(&worker.ProcessedCount, 1)
	atomic.AddInt64(&m.totalProcessed, 1)
	
	// Update rate limit counters
	worker.mutex.Lock()
	worker.messagesSentMinute++
	worker.messagesSentHour++
	worker.messagesSentDay++
	worker.mutex.Unlock()
	
	// Update broadcast message status in database
	broadcastRepo := repository.GetBroadcastRepository()
	_ = broadcastRepo.UpdateBroadcastStatus(msg.ID, "sent", "")
	
	// Log based on message type
	if msg.CampaignID != nil {
		logrus.Infof("Campaign message sent to %s via device %s", msg.RecipientJID, worker.DeviceID)
	} else if msg.SequenceID != nil {
		logrus.Infof("Sequence message sent to %s via device %s", msg.RecipientJID, worker.DeviceID)
	} else {
		logrus.Debugf("Message sent to %s via device %s", msg.RecipientJID, worker.DeviceID)
	}
}

// handleFailedMessage handles failed message send
func (m *OptimizedBroadcastManager) handleFailedMessage(worker *DeviceWorker, msg *BroadcastMessage, err error) {
	atomic.AddInt64(&worker.FailedCount, 1)
	atomic.AddInt64(&m.totalFailed, 1)
	
	msg.RetryCount++
	
	// Retry logic
	if msg.RetryCount < config.RetryAttempts {
		// Re-queue for retry
		go func() {
			time.Sleep(time.Duration(config.RetryDelaySeconds) * time.Second)
			select {
			case worker.Queue <- msg:
				atomic.AddInt64(&m.totalPending, 1)
			default:
				logrus.Warnf("Failed to re-queue message for retry")
			}
		}()
	} else {
		// Max retries reached, mark as failed
		broadcastRepo := repository.GetBroadcastRepository()
		_ = broadcastRepo.UpdateBroadcastStatus(msg.ID, "failed", err.Error())
		
		logrus.Errorf("Message to %s failed after %d retries: %v", msg.RecipientJID, msg.RetryCount, err)
	}
}

// checkRateLimits checks if worker can send more messages
func (worker *DeviceWorker) checkRateLimits() bool {
	worker.mutex.Lock()
	defer worker.mutex.Unlock()
	
	now := time.Now()
	
	// Reset counters if needed
	if now.Sub(worker.lastResetMinute) > time.Minute {
		worker.messagesSentMinute = 0
		worker.lastResetMinute = now
	}
	
	if now.Sub(worker.lastResetHour) > time.Hour {
		worker.messagesSentHour = 0
		worker.lastResetHour = now
	}
	
	if now.Sub(worker.lastResetDay) > 24*time.Hour {
		worker.messagesSentDay = 0
		worker.lastResetDay = now
	}
	
	// Check limits
	if worker.messagesSentMinute >= config.MessagesPerMinute {
		logrus.Warnf("Device %s hit minute rate limit", worker.DeviceID)
		return false
	}
	
	if worker.messagesSentHour >= config.MessagesPerHour {
		logrus.Warnf("Device %s hit hour rate limit", worker.DeviceID)
		return false
	}
	
	if worker.messagesSentDay >= config.MessagesPerDay {
		logrus.Warnf("Device %s hit day rate limit", worker.DeviceID)
		return false
	}
	
	return true
}

// QueueMessage queues a message for broadcasting
func (m *OptimizedBroadcastManager) QueueMessage(deviceID string, msg *BroadcastMessage) error {
	m.workersMutex.RLock()
	worker, exists := m.workers[deviceID]
	m.workersMutex.RUnlock()
	
	if !exists {
		return fmt.Errorf("worker not found for device %s", deviceID)
	}
	
	select {
	case worker.Queue <- msg:
		atomic.AddInt64(&m.totalPending, 1)
		
		// Log queue action
		if msg.CampaignID != nil {
			logrus.Debugf("Queued campaign message for device %s (queue size: %d)", deviceID, len(worker.Queue))
		} else if msg.SequenceID != nil {
			logrus.Debugf("Queued sequence message for device %s (queue size: %d)", deviceID, len(worker.Queue))
		}
		
		return nil
	default:
		return fmt.Errorf("queue full for device %s", deviceID)
	}
}

// healthCheckRoutine monitors worker health
func (m *OptimizedBroadcastManager) healthCheckRoutine() {
	ticker := time.NewTicker(time.Duration(config.WorkerHealthCheckSec) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.performHealthCheck()
		}
	}
}

// performHealthCheck checks all workers
func (m *OptimizedBroadcastManager) performHealthCheck() {
	m.workersMutex.Lock()
	defer m.workersMutex.Unlock()
	
	now := time.Now()
	for deviceID, worker := range m.workers {
		worker.mutex.Lock()
		lastActivity := worker.LastActivity
		status := worker.Status
		queueSize := len(worker.Queue)
		worker.mutex.Unlock()
		
		// Check if worker is stuck
		if status == "active" && now.Sub(lastActivity) > time.Duration(config.WorkerIdleTimeoutMin)*time.Minute && queueSize > 0 {
			logrus.Warnf("Worker for device %s appears stuck, restarting", deviceID)
			// Could implement restart logic here
		}
		
		// Update worker status in database
		workerRepo := repository.GetWorkerRepository()
		_ = workerRepo.UpdateWorkerStatus(deviceID, status, queueSize, worker.ProcessedCount, worker.FailedCount)
	}
}

// metricsRoutine collects and reports metrics
func (m *OptimizedBroadcastManager) metricsRoutine() {
	ticker := time.NewTicker(time.Duration(config.MetricsIntervalSec) * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.reportMetrics()
		}
	}
}

// reportMetrics logs current metrics
func (m *OptimizedBroadcastManager) reportMetrics() {
	activeWorkers := atomic.LoadInt32(&m.activeWorkers)
	totalProcessed := atomic.LoadInt64(&m.totalProcessed)
	totalFailed := atomic.LoadInt64(&m.totalFailed)
	totalPending := atomic.LoadInt64(&m.totalPending)
	
	logrus.WithFields(logrus.Fields{
		"active_workers":  activeWorkers,
		"total_processed": totalProcessed,
		"total_failed":    totalFailed,
		"total_pending":   totalPending,
		"max_workers":     m.maxWorkers,
	}).Info("Broadcast manager metrics")
}

// GetWorkerStatus returns status of all workers
func (m *OptimizedBroadcastManager) GetWorkerStatus() map[string]interface{} {
	m.workersMutex.RLock()
	defer m.workersMutex.RUnlock()
	
	workers := make([]map[string]interface{}, 0, len(m.workers))
	for _, worker := range m.workers {
		worker.mutex.Lock()
		workerInfo := map[string]interface{}{
			"device_id":      worker.DeviceID,
			"status":         worker.Status,
			"queue_size":     len(worker.Queue),
			"processed":      worker.ProcessedCount,
			"failed":         worker.FailedCount,
			"last_activity":  worker.LastActivity,
			"rate_limits": map[string]int{
				"minute": worker.messagesSentMinute,
				"hour":   worker.messagesSentHour,
				"day":    worker.messagesSentDay,
			},
		}
		worker.mutex.Unlock()
		workers = append(workers, workerInfo)
	}
	
	return map[string]interface{}{
		"active_workers":  atomic.LoadInt32(&m.activeWorkers),
		"max_workers":     m.maxWorkers,
		"total_processed": atomic.LoadInt64(&m.totalProcessed),
		"total_failed":    atomic.LoadInt64(&m.totalFailed),
		"total_pending":   atomic.LoadInt64(&m.totalPending),
		"workers":         workers,
	}
}

// Shutdown gracefully shuts down the broadcast manager
func (m *OptimizedBroadcastManager) Shutdown() {
	logrus.Info("Shutting down broadcast manager...")
	
	// Cancel context
	m.cancel()
	
	// Stop all workers
	m.workersMutex.Lock()
	for _, worker := range m.workers {
		worker.cancel()
	}
	m.workersMutex.Unlock()
	
	// Wait for all workers to finish
	m.wg.Wait()
	
	logrus.Info("Broadcast manager shutdown complete")
}

// Global instance
var globalBroadcastManager *OptimizedBroadcastManager
var once sync.Once

// GetBroadcastManager returns the global broadcast manager instance
func GetBroadcastManager() *OptimizedBroadcastManager {
	once.Do(func() {
		globalBroadcastManager = NewOptimizedBroadcastManager()
	})
	return globalBroadcastManager
}
