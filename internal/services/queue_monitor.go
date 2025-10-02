package services

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// QueueMonitor provides comprehensive monitoring and logging for queue processing bottlenecks
type QueueMonitor struct {
	mu                    sync.RWMutex
	processingTimes       []time.Duration
	queueSizes            map[string]int64
	workerUtilization     map[string]float64
	throughputCounter     int64
	errorCounter          int64
	lastResetTime         time.Time
	monitoringInterval    time.Duration
	performanceThresholds *PerformanceThresholds
	ctx                   context.Context
	cancel                context.CancelFunc
	isRunning             bool
}

// PerformanceThresholds defines performance warning and critical thresholds
type PerformanceThresholds struct {
	MaxProcessingTime      time.Duration // Warning if processing takes longer
	CriticalProcessingTime time.Duration // Critical if processing takes longer
	MaxQueueSize           int64         // Warning if queue size exceeds this
	CriticalQueueSize      int64         // Critical if queue size exceeds this
	MinThroughput          int64         // Warning if throughput drops below this per minute
	MaxErrorRate           float64       // Warning if error rate exceeds this percentage
	CriticalErrorRate      float64       // Critical if error rate exceeds this percentage
}

// QueueMetrics represents current queue performance metrics
type QueueMetrics struct {
	AverageProcessingTime time.Duration      `json:"average_processing_time"`
	MaxProcessingTime     time.Duration      `json:"max_processing_time"`
	MinProcessingTime     time.Duration      `json:"min_processing_time"`
	QueueSizes            map[string]int64   `json:"queue_sizes"`
	WorkerUtilization     map[string]float64 `json:"worker_utilization"`
	ThroughputPerMinute   int64              `json:"throughput_per_minute"`
	ErrorRate             float64            `json:"error_rate"`
	TotalProcessed        int64              `json:"total_processed"`
	TotalErrors           int64              `json:"total_errors"`
	Uptime                time.Duration      `json:"uptime"`
	Bottlenecks           []string           `json:"bottlenecks"`
	HealthStatus          string             `json:"health_status"`
}

// WorkerPoolStats represents statistics for a worker pool
type WorkerPoolStats struct {
	PoolName       string        `json:"pool_name"`
	ActiveWorkers  int           `json:"active_workers"`
	IdleWorkers    int           `json:"idle_workers"`
	TotalWorkers   int           `json:"total_workers"`
	QueuedJobs     int64         `json:"queued_jobs"`
	ProcessedJobs  int64         `json:"processed_jobs"`
	FailedJobs     int64         `json:"failed_jobs"`
	AverageJobTime time.Duration `json:"average_job_time"`
	Utilization    float64       `json:"utilization"`
}

// NewQueueMonitor creates a new queue monitor with default thresholds
func NewQueueMonitor() *QueueMonitor {
	ctx, cancel := context.WithCancel(context.Background())

	return &QueueMonitor{
		processingTimes:    make([]time.Duration, 0, 1000), // Keep last 1000 processing times
		queueSizes:         make(map[string]int64),
		workerUtilization:  make(map[string]float64),
		lastResetTime:      time.Now(),
		monitoringInterval: 30 * time.Second, // Monitor every 30 seconds
		performanceThresholds: &PerformanceThresholds{
			MaxProcessingTime:      5 * time.Second,  // Warning if processing > 5s
			CriticalProcessingTime: 15 * time.Second, // Critical if processing > 15s
			MaxQueueSize:           1000,             // Warning if queue > 1000 items
			CriticalQueueSize:      5000,             // Critical if queue > 5000 items
			MinThroughput:          100,              // Warning if < 100 items/minute
			MaxErrorRate:           5.0,              // Warning if error rate > 5%
			CriticalErrorRate:      15.0,             // Critical if error rate > 15%
		},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start begins monitoring queue performance
func (qm *QueueMonitor) Start() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if qm.isRunning {
		return
	}

	qm.isRunning = true
	qm.lastResetTime = time.Now()

	go qm.monitoringLoop()

	logrus.WithField("monitoring_interval", qm.monitoringInterval).Info("Queue monitor started")
}

// Stop stops the queue monitoring
func (qm *QueueMonitor) Stop() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	if !qm.isRunning {
		return
	}

	qm.cancel()
	qm.isRunning = false

	logrus.Info("Queue monitor stopped")
}

