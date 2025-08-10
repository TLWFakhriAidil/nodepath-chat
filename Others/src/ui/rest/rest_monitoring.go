package rest

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/gofiber/fiber/v2"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// RedisMetrics represents Redis queue metrics
type RedisMetrics struct {
	Connected       bool                       `json:"connected"`
	TotalQueues     int                        `json:"total_queues"`
	TotalMessages   int64                      `json:"total_messages"`
	QueueDetails    []QueueDetail              `json:"queue_details"`
	MemoryUsage     string                     `json:"memory_usage"`
	ConnectedClients int                       `json:"connected_clients"`
	WorkerMetrics   map[string]WorkerMetric    `json:"worker_metrics"`
}

type QueueDetail struct {
	QueueName    string `json:"queue_name"`
	DeviceID     string `json:"device_id"`
	Type         string `json:"type"` // campaign or sequence
	MessageCount int64  `json:"message_count"`
}

type WorkerMetric struct {
	DeviceID      string  `json:"device_id"`
	SuccessCount  int64   `json:"success_count"`
	FailedCount   int64   `json:"failed_count"`
	LastSuccess   string  `json:"last_success"`
	LastFailure   string  `json:"last_failure"`
	SuccessRate   float64 `json:"success_rate"`
}

// GetRedisMetrics returns comprehensive Redis metrics
func GetRedisMetrics(c *fiber.Ctx) error {
	// Get Redis client
	redisURL := config.RedisURL
	if redisURL == "" {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"message": "Redis not configured",
			"error":   false,
			"data": RedisMetrics{
				Connected: false,
			},
		})
	}
	
	// Parse Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"message": "Failed to parse Redis URL",
			"error":   true,
		})
	}
	
	client := redis.NewClient(opt)
	defer client.Close()
	
	ctx := c.Context()
	
	// Check connection
	if err := client.Ping(ctx).Err(); err != nil {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusOK,
			"message": "Redis not connected",
			"error":   false,
			"data": RedisMetrics{
				Connected: false,
			},
		})
	}
	
	metrics := RedisMetrics{
		Connected:     true,
		QueueDetails:  []QueueDetail{},
		WorkerMetrics: make(map[string]WorkerMetric),
	}
	
	// Get Redis info
	info, err := client.Info(ctx, "memory", "clients").Result()
	if err == nil {
		// Parse memory usage
		if memMatch := regexp.MustCompile(`used_memory_human:(\S+)`).FindStringSubmatch(info); len(memMatch) > 1 {
			metrics.MemoryUsage = memMatch[1]
		}
		// Parse connected clients
		if clientMatch := regexp.MustCompile(`connected_clients:(\d+)`).FindStringSubmatch(info); len(clientMatch) > 1 {
			fmt.Sscanf(clientMatch[1], "%d", &metrics.ConnectedClients)
		}
	}
	
	// Get all campaign queues
	campaignKeys, _ := client.Keys(ctx, "ultra:queue:campaign:*").Result()
	for _, key := range campaignKeys {
		count, _ := client.LLen(ctx, key).Result()
		deviceID := key[len("ultra:queue:campaign:"):]
		metrics.QueueDetails = append(metrics.QueueDetails, QueueDetail{
			QueueName:    key,
			DeviceID:     deviceID,
			Type:         "campaign",
			MessageCount: count,
		})
		metrics.TotalMessages += count
	}
	
	// Get all sequence queues
	sequenceKeys, _ := client.Keys(ctx, "ultra:queue:sequence:*").Result()
	for _, key := range sequenceKeys {
		count, _ := client.LLen(ctx, key).Result()
		deviceID := key[len("ultra:queue:sequence:"):]
		metrics.QueueDetails = append(metrics.QueueDetails, QueueDetail{
			QueueName:    key,
			DeviceID:     deviceID,
			Type:         "sequence",
			MessageCount: count,
		})
		metrics.TotalMessages += count
	}
	
	metrics.TotalQueues = len(metrics.QueueDetails)
	
	// Get worker metrics
	metricsKeys, _ := client.Keys(ctx, "broadcast:metrics:*").Result()
	for _, key := range metricsKeys {
		deviceID := key[len("broadcast:metrics:"):]
		
		// Get all metrics for this device
		metricsData, _ := client.HGetAll(ctx, key).Result()
		
		metric := WorkerMetric{
			DeviceID: deviceID,
		}
		
		// Parse metrics
		fmt.Sscanf(metricsData["success_count"], "%d", &metric.SuccessCount)
		fmt.Sscanf(metricsData["failed_count"], "%d", &metric.FailedCount)
		
		// Parse timestamps
		if lastSuccess, ok := metricsData["last_success"]; ok {
			if ts, _ := strconv.ParseInt(lastSuccess, 10, 64); ts > 0 {
				metric.LastSuccess = time.Unix(ts, 0).Format("2006-01-02 15:04:05")
			}
		}
		
		if lastFailure, ok := metricsData["last_failure"]; ok {
			if ts, _ := strconv.ParseInt(lastFailure, 10, 64); ts > 0 {
				metric.LastFailure = time.Unix(ts, 0).Format("2006-01-02 15:04:05")
			}
		}
		
		// Calculate success rate
		total := metric.SuccessCount + metric.FailedCount
		if total > 0 {
			metric.SuccessRate = float64(metric.SuccessCount) / float64(total) * 100
		}
		
		metrics.WorkerMetrics[deviceID] = metric
	}
	
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "Redis metrics retrieved",
		"error":   false,
		"data":    metrics,
	})
}

