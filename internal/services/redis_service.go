package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"nodepath-chat/internal/config"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// InitializeRedis initializes Redis connection with clustering support
func InitializeRedis(cfg *config.Config) redis.Cmdable {
	if cfg.RedisURL == "" {
		logrus.Warn("Redis URL not provided, Redis features will be disabled")
		return nil
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		logrus.WithError(err).Error("Failed to parse Redis URL")
		return nil
	}

	// Check if cluster addresses are provided in config
	var client redis.Cmdable
	if len(cfg.RedisClusterAddrs) > 0 {
		// Use Redis Cluster for high availability and performance
		client = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:    cfg.RedisClusterAddrs,
			Password: opt.Password,
			// Optimized for high concurrency
			PoolSize:     100, // Increased pool size
			MinIdleConns: 20,  // Minimum idle connections
			MaxRetries:   3,   // Retry failed commands
			DialTimeout:  5 * time.Second,
			ReadTimeout:  3 * time.Second,
			WriteTimeout: 3 * time.Second,
		})
		logrus.Info("Using Redis Cluster configuration")
	} else {
		// Single Redis instance with optimized settings
		opt.PoolSize = 50     // Increased pool size
		opt.MinIdleConns = 10 // Minimum idle connections
		opt.MaxRetries = 3    // Retry failed commands
		opt.DialTimeout = 5 * time.Second
		opt.ReadTimeout = 3 * time.Second
		opt.WriteTimeout = 3 * time.Second
		client = redis.NewClient(opt)
		logrus.Info("Using single Redis instance configuration")
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logrus.WithError(err).Error("Failed to connect to Redis")
		return nil
	}

	logrus.Info("Redis connection established")
	return client
}

// QueueService handles Redis-based job queuing with clustering support and monitoring
type QueueService struct {
	redis redis.Cmdable // Interface to support both single and cluster clients
	// WhatsApp service interface for flow continuation
	whatsappService WhatsAppServiceInterface
	queueMonitor    *QueueMonitor
}

// WhatsAppServiceInterface defines the interface for WhatsApp service methods needed by queue service
type WhatsAppServiceInterface interface {
	ProcessFlowContinuation(executionID, flowID, nodeID, phoneNumber, deviceID, userInput string) error
}

// NewQueueService creates a new queue service with monitoring
func NewQueueService(redis redis.Cmdable, queueMonitor *QueueMonitor) *QueueService {
	return &QueueService{
		redis:        redis,
		queueMonitor: queueMonitor,
	}
}

// SetWhatsAppService sets the WhatsApp service for flow continuation
func (s *QueueService) SetWhatsAppService(whatsappService WhatsAppServiceInterface) {
	s.whatsappService = whatsappService
}

