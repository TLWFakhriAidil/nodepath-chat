package broadcast

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

const (
	// Redis queue keys - using unique names to avoid conflicts
	ultraCampaignQueuePrefix  = "ultra:queue:campaign:"
	ultraSequenceQueuePrefix  = "ultra:queue:sequence:"
	ultraDeadLetterPrefix     = "ultra:queue:dead:"
	ultraMetricsPrefix        = "ultra:metrics:"
	ultraRateLimitPrefix      = "ultra:ratelimit:"
	ultraWorkerStatusKey      = "ultra:workers"
	ultraWorkerLockPrefix     = "ultra:lock:"
	
	// Performance settings for 3000 devices
	maxConcurrentWorkers = 3000
	workerBatchSize      = 100
	queueCheckInterval   = 100 * time.Millisecond
	metricsInterval      = 5 * time.Second
	healthCheckInterval  = 30 * time.Second
	lockTTL              = 5 * time.Minute
)

// UltraScaleRedisManager is optimized for 3000+ devices
type UltraScaleRedisManager struct {
	redisClient   *redis.Client
	workers       map[string]*DeviceWorker
	workersMutex  sync.RWMutex
	activeWorkers int32
	
	// Performance optimization
	workerPools   map[int]*sync.Pool // Worker pools by priority
	metricsBatch  map[string]int64
	metricsMutex  sync.Mutex
	
	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// UltraRedisMessage represents a message in Redis queue (unique name to avoid conflict)
type UltraRedisMessage struct {
	Message   domainBroadcast.BroadcastMessage `json:"message"`
	Priority  int                              `json:"priority"`
	Timestamp time.Time                        `json:"timestamp"`
	Retries   int                              `json:"retries"`
}

// NewUltraScaleRedisManager creates a manager optimized for 3000+ devices
func NewUltraScaleRedisManager() *UltraScaleRedisManager {
	// Initialize Redis client
	redisURL := config.RedisURL
	if redisURL == "" {
		redisHost := config.RedisHost
		redisPort := config.RedisPort
		redisPassword := config.RedisPassword
		
		if redisHost == "" {
			redisHost = "localhost"
		}
		if redisPort == "" {
			redisPort = "6379"
		}
		
		redisURL = fmt.Sprintf("redis://:%s@%s:%s/0", redisPassword, redisHost, redisPort)
	}
	
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logrus.Fatalf("Failed to parse Redis URL: %v", err)
	}
	
	// Optimize Redis client for high performance
	opt.PoolSize = 100           // Increase pool size for 3000 devices
	opt.MinIdleConns = 20        // Keep connections ready
	opt.MaxRetries = 3           // Retry on failure
	opt.DialTimeout = 5 * time.Second
	opt.ReadTimeout = 3 * time.Second
	opt.WriteTimeout = 3 * time.Second
	
	redisClient := redis.NewClient(opt)
	
	// Test connection with retry logic
	ctx := context.Background()
	var connectErr error
	for retries := 0; retries < 5; retries++ {
		if err := redisClient.Ping(ctx).Err(); err != nil {
			connectErr = err
			logrus.Warnf("Failed to connect to Redis (attempt %d/5): %v", retries+1, err)
			time.Sleep(time.Duration(retries+1) * time.Second) // Exponential backoff
			continue
		}
		connectErr = nil
		break
	}
	
	if connectErr != nil {
		logrus.Fatalf("Failed to connect to Redis after 5 attempts: %v", connectErr)
	}
	
	logrus.Info("Successfully connected to Redis (Ultra Scale Mode)")
	
	ctx, cancel := context.WithCancel(context.Background())
	manager := &UltraScaleRedisManager{
		redisClient:  redisClient,
		workers:      make(map[string]*DeviceWorker),
		workerPools:  make(map[int]*sync.Pool),
		metricsBatch: make(map[string]int64),
		ctx:          ctx,
		cancel:       cancel,
	}
	
	// Initialize worker pools for different priorities
	for priority := 1; priority <= 10; priority++ {
		manager.workerPools[priority] = &sync.Pool{
			New: func() interface{} {
				return &DeviceWorker{}
			},
		}
	}
	
	// Start manager routines
	manager.wg.Add(5)
	go manager.processQueues()
	go manager.monitorWorkers()
	go manager.cleanupDeadLetters()
	go manager.flushMetrics()
	go manager.cleanupOldMessages() // Add cleanup routine
	
	logrus.Infof("Ultra Scale Redis Manager started - Ready for %d devices", maxConcurrentWorkers)
	return manager
}

