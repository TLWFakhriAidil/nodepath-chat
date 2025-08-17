package usecase

import (
	"fmt"
	"sync"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
	"context"
)

const (
	MAX_CONCURRENT_WORKERS = 500  // Limit concurrent workers
	WORKER_CHECK_INTERVAL  = 10 * time.Second
	MESSAGE_BATCH_SIZE     = 100
)

type OptimizedBroadcastProcessor struct {
	redisClient   *redis.Client
	broadcastRepo *repository.BroadcastRepository
	userRepo      *repository.UserRepository
	manager       broadcast.BroadcastManagerInterface
	activeWorkers sync.Map // deviceID -> bool
	workerPool    chan struct{} // Semaphore for limiting concurrent workers
}

// StartOptimizedBroadcastProcessor starts the Redis-based broadcast processor for 3000+ devices
func StartOptimizedBroadcastProcessor() {
	logrus.Info("Starting OPTIMIZED broadcast processor for 3000+ devices...")
	
	// Initialize Redis
	opt, err := redis.ParseURL(config.RedisURL)
	if err != nil {
		logrus.Errorf("Failed to parse Redis URL: %v", err)
		// Fall back to basic processor
		StartBroadcastWorkerProcessor()
		return
	}
	
	redisClient := redis.NewClient(opt)
	ctx := context.Background()
	
	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logrus.Errorf("Failed to connect to Redis: %v", err)
		// Fall back to basic processor
		StartBroadcastWorkerProcessor()
		return
	}
	
	logrus.Info("Connected to Redis for optimized broadcast processing")
	
	processor := &OptimizedBroadcastProcessor{
		redisClient:   redisClient,
		broadcastRepo: repository.GetBroadcastRepository(),
		userRepo:      repository.GetUserRepository(),
		manager:       broadcast.GetBroadcastManager(),
		workerPool:    make(chan struct{}, MAX_CONCURRENT_WORKERS),
	}
	
	// Initialize worker pool
	for i := 0; i < MAX_CONCURRENT_WORKERS; i++ {
		processor.workerPool <- struct{}{}
	}
	
	// Start the main processor
	go processor.run()
	
	// Start the queue monitor
	go processor.monitorQueues()
	
	// Start the worker health checker
	go processor.checkWorkerHealth()
}

func (p *OptimizedBroadcastProcessor) run() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			p.processNewMessages()
		}
	}
}
func (p *OptimizedBroadcastProcessor) processNewMessages() {
	// Get devices with pending messages from database
	devices, err := p.broadcastRepo.GetDevicesWithPendingMessages()
	if err != nil {
		logrus.Errorf("Failed to get devices with pending messages: %v", err)
		return
	}
	
	if len(devices) == 0 {
		return
	}
	
	// logrus.Debugf("Found %d devices with pending messages", len(devices))
	
	// Process each device in parallel
	var wg sync.WaitGroup
	for _, deviceID := range devices {
		// Check if worker already exists for this device
		if _, exists := p.activeWorkers.Load(deviceID); exists {
			continue // Worker already processing this device
		}
		
		// Get a slot from the worker pool
		select {
		case <-p.workerPool:
			wg.Add(1)
			go func(devID string) {
				defer wg.Done()
				defer func() {
					p.workerPool <- struct{}{} // Return slot to pool
					p.activeWorkers.Delete(devID)
				}()
				
				// Mark as active
				p.activeWorkers.Store(devID, true)
				
				// Process messages for this device
				p.processDeviceMessages(devID)
			}(deviceID)
		default:
			// Worker pool is full, skip this device for now
			logrus.Debugf("Worker pool full, skipping device %s", deviceID)
		}
	}
	
	// Wait for all workers to complete
	wg.Wait()
}

func (p *OptimizedBroadcastProcessor) processDeviceMessages(deviceID string) {
	// Check device status
	device, err := p.userRepo.GetDeviceByID(deviceID)
	if err != nil {
		logrus.Errorf("Failed to get device %s: %v", deviceID, err)
		// Skip messages for non-existent device
		p.skipDeviceMessages(deviceID, "Device not found")
		return
	}
	
	// Check if device is online
	if device.Status != "online" && device.Status != "Online" && 
	   device.Status != "connected" && device.Status != "Connected" {
		logrus.Warnf("Device %s is not online (status: %s), skipping messages", deviceID, device.Status)
		// Skip messages for offline device
		p.skipDeviceMessages(deviceID, "Device offline")
		return
	}
	
	// Check if WhatsApp client exists for this device
	clientManager := whatsapp.GetClientManager()
	_, err = clientManager.GetClient(deviceID)
	if err != nil {
		logrus.Warnf("WhatsApp client not found for device %s, skipping messages", deviceID)
		// Skip messages for device without WhatsApp client
		p.skipDeviceMessages(deviceID, "WhatsApp client not available")
		return
	}
	
	// Get pending messages from database
	messages, err := p.broadcastRepo.GetPendingMessagesAndLock(deviceID, MESSAGE_BATCH_SIZE)
	if err != nil {
		logrus.Errorf("Failed to get pending messages for device %s: %v", deviceID, err)
		return
	}
	
	if len(messages) == 0 {
		return
	}
	
	logrus.Infof("Found %d pending messages for device %s, sending to broadcast manager", len(messages), deviceID)
	
	// Send each message to the broadcast manager
	// The UltraScaleRedisManager will:
	// 1. Add to Redis queue
	// 2. Create/ensure worker exists
	// 3. Worker will process and send
	for _, msg := range messages {
		// Send to broadcast manager
		err := p.manager.SendMessage(msg)
		if err != nil {
			logrus.Errorf("Failed to queue message %s: %v", msg.ID, err)
			// Direct update to failed
			db := database.GetDB()
			db.Exec(`UPDATE broadcast_messages SET status = 'failed', error_message = ?, updated_at = NOW() WHERE id = ?`, err.Error(), msg.ID)
		} else {
			// Mark as queued - direct update like skipped
			db := database.GetDB()
			_, err := db.Exec(`UPDATE broadcast_messages SET status = 'queued', updated_at = NOW() WHERE id = ? AND status = 'pending'`, msg.ID)
			if err != nil {
				logrus.Errorf("Failed to update message %s to queued status: %v", msg.ID, err)
			} else {
				logrus.Infof("Message %s marked as queued", msg.ID)
			}
		}
	}
}

