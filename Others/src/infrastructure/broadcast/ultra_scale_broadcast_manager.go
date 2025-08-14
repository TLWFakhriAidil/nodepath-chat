package broadcast

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/antipattern"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// DeviceWorkerGroup manages multiple workers for a single device
type DeviceWorkerGroup struct {
	deviceID      string
	workers       []*BroadcastWorker
	messageQueue  chan *domainBroadcast.BroadcastMessage
	currentWorker int32 // For round-robin distribution
	mu            sync.RWMutex
	
	// Rate limiting - ensures sequential sending
	lastSentTime  time.Time
	sendMutex     sync.Mutex  // Only one worker can send at a time
}

// BroadcastWorkerPool manages workers per broadcast (campaign/sequence)
type BroadcastWorkerPool struct {
	broadcastID   string
	broadcastType string // "campaign" or "sequence"
	deviceGroups  map[string]*DeviceWorkerGroup // key: deviceID -> group of workers
	maxWorkers    int
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	redisClient   *redis.Client
	
	// Statistics
	totalMessages    int64
	processedCount   int64
	failedCount      int64
	startTime        time.Time
	completionTime   *time.Time
}

// BroadcastWorker handles messages for a specific device within a broadcast
type BroadcastWorker struct {
	poolID        string
	deviceID      string
	workerID      int // Worker number within device group
	broadcastID   string
	broadcastType string
	messageSender *WhatsAppMessageSender // Real WhatsApp sender
	pool          *BroadcastWorkerPool   // Reference to parent pool
	
	// Message processing
	status        string
	processedCount int64
	failedCount    int64
	
	// Context management
	ctx            context.Context
	cancel         context.CancelFunc
	lastActivity   time.Time
	mu             sync.RWMutex
}

// Global manager instance
var (
	broadcastManager     *UltraScaleBroadcastManager
	broadcastManagerOnce sync.Once
)

// UltraScaleBroadcastManager manages broadcast pools for 3000+ devices
// This version creates broadcast-specific worker pools instead of global pools
type UltraScaleBroadcastManager struct {
	pools         map[string]*BroadcastWorkerPool // key: "campaign:123" or "sequence:abc-def"
	redisClient   *redis.Client
	mu            sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
	
	// Configuration
	maxPoolsPerType      int
	maxWorkersPerPool    int
	workerQueueSize      int
	messageQueueTimeout  time.Duration // Increased from 5s to 30s
	
	// Statistics
	activePools      int32
	totalMessages    int64
	processedMessages int64
	failedMessages    int64
}

// NewUltraScaleBroadcastManager creates a new broadcast manager optimized for scale
func NewUltraScaleBroadcastManager(redisClient *redis.Client) *UltraScaleBroadcastManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &UltraScaleBroadcastManager{
		pools:               make(map[string]*BroadcastWorkerPool),
		redisClient:         redisClient,
		ctx:                 ctx,
		cancel:              cancel,
		maxPoolsPerType:     100,
		maxWorkersPerPool:   config.MaxConcurrentWorkers,
		workerQueueSize:     config.WorkerQueueSize,
		messageQueueTimeout: 30 * time.Second, // Increased from 5s to 30s
	}
	
	// Start monitoring
	go manager.monitorPools()
	
	return manager
}

// GetBroadcastManager returns the singleton broadcast manager
func GetBroadcastManager() *UltraScaleBroadcastManager {
	broadcastManagerOnce.Do(func() {
		// Try to get Redis client, but work without it
		var redisClient *redis.Client
		if config.RedisURL != "" {
			redisClient = redis.NewClient(&redis.Options{
				Addr:     config.RedisURL,
				Password: config.RedisPassword,
			})
			
			// Test connection
			ctx := context.Background()
			if err := redisClient.Ping(ctx).Err(); err != nil {
				logrus.Warnf("Redis connection failed, continuing without Redis: %v", err)
				redisClient = nil
			}
		}
		
		broadcastManager = NewUltraScaleBroadcastManager(redisClient)
		logrus.Info("Ultra-scale broadcast manager initialized for 3000+ devices")
	})
	
	return broadcastManager
}

