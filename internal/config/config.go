package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all configuration for the application with high-performance settings
type Config struct {
	// Server configuration
	Port   int
	AppEnv string

	// Database configuration
	MySQLURI string // Primary MySQL connection URI

	// Redis configuration
	RedisURL          string
	RedisClusterAddrs []string // Support for Redis clustering

	// WhatsApp configuration
	WhatsAppStoragePath string
	WhatsAppSessionDir  string
	WhatsAppMaxDevices  int // Support for multiple devices

	// OpenRouter configuration
	OpenRouterDefaultKey string
	OpenRouterTimeout    int // Configurable timeout
	OpenRouterMaxRetries int // Max retries for AI requests

	// Security configuration
	JWTSecret     string
	SessionSecret string

	// Performance configuration
	MaxConcurrentUsers int    // Maximum concurrent users
	WebSocketEnabled   bool   // Enable WebSocket support
	CDNEnabled         bool   // Enable CDN for media files
	CDNBaseURL         string // CDN base URL
}

// Load loads configuration from environment variables with performance optimizations
func Load() *Config {
	cfg := &Config{
		// Server configuration - Railway sets PORT at runtime
		Port:   getEnvAsInt("PORT", 8080),
		AppEnv: getEnv("APP_ENV", "development"),

		// Database configuration
		MySQLURI: getEnv("MYSQL_URI", ""), // Primary MySQL connection

		// Redis configuration with clustering support
		RedisURL:          getEnv("REDIS_URL", ""),
		RedisClusterAddrs: getEnvAsSlice("REDIS_CLUSTER_ADDRS", ","),

		// WhatsApp configuration with multi-device support
		WhatsAppStoragePath: getEnv("WHATSAPP_STORAGE_PATH", "./whatsapp_sessions"),
		WhatsAppSessionDir:  getEnv("WHATSAPP_SESSION_DIR", "./whatsapp_sessions"),
		WhatsAppMaxDevices:  getEnvAsInt("WHATSAPP_MAX_DEVICES", 10),

		// OpenRouter configuration with performance settings
		OpenRouterDefaultKey: getEnv("OPENROUTER_DEFAULT_KEY", ""),
		OpenRouterTimeout:    getEnvAsInt("OPENROUTER_TIMEOUT", 15), // Reduced from 30s
		OpenRouterMaxRetries: getEnvAsInt("OPENROUTER_MAX_RETRIES", 2),

		// Security configuration
		JWTSecret:     getEnv("JWT_SECRET", "your-jwt-secret-key"),
		SessionSecret: getEnv("SESSION_SECRET", "your-session-secret-key"),

		// Performance configuration for 3000+ concurrent users
		MaxConcurrentUsers: getEnvAsInt("MAX_CONCURRENT_USERS", 5000),
		WebSocketEnabled:   getEnvAsBool("WEBSOCKET_ENABLED", true),
		CDNEnabled:         getEnvAsBool("CDN_ENABLED", false),
		CDNBaseURL:         getEnv("CDN_BASE_URL", ""),
	}

	return cfg
}

// GetDSN returns the MySQL DSN connection string
// Format: mysql://user:password@host:port/database
func (c *Config) GetDSN() string {
	if c.MySQLURI == "" {
		return "" // Return empty if no database URL provided
	}

	// Convert mysql:// to proper DSN format if needed
	if strings.HasPrefix(c.MySQLURI, "mysql://") {
		// Remove mysql:// prefix and add tcp() wrapper
		dsn := strings.TrimPrefix(c.MySQLURI, "mysql://")
		// Parse user:password@host:port/database format
		parts := strings.Split(dsn, "/")
		if len(parts) >= 2 {
			userHostPart := parts[0]
			databasePart := parts[1]
			// Split user:password@host:port
			atIndex := strings.LastIndex(userHostPart, "@")
			if atIndex > 0 {
				userPass := userHostPart[:atIndex]
				hostPort := userHostPart[atIndex+1:]
				// Reconstruct with tcp() wrapper for go-sql-driver/mysql
				dsn = userPass + "@tcp(" + hostPort + ")/" + databasePart
				if !strings.Contains(dsn, "?") {
					dsn += "?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
				} else {
					dsn += "&charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
				}
				return dsn
			}
		}
	}

	// Return as-is if already in proper format
	return c.MySQLURI
}

// IsProduction returns true if the app is running in production
func (c *Config) IsProduction() bool {
	return c.AppEnv == "production"
}

// IsDevelopment returns true if the app is running in development
func (c *Config) IsDevelopment() bool {
	return c.AppEnv == "development"
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as an integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// getEnvAsBool gets an environment variable as a boolean with a fallback value
func getEnvAsBool(key string, fallback bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

// getEnvAsSlice gets an environment variable as a slice with a separator
func getEnvAsSlice(key, separator string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, separator)
	}
	return []string{}
}
