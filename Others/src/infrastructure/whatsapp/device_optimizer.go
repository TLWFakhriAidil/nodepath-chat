package whatsapp

import (
	"sync"
	"time"
	"github.com/sirupsen/logrus"
)

// DeviceConnectionOptimizer optimizes connections for 3000+ devices
type DeviceConnectionOptimizer struct {
	mu                    sync.RWMutex
	maxConcurrentDevices  int
	connectionCooldown    time.Duration
	lastConnectionAttempt map[string]time.Time
}

var (
	optimizer     *DeviceConnectionOptimizer
	optimizerOnce sync.Once
)

// GetDeviceOptimizer returns singleton optimizer
func GetDeviceOptimizer() *DeviceConnectionOptimizer {
	optimizerOnce.Do(func() {
		optimizer = &DeviceConnectionOptimizer{
			maxConcurrentDevices:  1500, // Limit for stability
			connectionCooldown:    5 * time.Minute,
			lastConnectionAttempt: make(map[string]time.Time),
		}
	})
	return optimizer
}

// CanConnect checks if device can attempt connection
func (do *DeviceConnectionOptimizer) CanConnect(deviceID string) bool {
	do.mu.Lock()
	defer do.mu.Unlock()
	
	// Check cooldown
	if lastAttempt, exists := do.lastConnectionAttempt[deviceID]; exists {
		if time.Since(lastAttempt) < do.connectionCooldown {
			return false
		}
	}
	
	// Check concurrent limit
	cm := GetClientManager()
	activeCount := len(cm.GetAllClients())
	if activeCount >= do.maxConcurrentDevices {
		logrus.Warnf("Connection limit reached (%d/%d)", activeCount, do.maxConcurrentDevices)
		return false
	}
	
	// Record attempt
	do.lastConnectionAttempt[deviceID] = time.Now()
	return true
}

// OptimizeForScale adjusts settings for large scale
func (do *DeviceConnectionOptimizer) OptimizeForScale(deviceCount int) {
	do.mu.Lock()
	defer do.mu.Unlock()
	
	if deviceCount > 2000 {
		do.maxConcurrentDevices = 1000
		do.connectionCooldown = 10 * time.Minute
		logrus.Info("Optimized for 2000+ devices: stricter limits")
	} else if deviceCount > 1000 {
		do.maxConcurrentDevices = 1500
		do.connectionCooldown = 5 * time.Minute
		logrus.Info("Optimized for 1000+ devices: moderate limits")
	}
}

// GetRecommendedBatchSize returns optimal batch size for operations
func (do *DeviceConnectionOptimizer) GetRecommendedBatchSize() int {
	activeCount := len(GetClientManager().GetAllClients())
	
	switch {
	case activeCount > 2000:
		return 50
	case activeCount > 1000:
		return 100
	case activeCount > 500:
		return 200
	default:
		return 300
	}
}

// ShouldThrottle checks if system should slow down
func (do *DeviceConnectionOptimizer) ShouldThrottle() bool {
	cm := GetClientManager()
	activeCount := len(cm.GetAllClients())
	
	// Throttle if approaching limit
	return float64(activeCount) > float64(do.maxConcurrentDevices)*0.9
}