// GetOrCreatePool gets or creates a broadcast-specific worker pool
func (m *UltraScaleBroadcastManager) GetOrCreatePool(broadcastType string, broadcastID string) (*BroadcastWorkerPool, error) {
	poolKey := fmt.Sprintf("%s:%s", broadcastType, broadcastID)
	
	m.mu.RLock()
	pool, exists := m.pools[poolKey]
	m.mu.RUnlock()
	
	if exists {
		return pool, nil
	}
	
	// Create new pool
	m.mu.Lock()
	defer m.mu.Unlock()
	
	// Double-check
	if pool, exists = m.pools[poolKey]; exists {
		return pool, nil
	}
	
	// Create context for this pool
	ctx, cancel := context.WithCancel(m.ctx)
	
	pool = &BroadcastWorkerPool{
		broadcastID:   broadcastID,
		broadcastType: broadcastType,
		deviceGroups:  make(map[string]*DeviceWorkerGroup),
		maxWorkers:    m.maxWorkersPerPool,
		ctx:           ctx,
		cancel:        cancel,
		redisClient:   m.redisClient,
		startTime:     time.Now(),
	}
	
	m.pools[poolKey] = pool
	atomic.AddInt32(&m.activePools, 1)
	
	logrus.Infof("Created worker pool for %s %s", broadcastType, broadcastID)
	
	return pool, nil
}

// QueueMessage adds a message to the appropriate worker pool with improved timeout
func (m *UltraScaleBroadcastManager) QueueMessage(msg *domainBroadcast.BroadcastMessage) error {
	if msg == nil {
		return fmt.Errorf("message is nil")
	}
	
	// Determine broadcast type and ID
	var broadcastType, broadcastID string
	if msg.CampaignID != nil {
		broadcastType = "campaign"
		broadcastID = fmt.Sprintf("%d", *msg.CampaignID)
	} else if msg.SequenceID != nil {
		broadcastType = "sequence"
		broadcastID = *msg.SequenceID
	} else {
		return fmt.Errorf("message has no campaign or sequence ID")
	}
	
	// Get or create pool
	pool, err := m.GetOrCreatePool(broadcastType, broadcastID)
	if err != nil {
		return fmt.Errorf("failed to get broadcast pool: %w", err)
	}
	
	// Queue to pool with improved timeout
	return pool.QueueMessage(msg)
}

// QueueMessage adds a message to the broadcast pool with better timeout handling
func (bwp *BroadcastWorkerPool) QueueMessage(msg *domainBroadcast.BroadcastMessage) error {
	atomic.AddInt64(&bwp.totalMessages, 1)
	
	// Get or create device worker group
	group := bwp.getOrCreateWorkerGroup(msg.DeviceID)
	
	// Queue to group with increased timeout (30 seconds instead of 5)
	select {
	case group.messageQueue <- msg:
		// Update message status to queued
		db := database.GetDB()
		_, err := db.Exec(`UPDATE broadcast_messages SET STATUS = 'queued' WHERE id = ? AND status IN ('pending', 'processing')`, msg.ID)
		if err != nil {
			logrus.Errorf("Failed to update message status: %v", err)
		}
		return nil
	case <-time.After(30 * time.Second): // 30 second timeout
		atomic.AddInt64(&bwp.failedCount, 1)
		// Log detailed error for debugging
		logrus.Errorf("Timeout queueing message to device %s after 30s. Queue size: %d/%d", 
			msg.DeviceID, len(group.messageQueue), cap(group.messageQueue))
		return fmt.Errorf("timeout queueing message to worker after 30 seconds")
	}
}

