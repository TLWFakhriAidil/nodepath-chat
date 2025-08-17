package config

import (
	"os"
	"strconv"
)

// StabilityConfig contains all stability-related settings
type StabilityConfig struct {
	// UltraStableMode - when true, devices NEVER disconnect
	UltraStableMode bool
	
	// ForceReconnectAttempts - how many times to try reconnecting
	ForceReconnectAttempts int
	
	// KeepAliveInterval - how often to send keep-alive (seconds)
	KeepAliveInterval int
	
	// IgnoreRateLimits - when true, ignore all WhatsApp rate limits
	IgnoreRateLimits bool
	
	// MaxSpeedMode - when true, send at maximum possible speed
	MaxSpeedMode bool
	
	// DisableDelays - when true, remove ALL delays between messages
	DisableDelays bool
	
	// ForceOnlineStatus - when true, always report devices as online
	ForceOnlineStatus bool
}

var stabilityConfig *StabilityConfig

// GetStabilityConfig returns the stability configuration
func GetStabilityConfig() *StabilityConfig {
	if stabilityConfig == nil {
		stabilityConfig = &StabilityConfig{
			UltraStableMode:        getEnvBool("ULTRA_STABLE_MODE", true),
			ForceReconnectAttempts: getEnvInt("FORCE_RECONNECT_ATTEMPTS", 100),
			KeepAliveInterval:      getEnvInt("KEEP_ALIVE_INTERVAL", 5),
			IgnoreRateLimits:       getEnvBool("IGNORE_RATE_LIMITS", true),
			MaxSpeedMode:           getEnvBool("MAX_SPEED_MODE", true),
			DisableDelays:          getEnvBool("DISABLE_DELAYS", true),
			ForceOnlineStatus:      getEnvBool("FORCE_ONLINE_STATUS", true),
		}
	}
	return stabilityConfig
}

func getEnvBool(key string, defaultValue bool) bool {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	b, err := strconv.ParseBool(val)
	if err != nil {
		return defaultValue
	}
	return b
}

func getEnvInt(key string, defaultValue int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultValue
	}
	return i
}
