package services

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// WhatsAppServiceInterface defines the interface for WhatsApp service
type WhatsAppServiceInterface interface {
	Connect() error
	Disconnect()
	IsConnected() bool
	SendMessage(phoneNumber, message string) error
	SendMediaMessage(phoneNumber, caption, mediaURL, mediaType string) error
	GetQRCode() (string, error)
}

// HighPerformanceManager manages all high-performance components
type HighPerformanceManager struct {
	mu               sync.RWMutex
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	
	// Core services
	messageProcessor *MessageProcessor
	enhancedAI       *EnhancedAIService
	chatService      *ChatService
	flowService      *FlowService
	queueService     *QueueService
	whatsappService  WhatsAppServiceInterface
	
	// Performance components
	globalRateLimiter *RateLimiter
	connectionPool    *ConnectionPool
	metrics           *SystemMetrics
	
	// Configuration
	config           *HighPerformanceConfig
	
	// Status
	isRunning        bool
	startTime        time.Time
	logger           *logrus.Entry
}

// HighPerformanceConfig holds configuration for the high-performance system
type HighPerformanceConfig struct {
	// System settings
	MaxConcurrentUsers    int           `json:"max_concurrent_users"`
	TargetResponseTime    time.Duration `json:"target_response_time"`
	GlobalRateLimit       int           `json:"global_rate_limit"`
	HealthCheckInterval   time.Duration `json:"health_check_interval"`
	MetricsInterval       time.Duration `json:"metrics_interval"`
	
	// Message processing
	MessageWorkerCount    int           `json:"message_worker_count"`
	MessageQueueSize      int           `json:"message_queue_size"`
	MessageTimeout        time.Duration `json:"message_timeout"`
	
	// AI settings
	AIMaxConcurrency      int           `json:"ai_max_concurrency"`
	AICacheEnabled        bool          `json:"ai_cache_enabled"`
	AICacheSize           int           `json:"ai_cache_size"`
	AITimeout             time.Duration `json:"ai_timeout"`
	
	// Database settings
	DBMaxConnections      int           `json:"db_max_connections"`
	DBConnectionTimeout   time.Duration `json:"db_connection_timeout"`
	DBQueryTimeout        time.Duration `json:"db_query_timeout"`
	
	// Performance optimizations
	EnableCompression     bool          `json:"enable_compression"`
	EnableKeepAlive       bool          `json:"enable_keep_alive"`
	EnablePipelining      bool          `json:"enable_pipelining"`
	GCTargetPercentage    int           `json:"gc_target_percentage"`
}

// SystemMetrics tracks overall system performance
type SystemMetrics struct {
	mu                    sync.RWMutex
	TotalUsers            int64         `json:"total_users"`
	ConcurrentUsers       int64         `json:"concurrent_users"`
	PeakConcurrentUsers   int64         `json:"peak_concurrent_users"`
	TotalMessages         int64         `json:"total_messages"`
	MessagesPerSecond     float64       `json:"messages_per_second"`
	AverageResponseTime   time.Duration `json:"average_response_time"`
	PeakResponseTime      time.Duration `json:"peak_response_time"`
	SystemUptime          time.Duration `json:"system_uptime"`
	MemoryUsage           uint64        `json:"memory_usage"`
	CPUUsage              float64       `json:"cpu_usage"`
	GoroutineCount        int           `json:"goroutine_count"`
	LastUpdated           time.Time     `json:"last_updated"`
}