// SendMessage implements the interface - optimized for high throughput
func (um *UltraScaleRedisManager) SendMessage(msg domainBroadcast.BroadcastMessage) error {
	// Use device-specific queue for better distribution
	var queueKey string
	if msg.CampaignID != nil {
		queueKey = fmt.Sprintf("%s%s", ultraCampaignQueuePrefix, msg.DeviceID)
	} else {
		queueKey = fmt.Sprintf("%s%s", ultraSequenceQueuePrefix, msg.DeviceID)
	}
	
	// Create Redis message
	redisMsg := UltraRedisMessage{
		Message:   msg,
		Priority:  getPriority(msg),
		Timestamp: time.Now(),
		Retries:   0,
	}
	
	data, err := json.Marshal(redisMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}
	
	// Push to device-specific queue
	if err := um.redisClient.LPush(um.ctx, queueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to push to queue: %v", err)
	}
	
	// Ensure worker exists for this device
	um.ensureWorker(msg.DeviceID)
	
	// Update metrics asynchronously
	go um.incrementMetricBatch("messages_queued")
	
	return nil
}

// ensureWorker creates a worker if it doesn't exist
func (um *UltraScaleRedisManager) ensureWorker(deviceID string) {
	um.workersMutex.RLock()
	_, exists := um.workers[deviceID]
	um.workersMutex.RUnlock()
	
	if !exists {
		// First check if device exists before attempting to create worker
		userRepo := repository.GetUserRepository()
		device, err := userRepo.GetDeviceByID(deviceID)
		if err != nil || device == nil {
			// Device doesn't exist, clean it up
			um.CleanupNonExistentDevice(deviceID)
			return
		}
		
		// Check if device is online
		if device.Status != "online" && device.Status != "Online" && 
		   device.Status != "connected" && device.Status != "Connected" {
			logrus.Debugf("Device %s is not online (status: %s), skipping worker creation", deviceID, device.Status)
			return
		}
		
		// Try to acquire lock for this device
		lockKey := fmt.Sprintf("%s%s", ultraWorkerLockPrefix, deviceID)
		locked, err := um.redisClient.SetNX(um.ctx, lockKey, "1", lockTTL).Result()
		if err != nil {
			logrus.Errorf("Failed to acquire lock for device %s: %v", deviceID, err)
			return
		}
		
		if !locked {
			// Another server is handling this device
			logrus.Debugf("Another server is handling device %s", deviceID)
			return
		}
		
		// Now we have the lock, log that we're creating the worker
		logrus.Infof("Worker doesn't exist for device %s, creating worker...", deviceID)
		
		// Check worker limit
		currentWorkers := atomic.LoadInt32(&um.activeWorkers)
		if currentWorkers >= maxConcurrentWorkers {
			logrus.Warnf("Worker limit reached (%d/%d), device %s will wait", 
				currentWorkers, maxConcurrentWorkers, deviceID)
			return
		}
		
		// Create worker
		um.workersMutex.Lock()
		if _, exists := um.workers[deviceID]; !exists {
			worker := um.createDeviceWorker(deviceID)
			if worker != nil {
				um.workers[deviceID] = worker
				atomic.AddInt32(&um.activeWorkers, 1)
				
				// Start worker
				um.wg.Add(1)
				go um.runWorker(deviceID, worker)
				
				logrus.Infof("Successfully created and started worker for device %s (total: %d)", 
					deviceID, atomic.LoadInt32(&um.activeWorkers))
			} else {
				logrus.Warnf("Could not create worker for device %s - WhatsApp client not available", deviceID)
			}
		} else {
			logrus.Debugf("Worker for device %s was already created by another goroutine", deviceID)
		}
		um.workersMutex.Unlock()
	}
}