// getOrCreateWorkerGroup gets existing worker group or creates new one with multiple workers
func (bwp *BroadcastWorkerPool) getOrCreateWorkerGroup(deviceID string) *DeviceWorkerGroup {
	bwp.mu.RLock()
	group, exists := bwp.deviceGroups[deviceID]
	bwp.mu.RUnlock()
	
	if exists {
		return group
	}
	
	bwp.mu.Lock()
	defer bwp.mu.Unlock()
	
	// Double-check after acquiring write lock
	if group, exists = bwp.deviceGroups[deviceID]; exists {
		return group
	}
	
	// Create new worker group with multiple workers
	group = &DeviceWorkerGroup{
		deviceID:     deviceID,
		workers:      make([]*BroadcastWorker, 0, config.MaxWorkersPerDevice),
		messageQueue: make(chan *domainBroadcast.BroadcastMessage, config.WorkerQueueSize),
	}
	
	// Create multiple workers per device (5 as per config)
	for i := 0; i < config.MaxWorkersPerDevice; i++ {
		ctx, cancel := context.WithCancel(bwp.ctx)
		worker := &BroadcastWorker{
			poolID:        fmt.Sprintf("%s:%s", bwp.broadcastType, bwp.broadcastID),
			deviceID:      deviceID,
			workerID:      i,
			broadcastID:   bwp.broadcastID,
			broadcastType: bwp.broadcastType,
			messageSender: NewWhatsAppMessageSender(),
			pool:          bwp,
			status:        "idle",
			ctx:           ctx,
			cancel:        cancel,
			lastActivity:  time.Now(),
		}
		
		group.workers = append(group.workers, worker)
		
		// Start worker
		go worker.process(group.messageQueue)
		
		logrus.Infof("Started worker %d for device %s in %s %s", 
			i, deviceID, bwp.broadcastType, bwp.broadcastID)
	}
	
	bwp.deviceGroups[deviceID] = group
	
	logrus.Infof("Created worker group with %d workers for device %s", 
		config.MaxWorkersPerDevice, deviceID)
	
	return group
}

// process handles messages for this worker from the shared device queue
func (bw *BroadcastWorker) process(messageQueue <-chan *domainBroadcast.BroadcastMessage) {
	logrus.Infof("Worker %d started for device %s in %s", 
		bw.workerID, bw.deviceID, bw.poolID)
	
	for {
		select {
		case <-bw.ctx.Done():
			logrus.Infof("Worker %d stopped for device %s", bw.workerID, bw.deviceID)
			return
			
		case msg, ok := <-messageQueue:
			if !ok {
				logrus.Infof("Message queue closed for worker %d device %s", 
					bw.workerID, bw.deviceID)
				return
			}
			bw.processMessage(msg)
		}
	}
}

// processMessage sends a single message with rate limiting
func (bw *BroadcastWorker) processMessage(msg *domainBroadcast.BroadcastMessage) {
	bw.mu.Lock()
	bw.status = "processing"
	bw.lastActivity = time.Now()
	bw.mu.Unlock()
	
	// Get the device worker group
	bw.pool.mu.RLock()
	group, exists := bw.pool.deviceGroups[bw.deviceID]
	bw.pool.mu.RUnlock()
	
	if !exists {
		logrus.Errorf("Worker %d: Device group not found for %s", bw.workerID, bw.deviceID)
		return
	}
	
	// Log which broadcast this message belongs to
	broadcastInfo := "Unknown broadcast"
	if msg.CampaignID != nil {
		broadcastInfo = fmt.Sprintf("Campaign %d", *msg.CampaignID)
	} else if msg.SequenceID != nil {
		broadcastInfo = fmt.Sprintf("Sequence %s", *msg.SequenceID)
	}
	
	// CRITICAL: Acquire send permission (this enforces rate limiting)
	minDelay := msg.MinDelay
	maxDelay := msg.MaxDelay
	if minDelay <= 0 {
		minDelay = 5  // Default minimum
	}
	if maxDelay <= 0 {
		maxDelay = 15 // Default maximum
	}
	
	// SAFETY CHECK: Verify message wasn't already sent
	db := database.GetDB()
	var currentStatus string
	err := db.QueryRow("SELECT status FROM broadcast_messages WHERE id = ?", msg.ID).Scan(&currentStatus)
	if err == nil && currentStatus == "sent" {
		logrus.Warnf("Worker %d: Message %s already sent, skipping duplicate send", bw.workerID, msg.ID)
		return
	}
	
	// This will block until it's this worker's turn to send
	group.acquireSendPermission(minDelay, maxDelay)
	
	// Now we have exclusive permission to send
	logrus.Debugf("Worker %d on device %s sending message %s for %s to %s", 
		bw.workerID, bw.deviceID, msg.ID, broadcastInfo, msg.RecipientPhone)
	
	// Send via WhatsApp
	sendErr := bw.sendWhatsAppMessage(msg)
	
	// IMPORTANT: Release permission after sending
	group.releaseSendPermission()
	
	db2 := database.GetDB()
	if sendErr != nil {
		atomic.AddInt64(&bw.failedCount, 1)
		// Also increment pool's failed count
		if bw.pool != nil {
			atomic.AddInt64(&bw.pool.failedCount, 1)
		}
		// Update status to failed
		db2.Exec(`UPDATE broadcast_messages SET STATUS = 'failed', error_message = ?, updated_at = NOW() WHERE id = ?`, 
			sendErr.Error(), msg.ID)
		logrus.Errorf("Failed to send message %s: %v", msg.ID, sendErr)
	} else {
		atomic.AddInt64(&bw.processedCount, 1)
		// Also increment pool's processed count
		if bw.pool != nil {
			atomic.AddInt64(&bw.pool.processedCount, 1)
		}
		// Update status to sent (preserve processing_worker_id for audit trail)
		db2.Exec(`UPDATE broadcast_messages SET status = 'sent', sent_at = NOW() WHERE id = ? AND status IN ('queued', 'processing')`, msg.ID)
		
		// Update sequence progress if this is a sequence message
		if msg.SequenceID != nil {
			db2.Exec(`UPDATE sequence_contacts SET last_message_at = NOW() WHERE sequence_id = ? AND contact_phone = ?`,
				*msg.SequenceID, msg.RecipientPhone)
			// Call the progress update function
			db2.Exec(`SELECT update_sequence_progress(?)`, *msg.SequenceID)
		}
		
		// Successfully sent message
	}
	
	bw.mu.Lock()
	bw.status = "idle"
	bw.lastActivity = time.Now()
	bw.mu.Unlock()
}

