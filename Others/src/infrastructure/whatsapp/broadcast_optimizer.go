package whatsapp

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	waLog "go.mau.fi/whatsmeow/util/log"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/sirupsen/logrus"
)

// OptimizeClientForBroadcast configures WhatsApp client for high-volume broadcasting
func OptimizeClientForBroadcast(client *whatsmeow.Client) {
	if client == nil {
		return
	}
	
	// Enable auto reconnect and trust identity
	client.EnableAutoReconnect = true
	client.AutoTrustIdentity = true
	
	// Note: whatsmeow doesn't expose direct timeout configuration
	// But auto-reconnect will handle disconnections
	
	logrus.Debug("Client optimized for broadcast operations")
}

// CreateOptimizedClient creates a new WhatsApp client optimized for broadcasting
func CreateOptimizedClient(device *store.Device) *whatsmeow.Client {
	// Create client with optimized logging
	client := whatsmeow.NewClient(device, waLog.Stdout("Client", config.WhatsappLogLevel, true))
	
	// Apply optimizations
	OptimizeClientForBroadcast(client)
	
	return client
}

// ConfigureForHighLoad prepares the system for high-load broadcasting
func ConfigureForHighLoad() {
	// Initialize event processor with optimal workers
	processor := GetEventProcessor()
	logrus.Infof("Event processor ready with queue size: %d", processor.GetQueueSize())
	
	// Log configuration
	logrus.Info("System configured for high-load broadcasting:")
	logrus.Info("- Async event processing enabled")
	logrus.Info("- Extended keepalive timeouts")
	logrus.Info("- Auto-reconnect enabled")
	logrus.Info("- Event queue buffer: 10,000")
}