// createDeviceWorker creates a new device worker
func (um *UltraScaleRedisManager) createDeviceWorker(deviceID string) *DeviceWorker {
	// First check if device is online
	userRepo := repository.GetUserRepository()
	device, err := userRepo.GetDeviceByID(deviceID)
	if err != nil {
		logrus.Errorf("Failed to get device info for %s: %v", deviceID, err)
		return nil
	}
	
	// Skip if device is not online
	if device.Status != "online" && device.Status != "Online" && 
	   device.Status != "connected" && device.Status != "Connected" {
		logrus.Warnf("Device %s is not online (status: %s), skipping worker creation", deviceID, device.Status)
		// Mark all pending messages for this device as skipped
		go um.skipOfflineDeviceMessages(deviceID)
		return nil
	}
	
	// Get WhatsApp client
	clientManager := whatsapp.GetClientManager()
	logrus.Infof("Getting WhatsApp client for device %s from ClientManager", deviceID)
	
	client, err := clientManager.GetClient(deviceID)
	if err != nil || client == nil {
		logrus.Errorf("Failed to get WhatsApp client for device %s: %v", deviceID, err)
		
		// Debug: Check if any clients are registered
		clientCount := clientManager.GetClientCount()
		logrus.Warnf("Total registered clients in ClientManager: %d", clientCount)
		
		// Debug: List all registered device IDs
		allClients := clientManager.GetAllClients()
		for id := range allClients {
			logrus.Warnf("Registered device ID: %s", id)
		}
		
		return nil
	}
	
	logrus.Infof("Successfully got WhatsApp client for device %s", deviceID)
	
	// Create worker using the existing constructor
	// Pass default delays that will be overridden by message-specific delays
	worker := NewDeviceWorker(deviceID, client, 10, 30)
	
	if worker == nil {
		logrus.Errorf("Failed to create worker for device %s", deviceID)
		return nil
	}
	
	return worker
}

// runWorker runs a device worker with optimizations
func (um *UltraScaleRedisManager) runWorker(deviceID string, worker *DeviceWorker) {
	defer um.wg.Done()
	
	// Safety check
	if worker == nil {
		logrus.Errorf("Cannot run nil worker for device %s", deviceID)
		atomic.AddInt32(&um.activeWorkers, -1)
		um.workersMutex.Lock()
		delete(um.workers, deviceID)
		um.workersMutex.Unlock()
		return
	}
	
	defer func() {
		atomic.AddInt32(&um.activeWorkers, -1)
		um.workersMutex.Lock()
		delete(um.workers, deviceID)
		um.workersMutex.Unlock()
		
		// Release lock
		lockKey := fmt.Sprintf("%s%s", ultraWorkerLockPrefix, deviceID)
		um.redisClient.Del(um.ctx, lockKey)
		
		// Stop the worker
		worker.Stop()
	}()
	
	// Start the worker
	worker.Start()
	
	// Queue keys for this device
	campaignQueue := fmt.Sprintf("%s%s", ultraCampaignQueuePrefix, deviceID)
	sequenceQueue := fmt.Sprintf("%s%s", ultraSequenceQueuePrefix, deviceID)
	deadLetterQueue := fmt.Sprintf("%s%s", ultraDeadLetterPrefix, deviceID)
	
	idleCount := 0
	maxIdle := 100 // Stop worker after 100 idle checks (10 seconds)
	
	for {
		select {
		case <-um.ctx.Done():
			return
		case <-worker.ctx.Done():
			return
		default:
			// Try campaign queue first (higher priority)
			processed := false
			
			// Check campaign queue
			if msg := um.popMessage(campaignQueue); msg != nil {
				if err := um.processMessage(worker, msg); err != nil {
					um.handleFailedMessage(msg, deadLetterQueue, err)
				}
				processed = true
				idleCount = 0
			} else if msg := um.popMessage(sequenceQueue); msg != nil {
				// Then check sequence queue
				if err := um.processMessage(worker, msg); err != nil {
					um.handleFailedMessage(msg, deadLetterQueue, err)
				}
				processed = true
				idleCount = 0
			}
			
			if !processed {
				idleCount++
				if idleCount >= maxIdle {
					// No messages for 10 seconds, stop worker
					logrus.Infof("Worker for device %s idle, stopping", deviceID)
					return
				}
				time.Sleep(queueCheckInterval)
			}
			
			// Extend lock periodically
			if idleCount%20 == 0 {
				lockKey := fmt.Sprintf("%s%s", ultraWorkerLockPrefix, deviceID)
				um.redisClient.Expire(um.ctx, lockKey, lockTTL)
			}
		}
	}
}

