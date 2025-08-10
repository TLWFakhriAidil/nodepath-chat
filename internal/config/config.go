package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	Port   int
	AppEnv string

	// Database configuration
	MySQLHost     string
	MySQLPort     int
	MySQLUser     string
	MySQLPassword string
	MySQLDatabase string

	// Redis configuration
	RedisURL string

	// WhatsApp configuration
	WhatsAppStoragePath string
	WhatsAppSessionDir  string

	// OpenRouter configuration
	OpenRouterDefaultKey string

	// Security configuration
	JWTSecret     string
	SessionSecret string
}

// Load loads configuration from environment variables
func Load() *Config {
	cfg := &Config{
		// Server configuration - Railway sets PORT at runtime
		Port:   getEnvAsInt("PORT", 8080),
		AppEnv: getEnv("APP_ENV", "development"),

		// Database configuration
		MySQLHost:     getEnv("MYSQL_HOST", "159.89.198.71"),
		MySQLPort:     getEnvAsInt("MYSQL_PORT", 3306),
		MySQLUser:     getEnv("MYSQL_USER", "admin_aqil"),
		MySQLPassword: getEnv("MYSQL_PASSWORD", "admin_aqil"),
		MySQLDatabase: getEnv("MYSQL_DATABASE", "admin_railway"),

		// Redis configuration
		RedisURL: getEnv("REDIS_URL", ""),

		// WhatsApp configuration
		WhatsAppStoragePath: getEnv("WHATSAPP_STORAGE_PATH", "./whatsapp_sessions"),
		WhatsAppSessionDir:  getEnv("WHATSAPP_SESSION_DIR", "./whatsapp_sessions"),

		// OpenRouter configuration
		OpenRouterDefaultKey: getEnv("OPENROUTER_DEFAULT_KEY", ""),

		// Security configuration
		JWTSecret:     getEnv("JWT_SECRET", "your-jwt-secret-key"),
		SessionSecret: getEnv("SESSION_SECRET", "your-session-secret-key"),
	}

	return cfg
}

// GetDSN returns the MySQL DSN connection string
func (c *Config) GetDSN() string {
	return c.MySQLUser + ":" + c.MySQLPassword + "@tcp(" + c.MySQLHost + ":" + strconv.Itoa(c.MySQLPort) + ")/" + c.MySQLDatabase + "?charset=utf8mb4&parseTime=True&loc=Local&collation=utf8mb4_unicode_ci"
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