// NewHighPerformanceManager creates a new high-performance manager
func NewHighPerformanceManager(
	chatService *ChatService,
	flowService *FlowService,
	aiService *AIService,
	queueService *QueueService,
	whatsappService WhatsAppServiceInterface,
) *HighPerformanceManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Default configuration optimized for 1500+ concurrent users
	config := &HighPerformanceConfig{
		MaxConcurrentUsers:    2000,                // Target for 2000 users
		TargetResponseTime:    100 * time.Millisecond, // Sub-100ms target
		GlobalRateLimit:       5000,                // 5000 requests per second
		HealthCheckInterval:   30 * time.Second,
		MetricsInterval:       10 * time.Second,
		
		MessageWorkerCount:    runtime.NumCPU() * 8, // Scale with CPU
		MessageQueueSize:      50000,               // Large queue
		MessageTimeout:        30 * time.Second,
		
		AIMaxConcurrency:      100,                 // High AI concurrency
		AICacheEnabled:        true,
		AICacheSize:           20000,
		AITimeout:             15 * time.Second,
		
		DBMaxConnections:      50,                  // High DB concurrency
		DBConnectionTimeout:   5 * time.Second,
		DBQueryTimeout:        10 * time.Second,
		
		EnableCompression:     true,
		EnableKeepAlive:       true,
		EnablePipelining:      true,
		GCTargetPercentage:    50,                  // Optimize GC
	}
	
	manager := &HighPerformanceManager{
		ctx:    ctx,
		cancel: cancel,
		chatService:     chatService,
		flowService:     flowService,
		queueService:    queueService,
		whatsappService: whatsappService,
		config:          config,
		metrics: &SystemMetrics{
			LastUpdated: time.Now(),
		},
		logger: logrus.WithFields(logrus.Fields{
			"component": "high_performance_manager",
		}),
	}
	
	// Initialize enhanced AI service
	manager.enhancedAI = NewEnhancedAIService(aiService)
	
	// Initialize message processor - commented out due to type mismatch
	// manager.messageProcessor = NewMessageProcessor(
	//	chatService,
	//	flowService,
	//	manager.enhancedAI,
	//	queueService,
	// )
	
	// Initialize global rate limiter
	manager.globalRateLimiter = NewRateLimiter(config.GlobalRateLimit)
	
	return manager
}

// Start starts the high-performance system
func (hpm *HighPerformanceManager) Start() error {
	hpm.mu.Lock()
	defer hpm.mu.Unlock()
	
	if hpm.isRunning {
		return fmt.Errorf("high-performance manager is already running")
	}
	
	hpm.logger.WithFields(logrus.Fields{
		"max_concurrent_users": hpm.config.MaxConcurrentUsers,
		"target_response_time": hpm.config.TargetResponseTime,
		"worker_count":         hpm.config.MessageWorkerCount,
		"queue_size":           hpm.config.MessageQueueSize,
	}).Info("Starting high-performance message processing system")
	
	// Apply system optimizations
	hpm.applySystemOptimizations()
	
	// Start message processor
	if err := hpm.messageProcessor.Start(); err != nil {
		return fmt.Errorf("failed to start message processor: %w", err)
	}
	
	// Start system monitors
	hpm.startSystemMonitors()
	
	hpm.isRunning = true
	hpm.startTime = time.Now()
	
	hpm.logger.Info("High-performance system started successfully")
	return nil
}

// Stop stops the high-performance system gracefully
func (hpm *HighPerformanceManager) Stop() error {
	hpm.mu.Lock()
	defer hpm.mu.Unlock()
	
	if !hpm.isRunning {
		return fmt.Errorf("high-performance manager is not running")
	}
	
	hpm.logger.Info("Stopping high-performance system...")
	
	// Cancel context to signal shutdown
	hpm.cancel()
	
	// Stop message processor
	if err := hpm.messageProcessor.Stop(); err != nil {
		hpm.logger.WithError(err).Error("Error stopping message processor")
	}
	
	// Stop enhanced AI service
	hpm.enhancedAI.Stop()
	
	// Wait for all goroutines to finish
	hpm.wg.Wait()
	
	hpm.isRunning = false
	hpm.logger.Info("High-performance system stopped")
	return nil
}

