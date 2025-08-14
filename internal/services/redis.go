package services

import (
	"context"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"nodepath-chat/internal/config"
)

// InitializeRedis initializes and returns a Redis client
func InitializeRedis(cfg *config.Config) *redis.Client {
	// Use Redis URL from config
	redisURL := cfg.RedisURL
	if redisURL == "" {
		// Fallback to localhost
		redisURL = "redis://localhost:6379/0"
	}
	
	// Parse Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse Redis URL")
	}
	
	// Create Redis client
	redisClient := redis.NewClient(opt)
	
	// Test connection
	ctx := context.Background()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logrus.WithError(err).Fatal("Failed to connect to Redis")
	}
	
	logrus.Info("Successfully connected to Redis")
	return redisClient
}