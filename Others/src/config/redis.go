package config

import "fmt"

// GetRedisURL returns the Redis connection URL
func GetRedisURL() string {
	// First check for full Redis URL
	if RedisURL != "" {
		return RedisURL
	}
	
	// Build URL from components
	if RedisPassword != "" {
		return fmt.Sprintf("redis://default:%s@%s:%s", RedisPassword, RedisHost, RedisPort)
	}
	
	return fmt.Sprintf("redis://%s:%s", RedisHost, RedisPort)
}