// GetQueueMessages returns messages in a specific queue
func GetQueueMessages(c *fiber.Ctx) error {
	queueName := c.Params("queue")
	if queueName == "" {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": "Queue name required",
			"error":   true,
		})
	}
	
	// Get Redis client
	redisURL := config.RedisURL
	if redisURL == "" {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusServiceUnavailable,
			"message": "Redis not configured",
			"error":   true,
		})
	}
	
	opt, _ := redis.ParseURL(redisURL)
	client := redis.NewClient(opt)
	defer client.Close()
	
	ctx := c.Context()
	
	// Get messages from queue
	messages, err := client.LRange(ctx, queueName, 0, 99).Result() // Limit to 100 messages
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"message": "Failed to get queue messages",
			"error":   true,
		})
	}
	
	// Parse messages
	var parsedMessages []map[string]interface{}
	for _, msgData := range messages {
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(msgData), &msg); err == nil {
			parsedMessages = append(parsedMessages, msg)
		}
	}
	
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "Queue messages retrieved",
		"error":   false,
		"data": map[string]interface{}{
			"queue_name": queueName,
			"count":      len(parsedMessages),
			"messages":   parsedMessages,
		},
	})
}

// ClearQueue removes all messages from a queue
func ClearQueue(c *fiber.Ctx) error {
	// Auth check
	if err := checkAdminAuth(c); err != nil {
		return err
	}
	
	queueName := c.Params("queue")
	if queueName == "" {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusBadRequest,
			"message": "Queue name required",
			"error":   true,
		})
	}
	
	// Get Redis client
	redisURL := config.RedisURL
	if redisURL == "" {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusServiceUnavailable,
			"message": "Redis not configured",
			"error":   true,
		})
	}
	
	opt, _ := redis.ParseURL(redisURL)
	client := redis.NewClient(opt)
	defer client.Close()
	
	ctx := c.Context()
	
	// Delete the queue
	err := client.Del(ctx, queueName).Err()
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"message": "Failed to clear queue",
			"error":   true,
		})
	}
	
	logrus.Warnf("Queue %s cleared by admin", queueName)
	
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": "Queue cleared successfully",
		"error":   false,
	})
}

// checkAdminAuth verifies admin permissions
func checkAdminAuth(c *fiber.Ctx) error {
	// Simple admin check - you can enhance this
	adminToken := c.Get("X-Admin-Token")
	if adminToken != "your-secure-admin-token" {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusUnauthorized,
			"message": "Admin authorization required",
			"error":   true,
		})
	}
	return nil
}

// ExpireOldMessages manually triggers message expiration
func ExpireOldMessages(c *fiber.Ctx) error {
	// Auth check
	if err := checkAdminAuth(c); err != nil {
		return err
	}
	
	hoursStr := c.Query("hours", "24")
	hours, _ := strconv.Atoi(hoursStr)
	
	if hours < 1 || hours > 168 { // Max 1 week
		hours = 24
	}
	
	// Expire messages in database
	query := fmt.Sprintf(`
		UPDATE broadcast_messages SET status = 'expired', 
		    error_message = 'Manually expired (older than %d hours)' 
		WHERE status IN ('pending', 'queued') 
		AND created_at < NOW() - INTERVAL '%d hours'
	`, hours, hours)
	
	db := database.GetDB()
	result, err := db.Exec(query)
	if err != nil {
		return c.JSON(fiber.Map{
			"code":    fiber.StatusInternalServerError,
			"message": "Failed to expire messages",
			"error":   true,
		})
	}
	
	rowsAffected, _ := result.RowsAffected()
	
	return c.JSON(fiber.Map{
		"code":    fiber.StatusOK,
		"message": fmt.Sprintf("Expired %d messages older than %d hours", rowsAffected, hours),
		"error":   false,
		"data": map[string]interface{}{
			"expired_count": rowsAffected,
			"hours":         hours,
		},
	})
}
