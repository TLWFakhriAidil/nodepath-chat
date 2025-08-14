package broadcast

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

const (
	// Redis queue keys
	campaignQueueKey  = "broadcast:queue:campaign"
	sequenceQueueKey  = "broadcast:queue:sequence"
	deadLetterKey     = "broadcast:queue:dead"
	metricsPrefix     = "broadcast:metrics:"
	rateLimitPrefix   = "broadcast:ratelimit:"
	workerStatusKey   = "broadcast:workers"
	
	// Queue priorities
	priorityHigh   = 1
	priorityNormal = 5
	priorityLow    = 10
)

// RedisOptimizedBroadcastManager handles broadcasting with Redis queues for ultimate scale
type RedisOptimizedBroadcastManager struct {
	redisClient   *redis.Client
	workers       map[string]*DeviceWorker
	workersMutex  sync.RWMutex
	activeWorkers int32
	maxWorkers    int32
	
	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// RedisMessage represents a message in Redis queue
type RedisMessage struct {
	Message   domainBroadcast.BroadcastMessage `json:"message"`
	Priority  int                              `json:"priority"`
	Timestamp time.Time                        `json:"timestamp"`
	Retries   int                              `json:"retries"`
}

// NewRedisOptimizedBroadcastManager creates a new Redis-based broadcast manager
func NewRedisOptimizedBroadcastManager() *RedisOptimizedBroadcastManager {
	// Initialize Redis client
	redisURL := config.RedisURL
	if redisURL == "" {
		// Fallback to individual settings
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
	
	redisClient := redis.NewClient(opt)
	
	// Test connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logrus.Fatalf("Failed to connect to Redis: %v", err)
	}
	
	logrus.Info("Successfully connected to Redis")
	
	ctx, cancel := context.WithCancel(context.Background())
	manager := &RedisOptimizedBroadcastManager{
		redisClient: redisClient,
		workers:     make(map[string]*DeviceWorker),
		maxWorkers:  3000, // Can handle 3000 devices with Redis
		ctx:         ctx,
		cancel:      cancel,
	}
	
	// Start manager routines
	go manager.processQueues()
	go manager.monitorWorkers()
	go manager.cleanupDeadLetters()
	
	return manager
}

// SendMessage adds a message to the appropriate Redis queue
func (rm *RedisOptimizedBroadcastManager) SendMessage(msg domainBroadcast.BroadcastMessage) error {
	// Determine queue and priority
	queueKey := sequenceQueueKey
	priority := priorityNormal
	
	if msg.CampaignID != nil {
		queueKey = campaignQueueKey
		priority = priorityHigh // Campaigns get higher priority
	}
	
	// Create Redis message
	redisMsg := RedisMessage{
		Message:   msg,
		Priority:  priority,
		Timestamp: time.Now(),
		Retries:   0,
	}
	
	// Serialize message
	data, err := json.Marshal(redisMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}
	
	// Add to Redis queue with priority score
	score := float64(priority)*1e10 + float64(time.Now().Unix())
	if err := rm.redisClient.ZAdd(rm.ctx, queueKey, &redis.Z{
		Score:  score,
		Member: string(data),
	}).Err(); err != nil {
		return fmt.Errorf("failed to add message to queue: %v", err)
	}
	
	// Update metrics
	rm.incrementMetric("messages:queued")
	
	return nil
}

// processQueues continuously processes messages from Redis queues
func (rm *RedisOptimizedBroadcastManager) processQueues() {
	for {
		select {
		case <-rm.ctx.Done():
			return
		default:
			// Process campaign queue first (higher priority)
			rm.processQueue(campaignQueueKey)
			
			// Then process sequence queue
			rm.processQueue(sequenceQueueKey)
			
			// Small delay to prevent CPU spinning
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// processQueue processes messages from a specific queue
func (rm *RedisOptimizedBroadcastManager) processQueue(queueKey string) {
	// Get batch of messages
	results, err := rm.redisClient.ZPopMin(rm.ctx, queueKey, 10).Result()
	if err != nil || len(results) == 0 {
		return
	}
	
	for _, result := range results {
		// Deserialize message
		var redisMsg RedisMessage
		if err := json.Unmarshal([]byte(result.Member.(string)), &redisMsg); err != nil {
			logrus.Errorf("Failed to unmarshal message: %v", err)
			continue
		}
		
		// Process message
		if err := rm.processMessage(redisMsg); err != nil {
			logrus.Errorf("Failed to process message: %v", err)
			// Add to retry queue
			rm.retryMessage(redisMsg)
		}
	}
}

// processMessage processes a single message
func (rm *RedisOptimizedBroadcastManager) processMessage(redisMsg RedisMessage) error {
	msg := redisMsg.Message
	
	// Get or create worker
	worker := rm.getOrCreateWorker(msg.DeviceID)
	if worker == nil {
		return fmt.Errorf("no worker available for device %s", msg.DeviceID)
	}
	
	// Rate limiting removed - using campaign/sequence delays instead
	
	// Send message through worker
	startTime := time.Now()
	err := worker.SendMessage(msg)
	processingTime := time.Since(startTime)
	
	// Update metrics
	if err == nil {
		rm.incrementMetric("messages:sent")
		rm.updateProcessingTime(msg.DeviceID, processingTime)
		
		// Update message status in database
		repo := repository.GetBroadcastRepository()
		repo.UpdateMessageStatus(msg.ID, "sent", "")
	} else {
		rm.incrementMetric("messages:failed")
		return err
	}
	
	return nil
}

// getOrCreateWorker gets or creates a worker for a device
func (rm *RedisOptimizedBroadcastManager) getOrCreateWorker(deviceID string) *DeviceWorker {
	rm.workersMutex.RLock()
	worker, exists := rm.workers[deviceID]
	rm.workersMutex.RUnlock()
	
	if exists && worker.IsHealthy() {
		return worker
	}
	
	// Create new worker
	rm.workersMutex.Lock()
	defer rm.workersMutex.Unlock()
	
	// Check if we've hit max workers
	currentWorkers := atomic.LoadInt32(&rm.activeWorkers)
	if currentWorkers >= rm.maxWorkers {
		logrus.Warnf("Max workers reached (%d), cannot create worker for device %s", rm.maxWorkers, deviceID)
		return nil
	}
	
	// Get device
	clientManager := whatsapp.GetClientManager()
	client, err := clientManager.GetClient(deviceID)
	if err != nil || client == nil {
		logrus.Errorf("Failed to get device %s: %v", deviceID, err)
		return nil
	}
	
	// Create new worker
	worker = NewDeviceWorker(deviceID, client, 10, 30)
	rm.workers[deviceID] = worker
	
	// Update worker count
	atomic.AddInt32(&rm.activeWorkers, 1)
	
	// Start worker
	go func() {
		worker.Run()
		// When worker stops, update count
		atomic.AddInt32(&rm.activeWorkers, -1)
	}()
	
	// Update worker status in Redis
	rm.updateWorkerStatus(deviceID, "active")
	
	logrus.Infof("Created new worker for device %s", deviceID)
	return worker
}

// checkRateLimit checks if device can send another message - DISABLED
// Using campaign/sequence min/max delays instead
/*
func (rm *RedisOptimizedBroadcastManager) checkRateLimit(deviceID string) bool {
	key := fmt.Sprintf("%s%s", rateLimitPrefix, deviceID)
	
	// Check minute rate limit (20/min)
	minuteKey := fmt.Sprintf("%s:minute:%d", key, time.Now().Unix()/60)
	count, _ := rm.redisClient.Incr(rm.ctx, minuteKey).Result()
	rm.redisClient.Expire(rm.ctx, minuteKey, 2*time.Minute)
	
	if count > 20 {
		return false
	}
	
	// Check hour rate limit (500/hour)
	hourKey := fmt.Sprintf("%s:hour:%d", key, time.Now().Unix()/3600)
	count, _ = rm.redisClient.Incr(rm.ctx, hourKey).Result()
	rm.redisClient.Expire(rm.ctx, hourKey, 2*time.Hour)
	
	if count > 500 {
		return false
	}
	
	// Check day rate LIMIT (5000/day)
	dayKey := fmt.Sprintf("%s:day:%s", key, time.Now().Format("2006-01-02"))
	count, _ = rm.redisClient.Incr(rm.ctx, dayKey).Result()
	rm.redisClient.Expire(rm.ctx, dayKey, 25*time.Hour)
	
	if count > 5000 {
		return false
	}
	
	return true
}
*/

// retryMessage adds message back to queue with exponential backoff
func (rm *RedisOptimizedBroadcastManager) retryMessage(redisMsg RedisMessage) {
	redisMsg.Retries++
	
	// Max 3 retries
	if redisMsg.Retries > 3 {
		// Move to dead letter queue
		rm.moveToDeadLetter(redisMsg)
		return
	}
	
	// Calculate backoff delay (1min, 4min, 9min)
	delay := time.Duration(redisMsg.Retries*redisMsg.Retries) * time.Minute
	
	// Re-add to queue with future timestamp
	redisMsg.Timestamp = time.Now().Add(delay)
	
	data, _ := json.Marshal(redisMsg)
	score := float64(redisMsg.Priority)*1e10 + float64(redisMsg.Timestamp.Unix())
	
	queueKey := sequenceQueueKey
	if redisMsg.Message.CampaignID != nil {
		queueKey = campaignQueueKey
	}
	
	rm.redisClient.ZAdd(rm.ctx, queueKey, &redis.Z{
		Score:  score,
		Member: string(data),
	})
}

// moveToDeadLetter moves failed message to dead letter queue
func (rm *RedisOptimizedBroadcastManager) moveToDeadLetter(redisMsg RedisMessage) {
	data, _ := json.Marshal(redisMsg)
	rm.redisClient.LPush(rm.ctx, deadLetterKey, string(data))
	rm.incrementMetric("messages:dead_letter")
	
	// Update message status in database
	repo := repository.GetBroadcastRepository()
	repo.UpdateMessageStatus(redisMsg.Message.ID, "failed", "Max retries exceeded")
}

// monitorWorkers monitors worker health
func (rm *RedisOptimizedBroadcastManager) monitorWorkers() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.checkWorkerHealth()
		}
	}
}

// checkWorkerHealth checks health of all workers
func (rm *RedisOptimizedBroadcastManager) checkWorkerHealth() {
	rm.workersMutex.RLock()
	deviceIDs := make([]string, 0, len(rm.workers))
	for deviceID := range rm.workers {
		deviceIDs = append(deviceIDs, deviceID)
	}
	rm.workersMutex.RUnlock()
	
	for _, deviceID := range deviceIDs {
		rm.workersMutex.RLock()
		worker := rm.workers[deviceID]
		rm.workersMutex.RUnlock()
		
		if worker != nil && !worker.IsHealthy() {
			logrus.Warnf("Worker for device %s is unhealthy, removing", deviceID)
			
			// Stop worker
			worker.Stop()
			
			// Remove from map
			rm.workersMutex.Lock()
			delete(rm.workers, deviceID)
			rm.workersMutex.Unlock()
			
			// Update status in Redis
			rm.updateWorkerStatus(deviceID, "failed")
		}
	}
}

// cleanupDeadLetters periodically processes dead letter queue
func (rm *RedisOptimizedBroadcastManager) cleanupDeadLetters() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			// Process old dead letters (could implement retry logic here)
			count, _ := rm.redisClient.LLen(rm.ctx, deadLetterKey).Result()
			if count > 1000 {
				// Trim old messages
				rm.redisClient.LTrim(rm.ctx, deadLetterKey, 0, 999)
			}
		}
	}
}