func (p *OptimizedBroadcastProcessor) calculateDelay(minDelay, maxDelay int) time.Duration {
	if minDelay <= 0 {
		minDelay = 5
	}
	if maxDelay <= 0 || maxDelay < minDelay {
		maxDelay = minDelay + 10
	}
	
	// Use average for now (could add randomization)
	avgDelay := (minDelay + maxDelay) / 2
	return time.Duration(avgDelay) * time.Second
}

func (p *OptimizedBroadcastProcessor) updateMetrics(deviceID string, success bool, duration time.Duration) {
	ctx := context.Background()
	metricsKey := fmt.Sprintf("broadcast:metrics:%s", deviceID)
	
	if success {
		p.redisClient.HIncrBy(ctx, metricsKey, "success_count", 1)
		p.redisClient.HSet(ctx, metricsKey, "last_success", time.Now().Unix())
	} else {
		p.redisClient.HIncrBy(ctx, metricsKey, "failed_count", 1)
		p.redisClient.HSet(ctx, metricsKey, "last_failure", time.Now().Unix())
	}
	
	// Update average send time
	p.redisClient.HSet(ctx, metricsKey, "last_send_duration_ms", duration.Milliseconds())
	
	// Set expiry on metrics
	p.redisClient.Expire(ctx, metricsKey, 24*time.Hour)
}

func (p *OptimizedBroadcastProcessor) monitorQueues() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	ctx := context.Background()
	
	for {
		select {
		case <-ticker.C:
			// Get all queue keys
			keys, err := p.redisClient.Keys(ctx, "broadcast:queue:*").Result()
			if err != nil {
				continue
			}
			
			totalQueued := 0
			for _, key := range keys {
				count, _ := p.redisClient.ZCard(ctx, key).Result()
				totalQueued += int(count)
			}
			
			if totalQueued > 0 {
				logrus.Infof("Total messages in queues: %d across %d devices", totalQueued, len(keys))
			}
			
			// Check active workers
			activeCount := 0
			p.activeWorkers.Range(func(key, value interface{}) bool {
				activeCount++
				return true
			})
			
			logrus.Infof("Active workers: %d/%d", activeCount, MAX_CONCURRENT_WORKERS)
		}
	}
}

func (p *OptimizedBroadcastProcessor) checkWorkerHealth() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	ctx := context.Background()
	
	for {
		select {
		case <-ticker.C:
			// Check for stuck messages in processing sets
			keys, err := p.redisClient.Keys(ctx, "broadcast:processing:*").Result()
			if err != nil {
				continue
			}
			
			for _, key := range keys {
				members, _ := p.redisClient.SMembers(ctx, key).Result()
				if len(members) > 0 {
					// Extract device ID from key
					deviceID := key[len("broadcast:processing:"):]
					
					// Check if worker is still active
					if _, exists := p.activeWorkers.Load(deviceID); !exists {
						// Worker died, move messages back to queue
						queueKey := fmt.Sprintf("broadcast:queue:%s", deviceID)
						for _, member := range members {
							p.redisClient.ZAdd(ctx, queueKey, &redis.Z{
								Score:  float64(time.Now().Unix()),
								Member: member,
							})
							p.redisClient.SRem(ctx, key, member)
						}
						logrus.Warnf("Recovered %d stuck messages for device %s", len(members), deviceID)
					}
				}
			}
		}
	}
}


// skipDeviceMessages marks all pending messages for a device as skipped
func (p *OptimizedBroadcastProcessor) skipDeviceMessages(deviceID string, reason string) {
	query := `
		UPDATE broadcast_messages SET status = 'skipped', error_message = ? 
		WHERE device_id = ? AND status = 'pending'
	`
	
	db := database.GetDB()
	result, err := db.Exec(query, reason, deviceID)
	if err != nil {
		logrus.Errorf("Failed to skip messages for device %s: %v", deviceID, err)
		return
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		logrus.Infof("Skipped %d messages for device %s (reason: %s)", rowsAffected, deviceID, reason)
	}
}