// sendWhatsAppMessage sends the actual message via WhatsApp
func (bw *BroadcastWorker) sendWhatsAppMessage(msg *domainBroadcast.BroadcastMessage) error {
	if bw.messageSender == nil {
		return fmt.Errorf("message sender not initialized")
	}
	
	// Check if content is empty and use Message field as fallback
	if msg.Content == "" && msg.Message != "" {
		msg.Content = msg.Message
	}
	
	// Apply anti-spam for ALL devices (both WhatsApp Web and Platform)
	// Create message randomizer and greeting processor
	messageRandomizer := antipattern.NewMessageRandomizer()
	greetingProcessor := antipattern.NewGreetingProcessor()
	
	// STEP 1: Apply randomization to CONTENT ONLY
	randomizedContent := messageRandomizer.RandomizeMessage(msg.Content)
	
	// STEP 2: Add greeting to the randomized content
	finalContent := greetingProcessor.PrepareMessageWithGreeting(
		randomizedContent,
		msg.RecipientName,
		bw.deviceID,
		msg.RecipientPhone,
	)
	
	// Update the message content
	msg.Content = finalContent
	msg.Message = finalContent
	
	logrus.Debugf("Applied anti-spam for device %s: randomized and added greeting", bw.deviceID)
	
	// Use the self-healing message sender with modified content
	return bw.messageSender.SendMessage(bw.deviceID, msg)
}

// monitorPools monitors and cleans up idle pools
func (m *UltraScaleBroadcastManager) monitorPools() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.cleanupIdlePools()
		}
	}
}

// cleanupIdlePools removes idle broadcast pools
func (m *UltraScaleBroadcastManager) cleanupIdlePools() {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for poolKey, pool := range m.pools {
		// Check if pool is idle (no activity for 30 minutes)
		pool.mu.RLock()
		idleTime := time.Since(pool.startTime)
		if pool.completionTime != nil {
			idleTime = time.Since(*pool.completionTime)
		}
		hasActiveWorkers := false
		for _, group := range pool.deviceGroups {
			if len(group.messageQueue) > 0 {
				hasActiveWorkers = true
				break
			}
		}
		pool.mu.RUnlock()
		
		if idleTime > 30*time.Minute && !hasActiveWorkers {
			// Shutdown pool
			pool.Shutdown()
			delete(m.pools, poolKey)
			atomic.AddInt32(&m.activePools, -1)
			logrus.Infof("Cleaned up idle pool: %s", poolKey)
		}
	}
}