// RecordProcessingTime records the time taken to process a queue item
func (qm *QueueMonitor) RecordProcessingTime(duration time.Duration) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	// Keep only the last 1000 processing times to prevent memory growth
	if len(qm.processingTimes) >= 1000 {
		qm.processingTimes = qm.processingTimes[1:]
	}

	qm.processingTimes = append(qm.processingTimes, duration)
	qm.throughputCounter++

	// Log slow processing times
	if duration > qm.performanceThresholds.MaxProcessingTime {
		level := logrus.WarnLevel
		if duration > qm.performanceThresholds.CriticalProcessingTime {
			level = logrus.ErrorLevel
		}

		logrus.WithFields(logrus.Fields{
			"processing_time":    duration,
			"threshold_max":      qm.performanceThresholds.MaxProcessingTime,
			"threshold_critical": qm.performanceThresholds.CriticalProcessingTime,
		}).Log(level, "Slow queue processing detected")
	}
}

// RecordQueueSize records the current size of a specific queue
func (qm *QueueMonitor) RecordQueueSize(queueName string, size int64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.queueSizes[queueName] = size

	// Log large queue sizes
	if size > qm.performanceThresholds.MaxQueueSize {
		level := logrus.WarnLevel
		if size > qm.performanceThresholds.CriticalQueueSize {
			level = logrus.ErrorLevel
		}

		logrus.WithFields(logrus.Fields{
			"queue_name":         queueName,
			"queue_size":         size,
			"threshold_max":      qm.performanceThresholds.MaxQueueSize,
			"threshold_critical": qm.performanceThresholds.CriticalQueueSize,
		}).Log(level, "Large queue size detected - potential bottleneck")
	}
}

// RecordWorkerUtilization records the utilization percentage of a worker pool
func (qm *QueueMonitor) RecordWorkerUtilization(poolName string, utilization float64) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.workerUtilization[poolName] = utilization

	// Log high utilization
	if utilization > 90.0 {
		logrus.WithFields(logrus.Fields{
			"pool_name":   poolName,
			"utilization": utilization,
		}).Warn("High worker pool utilization - consider scaling")
	}
}

// RecordError records a processing error
func (qm *QueueMonitor) RecordError() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.errorCounter++
}

// GetMetrics returns current queue performance metrics
func (qm *QueueMonitor) GetMetrics() *QueueMetrics {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	metrics := &QueueMetrics{
		QueueSizes:        make(map[string]int64),
		WorkerUtilization: make(map[string]float64),
		TotalProcessed:    qm.throughputCounter,
		TotalErrors:       qm.errorCounter,
		Uptime:            time.Since(qm.lastResetTime),
		Bottlenecks:       make([]string, 0),
	}

	// Copy maps to avoid race conditions
	for k, v := range qm.queueSizes {
		metrics.QueueSizes[k] = v
	}
	for k, v := range qm.workerUtilization {
		metrics.WorkerUtilization[k] = v
	}

	// Calculate processing time statistics
	if len(qm.processingTimes) > 0 {
		total := time.Duration(0)
		metrics.MaxProcessingTime = qm.processingTimes[0]
		metrics.MinProcessingTime = qm.processingTimes[0]

		for _, pt := range qm.processingTimes {
			total += pt
			if pt > metrics.MaxProcessingTime {
				metrics.MaxProcessingTime = pt
			}
			if pt < metrics.MinProcessingTime {
				metrics.MinProcessingTime = pt
			}
		}

		metrics.AverageProcessingTime = total / time.Duration(len(qm.processingTimes))
	}

	// Calculate throughput per minute
	uptime := time.Since(qm.lastResetTime)
	if uptime > 0 {
		metrics.ThroughputPerMinute = int64(float64(qm.throughputCounter) / uptime.Minutes())
	}

	// Calculate error rate
	if qm.throughputCounter > 0 {
		metrics.ErrorRate = float64(qm.errorCounter) / float64(qm.throughputCounter) * 100
	}

	// Identify bottlenecks
	metrics.Bottlenecks = qm.identifyBottlenecks(metrics)

	// Determine health status
	metrics.HealthStatus = qm.determineHealthStatus(metrics)

	return metrics
}

