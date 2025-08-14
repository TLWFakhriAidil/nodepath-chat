package services

import (
	"sync"
	"sync/atomic"
	"time"

	"nodepath-chat/internal/repository"

	"github.com/sirupsen/logrus"
)

// MetricsService provides performance monitoring and statistics
type MetricsService struct {
	mu                    sync.RWMutex
	messagesProcessed     int64
	messagesFailed        int64
	messagesQueued        int64
	processingTimes       []time.Duration
	deviceMetrics         map[string]*DeviceMetrics
	startTime             time.Time
	lastResetTime         time.Time
	repo                  *repository.BroadcastRepository
	maxProcessingTimes    int // Limit stored processing times
}

// DeviceMetrics tracks metrics for individual devices
type DeviceMetrics struct {
	DeviceID          string
	MessagesProcessed int64
	MessagesFailed    int64
	LastActivity      time.Time
	AverageDelay      time.Duration
	Status            string
}

// BroadcastMetrics represents overall system metrics
type BroadcastMetrics struct {
	TotalProcessed       int64             `json:"total_processed"`
	TotalFailed          int64             `json:"total_failed"`
	TotalQueued          int64             `json:"total_queued"`
	MessagesPerSecond    float64           `json:"messages_per_second"`
	AverageProcessingTime time.Duration     `json:"average_processing_time"`
	Uptime               time.Duration     `json:"uptime"`
	ActiveDevices        int               `json:"active_devices"`
	DeviceMetrics        []*DeviceMetrics  `json:"device_metrics"`
	QueueStats           map[string]int    `json:"queue_stats"`
	LastUpdated          time.Time         `json:"last_updated"`
}

// NewMetricsService creates a new metrics service
func NewMetricsService() *MetricsService {
	now := time.Now()
	return &MetricsService{
		deviceMetrics:      make(map[string]*DeviceMetrics),
		startTime:          now,
		lastResetTime:      now,
		repo:               repository.GetBroadcastRepository(),
		maxProcessingTimes: 1000, // Keep last 1000 processing times
	}
}

// RecordMessageProcessed increments the processed message counter
func (ms *MetricsService) RecordMessageProcessed(deviceID string, processingTime time.Duration) {
	atomic.AddInt64(&ms.messagesProcessed, 1)
	
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	// Record processing time
	ms.processingTimes = append(ms.processingTimes, processingTime)
	if len(ms.processingTimes) > ms.maxProcessingTimes {
		// Remove oldest entries to prevent memory bloat
		ms.processingTimes = ms.processingTimes[len(ms.processingTimes)-ms.maxProcessingTimes:]
	}
	
	// Update device metrics
	if deviceMetric, exists := ms.deviceMetrics[deviceID]; exists {
		atomic.AddInt64(&deviceMetric.MessagesProcessed, 1)
		deviceMetric.LastActivity = time.Now()
		deviceMetric.Status = "active"
	} else {
		ms.deviceMetrics[deviceID] = &DeviceMetrics{
			DeviceID:          deviceID,
			MessagesProcessed: 1,
			LastActivity:      time.Now(),
			Status:            "active",
		}
	}
}

// RecordMessageFailed increments the failed message counter
func (ms *MetricsService) RecordMessageFailed(deviceID string, reason string) {
	atomic.AddInt64(&ms.messagesFailed, 1)
	
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	// Update device metrics
	if deviceMetric, exists := ms.deviceMetrics[deviceID]; exists {
		atomic.AddInt64(&deviceMetric.MessagesFailed, 1)
		deviceMetric.LastActivity = time.Now()
		deviceMetric.Status = "error"
	} else {
		ms.deviceMetrics[deviceID] = &DeviceMetrics{
			DeviceID:       deviceID,
			MessagesFailed: 1,
			LastActivity:   time.Now(),
			Status:         "error",
		}
	}
	
	logrus.Warnf("Message failed on device %s: %s", deviceID, reason)
}

// RecordMessageQueued increments the queued message counter
func (ms *MetricsService) RecordMessageQueued(deviceID string) {
	atomic.AddInt64(&ms.messagesQueued, 1)
	
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	// Update device metrics
	if deviceMetric, exists := ms.deviceMetrics[deviceID]; exists {
		deviceMetric.LastActivity = time.Now()
		if deviceMetric.Status == "idle" {
			deviceMetric.Status = "queued"
		}
	} else {
		ms.deviceMetrics[deviceID] = &DeviceMetrics{
			DeviceID:     deviceID,
			LastActivity: time.Now(),
			Status:       "queued",
		}
	}
}

