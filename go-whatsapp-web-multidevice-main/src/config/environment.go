package config

import (
	"os"
	"strings"
)

func InitEnvironment() {
	// Override with environment variables if they exist
	if port := os.Getenv("APP_PORT"); port != "" {
		AppPort = port
	}
	
	if port := os.Getenv("PORT"); port != "" {
		// Railway provides PORT env var
		AppPort = port
	}
	
	if debug := os.Getenv("APP_DEBUG"); debug == "true" {
		AppDebug = true
	}
	
	if dbUri := os.Getenv("DB_URI"); dbUri != "" {
		DBURI = dbUri
	}
	
	// For Railway deployment, use DATABASE_URL for WhatsApp if DB_URI is MySQL
	if strings.Contains(DBURI, "mysql") {
		if postgresURL := os.Getenv("DATABASE_URL"); postgresURL != "" {
			// Override DBURI with PostgreSQL for WhatsApp initialization
			DBURI = postgresURL
		}
	}
	
	if basicAuth := os.Getenv("APP_BASIC_AUTH"); basicAuth != "" {
		AppBasicAuthCredential = strings.Split(basicAuth, ",")
	}
	
	if autoReply := os.Getenv("WHATSAPP_AUTO_REPLY"); autoReply != "" {
		WhatsappAutoReplyMessage = autoReply
	}
	
	if webhook := os.Getenv("WHATSAPP_WEBHOOK"); webhook != "" {
		WhatsappWebhook = strings.Split(webhook, ",")
	}
	
	if webhookSecret := os.Getenv("WHATSAPP_WEBHOOK_SECRET"); webhookSecret != "" {
		WhatsappWebhookSecret = webhookSecret
	}
	
	if validation := os.Getenv("WHATSAPP_ACCOUNT_VALIDATION"); validation == "false" {
		WhatsappAccountValidation = false
	}
	
	if storage := os.Getenv("WHATSAPP_CHAT_STORAGE"); storage == "true" {
		WhatsappChatStorage = true
	} else {
		WhatsappChatStorage = false
	}
	
	// Redis settings - check multiple possible env var names
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = os.Getenv("redis_url")
	}
	if redisURL == "" {
		redisURL = os.Getenv("RedisURL")
	}
	// Only set if it's not a template variable
	if redisURL != "" && !strings.Contains(redisURL, "${{") {
		RedisURL = redisURL
	}
	
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		RedisPassword = redisPassword
	}
	
	// Check both REDIS_HOST and REDISHOST
	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		RedisHost = redisHost
	}
	if redisHost := os.Getenv("REDISHOST"); redisHost != "" {
		RedisHost = redisHost
	}
	
	// Check both REDIS_PORT and REDISPORT
	if redisPort := os.Getenv("REDIS_PORT"); redisPort != "" {
		RedisPort = redisPort
	}
	if redisPort := os.Getenv("REDISPORT"); redisPort != "" {
		RedisPort = redisPort
	}
}