// popMessage pops a message from queue
func (um *UltraScaleRedisManager) popMessage(queueKey string) *UltraRedisMessage {
	result, err := um.redisClient.RPop(um.ctx, queueKey).Result()
	if err != nil || result == "" {
		return nil
	}
	
	logrus.Debugf("Popped message from queue %s", queueKey)
	
	var msg UltraRedisMessage
	if err := json.Unmarshal([]byte(result), &msg); err != nil {
		logrus.Errorf("Failed to unmarshal message: %v", err)
		return nil
	}
	
	return &msg
}

// processMessage processes a single message
func (um *UltraScaleRedisManager) processMessage(worker *DeviceWorker, msg *UltraRedisMessage) error {
	start := time.Now()
	
	// Queue the message to the worker's internal queue
	err := worker.QueueMessage(msg.Message)
	if err != nil {
		logrus.Errorf("Failed to queue message to worker: %v", err)
		um.incrementMetricBatch("messages_failed")
		
		// Update database status
		broadcastRepo := repository.GetBroadcastRepository()
		broadcastRepo.UpdateMessageStatus(msg.Message.ID, "failed", err.Error())
		return err
	}
	
	// Message queued successfully - the worker will handle status updates
	duration := time.Since(start)
	um.updateProcessingTime(msg.Message.DeviceID, duration)
	um.incrementMetricBatch("messages_queued_to_worker")
	
	logrus.Infof("Message %s queued to worker for %s", msg.Message.ID, msg.Message.RecipientPhone)
	return nil
}

// updateProcessingTime updates processing time metrics
func (um *UltraScaleRedisManager) updateProcessingTime(deviceID string, duration time.Duration) {
	key := fmt.Sprintf("%s%s:processing_time", ultraMetricsPrefix, deviceID)
	um.redisClient.LPush(um.ctx, key, duration.Milliseconds())
	um.redisClient.LTrim(um.ctx, key, 0, 99) // Keep last 100 values
}

// handleFailedMessage handles failed messages
func (um *UltraScaleRedisManager) handleFailedMessage(msg *UltraRedisMessage, deadLetterQueue string, err error) {
	msg.Retries++
	
	// If retries exceeded, move to dead letter
	if msg.Retries > 3 {
		data, _ := json.Marshal(msg)
		um.redisClient.LPush(um.ctx, deadLetterQueue, data)
		um.incrementMetricBatch("messages_dead_letter")
		return
	}
	
	// Otherwise, requeue with exponential backoff
	backoff := time.Duration(msg.Retries) * time.Minute
	time.Sleep(backoff)
	
	um.SendMessage(msg.Message)
}

// processQueues monitors all device queues
func (um *UltraScaleRedisManager) processQueues() {
	defer um.wg.Done()
	
	ticker := time.NewTicker(queueCheckInterval) // Changed from metricsInterval to queueCheckInterval
	defer ticker.Stop()
	
	for {
		select {
		case <-um.ctx.Done():
			return
		case <-ticker.C:
			// Check for devices with pending messages but no workers
			um.checkPendingQueues()
		}
	}
}

