package broadcast

import (
	"os"
	"strings"
	"sync"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/sirupsen/logrus"
)

// BroadcastManagerInterface defines the interface for broadcast managers
type BroadcastManagerInterface interface {
	SendMessage(msg domainBroadcast.BroadcastMessage) error
	GetWorkerStatus(deviceID string) (domainBroadcast.WorkerStatus, bool)
	GetAllWorkerStatus() []domainBroadcast.WorkerStatus
	StopAllWorkers() error
	ResumeFailedWorkers() error
}

var (
	unifiedManager BroadcastManagerInterface
	umOnce        sync.Once
)

// GetBroadcastManager returns the appropriate broadcast manager based on configuration
func GetBroadcastManager() BroadcastManagerInterface {
	umOnce.Do(func() {
		// Initialize config first
		config.InitEnvironment()
		
		// Check if Redis is available
		redisURL := config.RedisURL
		if redisURL == "" {
			redisURL = os.Getenv("REDIS_URL")
		}
		
		// Log what we found
		logrus.Infof("Checking for Redis - URL: '%s'", redisURL)
		
		// Check if we have a valid Redis URL
		if redisURL != "" && 
		   !strings.Contains(redisURL, "${{") && 
		   !strings.Contains(redisURL, "localhost") && 
		   !strings.Contains(redisURL, "[::1]") &&
		   (strings.Contains(redisURL, "redis://") || strings.Contains(redisURL, "rediss://")) {
			logrus.Info("Valid Redis URL found, initializing Ultra Scale Redis broadcast manager for 3000+ devices")
			unifiedManager = NewUltraScaleRedisManager()
		} else {
			logrus.Info("No valid Redis URL found, using in-memory broadcast manager")
			unifiedManager = NewBasicBroadcastManager()
		}
	})
	return unifiedManager
}