// Shutdown gracefully shuts down a worker pool
func (bwp *BroadcastWorkerPool) Shutdown() {
	bwp.cancel()
	
	bwp.mu.Lock()
	defer bwp.mu.Unlock()
	
	// Shutdown all worker groups
	for _, group := range bwp.deviceGroups {
		close(group.messageQueue)
		for _, worker := range group.workers {
			worker.cancel()
		}
	}
	
	if bwp.completionTime == nil {
		now := time.Now()
		bwp.completionTime = &now
	}
	
	logrus.Infof("Shut down worker pool for %s %s. Processed: %d, Failed: %d",
		bwp.broadcastType, bwp.broadcastID, 
		atomic.LoadInt64(&bwp.processedCount),
		atomic.LoadInt64(&bwp.failedCount))
}

// GetStatistics returns current statistics for the manager
func (m *UltraScaleBroadcastManager) GetStatistics() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	stats := map[string]interface{}{
		"active_pools":       atomic.LoadInt32(&m.activePools),
		"total_messages":     atomic.LoadInt64(&m.totalMessages),
		"processed_messages": atomic.LoadInt64(&m.processedMessages),
		"failed_messages":    atomic.LoadInt64(&m.failedMessages),
		"pools":              make([]map[string]interface{}, 0),
	}
	
	for poolKey, pool := range m.pools {
		poolStats := map[string]interface{}{
			"key":            poolKey,
			"total_messages": atomic.LoadInt64(&pool.totalMessages),
			"processed":      atomic.LoadInt64(&pool.processedCount),
			"failed":         atomic.LoadInt64(&pool.failedCount),
			"device_count":   len(pool.deviceGroups),
			"worker_count":   len(pool.deviceGroups) * config.MaxWorkersPerDevice,
		}
		stats["pools"] = append(stats["pools"].([]map[string]interface{}), poolStats)
	}
	
	return stats
}

// calculateRandomDelay calculates a random delay between min and max
func calculateRandomDelay(minSeconds, maxSeconds int) time.Duration {
	if minSeconds >= maxSeconds {
		return time.Duration(minSeconds) * time.Second
	}
	
	// Generate truly random delay between min and max
	rand.Seed(time.Now().UnixNano())
	delayRange := maxSeconds - minSeconds
	randomDelay := minSeconds + rand.Intn(delayRange+1) // +1 to include maxSeconds
	
	logrus.Debugf("Random delay: %d seconds (between %d-%d)", randomDelay, minSeconds, maxSeconds)
	return time.Duration(randomDelay) * time.Second
}

// acquireSendPermission ensures only one worker sends at a time with proper delay
func (dwg *DeviceWorkerGroup) acquireSendPermission(minDelay, maxDelay int) {
	dwg.sendMutex.Lock()
	// Don't unlock here - the worker will unlock after sending
	
	// Calculate time since last send
	timeSinceLastSend := time.Since(dwg.lastSentTime)
	
	// Calculate required delay
	requiredDelay := calculateRandomDelay(minDelay, maxDelay)
	
	// If not enough time has passed, wait
	if timeSinceLastSend < requiredDelay {
		waitTime := requiredDelay - timeSinceLastSend
		logrus.Debugf("Device %s: Waiting %v before next send (rate limiting)", dwg.deviceID, waitTime)
		time.Sleep(waitTime)
	}
}

// releaseSendPermission updates last sent time and releases the mutex
func (dwg *DeviceWorkerGroup) releaseSendPermission() {
	dwg.lastSentTime = time.Now()
	dwg.sendMutex.Unlock()
}

// Backward compatibility functions

// GetUltraScaleBroadcastManager returns the global broadcast manager
func GetUltraScaleBroadcastManager() *UltraScaleBroadcastManager {
	return GetBroadcastManager()
}

// StartBroadcastPool starts a broadcast pool (backward compatibility)
func (m *UltraScaleBroadcastManager) StartBroadcastPool(broadcastType, broadcastID string) (*BroadcastWorkerPool, error) {
	return m.GetOrCreatePool(broadcastType, broadcastID)
}