// QueueMessage represents a queued message
type QueueMessage struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	PhoneNumber string                 `json:"phone_number"`
	Content     string                 `json:"content"`
	MediaURL    string                 `json:"media_url,omitempty"`
	MediaType   string                 `json:"media_type,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Retries     int                    `json:"retries"`
	MaxRetries  int                    `json:"max_retries"`
	CreatedAt   time.Time              `json:"created_at"`
	ScheduledAt time.Time              `json:"scheduled_at,omitempty"`
	// Additional fields for flow continuation
	DeviceID    string        `json:"device_id,omitempty"`
	MessageType string        `json:"message_type,omitempty"`
	FlowID      string        `json:"flow_id,omitempty"`
	ExecutionID string        `json:"execution_id,omitempty"`
	NodeID      string        `json:"node_id,omitempty"`
	Delay       time.Duration `json:"delay,omitempty"`
}

const (
	queueKeyOutbound = "queue:outbound"
	queueKeyFailed   = "queue:failed"
	queueKeyDelay    = "queue:delay"
)

// EnqueueOutboundMessage queues an outbound WhatsApp message with monitoring
func (s *QueueService) EnqueueOutboundMessage(phoneNumber, content, mediaURL, mediaType string, metadata map[string]interface{}) error {
	if s.redis == nil {
		logrus.Warn("Redis not available, message will be sent immediately")
		return nil
	}

	startTime := time.Now()

	message := QueueMessage{
		ID:          fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		Type:        "whatsapp_outbound",
		PhoneNumber: phoneNumber,
		Content:     content,
		MediaURL:    mediaURL,
		MediaType:   mediaType,
		Metadata:    metadata,
		Retries:     0,
		MaxRetries:  3,
		CreatedAt:   time.Now(),
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		if s.queueMonitor != nil {
			s.queueMonitor.RecordError()
		}
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx := context.Background()
	err = s.redis.LPush(ctx, queueKeyOutbound, messageJSON).Err()
	if err != nil {
		if s.queueMonitor != nil {
			s.queueMonitor.RecordError()
		}
		return fmt.Errorf("failed to enqueue message: %w", err)
	}

	// Record queue metrics
	if s.queueMonitor != nil {
		s.queueMonitor.RecordProcessingTime(time.Since(startTime))
		// Get current queue size
		queueSize, _ := s.redis.LLen(ctx, queueKeyOutbound).Result()
		s.queueMonitor.RecordQueueSize(queueKeyOutbound, queueSize)
	}

	logrus.WithFields(logrus.Fields{
		"message_id":   message.ID,
		"phone_number": phoneNumber,
		"content_len":  len(content),
		"enqueue_time": time.Since(startTime),
	}).Info("Message queued for sending")

	return nil
}

// DequeueOutboundMessage dequeues the next outbound message with monitoring
func (s *QueueService) DequeueOutboundMessage() (*QueueMessage, error) {
	if s.redis == nil {
		return nil, nil
	}

	startTime := time.Now()
	ctx := context.Background()
	result, err := s.redis.BRPop(ctx, 5*time.Second, queueKeyOutbound).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No messages available
		}
		if s.queueMonitor != nil {
			s.queueMonitor.RecordError()
		}
		return nil, fmt.Errorf("failed to dequeue message: %w", err)
	}

	if len(result) < 2 {
		if s.queueMonitor != nil {
			s.queueMonitor.RecordError()
		}
		return nil, fmt.Errorf("invalid queue result")
	}

	var message QueueMessage
	err = json.Unmarshal([]byte(result[1]), &message)
	if err != nil {
		if s.queueMonitor != nil {
			s.queueMonitor.RecordError()
		}
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Record queue metrics
	if s.queueMonitor != nil {
		s.queueMonitor.RecordProcessingTime(time.Since(startTime))
		// Get current queue size after dequeue
		queueSize, _ := s.redis.LLen(ctx, queueKeyOutbound).Result()
		s.queueMonitor.RecordQueueSize(queueKeyOutbound, queueSize)
	}

	return &message, nil
}

// RequeueFailedMessage requeues a failed message with retry logic
func (s *QueueService) RequeueFailedMessage(message *QueueMessage, err error) error {
	if s.redis == nil {
		return nil
	}

	message.Retries++

	logrus.WithFields(logrus.Fields{
		"message_id":  message.ID,
		"retries":     message.Retries,
		"max_retries": message.MaxRetries,
		"error":       err.Error(),
	}).Warn("Message failed, requeuing")

	ctx := context.Background()

	if message.Retries >= message.MaxRetries {
		// Move to failed queue
		messageJSON, marshalErr := json.Marshal(message)
		if marshalErr != nil {
			return fmt.Errorf("failed to marshal failed message: %w", marshalErr)
		}

		err = s.redis.LPush(ctx, queueKeyFailed, messageJSON).Err()
		if err != nil {
			return fmt.Errorf("failed to enqueue failed message: %w", err)
		}

		logrus.WithField("message_id", message.ID).Error("Message moved to failed queue")
		return nil
	}

	// Calculate delay for retry (exponential backoff)
	delay := time.Duration(message.Retries*message.Retries) * time.Minute
	message.ScheduledAt = time.Now().Add(delay)

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal retry message: %w", err)
	}

	// Add to delay queue with score as timestamp
	score := float64(message.ScheduledAt.Unix())
	err = s.redis.ZAdd(ctx, queueKeyDelay, redis.Z{
		Score:  score,
		Member: string(messageJSON),
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to enqueue delayed message: %w", err)
	}

	return nil
}

// EnqueueDelayedMessage queues a message for delayed processing
func (s *QueueService) EnqueueDelayedMessage(message *QueueMessage) error {
	if s.redis == nil {
		logrus.Warn("Redis not available, delayed message cannot be queued")
		return fmt.Errorf("redis not available")
	}

	// Set scheduled time based on delay
	message.ScheduledAt = time.Now().Add(message.Delay)
	message.MaxRetries = 3
	message.Retries = 0

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal delayed message: %w", err)
	}

	ctx := context.Background()
	// Add to delay queue with score as timestamp
	score := float64(message.ScheduledAt.Unix())
	err = s.redis.ZAdd(ctx, queueKeyDelay, redis.Z{
		Score:  score,
		Member: string(messageJSON),
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to enqueue delayed message: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"message_id":     message.ID,
		"execution_id":   message.ExecutionID,
		"scheduled_time": message.ScheduledAt,
		"delay_seconds":  message.Delay.Seconds(),
	}).Info("🕐 QUEUE: Delayed message queued successfully")

	return nil
}

// ProcessDelayedMessages moves ready delayed messages back to the main queue
func (s *QueueService) ProcessDelayedMessages() error {
	if s.redis == nil {
		return nil
	}

	ctx := context.Background()
	now := float64(time.Now().Unix())

	// Get messages that are ready to be processed
	result, err := s.redis.ZRangeByScore(ctx, queueKeyDelay, &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to get delayed messages: %w", err)
	}

	for _, messageJSON := range result {
		// Parse message to check if it's a flow continuation
		var message QueueMessage
		err = json.Unmarshal([]byte(messageJSON), &message)
		if err != nil {
			logrus.WithError(err).Error("Failed to unmarshal delayed message")
			continue
		}

		// Handle flow continuation messages differently
		if message.MessageType == "flow_continuation" {
			// Process flow continuation directly
			err = s.processFlowContinuation(&message)
			if err != nil {
				logrus.WithError(err).Error("Failed to process flow continuation")
				continue
			}
		} else {
			// Move regular message back to main queue
			err = s.redis.LPush(ctx, queueKeyOutbound, messageJSON).Err()
			if err != nil {
				logrus.WithError(err).Error("Failed to move delayed message to main queue")
				continue
			}
		}

		// Remove from delay queue
		err = s.redis.ZRem(ctx, queueKeyDelay, messageJSON).Err()
		if err != nil {
			logrus.WithError(err).Error("Failed to remove message from delay queue")
		}
	}

	if len(result) > 0 {
		logrus.WithField("count", len(result)).Info("Processed delayed messages")
	}

	return nil
}

// GetQueueStats returns queue statistics
func (s *QueueService) GetQueueStats() (map[string]int64, error) {
	if s.redis == nil {
		return map[string]int64{
			"outbound": 0,
			"failed":   0,
			"delayed":  0,
		}, nil
	}

	ctx := context.Background()

	outbound, err := s.redis.LLen(ctx, queueKeyOutbound).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get outbound queue length: %w", err)
	}

	failed, err := s.redis.LLen(ctx, queueKeyFailed).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get failed queue length: %w", err)
	}

	delayed, err := s.redis.ZCard(ctx, queueKeyDelay).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get delayed queue length: %w", err)
	}

	return map[string]int64{
		"outbound": outbound,
		"failed":   failed,
		"delayed":  delayed,
	}, nil
}

// processFlowContinuation processes a flow continuation message after delay
func (s *QueueService) processFlowContinuation(message *QueueMessage) error {
	logrus.WithFields(logrus.Fields{
		"message_id":   message.ID,
		"execution_id": message.ExecutionID,
		"flow_id":      message.FlowID,
		"node_id":      message.NodeID,
	}).Info("🔄 QUEUE: Processing flow continuation after delay")

	// Check if WhatsApp service is available
	if s.whatsappService == nil {
		logrus.Error("🔄 QUEUE: WhatsApp service not available for flow continuation")
		return fmt.Errorf("whatsapp service not available")
	}

	// Call WhatsApp service to continue flow processing
	err := s.whatsappService.ProcessFlowContinuation(
		message.ExecutionID,
		message.FlowID,
		message.NodeID,
		message.PhoneNumber,
		message.DeviceID,
		message.Content,
	)

	if err != nil {
		// Check if this is an "execution not found" error - these are expected for cleaned up executions
		if strings.Contains(err.Error(), "execution not found") {
			// Log as debug level instead of error to reduce noise
			logrus.WithFields(logrus.Fields{
				"execution_id": message.ExecutionID,
				"message_id":   message.ID,
			}).Debug("🔄 QUEUE: Execution not found for delayed message (likely cleaned up)")
			// Return nil to prevent retries and remove from queue
			return nil
		}

		// For other errors, log as error
		logrus.WithError(err).Error("🔄 QUEUE: Failed to process flow continuation")
		return fmt.Errorf("failed to process flow continuation: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"execution_id": message.ExecutionID,
		"node_id":      message.NodeID,
	}).Info("✅ QUEUE: Flow continuation processed successfully")

	return nil
}

// ClearFailedMessages clears the failed message queue
func (s *QueueService) ClearFailedMessages() error {
	if s.redis == nil {
		return nil
	}

	ctx := context.Background()
	err := s.redis.Del(ctx, queueKeyFailed).Err()
	if err != nil {
		return fmt.Errorf("failed to clear failed queue: %w", err)
	}

	logrus.Info("Failed message queue cleared")
	return nil
}