// monitoringLoop runs the periodic monitoring and logging
func (qm *QueueMonitor) monitoringLoop() {
	ticker := time.NewTicker(qm.monitoringInterval)
	defer ticker.Stop()

	for {
		select {
		case <-qm.ctx.Done():
			return
		case <-ticker.C:
			qm.logPerformanceMetrics()
		}
	}
}

// logPerformanceMetrics logs current performance metrics
func (qm *QueueMonitor) logPerformanceMetrics() {
	metrics := qm.GetMetrics()

	logFields := logrus.Fields{
		"avg_processing_time": metrics.AverageProcessingTime,
		"max_processing_time": metrics.MaxProcessingTime,
		"throughput_per_min":  metrics.ThroughputPerMinute,
		"error_rate":          metrics.ErrorRate,
		"total_processed":     metrics.TotalProcessed,
		"total_errors":        metrics.TotalErrors,
		"health_status":       metrics.HealthStatus,
		"uptime":              metrics.Uptime,
	}

	// Add queue sizes to log
	for queueName, size := range metrics.QueueSizes {
		logFields["queue_size_"+queueName] = size
	}

	// Add worker utilization to log
	for poolName, util := range metrics.WorkerUtilization {
		logFields["worker_util_"+poolName] = util
	}

	// Add bottlenecks if any
	if len(metrics.Bottlenecks) > 0 {
		logFields["bottlenecks"] = metrics.Bottlenecks
	}

	// Choose log level based on health status
	var logLevel logrus.Level
	switch metrics.HealthStatus {
	case "critical":
		logLevel = logrus.ErrorLevel
	case "warning":
		logLevel = logrus.WarnLevel
	default:
		logLevel = logrus.InfoLevel
	}

	logrus.WithFields(logFields).Log(logLevel, "Queue performance metrics")
}

// identifyBottlenecks analyzes metrics to identify potential bottlenecks
func (qm *QueueMonitor) identifyBottlenecks(metrics *QueueMetrics) []string {
	bottlenecks := make([]string, 0)

	// Check processing time bottlenecks
	if metrics.AverageProcessingTime > qm.performanceThresholds.MaxProcessingTime {
		bottlenecks = append(bottlenecks, "slow_processing")
	}

	// Check queue size bottlenecks
	for queueName, size := range metrics.QueueSizes {
		if size > qm.performanceThresholds.MaxQueueSize {
			bottlenecks = append(bottlenecks, "large_queue_"+queueName)
		}
	}

	// Check throughput bottlenecks
	if metrics.ThroughputPerMinute < qm.performanceThresholds.MinThroughput {
		bottlenecks = append(bottlenecks, "low_throughput")
	}

	// Check error rate bottlenecks
	if metrics.ErrorRate > qm.performanceThresholds.MaxErrorRate {
		bottlenecks = append(bottlenecks, "high_error_rate")
	}

	// Check worker utilization bottlenecks
	for poolName, util := range metrics.WorkerUtilization {
		if util > 95.0 {
			bottlenecks = append(bottlenecks, "overloaded_workers_"+poolName)
		}
	}

	return bottlenecks
}

// determineHealthStatus determines overall system health based on metrics
func (qm *QueueMonitor) determineHealthStatus(metrics *QueueMetrics) string {
	// Critical conditions
	if metrics.AverageProcessingTime > qm.performanceThresholds.CriticalProcessingTime ||
		metrics.ErrorRate > qm.performanceThresholds.CriticalErrorRate {
		return "critical"
	}

	for _, size := range metrics.QueueSizes {
		if size > qm.performanceThresholds.CriticalQueueSize {
			return "critical"
		}
	}

	// Warning conditions
	if len(metrics.Bottlenecks) > 0 {
		return "warning"
	}

	return "healthy"
}

// ResetMetrics resets all counters and metrics
func (qm *QueueMonitor) ResetMetrics() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.processingTimes = qm.processingTimes[:0]
	qm.throughputCounter = 0
	qm.errorCounter = 0
	qm.lastResetTime = time.Now()

	logrus.Info("Queue monitor metrics reset")
}