// checkPendingQueues checks for queues with messages but no workers
func (um *UltraScaleRedisManager) checkPendingQueues() {
	// Get all queue keys
	campaignQueues, _ := um.redisClient.Keys(um.ctx, ultraCampaignQueuePrefix+"*").Result()
	sequenceQueues, _ := um.redisClient.Keys(um.ctx, ultraSequenceQueuePrefix+"*").Result()
	
	allQueues := append(campaignQueues, sequenceQueues...)
	
	// No spam logging unless there are queues
	if len(allQueues) == 0 {
		return
	}
	
	// Get all valid devices once to avoid repeated DB queries
	userRepo := repository.GetUserRepository()
	validDevices := make(map[string]bool)
	
	for _, queueKey := range allQueues {
		// Extract device ID from queue key
		var deviceID string
		if strings.HasPrefix(queueKey, ultraCampaignQueuePrefix) {
			deviceID = strings.TrimPrefix(queueKey, ultraCampaignQueuePrefix)
		} else {
			deviceID = strings.TrimPrefix(queueKey, ultraSequenceQueuePrefix)
		}
		
		// Check if we already validated this device
		if validated, exists := validDevices[deviceID]; exists {
			if !validated {
				// Device doesn't exist, clean it up
				um.CleanupNonExistentDevice(deviceID)
				continue
			}
		} else {
			// Validate device exists
			device, err := userRepo.GetDeviceByID(deviceID)
			if err != nil || device == nil {
				validDevices[deviceID] = false
				um.CleanupNonExistentDevice(deviceID)
				continue
			}
			validDevices[deviceID] = true
		}
		
		// Check if queue has messages
		count, _ := um.redisClient.LLen(um.ctx, queueKey).Result()
		if count > 0 {
			// Only log if there are many messages
			if count > 10 {
				logrus.Debugf("Queue %s has %d messages, ensuring worker for device %s", queueKey, count, deviceID)
			}
			// Ensure worker exists
			um.ensureWorker(deviceID)
		}
	}
}

// monitorWorkers monitors worker health
func (um *UltraScaleRedisManager) monitorWorkers() {
	defer um.wg.Done()
	
	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-um.ctx.Done():
			return
		case <-ticker.C:
			um.performHealthCheck()
		}
	}
}

// performHealthCheck checks worker health
func (um *UltraScaleRedisManager) performHealthCheck() {
	um.workersMutex.RLock()
	workerCount := len(um.workers)
	deviceIDs := make([]string, 0, workerCount)
	for deviceID := range um.workers {
		deviceIDs = append(deviceIDs, deviceID)
	}
	um.workersMutex.RUnlock()
	
	// Log statistics only if there are active workers or every 10th check
	if workerCount > 0 {
		logrus.Infof("Health check: %d active workers out of %d max", 
			workerCount, maxConcurrentWorkers)
	}
	
	// Check and extend locks
	for _, deviceID := range deviceIDs {
		lockKey := fmt.Sprintf("%s%s", ultraWorkerLockPrefix, deviceID)
		um.redisClient.Expire(um.ctx, lockKey, lockTTL)
	}
	
	// Update global metrics
	um.redisClient.Set(um.ctx, "ultra:stats:active_workers", workerCount, 0)
	um.redisClient.Set(um.ctx, "ultra:stats:max_workers", maxConcurrentWorkers, 0)
}

// flushMetrics periodically flushes batched metrics to Redis
func (um *UltraScaleRedisManager) flushMetrics() {
	defer um.wg.Done()
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-um.ctx.Done():
			return
		case <-ticker.C:
			um.flushMetricsBatch()
		}
	}
}

// flushMetricsBatch flushes batched metrics
func (um *UltraScaleRedisManager) flushMetricsBatch() {
	um.metricsMutex.Lock()
	batch := um.metricsBatch
	um.metricsBatch = make(map[string]int64)
	um.metricsMutex.Unlock()
	
	// Use pipeline for efficiency
	pipe := um.redisClient.Pipeline()
	for metric, count := range batch {
		key := fmt.Sprintf("%s%s", ultraMetricsPrefix, metric)
		pipe.IncrBy(um.ctx, key, count)
	}
	pipe.Exec(um.ctx)
}

// incrementMetricBatch increments a metric in batch
func (um *UltraScaleRedisManager) incrementMetricBatch(metric string) {
	um.metricsMutex.Lock()
	um.metricsBatch[metric]++
	um.metricsMutex.Unlock()
}

// cleanupDeadLetters cleans up old dead letter messages
func (um *UltraScaleRedisManager) cleanupDeadLetters() {
	defer um.wg.Done()
	
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-um.ctx.Done():
			return
		case <-ticker.C:
			// Clean up old dead letters
			deadLetterQueues, _ := um.redisClient.Keys(um.ctx, ultraDeadLetterPrefix+"*").Result()
			for _, queue := range deadLetterQueues {
				count, _ := um.redisClient.LLen(um.ctx, queue).Result()
				if count > 1000 {
					// Keep only recent 1000
					um.redisClient.LTrim(um.ctx, queue, 0, 999)
				}
			}
		}
	}
}