// GetMetrics returns current system metrics
func (ms *MetricsService) GetMetrics() *BroadcastMetrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	now := time.Now()
	uptime := now.Sub(ms.startTime)
	timeSinceReset := now.Sub(ms.lastResetTime)
	
	// Calculate messages per second
	var messagesPerSecond float64
	if timeSinceReset.Seconds() > 0 {
		messagesPerSecond = float64(atomic.LoadInt64(&ms.messagesProcessed)) / timeSinceReset.Seconds()
	}
	
	// Calculate average processing time
	var avgProcessingTime time.Duration
	if len(ms.processingTimes) > 0 {
		var total time.Duration
		for _, pt := range ms.processingTimes {
			total += pt
		}
		avgProcessingTime = total / time.Duration(len(ms.processingTimes))
	}
	
	// Count active devices (active in last 5 minutes)
	activeThreshold := now.Add(-5 * time.Minute)
	activeDevices := 0
	deviceMetricsList := make([]*DeviceMetrics, 0, len(ms.deviceMetrics))
	
	for _, dm := range ms.deviceMetrics {
		if dm.LastActivity.After(activeThreshold) {
			activeDevices++
		}
		deviceMetricsList = append(deviceMetricsList, dm)
	}
	
	// Get queue stats from database
	queueStats := ms.getQueueStatsFromDB()
	
	return &BroadcastMetrics{
		TotalProcessed:        atomic.LoadInt64(&ms.messagesProcessed),
		TotalFailed:           atomic.LoadInt64(&ms.messagesFailed),
		TotalQueued:           atomic.LoadInt64(&ms.messagesQueued),
		MessagesPerSecond:     messagesPerSecond,
		AverageProcessingTime: avgProcessingTime,
		Uptime:                uptime,
		ActiveDevices:         activeDevices,
		DeviceMetrics:         deviceMetricsList,
		QueueStats:            queueStats,
		LastUpdated:           now,
	}
}

// getQueueStatsFromDB retrieves current queue statistics from database
func (ms *MetricsService) getQueueStatsFromDB() map[string]int {
	stats := make(map[string]int)
	
	query := `
		SELECT status, COUNT(*) as count 
		FROM broadcast_messages 
		WHERE created_at > DATE_SUB(NOW(), INTERVAL 24 HOUR)
		GROUP BY status
	`
	
	rows, err := ms.repo.DB().Query(query)
	if err != nil {
		logrus.Errorf("Failed to get queue stats: %v", err)
		return stats
	}
	defer rows.Close()
	
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			logrus.Errorf("Failed to scan queue stats: %v", err)
			continue
		}
		stats[status] = count
	}
	
	return stats
}

// GetDeviceMetrics returns metrics for a specific device
func (ms *MetricsService) GetDeviceMetrics(deviceID string) *DeviceMetrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if metric, exists := ms.deviceMetrics[deviceID]; exists {
		return metric
	}
	return nil
}

// GetTopPerformingDevices returns the top N performing devices by message count
func (ms *MetricsService) GetTopPerformingDevices(limit int) []*DeviceMetrics {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	devices := make([]*DeviceMetrics, 0, len(ms.deviceMetrics))
	for _, dm := range ms.deviceMetrics {
		devices = append(devices, dm)
	}
	
	// Simple bubble sort for top performers (good enough for small datasets)
	for i := 0; i < len(devices)-1; i++ {
		for j := 0; j < len(devices)-i-1; j++ {
			if devices[j].MessagesProcessed < devices[j+1].MessagesProcessed {
				devices[j], devices[j+1] = devices[j+1], devices[j]
			}
		}
	}
	
	if limit > 0 && limit < len(devices) {
		return devices[:limit]
	}
	return devices
}

// ResetMetrics resets all counters (useful for testing or periodic resets)
func (ms *MetricsService) ResetMetrics() {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	atomic.StoreInt64(&ms.messagesProcessed, 0)
	atomic.StoreInt64(&ms.messagesFailed, 0)
	atomic.StoreInt64(&ms.messagesQueued, 0)
	ms.processingTimes = nil
	ms.deviceMetrics = make(map[string]*DeviceMetrics)
	ms.lastResetTime = time.Now()
	
	logrus.Info("Metrics reset")
}

// LogPerformanceReport logs a detailed performance report
func (ms *MetricsService) LogPerformanceReport() {
	metrics := ms.GetMetrics()
	
	logrus.WithFields(logrus.Fields{
		"total_processed":         metrics.TotalProcessed,
		"total_failed":            metrics.TotalFailed,
		"total_queued":            metrics.TotalQueued,
		"messages_per_second":     metrics.MessagesPerSecond,
		"average_processing_time": metrics.AverageProcessingTime,
		"uptime":                  metrics.Uptime,
		"active_devices":          metrics.ActiveDevices,
	}).Info("Broadcast system performance report")
	
	// Log top performing devices
	topDevices := ms.GetTopPerformingDevices(5)
	for i, device := range topDevices {
		logrus.WithFields(logrus.Fields{
			"rank":               i + 1,
			"device_id":          device.DeviceID,
			"messages_processed": device.MessagesProcessed,
			"messages_failed":    device.MessagesFailed,
			"status":             device.Status,
		}).Info("Top performing device")
	}
}

// StartPeriodicReporting starts a goroutine that logs performance reports periodically
func (ms *MetricsService) StartPeriodicReporting(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		
		for range ticker.C {
			ms.LogPerformanceReport()
		}
	}()
}

// GetHealthStatus returns the overall health status of the broadcast system
func (ms *MetricsService) GetHealthStatus() string {
	metrics := ms.GetMetrics()
	
	// Calculate failure rate
	totalMessages := metrics.TotalProcessed + metrics.TotalFailed
	var failureRate float64
	if totalMessages > 0 {
		failureRate = float64(metrics.TotalFailed) / float64(totalMessages)
	}
	
	// Determine health based on various factors
	if metrics.ActiveDevices == 0 {
		return "critical" // No active devices
	}
	
	if failureRate > 0.1 { // More than 10% failure rate
		return "warning"
	}
	
	if metrics.MessagesPerSecond < 0.1 && metrics.TotalQueued > 100 {
		return "warning" // Low throughput with high queue
	}
	
	return "healthy"
}