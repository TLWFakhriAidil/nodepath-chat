package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// QueueMessage represents a message in the queue
type QueueMessage struct {
	ID          string    `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	Content     string    `json:"content"`
	MediaURL    string    `json:"media_url,omitempty"`
	MediaType   string    `json:"media_type,omitempty"`
	RetryCount  int       `json:"retry_count"`
	MaxRetries  int       `json:"max_retries"`
	CreatedAt   time.Time `json:"created_at"`
	ScheduledAt time.Time `json:"scheduled_at"`
}

// QueueService handles message queuing using Redis
type QueueService struct {
	redis *redis.Client
	ctx   context.Context
}

// NewQueueService creates a new queue service
func NewQueueService(redisClient *redis.Client) *QueueService {
	return &QueueService{
		redis: redisClient,
		ctx:   context.Background(),
	}
}

// QueueMessage queues a message for processing
func (qs *QueueService) QueueMessage(phoneNumber, content string) error {
	message := &QueueMessage{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		PhoneNumber: phoneNumber,
		Content:     content,
		RetryCount:  0,
		MaxRetries:  3,
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now(),
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = qs.redis.LPush(qs.ctx, "outbound_messages", messageJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to queue message: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"message_id":   message.ID,
	}).Debug("Message queued successfully")

	return nil
}

// QueueMediaMessage queues a media message for processing
func (qs *QueueService) QueueMediaMessage(phoneNumber, content, mediaURL, mediaType string) error {
	message := &QueueMessage{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		PhoneNumber: phoneNumber,
		Content:     content,
		MediaURL:    mediaURL,
		MediaType:   mediaType,
		RetryCount:  0,
		MaxRetries:  3,
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now(),
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = qs.redis.LPush(qs.ctx, "outbound_messages", messageJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to queue message: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"message_id":   message.ID,
		"media_type":   mediaType,
	}).Debug("Media message queued successfully")

	return nil
}

// QueueDelayedMessage queues a message with a delay
func (qs *QueueService) QueueDelayedMessage(phoneNumber, content string, delay time.Duration) error {
	message := &QueueMessage{
		ID:          fmt.Sprintf("%d", time.Now().UnixNano()),
		PhoneNumber: phoneNumber,
		Content:     content,
		RetryCount:  0,
		MaxRetries:  3,
		CreatedAt:   time.Now(),
		ScheduledAt: time.Now().Add(delay),
	}

	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Use sorted set for delayed messages with timestamp as score
	score := float64(message.ScheduledAt.Unix())
	err = qs.redis.ZAdd(qs.ctx, "delayed_messages", redis.Z{
		Score:  score,
		Member: string(messageJSON),
	}).Err()

	if err != nil {
		return fmt.Errorf("failed to queue delayed message: %w", err)
	}

	logrus.WithFields(logrus.Fields{
		"phone_number": phoneNumber,
		"message_id":   message.ID,
		"delay":        delay,
	}).Debug("Delayed message queued successfully")

	return nil
}

// DequeueOutboundMessage dequeues a message for processing
func (qs *QueueService) DequeueOutboundMessage() (*QueueMessage, error) {
	result, err := qs.redis.BRPop(qs.ctx, 1*time.Second, "outbound_messages").Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // No messages available
		}
		return nil, fmt.Errorf("failed to dequeue message: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid result from Redis")
	}

	var message QueueMessage
	err = json.Unmarshal([]byte(result[1]), &message)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &message, nil
}

// ProcessDelayedMessages moves ready delayed messages to the main queue
func (qs *QueueService) ProcessDelayedMessages() error {
	now := float64(time.Now().Unix())
	
	// Get messages that are ready to be processed
	result, err := qs.redis.ZRangeByScore(qs.ctx, "delayed_messages", &redis.ZRangeBy{
		Min: "0",
		Max: fmt.Sprintf("%f", now),
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to get delayed messages: %w", err)
	}

	for _, messageJSON := range result {
		// Move to main queue
		err = qs.redis.LPush(qs.ctx, "outbound_messages", messageJSON).Err()
		if err != nil {
			logrus.WithError(err).Error("Failed to move delayed message to main queue")
			continue
		}

		// Remove from delayed queue
		err = qs.redis.ZRem(qs.ctx, "delayed_messages", messageJSON).Err()
		if err != nil {
			logrus.WithError(err).Error("Failed to remove message from delayed queue")
		}
	}

	if len(result) > 0 {
		logrus.WithField("count", len(result)).Debug("Processed delayed messages")
	}

	return nil
}

// RequeueFailedMessage requeues a failed message with retry logic
func (qs *QueueService) RequeueFailedMessage(message *QueueMessage, err error) {
	message.RetryCount++

	logrus.WithFields(logrus.Fields{
		"message_id":   message.ID,
		"phone_number": message.PhoneNumber,
		"retry_count":  message.RetryCount,
		"error":        err.Error(),
	}).Warn("Message failed, attempting retry")

	if message.RetryCount >= message.MaxRetries {
		logrus.WithFields(logrus.Fields{
			"message_id":   message.ID,
			"phone_number": message.PhoneNumber,
			"retry_count":  message.RetryCount,
		}).Error("Message exceeded max retries, discarding")
		return
	}

	// Exponential backoff: 2^retry_count minutes
	delay := time.Duration(1<<message.RetryCount) * time.Minute
	message.ScheduledAt = time.Now().Add(delay)

	messageJSON, marshalErr := json.Marshal(message)
	if marshalErr != nil {
		logrus.WithError(marshalErr).Error("Failed to marshal retry message")
		return
	}

	// Add back to delayed queue
	score := float64(message.ScheduledAt.Unix())
	redisErr := qs.redis.ZAdd(qs.ctx, "delayed_messages", redis.Z{
		Score:  score,
		Member: string(messageJSON),
	}).Err()

	if redisErr != nil {
		logrus.WithError(redisErr).Error("Failed to requeue message")
	}
}

// GetQueueStats returns statistics about the queues
func (qs *QueueService) GetQueueStats() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Get outbound queue length
	outboundLen, err := qs.redis.LLen(qs.ctx, "outbound_messages").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get outbound queue length: %w", err)
	}
	stats["outbound_queue"] = outboundLen

	// Get delayed queue length
	delayedLen, err := qs.redis.ZCard(qs.ctx, "delayed_messages").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get delayed queue length: %w", err)
	}
	stats["delayed_queue"] = delayedLen

	return stats, nil
}

// ClearQueues clears all queues (for testing/maintenance)
func (qs *QueueService) ClearQueues() error {
	err := qs.redis.Del(qs.ctx, "outbound_messages", "delayed_messages").Err()
	if err != nil {
		return fmt.Errorf("failed to clear queues: %w", err)
	}

	logrus.Info("All queues cleared")
	return nil
}