// GetWorkerStatus returns status for a specific device
func (um *UltraScaleRedisManager) GetWorkerStatus(deviceID string) (domainBroadcast.WorkerStatus, bool) {
	// Check in-memory first
	um.workersMutex.RLock()
	worker, exists := um.workers[deviceID]
	um.workersMutex.RUnlock()
	
	if exists && worker != nil {
		return worker.GetStatus(), true
	}
	
	// Check Redis for distributed status
	data, err := um.redisClient.HGet(um.ctx, ultraWorkerStatusKey, deviceID).Result()
	if err != nil {
		return domainBroadcast.WorkerStatus{}, false
	}
	
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return domainBroadcast.WorkerStatus{}, false
	}
	
	// Get queue lengths
	campaignQueue := fmt.Sprintf("%s%s", ultraCampaignQueuePrefix, deviceID)
	sequenceQueue := fmt.Sprintf("%s%s", ultraSequenceQueuePrefix, deviceID)
	
	campaignCount, _ := um.redisClient.LLen(um.ctx, campaignQueue).Result()
	sequenceCount, _ := um.redisClient.LLen(um.ctx, sequenceQueue).Result()
	
	return domainBroadcast.WorkerStatus{
		DeviceID:     deviceID,
		Status:       status["status"].(string),
		QueueSize:    int(campaignCount + sequenceCount),
		LastActivity: time.Unix(int64(status["last_activity"].(float64)), 0),
	}, true
}

// GetAllWorkerStatus returns status for all workers
func (um *UltraScaleRedisManager) GetAllWorkerStatus() []domainBroadcast.WorkerStatus {
	var statuses []domainBroadcast.WorkerStatus
	
	// Get all workers from Redis (includes workers on other servers)
	workers, _ := um.redisClient.HGetAll(um.ctx, ultraWorkerStatusKey).Result()
	
	for deviceID, data := range workers {
		var status map[string]interface{}
		if err := json.Unmarshal([]byte(data), &status); err != nil {
			continue
		}
		
		// Get queue lengths
		campaignQueue := fmt.Sprintf("%s%s", ultraCampaignQueuePrefix, deviceID)
		sequenceQueue := fmt.Sprintf("%s%s", ultraSequenceQueuePrefix, deviceID)
		
		campaignCount, _ := um.redisClient.LLen(um.ctx, campaignQueue).Result()
		sequenceCount, _ := um.redisClient.LLen(um.ctx, sequenceQueue).Result()
		
		statuses = append(statuses, domainBroadcast.WorkerStatus{
			DeviceID:     deviceID,
			Status:       status["status"].(string),
			QueueSize:    int(campaignCount + sequenceCount),
			LastActivity: time.Unix(int64(status["last_activity"].(float64)), 0),
		})
	}
	
	return statuses
}

// StopAllWorkers stops all workers gracefully
func (um *UltraScaleRedisManager) StopAllWorkers() error {
	logrus.Info("Stopping all workers...")
	
	um.workersMutex.Lock()
	for deviceID, worker := range um.workers {
		worker.Stop()
		logrus.Infof("Stopped worker for device %s", deviceID)
	}
	um.workersMutex.Unlock()
	
	// Wait for all to stop
	time.Sleep(2 * time.Second)
	
	return nil
}

// ResumeFailedWorkers resumes workers that have failed
func (um *UltraScaleRedisManager) ResumeFailedWorkers() error {
	logrus.Info("Resuming failed workers...")
	
	// Check all queues for pending messages
	um.checkPendingQueues()
	
	return nil
}

// Stop gracefully stops the manager
func (um *UltraScaleRedisManager) Stop() {
	logrus.Info("Stopping Ultra Scale Redis Manager...")
	
	// Stop accepting new messages
	um.cancel()
	
	// Stop all workers
	um.StopAllWorkers()
	
	// Wait for all goroutines
	um.wg.Wait()
	
	// Close Redis connection
	um.redisClient.Close()
	
	logrus.Info("Ultra Scale Redis Manager stopped")
}

// Helper function to determine message priority
func getPriority(msg domainBroadcast.BroadcastMessage) int {
	if msg.CampaignID != nil {
		return 1 // High priority for campaigns
	}
	return 5 // Normal priority for sequences
}