// QueueMessageToBroadcast queues a message to broadcast (backward compatibility)
func (m *UltraScaleBroadcastManager) QueueMessageToBroadcast(broadcastType, broadcastID string, msg *domainBroadcast.BroadcastMessage) error {
	// Set the appropriate ID based on type
	if broadcastType == "campaign" {
		if msg.CampaignID == nil {
			// Try to parse broadcastID as int
			var campaignID int
			if _, err := fmt.Sscanf(broadcastID, "%d", &campaignID); err == nil {
				msg.CampaignID = &campaignID
			}
		}
	} else if broadcastType == "sequence" {
		if msg.SequenceID == nil {
			msg.SequenceID = &broadcastID
		}
	}
	
	return m.QueueMessage(msg)
}

// BroadcastManagerInterface defines the interface for broadcast managers (backward compatibility)
type BroadcastManagerInterface interface {
	SendMessage(msg domainBroadcast.BroadcastMessage) error
	GetOrCreateWorker(deviceID string) interface{}
	GetWorkerStatus(deviceID string) (domainBroadcast.WorkerStatus, bool)
	GetAllWorkerStatus() []domainBroadcast.WorkerStatus
	StopAllWorkers() error
	StopWorker(deviceID string) error
	ResumeFailedWorkers() error
	CheckWorkerHealth()
}
// Implement BroadcastManagerInterface methods for backward compatibility

func (m *UltraScaleBroadcastManager) SendMessage(msg domainBroadcast.BroadcastMessage) error {
	return m.QueueMessage(&msg)
}

func (m *UltraScaleBroadcastManager) GetOrCreateWorker(deviceID string) interface{} {
	// Return nil as we manage workers internally per pool
	return nil
}

func (m *UltraScaleBroadcastManager) GetWorkerStatus(deviceID string) (domainBroadcast.WorkerStatus, bool) {
	// Return a generic status
	return domainBroadcast.WorkerStatus{
		DeviceID: deviceID,
		Status:   "managed",
	}, true
}

func (m *UltraScaleBroadcastManager) GetAllWorkerStatus() []domainBroadcast.WorkerStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	var statuses []domainBroadcast.WorkerStatus
	for _, pool := range m.pools {
		pool.mu.RLock()
		for deviceID := range pool.deviceGroups {
			statuses = append(statuses, domainBroadcast.WorkerStatus{
				DeviceID: deviceID,
				Status:   "active",
			})
		}
		pool.mu.RUnlock()
	}
	return statuses
}

func (m *UltraScaleBroadcastManager) StopAllWorkers() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, pool := range m.pools {
		pool.Shutdown()
	}
	return nil
}

func (m *UltraScaleBroadcastManager) StopWorker(deviceID string) error {
	// Workers are managed per pool, not globally
	return nil
}

func (m *UltraScaleBroadcastManager) ResumeFailedWorkers() error {
	// Workers auto-resume
	return nil
}

func (m *UltraScaleBroadcastManager) CheckWorkerHealth() {
	// Health is monitored internally
}
// GetPoolStatus returns the status of a specific pool
func (m *UltraScaleBroadcastManager) GetPoolStatus(poolKey string) (map[string]interface{}, error) {
	m.mu.RLock()
	pool, exists := m.pools[poolKey]
	m.mu.RUnlock()
	
	if !exists {
		return nil, fmt.Errorf("pool %s not found", poolKey)
	}
	
	pool.mu.RLock()
	defer pool.mu.RUnlock()
	
	status := map[string]interface{}{
		"pool_key":        poolKey,
		"broadcast_type":  pool.broadcastType,
		"broadcast_id":    pool.broadcastID,
		"total_messages":  atomic.LoadInt64(&pool.totalMessages),
		"processed":       atomic.LoadInt64(&pool.processedCount),
		"failed":          atomic.LoadInt64(&pool.failedCount),
		"device_count":    len(pool.deviceGroups),
		"worker_count":    len(pool.deviceGroups) * config.MaxWorkersPerDevice,
		"start_time":      pool.startTime,
		"completion_time": pool.completionTime,
		"devices":         make([]map[string]interface{}, 0),
	}
	
	// Add device details
	for deviceID, group := range pool.deviceGroups {
		deviceInfo := map[string]interface{}{
			"device_id":     deviceID,
			"worker_count":  len(group.workers),
			"queue_size":    len(group.messageQueue),
			"queue_capacity": cap(group.messageQueue),
		}
		status["devices"] = append(status["devices"].([]map[string]interface{}), deviceInfo)
	}
	
	return status, nil
}
