package services

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// MessageType represents different types of messages
type MessageType string

const (
	MessageTypeUserReply     MessageType = "user_reply"
	MessageTypeCustomerReply MessageType = "customer_reply"
	MessageTypeAIReply       MessageType = "ai_reply"
	MessageTypeSystem        MessageType = "system"
)

// IncomingMessage represents a message to be processed
type IncomingMessage struct {
	ID           string                 `json:"id"`
	Type         MessageType            `json:"type"`
	DeviceID     string                 `json:"device_id"`
	PhoneNumber  string                 `json:"phone_number"`
	Content      string                 `json:"content"`
	MediaURL     string                 `json:"media_url,omitempty"`
	MediaType    string                 `json:"media_type,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	Priority     int                    `json:"priority"` // 1=highest, 5=lowest
	RetryCount   int                    `json:"retry_count"`
	MaxRetries   int                    `json:"max_retries"`
}

// ProcessingResult represents the result of message processing
type ProcessingResult struct {
	MessageID    string    `json:"message_id"`
	Success      bool      `json:"success"`
	Response     string    `json:"response,omitempty"`
	Error        string    `json:"error,omitempty"`
	ProcessedAt  time.Time `json:"processed_at"`
	ProcessingTime time.Duration `json:"processing_time"`
}

// MessageProcessor handles high-performance message processing
type MessageProcessor struct {
	mu                sync.RWMutex
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	
	// Services
	chatService       *ChatService
	flowService       *FlowService
	aiService         *AIService
	queueService      *QueueService
	
	// Processing channels
	incomingChan      chan *IncomingMessage
	processingChan    chan *IncomingMessage
	resultChan        chan *ProcessingResult
	
	// Worker pools
	workerCount       int
	processorWorkers  []*MessageWorker
	
	// Performance metrics
	metrics           *ProcessingMetrics
	
	// Configuration
	config            *ProcessorConfig
	
	// Rate limiting
	rateLimiter       *RateLimiter
	
	// Circuit breaker for AI service
	circuitBreaker    *CircuitBreaker
	
	// Message cache for deduplication
	messageCache      *MessageCache
	
	isRunning         bool
}

// ProcessorConfig holds configuration for the message processor
type ProcessorConfig struct {
	WorkerCount          int           `json:"worker_count"`
	ChannelBufferSize    int           `json:"channel_buffer_size"`
	ProcessingTimeout    time.Duration `json:"processing_timeout"`
	MaxRetries           int           `json:"max_retries"`
	RetryDelay           time.Duration `json:"retry_delay"`
	RateLimitPerSecond   int           `json:"rate_limit_per_second"`
	EnableCircuitBreaker bool          `json:"enable_circuit_breaker"`
	EnableDeduplication  bool          `json:"enable_deduplication"`
	CacheSize            int           `json:"cache_size"`
	CacheTTL             time.Duration `json:"cache_ttl"`
}

// ProcessingMetrics tracks performance metrics
type ProcessingMetrics struct {
	mu                    sync.RWMutex
	TotalProcessed        int64         `json:"total_processed"`
	TotalFailed           int64         `json:"total_failed"`
	UserRepliesProcessed  int64         `json:"user_replies_processed"`
	CustomerRepliesProcessed int64      `json:"customer_replies_processed"`
	AIRepliesProcessed    int64         `json:"ai_replies_processed"`
	AverageProcessingTime time.Duration `json:"average_processing_time"`
	PeakProcessingTime    time.Duration `json:"peak_processing_time"`
	CurrentLoad           int           `json:"current_load"`
	MaxLoad               int           `json:"max_load"`
	StartTime             time.Time     `json:"start_time"`
	LastProcessedAt       time.Time     `json:"last_processed_at"`
}

// NewMessageProcessor creates a new high-performance message processor
func NewMessageProcessor(
	chatService *ChatService,
	flowService *FlowService,
	aiService *AIService,
	queueService *QueueService,
) *MessageProcessor {
	ctx, cancel := context.WithCancel(context.Background())
	
	// Default configuration optimized for 1500+ concurrent users
	config := &ProcessorConfig{
		WorkerCount:          runtime.NumCPU() * 4, // Scale with CPU cores
		ChannelBufferSize:    10000,                // Large buffer for high throughput
		ProcessingTimeout:    30 * time.Second,     // Reasonable timeout
		MaxRetries:           3,
		RetryDelay:           1 * time.Second,
		RateLimitPerSecond:   1000,                 // 1000 messages per second
		EnableCircuitBreaker: true,
		EnableDeduplication:  true,
		CacheSize:            50000,                // Cache for 50k messages
		CacheTTL:             5 * time.Minute,
	}
	
	processor := &MessageProcessor{
		ctx:            ctx,
		cancel:         cancel,
		chatService:    chatService,
		flowService:    flowService,
		aiService:      aiService,
		queueService:   queueService,
		incomingChan:   make(chan *IncomingMessage, config.ChannelBufferSize),
		processingChan: make(chan *IncomingMessage, config.ChannelBufferSize),
		resultChan:     make(chan *ProcessingResult, config.ChannelBufferSize),
		workerCount:    config.WorkerCount,
		config:         config,
		metrics: &ProcessingMetrics{
			StartTime: time.Now(),
		},
	}
	
	// Initialize components
	processor.rateLimiter = NewRateLimiter(config.RateLimitPerSecond)
	processor.circuitBreaker = NewCircuitBreaker()
	processor.messageCache = NewMessageCache(config.CacheSize, config.CacheTTL)
	
	return processor
}

// Start starts the message processor
func (mp *MessageProcessor) Start() error {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	if mp.isRunning {
		return fmt.Errorf("message processor is already running")
	}
	
	logrus.WithFields(logrus.Fields{
		"worker_count":       mp.workerCount,
		"channel_buffer":     mp.config.ChannelBufferSize,
		"rate_limit":         mp.config.RateLimitPerSecond,
		"processing_timeout": mp.config.ProcessingTimeout,
	}).Info("Starting high-performance message processor")
	
	// Start worker pool
	mp.processorWorkers = make([]*MessageWorker, mp.workerCount)
	for i := 0; i < mp.workerCount; i++ {
		worker := NewMessageWorker(i, mp)
		mp.processorWorkers[i] = worker
		
		mp.wg.Add(1)
		go func(w *MessageWorker) {
			defer mp.wg.Done()
			w.Start(mp.ctx)
		}(worker)
	}
	
	// Start result processor
	mp.wg.Add(1)
	go func() {
		defer mp.wg.Done()
		mp.processResults()
	}()
	
	// Start metrics collector
	mp.wg.Add(1)
	go func() {
		defer mp.wg.Done()
		mp.collectMetrics()
	}()
	
	// Start health monitor
	mp.wg.Add(1)
	go func() {
		defer mp.wg.Done()
		mp.healthMonitor()
	}()
	
	mp.isRunning = true
	logrus.Info("Message processor started successfully")
	return nil
}

// Stop stops the message processor gracefully
func (mp *MessageProcessor) Stop() error {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	
	if !mp.isRunning {
		return fmt.Errorf("message processor is not running")
	}
	
	logrus.Info("Stopping message processor...")
	
	// Cancel context to signal shutdown
	mp.cancel()
	
	// Close channels
	close(mp.incomingChan)
	close(mp.processingChan)
	
	// Wait for all workers to finish
	mp.wg.Wait()
	
	// Close result channel
	close(mp.resultChan)
	
	mp.isRunning = false
	logrus.Info("Message processor stopped")
	return nil
}

// ProcessMessage queues a message for processing
func (mp *MessageProcessor) ProcessMessage(msg *IncomingMessage) error {
	if !mp.isRunning {
		return fmt.Errorf("message processor is not running")
	}
	
	// Check rate limit
	if !mp.rateLimiter.Allow() {
		return fmt.Errorf("rate limit exceeded")
	}
	
	// Check for duplicate messages
	if mp.config.EnableDeduplication {
		if mp.messageCache.Exists(msg.ID) {
			logrus.WithField("message_id", msg.ID).Debug("Duplicate message detected, skipping")
			return nil
		}
		mp.messageCache.Set(msg.ID, true)
	}
	
	// Set defaults
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	if msg.MaxRetries == 0 {
		msg.MaxRetries = mp.config.MaxRetries
	}
	if msg.Priority == 0 {
		msg.Priority = 3 // Default priority
	}
	
	// Queue message for processing
	select {
	case mp.incomingChan <- msg:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("message queue is full, try again later")
	}
}

// GetMetrics returns current processing metrics
func (mp *MessageProcessor) GetMetrics() *ProcessingMetrics {
	mp.metrics.mu.RLock()
	defer mp.metrics.mu.RUnlock()
	
	// Create a copy to avoid race conditions
	metrics := *mp.metrics
	return &metrics
}

// GetStatus returns the current status of the processor
func (mp *MessageProcessor) GetStatus() map[string]interface{} {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	
	metrics := mp.GetMetrics()
	
	return map[string]interface{}{
		"running":              mp.isRunning,
		"worker_count":         mp.workerCount,
		"incoming_queue_size":  len(mp.incomingChan),
		"processing_queue_size": len(mp.processingChan),
		"result_queue_size":    len(mp.resultChan),
		"total_processed":      metrics.TotalProcessed,
		"total_failed":         metrics.TotalFailed,
		"current_load":         metrics.CurrentLoad,
		"max_load":             metrics.MaxLoad,
		"average_processing_time": metrics.AverageProcessingTime.String(),
		"peak_processing_time": metrics.PeakProcessingTime.String(),
		"uptime":               time.Since(metrics.StartTime).String(),
		"rate_limiter_status": mp.rateLimiter.GetStatus(),
		"circuit_breaker_status": mp.circuitBreaker.GetStatus(),
	}
}

// processResults processes the results from workers
func (mp *MessageProcessor) processResults() {
	for {
		select {
		case result := <-mp.resultChan:
			if result == nil {
				return
			}
			
			// Update metrics
			mp.updateMetrics(result)
			
			// Log result
			if result.Success {
				logrus.WithFields(logrus.Fields{
					"message_id":      result.MessageID,
					"processing_time": result.ProcessingTime,
				}).Debug("Message processed successfully")
			} else {
				logrus.WithFields(logrus.Fields{
					"message_id":      result.MessageID,
					"error":           result.Error,
					"processing_time": result.ProcessingTime,
				}).Error("Message processing failed")
			}
			
		case <-mp.ctx.Done():
			return
		}
	}
}

// updateMetrics updates processing metrics
func (mp *MessageProcessor) updateMetrics(result *ProcessingResult) {
	mp.metrics.mu.Lock()
	defer mp.metrics.mu.Unlock()
	
	if result.Success {
		mp.metrics.TotalProcessed++
	} else {
		mp.metrics.TotalFailed++
	}
	
	// Update processing time metrics
	if result.ProcessingTime > mp.metrics.PeakProcessingTime {
		mp.metrics.PeakProcessingTime = result.ProcessingTime
	}
	
	// Calculate average processing time
	total := mp.metrics.TotalProcessed + mp.metrics.TotalFailed
	if total > 0 {
		currentAvg := mp.metrics.AverageProcessingTime
		mp.metrics.AverageProcessingTime = time.Duration(
			(int64(currentAvg)*(total-1) + int64(result.ProcessingTime)) / total,
		)
	}
	
	mp.metrics.LastProcessedAt = result.ProcessedAt
}

// collectMetrics periodically collects system metrics
func (mp *MessageProcessor) collectMetrics() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			mp.metrics.mu.Lock()
			currentLoad := len(mp.incomingChan) + len(mp.processingChan)
			mp.metrics.CurrentLoad = currentLoad
			if currentLoad > mp.metrics.MaxLoad {
				mp.metrics.MaxLoad = currentLoad
			}
			mp.metrics.mu.Unlock()
			
		case <-mp.ctx.Done():
			return
		}
	}
}

// healthMonitor monitors the health of the processor
func (mp *MessageProcessor) healthMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			metrics := mp.GetMetrics()
			
			// Check for potential issues
			if metrics.CurrentLoad > mp.config.ChannelBufferSize*80/100 {
				logrus.WithField("current_load", metrics.CurrentLoad).Warn("High message queue load detected")
			}
			
			if time.Since(metrics.LastProcessedAt) > 5*time.Minute && metrics.TotalProcessed > 0 {
				logrus.Warn("No messages processed in the last 5 minutes")
			}
			
			// Log health status
			logrus.WithFields(logrus.Fields{
				"total_processed": metrics.TotalProcessed,
				"total_failed":    metrics.TotalFailed,
				"current_load":    metrics.CurrentLoad,
				"success_rate":    float64(metrics.TotalProcessed) / float64(metrics.TotalProcessed+metrics.TotalFailed) * 100,
			}).Info("Message processor health check")
			
		case <-mp.ctx.Done():
			return
		}
	}
}