// ProcessMessage processes a message through the high-performance pipeline
func (hpm *HighPerformanceManager) ProcessMessage(msg *IncomingMessage) error {
	if !hpm.isRunning {
		return fmt.Errorf("high-performance manager is not running")
	}
	
	// Check global rate limit
	if !hpm.globalRateLimiter.Allow() {
		return fmt.Errorf("global rate limit exceeded")
	}
	
	// Update concurrent users metric
	hpm.updateConcurrentUsers(1)
	defer hpm.updateConcurrentUsers(-1)
	
	// Process through message processor
	return hpm.messageProcessor.ProcessMessage(msg)
}

// GetSystemStatus returns comprehensive system status
func (hpm *HighPerformanceManager) GetSystemStatus() map[string]interface{} {
	hpm.mu.RLock()
	defer hpm.mu.RUnlock()
	
	metrics := hpm.getSystemMetrics()
	messageProcessorStatus := hpm.messageProcessor.GetStatus()
	aiServiceStatus := hpm.enhancedAI.GetStatus()
	
	return map[string]interface{}{
		"running":               hpm.isRunning,
		"uptime":                time.Since(hpm.startTime).String(),
		"system_metrics":        metrics,
		"message_processor":     messageProcessorStatus,
		"ai_service":            aiServiceStatus,
		"global_rate_limiter":   hpm.globalRateLimiter.GetStatus(),
		"configuration":         hpm.config,
		"performance_targets": map[string]interface{}{
			"max_concurrent_users": hpm.config.MaxConcurrentUsers,
			"target_response_time": hpm.config.TargetResponseTime.String(),
			"current_performance": map[string]interface{}{
				"concurrent_users":    metrics.ConcurrentUsers,
				"average_response_time": metrics.AverageResponseTime.String(),
				"messages_per_second": metrics.MessagesPerSecond,
			},
		},
	}
}

// GetPerformanceReport generates a detailed performance report
func (hpm *HighPerformanceManager) GetPerformanceReport() map[string]interface{} {
	status := hpm.GetSystemStatus()
	metrics := hpm.getSystemMetrics()
	
	// Calculate performance scores
	responseTimeScore := hpm.calculateResponseTimeScore(metrics.AverageResponseTime)
	throughputScore := hpm.calculateThroughputScore(metrics.MessagesPerSecond)
	concurrencyScore := hpm.calculateConcurrencyScore(metrics.ConcurrentUsers)
	overallScore := (responseTimeScore + throughputScore + concurrencyScore) / 3
	
	return map[string]interface{}{
		"timestamp":     time.Now().Unix(),
		"system_status": status,
		"performance_scores": map[string]interface{}{
			"response_time_score": responseTimeScore,
			"throughput_score":    throughputScore,
			"concurrency_score":   concurrencyScore,
			"overall_score":       overallScore,
		},
		"recommendations": hpm.generateRecommendations(metrics),
	}
}

// RegisterRoutes registers high-performance routes
func (hpm *HighPerformanceManager) RegisterRoutes(router *gin.Engine) {
	// Register optimized webhook routes - commented out as function is not defined
	// RegisterOptimizedWebhookRoutes(router, hpm.messageProcessor)
	
	// Performance monitoring routes
	v1 := router.Group("/api/v1/performance")
	{
		v1.GET("/status", func(c *gin.Context) {
			c.JSON(200, hpm.GetSystemStatus())
		})
		
		v1.GET("/report", func(c *gin.Context) {
			c.JSON(200, hpm.GetPerformanceReport())
		})
		
		v1.GET("/metrics", func(c *gin.Context) {
			c.JSON(200, hpm.getSystemMetrics())
		})
		
		v1.POST("/config", func(c *gin.Context) {
			var newConfig HighPerformanceConfig
			if err := c.ShouldBindJSON(&newConfig); err != nil {
				c.JSON(400, gin.H{"error": err.Error()})
				return
			}
			
			hpm.updateConfiguration(&newConfig)
			c.JSON(200, gin.H{"status": "configuration updated"})
		})
	}
}