// Interface compliance check
// var _ BroadcastManagerInterface = (*UltraScaleRedisManager)(nil)


// updateMessagesToNewDevice updates pending messages to use a new device ID
func (um *UltraScaleRedisManager) updateMessagesToNewDevice(oldDeviceID, newDeviceID string) {
	query := `
		UPDATE broadcast_messages 
		SET device_id = ? 
		WHERE device_id = ? AND STATUS = 'pending'
	`
	
	db := database.GetDB()
	result, err := db.Exec(query, newDeviceID, oldDeviceID)
	if err != nil {
		logrus.Errorf("Failed to update messages to new device: %v", err)
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	logrus.Infof("Updated %d messages from device %s to %s", rowsAffected, oldDeviceID, newDeviceID)
}

// skipOfflineDeviceMessages marks messages as skipped for offline devices
func (um *UltraScaleRedisManager) skipOfflineDeviceMessages(deviceID string) {
	query := `
		UPDATE broadcast_messages SET STATUS = 'skipped', error_message = 'Device offline' 
		WHERE device_id = ? AND STATUS = 'pending'
	`
	
	db := database.GetDB()
	result, err := db.Exec(query, deviceID)
	if err != nil {
		logrus.Errorf("Failed to skip messages for offline device: %v", err)
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	logrus.Infof("Skipped %d messages for offline device %s", rowsAffected, deviceID)
}
// cleanupOldMessages removes messages older than 24 hours
func (um *UltraScaleRedisManager) cleanupOldMessages() {
	defer um.wg.Done()
	
	ticker := time.NewTicker(1 * time.Hour) // Run every hour
	defer ticker.Stop()
	
	for {
		select {
		case <-um.ctx.Done():
			return
		case <-ticker.C:
			um.performMessageCleanup()
		}
	}
}

// performMessageCleanup does the actual cleanup
func (um *UltraScaleRedisManager) performMessageCleanup() {
	// Clean up old messages from database
	broadcastRepo := repository.GetBroadcastRepository()
	
	// Mark messages older than 24 hours as expired
	query := `
		UPDATE broadcast_messages SET STATUS = 'expired', 
		    error_message = 'Message expired (older than 24 hours)' 
		WHERE status IN ('pending', 'queued') 
		AND created_at < DATE_SUB(NOW(), INTERVAL 24 HOUR)
	`
	
	db := database.GetDB()
	result, err := db.Exec(query)
	if err != nil {
		logrus.Errorf("Failed to expire old messages: %v", err)
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		logrus.Infof("Expired %d old messages (>24 hours)", rowsAffected)
	}
	
	// Clean up old Redis queues
	pattern := fmt.Sprintf("%s*", ultraCampaignQueuePrefix)
	keys, err := um.redisClient.Keys(um.ctx, pattern).Result()
	if err != nil {
		return
	}
	
	// Also check sequence queues
	pattern2 := fmt.Sprintf("%s*", ultraSequenceQueuePrefix)
	keys2, err := um.redisClient.Keys(um.ctx, pattern2).Result()
	if err == nil {
		keys = append(keys, keys2...)
	}
	
	// Check each queue
	for _, key := range keys {
		// Get all messages from queue
		messages, err := um.redisClient.LRange(um.ctx, key, 0, -1).Result()
		if err != nil {
			continue
		}
		
		var toRemove []string
		for _, msgData := range messages {
			var msg UltraRedisMessage
			if err := json.Unmarshal([]byte(msgData), &msg); err != nil {
				continue
			}
			
			// Check if message is older than 24 hours
			if time.Since(msg.Timestamp) > 24*time.Hour {
				toRemove = append(toRemove, msgData)
				// Update database status
				broadcastRepo.UpdateMessageStatus(msg.Message.ID, "expired", "Message expired in queue")
			}
		}
		
		// Remove expired messages from Redis
		for _, msgData := range toRemove {
			um.redisClient.LRem(um.ctx, key, 1, msgData)
		}
		
		if len(toRemove) > 0 {
			logrus.Infof("Removed %d expired messages from queue %s", len(toRemove), key)
		}
	}
}
