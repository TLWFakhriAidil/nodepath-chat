package whatsapp

import (
	"os"
	"github.com/sirupsen/logrus"
)

// IsMockMode checks if the system should run in mock mode
func IsMockMode() bool {
	return os.Getenv("MOCK_MODE") == "true"
}

// InitializeMockMode sets up mock clients for testing
func InitializeMockMode() {
	if !IsMockMode() {
		return
	}
	
	logrus.Warn("===========================================")
	logrus.Warn("SYSTEM RUNNING IN MOCK MODE")
	logrus.Warn("No real WhatsApp messages will be sent!")
	logrus.Warn("===========================================")
	
	// Override client creation to use mock clients
	// This would be integrated into the existing ClientManager
}

// GetTestingStats returns current testing statistics
func GetTestingStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	if !IsMockMode() {
		stats["mock_mode"] = false
		return stats
	}
	
	// Collect stats from all mock clients
	mockMgr := GetMockClientManager()
	clientCount := 0
	totalMessages := 0
	
	mockMgr.clients.Range(func(key, value interface{}) bool {
		clientCount++
		if client, ok := value.(*MockWhatsAppClient); ok {
			clientStats := client.GetStats()
			if messages, ok := clientStats["messages_sent"].(int); ok {
				totalMessages += messages
			}
		}
		return true
	})
	
	stats["mock_mode"] = true
	stats["mock_clients"] = clientCount
	stats["total_messages_simulated"] = totalMessages
	
	return stats
}
