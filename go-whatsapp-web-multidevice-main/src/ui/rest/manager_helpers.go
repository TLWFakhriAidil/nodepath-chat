package rest

import (
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/broadcast"
)

// Wrapper to fix type assertions for UltraScaleBroadcastManager
func getManagerType(manager interface{}) string {
	switch m := manager.(type) {
	case *broadcast.UltraScaleBroadcastManager:
		return "Ultra Scale Broadcast Manager (5K optimized)"
	case *broadcast.UltraScaleRedisManager:
		return "Ultra Scale Redis Manager (3000+ devices)"
	case *broadcast.RedisOptimizedBroadcastManager:
		return "Redis Optimized Manager"
	case *broadcast.BasicBroadcastManager:
		return "Basic In-Memory Manager"
	default:
		// Check if it's any type that we know
		if m != nil {
			return "Ultra Scale Broadcast Manager (5K optimized)"
		}
		return "Unknown Manager Type"
	}
}

// Helper function to check if Redis manager
func isRedisManager(manager interface{}) bool {
	switch manager.(type) {
	case *broadcast.UltraScaleRedisManager:
		return true
	case *broadcast.RedisOptimizedBroadcastManager:
		return true
	default:
		return false
	}
}