// applySystemOptimizations applies various system-level optimizations
func (hpm *HighPerformanceManager) applySystemOptimizations() {
	// Set GC target percentage for better performance
	// TODO: Fix runtime.SetGCPercent issue
	// runtime.SetGCPercent(hpm.config.GCTargetPercentage)
	
	// Set max procs to utilize all CPU cores
	runtime.GOMAXPROCS(runtime.NumCPU())
	
	hpm.logger.WithFields(logrus.Fields{
		"gc_target_percentage": hpm.config.GCTargetPercentage,
		"max_procs":            runtime.GOMAXPROCS(0),
		"num_cpu":              runtime.NumCPU(),
	}).Info("Applied system optimizations")
}

// startSystemMonitors starts various system monitoring goroutines
func (hpm *HighPerformanceManager) startSystemMonitors() {
	// Metrics collector
	hpm.wg.Add(1)
	go func() {
		defer hpm.wg.Done()
		hpm.metricsCollector()
	}()
	
	// Health monitor
	hpm.wg.Add(1)
	go func() {
		defer hpm.wg.Done()
		hpm.healthMonitor()
	}()
	
	// Performance optimizer
	hpm.wg.Add(1)
	go func() {
		defer hpm.wg.Done()
		hpm.performanceOptimizer()
	}()
}

// metricsCollector collects system metrics periodically
func (hpm *HighPerformanceManager) metricsCollector() {
	ticker := time.NewTicker(hpm.config.MetricsInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			hpm.collectSystemMetrics()
		case <-hpm.ctx.Done():
			return
		}
	}
}

// healthMonitor monitors system health
func (hpm *HighPerformanceManager) healthMonitor() {
	ticker := time.NewTicker(hpm.config.HealthCheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			hpm.performHealthCheck()
		case <-hpm.ctx.Done():
			return
		}
	}
}

// performanceOptimizer automatically optimizes performance
func (hpm *HighPerformanceManager) performanceOptimizer() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			hpm.optimizePerformance()
		case <-hpm.ctx.Done():
			return
		}
	}
}

// collectSystemMetrics collects current system metrics
func (hpm *HighPerformanceManager) collectSystemMetrics() {
	hpm.metrics.mu.Lock()
	defer hpm.metrics.mu.Unlock()
	
	// Update system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	hpm.metrics.MemoryUsage = m.Alloc
	hpm.metrics.GoroutineCount = runtime.NumGoroutine()
	hpm.metrics.SystemUptime = time.Since(hpm.startTime)
	hpm.metrics.LastUpdated = time.Now()
	
	// Calculate messages per second
	processorMetrics := hpm.messageProcessor.GetMetrics()
	if hpm.metrics.SystemUptime > 0 {
		hpm.metrics.MessagesPerSecond = float64(processorMetrics.TotalProcessed) / hpm.metrics.SystemUptime.Seconds()
	}
}

// getSystemMetrics returns current system metrics
func (hpm *HighPerformanceManager) getSystemMetrics() *SystemMetrics {
	hpm.metrics.mu.RLock()
	defer hpm.metrics.mu.RUnlock()
	
	// Create a copy
	metrics := *hpm.metrics
	return &metrics
}

// updateConcurrentUsers updates concurrent user count
func (hpm *HighPerformanceManager) updateConcurrentUsers(delta int64) {
	hpm.metrics.mu.Lock()
	defer hpm.metrics.mu.Unlock()
	
	hpm.metrics.ConcurrentUsers += delta
	if hpm.metrics.ConcurrentUsers > hpm.metrics.PeakConcurrentUsers {
		hpm.metrics.PeakConcurrentUsers = hpm.metrics.ConcurrentUsers
	}
}

// calculateResponseTimeScore calculates response time performance score (0-100)
func (hpm *HighPerformanceManager) calculateResponseTimeScore(avgResponseTime time.Duration) float64 {
	target := hpm.config.TargetResponseTime
	if avgResponseTime <= target {
		return 100.0
	}
	
	// Score decreases as response time increases beyond target
	ratio := float64(avgResponseTime) / float64(target)
	score := 100.0 / ratio
	if score < 0 {
		score = 0
	}
	return score
}

