package broadcast

import (
	"context"
	"fmt"
	"sync"
	"time"
	
	"github.com/sirupsen/logrus"
)

// DeviceCleanupManager handles cleanup of non-existent devices
type DeviceCleanupManager struct {
	cleanedDevices map[string]time.Time
	mu             sync.RWMutex
}

var (
	cleanupManager     *DeviceCleanupManager
	cleanupManagerOnce sync.Once
)

// GetCleanupManager returns singleton cleanup manager
func GetCleanupManager() *DeviceCleanupManager {
	cleanupManagerOnce.Do(func() {
		cleanupManager = &DeviceCleanupManager{
			cleanedDevices: make(map[string]time.Time),
		}
		// Start cleanup routine
		go cleanupManager.periodicCleanup()
	})
	return cleanupManager
}

// IsDeviceCleaned checks if device was already cleaned up recently
func (dcm *DeviceCleanupManager) IsDeviceCleaned(deviceID string) bool {
	dcm.mu.RLock()
	defer dcm.mu.RUnlock()
	
	cleanedAt, exists := dcm.cleanedDevices[deviceID]
	if !exists {
		return false
	}
	
	// Consider cleaned if cleaned within last hour
	return time.Since(cleanedAt) < time.Hour
}

// MarkDeviceCleaned marks device as cleaned
func (dcm *DeviceCleanupManager) MarkDeviceCleaned(deviceID string) {
	dcm.mu.Lock()
	defer dcm.mu.Unlock()
	dcm.cleanedDevices[deviceID] = time.Now()
}

// periodicCleanup removes old entries from cleaned devices map
func (dcm *DeviceCleanupManager) periodicCleanup() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		dcm.mu.Lock()
		for deviceID, cleanedAt := range dcm.cleanedDevices {
			if time.Since(cleanedAt) > 2*time.Hour {
				delete(dcm.cleanedDevices, deviceID)
			}
		}
		dcm.mu.Unlock()
	}
}

// CleanupNonExistentDevice removes all traces of a non-existent device
func (um *UltraScaleRedisManager) CleanupNonExistentDevice(deviceID string) {
	// Check if already cleaned recently to avoid spam
	cleanupMgr := GetCleanupManager()
	if cleanupMgr.IsDeviceCleaned(deviceID) {
		return // Already cleaned, skip
	}
	
	logrus.Infof("Cleaning up non-existent device %s from Redis", deviceID)
	
	ctx := context.Background()
	
	// Remove all queues
	queues := []string{
		fmt.Sprintf("%s%s", ultraCampaignQueuePrefix, deviceID),
		fmt.Sprintf("%s%s", ultraSequenceQueuePrefix, deviceID),
		fmt.Sprintf("%s%s", ultraDeadLetterPrefix, deviceID),
	}
	
	for _, queue := range queues {
		count, _ := um.redisClient.LLen(ctx, queue).Result()
		if count > 0 {
			if err := um.redisClient.Del(ctx, queue).Err(); err != nil {
				logrus.Errorf("Failed to delete queue %s: %v", queue, err)
			} else {
				logrus.Infof("Deleted queue %s with %d messages", queue, count)
			}
		}
	}
	
	// Remove worker lock
	lockKey := fmt.Sprintf("%s%s", ultraWorkerLockPrefix, deviceID)
	um.redisClient.Del(ctx, lockKey)
	
	// Remove from worker status
	um.redisClient.HDel(ctx, ultraWorkerStatusKey, deviceID)
	
	// Remove metrics
	metricsKey := fmt.Sprintf("%s%s", ultraMetricsPrefix, deviceID)
	um.redisClient.Del(ctx, metricsKey)
	
	// Remove rate limit
	rateLimitKey := fmt.Sprintf("%s%s", ultraRateLimitPrefix, deviceID)
	um.redisClient.Del(ctx, rateLimitKey)
	
	// Remove worker if exists
	um.workersMutex.Lock()
	if worker, exists := um.workers[deviceID]; exists {
		if worker != nil {
			worker.Stop()
		}
		delete(um.workers, deviceID)
		logrus.Infof("Removed worker for non-existent device %s", deviceID)
	}
	um.workersMutex.Unlock()
	
	// Mark as cleaned
	cleanupMgr.MarkDeviceCleaned(deviceID)
	
	logrus.Infof("Completed cleanup for device %s", deviceID)
}
