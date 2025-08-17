package config

// Worker configuration optimized for high-volume messaging (5K per device)
const (
	// Worker Pool Settings - OPTIMIZED FOR 5K MESSAGES PER DEVICE
	MaxWorkersPerDevice   = 5      // Increased from 1 to handle parallel processing
	MaxConcurrentWorkers  = 2000   // Increased from 500 for better throughput
	WorkerQueueSize       = 10000  // Increased from 1000 to handle 5K+ messages
	WorkerHealthCheckSec  = 60     // Increased from 30 to reduce overhead
	WorkerIdleTimeoutMin  = 30     // Increased from 10 to keep workers active longer
	MessageQueueTimeout   = 30     // Timeout for queueing messages (seconds)
	
	// Message Processing - OPTIMIZED FOR VOLUME
	DefaultMinDelaySeconds = 5     // Min delay between messages
	DefaultMaxDelaySeconds = 15    // Max delay between messages
	BatchSize             = 500    // Increased from 100 for bulk processing
	RetryAttempts         = 3      // Retry failed messages
	RetryDelaySeconds     = 60     // Delay between retries
	
	// Campaign & Sequence Processing
	CampaignTriggerIntervalSec = 60  // Check for campaigns every minute
	SequenceTriggerIntervalSec = 300 // Process sequences every 5 minutes
	
	// Performance Tuning - INCREASED FOR HIGH VOLUME
	DatabaseMaxConnections = 500   // Increased from 200
	DatabaseMaxIdleConns   = 100   // Increased from 50
	DatabaseConnLifetime   = 3600  // Connection lifetime in seconds
	
	// Rate Limiting per Device - DEPRECATED
	// These are not used anymore - campaigns/sequences have their own min/max delays
	// MessagesPerMinute     = 3      // Not used
	// MessagesPerHour       = 80     // Not used  
	// MessagesPerDay        = 800    // Not used
	
	// System Limits - INCREASED FOR SCALE
	MaxDevicesPerUser     = 50     // Increased from 20
	MaxActiveUsers        = 500    // Increased from 250
	MaxTotalDevices       = 10000  // Increased from 5000
	
	// Memory Management
	GCPercent            = 50      // Garbage collection percentage
	MaxMemoryUsageGB     = 64      // Increased from 32
	WorkerMemoryLimitMB  = 200     // Increased from 100
	
	// Monitoring
	MetricsIntervalSec   = 10      // Metrics collection interval
	LogLevel             = "INFO"  // Log level
	EnableProfiling      = true    // Enable performance profiling
)

// GetWorkerConfig returns optimized worker configuration
func GetWorkerConfig() map[string]interface{} {
	return map[string]interface{}{
		"max_workers_per_device":   MaxWorkersPerDevice,
		"max_concurrent_workers":   MaxConcurrentWorkers,
		"worker_queue_size":        WorkerQueueSize,
		"worker_health_check_sec":  WorkerHealthCheckSec,
		"worker_idle_timeout_min":  WorkerIdleTimeoutMin,
		"default_min_delay_sec":    DefaultMinDelaySeconds,
		"default_max_delay_sec":    DefaultMaxDelaySeconds,
		"batch_size":               BatchSize,
		"retry_attempts":           RetryAttempts,
		"retry_delay_sec":          RetryDelaySeconds,
		// Rate limits removed - using campaign/sequence delays
		// "messages_per_minute":      MessagesPerMinute,
		// "messages_per_hour":        MessagesPerHour,
		// "messages_per_day":         MessagesPerDay,
	}
}

// CalculateOptimalWorkers calculates optimal worker count based on active devices
func CalculateOptimalWorkers(activeDevices int) int {
	// Formula: min(activeDevices, MaxConcurrentWorkers)
	if activeDevices > MaxConcurrentWorkers {
		return MaxConcurrentWorkers
	}
	return activeDevices
}

// GetDelayForDevice returns random delay between min and max for a device
func GetDelayForDevice(minDelay, maxDelay int) int {
	if minDelay >= maxDelay {
		return minDelay
	}
	return minDelay + (maxDelay-minDelay)/2 // Use middle value for consistency
}
