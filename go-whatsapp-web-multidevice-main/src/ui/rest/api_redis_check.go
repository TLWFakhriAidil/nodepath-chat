package rest

import (
	"os"
	"strings"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/pkg/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
)

func (rest *App) CheckRedisStatus(c *fiber.Ctx) error {
	// Get all Redis-related env vars
	envVars := map[string]string{
		"REDIS_URL":         os.Getenv("REDIS_URL"),
		"redis_url":         os.Getenv("redis_url"),
		"RedisURL":          os.Getenv("RedisURL"),
		"REDIS_PASSWORD":    os.Getenv("REDIS_PASSWORD"),
		"REDISHOST":         os.Getenv("REDISHOST"),
		"REDIS_HOST":        os.Getenv("REDIS_HOST"),
		"REDISPORT":         os.Getenv("REDISPORT"),
		"REDIS_PORT":        os.Getenv("REDIS_PORT"),
		"config.RedisURL":   config.RedisURL,
		"config.RedisHost":  config.RedisHost,
		"config.RedisPort":  config.RedisPort,
	}
	
	// Check which broadcast manager is being used
	manager := broadcast.GetBroadcastManager()
	managerType := getManagerType(manager)
	
	// Check Redis URL validation
	redisURL := config.RedisURL
	if redisURL == "" {
		redisURL = os.Getenv("REDIS_URL")
	}
	
	validationChecks := map[string]bool{
		"not_empty":          redisURL != "",
		"no_template_vars":   !strings.Contains(redisURL, "${{"),
		"not_localhost":      !strings.Contains(redisURL, "localhost") && !strings.Contains(redisURL, "[::1]"),
		"has_redis_scheme":   strings.Contains(redisURL, "redis://") || strings.Contains(redisURL, "rediss://"),
	}
	
	// Log for debugging
	logrus.WithFields(logrus.Fields{
		"redis_url":     redisURL,
		"manager_type":  managerType,
		"validations":   validationChecks,
	}).Info("Redis status check")
	
	isRedisEnabled := managerType == "Ultra Scale Redis Manager (3000+ devices)" || managerType == "Redis Optimized Manager"
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Redis status check",
		Results: map[string]interface{}{
			"status":           "ok",
			"manager_type":     managerType,
			"environment_vars": envVars,
			"validation_checks": validationChecks,
			"final_redis_url":  redisURL,
			"is_redis_enabled": isRedisEnabled,
			"message": func() string {
				if managerType == "Ultra Scale Redis Manager (3000+ devices)" {
					return "✅ Ultra Scale Redis is properly configured and running! Ready for 3000+ devices!"
				} else if managerType == "Redis Optimized Manager" {
					return "✅ Redis is properly configured and running!"
				}
				return "❌ Redis is not being used. Check your environment variables."
			}(),
		},
	})
}