// Metric helpers
func (rm *RedisOptimizedBroadcastManager) incrementMetric(metric string) {
	key := fmt.Sprintf("%s%s", metricsPrefix, metric)
	rm.redisClient.Incr(rm.ctx, key)
}

func (rm *RedisOptimizedBroadcastManager) updateProcessingTime(deviceID string, duration time.Duration) {
	key := fmt.Sprintf("%s%s:processing_time", metricsPrefix, deviceID)
	rm.redisClient.LPush(rm.ctx, key, duration.Milliseconds())
	rm.redisClient.LTrim(rm.ctx, key, 0, 99) // Keep last 100 values
}

func (rm *RedisOptimizedBroadcastManager) updateWorkerStatus(deviceID, status string) {
	data := map[string]interface{}{
		"device_id":     deviceID,
		"status":        status,
		"last_activity": time.Now().Unix(),
	}
	jsonData, _ := json.Marshal(data)
	rm.redisClient.HSet(rm.ctx, workerStatusKey, deviceID, string(jsonData))
}

// Interface implementations
func (rm *RedisOptimizedBroadcastManager) GetWorkerStatus(deviceID string) (domainBroadcast.WorkerStatus, bool) {
	// First check in-memory
	rm.workersMutex.RLock()
	worker, exists := rm.workers[deviceID]
	rm.workersMutex.RUnlock()
	
	if exists && worker != nil {
		return worker.GetStatus(), true
	}
	
	// Check Redis for status
	data, err := rm.redisClient.HGet(rm.ctx, workerStatusKey, deviceID).Result()
	if err != nil {
		return domainBroadcast.WorkerStatus{}, false
	}
	
	var status map[string]interface{}
	if err := json.Unmarshal([]byte(data), &status); err != nil {
		return domainBroadcast.WorkerStatus{}, false
	}
	
	return domainBroadcast.WorkerStatus{
		DeviceID:     deviceID,
		Status:       status["status"].(string),
		LastActivity: time.Unix(int64(status["last_activity"].(float64)), 0),
	}, true
}