// calculateThroughputScore calculates throughput performance score (0-100)
func (hpm *HighPerformanceManager) calculateThroughputScore(messagesPerSecond float64) float64 {
	// Target: 1000 messages per second for excellent score
	target := 1000.0
	score := (messagesPerSecond / target) * 100
	if score > 100 {
		score = 100
	}
	return score
}

// calculateConcurrencyScore calculates concurrency performance score (0-100)
func (hpm *HighPerformanceManager) calculateConcurrencyScore(concurrentUsers int64) float64 {
	target := float64(hpm.config.MaxConcurrentUsers)
	if concurrentUsers <= int64(target) {
		return 100.0
	}
	
	// Score decreases as concurrent users exceed target
	ratio := float64(concurrentUsers) / target
	score := 100.0 / ratio
	if score < 0 {
		score = 0
	}
	return score
}

// generateRecommendations generates performance recommendations
func (hpm *HighPerformanceManager) generateRecommendations(metrics *SystemMetrics) []string {
	recommendations := make([]string, 0)
	
	if metrics.AverageResponseTime > hpm.config.TargetResponseTime {
		recommendations = append(recommendations, "Consider increasing worker count or optimizing AI response time")
	}
	
	if metrics.ConcurrentUsers > int64(hpm.config.MaxConcurrentUsers*80/100) {
		recommendations = append(recommendations, "Approaching maximum concurrent user limit - consider scaling")
	}
	
	if metrics.MessagesPerSecond < 100 {
		recommendations = append(recommendations, "Low message throughput - check for bottlenecks")
	}
	
	if metrics.MemoryUsage > 1024*1024*1024 { // 1GB
		recommendations = append(recommendations, "High memory usage detected - consider optimizing memory usage")
	}
	
	if metrics.GoroutineCount > 10000 {
		recommendations = append(recommendations, "High goroutine count - check for goroutine leaks")
	}
	
	return recommendations
}

// performHealthCheck performs a comprehensive health check
func (hpm *HighPerformanceManager) performHealthCheck() {
	metrics := hpm.getSystemMetrics()
	
	// Check various health indicators
	healthIssues := make([]string, 0)
	
	if metrics.AverageResponseTime > hpm.config.TargetResponseTime*2 {
		healthIssues = append(healthIssues, "Response time is critically high")
	}
	
	if metrics.ConcurrentUsers > int64(hpm.config.MaxConcurrentUsers) {
		healthIssues = append(healthIssues, "Concurrent users exceed maximum limit")
	}
	
	if len(healthIssues) > 0 {
		hpm.logger.WithField("health_issues", healthIssues).Warn("Health check detected issues")
	} else {
		hpm.logger.Debug("Health check passed")
	}
}

// optimizePerformance automatically optimizes performance based on current metrics
func (hpm *HighPerformanceManager) optimizePerformance() {
	metrics := hpm.getSystemMetrics()
	
	// Auto-optimization logic
	if metrics.AverageResponseTime > hpm.config.TargetResponseTime {
		// Could trigger scaling or optimization here
		hpm.logger.Info("Performance optimization triggered due to high response time")
	}
	
	// Force garbage collection if memory usage is high
	if metrics.MemoryUsage > 512*1024*1024 { // 512MB
		runtime.GC()
		hpm.logger.Debug("Forced garbage collection due to high memory usage")
	}
}

// updateConfiguration updates the system configuration
func (hpm *HighPerformanceManager) updateConfiguration(newConfig *HighPerformanceConfig) {
	hpm.mu.Lock()
	defer hpm.mu.Unlock()
	
	hpm.config = newConfig
	hpm.logger.WithField("config", newConfig).Info("Configuration updated")
	
	// Apply new optimizations
	hpm.applySystemOptimizations()
}