func (rm *RedisOptimizedBroadcastManager) GetAllWorkerStatus() []domainBroadcast.WorkerStatus {
	statuses := []domainBroadcast.WorkerStatus{}
	
	// Get all from Redis
	allData, err := rm.redisClient.HGetAll(rm.ctx, workerStatusKey).Result()
	if err != nil {
		return statuses
	}
	
	for deviceID, data := range allData {
		var status map[string]interface{}
		if err := json.Unmarshal([]byte(data), &status); err != nil {
			continue
		}
		
		// Get queue sizes from Redis
		campaignCount, _ := rm.redisClient.ZCard(rm.ctx, campaignQueueKey).Result()
		sequenceCount, _ := rm.redisClient.ZCard(rm.ctx, sequenceQueueKey).Result()
		
		workerStatus := domainBroadcast.WorkerStatus{
			DeviceID:     deviceID,
			Status:       status["status"].(string),
			QueueSize:    int(campaignCount + sequenceCount),
			LastActivity: time.Unix(int64(status["last_activity"].(float64)), 0),
		}
		
		// Get metrics from Redis
		sentKey := fmt.Sprintf("%s%s:sent", metricsPrefix, deviceID)
		failedKey := fmt.Sprintf("%s%s:failed", metricsPrefix, deviceID)
		
		sent, _ := rm.redisClient.Get(rm.ctx, sentKey).Int()
		failed, _ := rm.redisClient.Get(rm.ctx, failedKey).Int()
		
		workerStatus.ProcessedCount = sent
		workerStatus.FailedCount = failed
		
		statuses = append(statuses, workerStatus)
	}
	
	return statuses
}

func (rm *RedisOptimizedBroadcastManager) StopAllWorkers() error {
	rm.workersMutex.Lock()
	defer rm.workersMutex.Unlock()
	
	for deviceID, worker := range rm.workers {
		if worker != nil {
			worker.Stop()
			rm.updateWorkerStatus(deviceID, "stopped")
		}
	}
	
	rm.workers = make(map[string]*DeviceWorker)
	atomic.StoreInt32(&rm.activeWorkers, 0)
	
	logrus.Info("All workers stopped")
	return nil
}

func (rm *RedisOptimizedBroadcastManager) ResumeFailedWorkers() error {
	// Get all failed workers from Redis
	allData, err := rm.redisClient.HGetAll(rm.ctx, workerStatusKey).Result()
	if err != nil {
		return err
	}
	
	resumed := 0
	for deviceID, data := range allData {
		var status map[string]interface{}
		if err := json.Unmarshal([]byte(data), &status); err != nil {
			continue
		}
		
		if status["status"] == "failed" {
			// Try to create worker
			if worker := rm.getOrCreateWorker(deviceID); worker != nil {
				resumed++
			}
		}
	}
	
	logrus.Infof("Resumed %d failed workers", resumed)
	return nil
}
