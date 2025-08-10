// Update WhatsAppMessageSender to use DeviceManager directly
// In whatsapp_message_sender.go, change this:

func (w *WhatsAppMessageSender) sendViaWhatsApp(deviceID string, msg *broadcast.BroadcastMessage) error {
	// 🔄 SELF-HEALING: Use DeviceManager for automatic refresh
	dm := multidevice.GetDeviceManager()
	waClient, err := dm.GetOrRefreshClient(deviceID)
	if err != nil {
		return fmt.Errorf("failed to get/refresh client for device %s: %v", deviceID, err)
	}
	
	// Double-check client health before sending
	if !dm.IsClientHealthy(waClient) {
		return fmt.Errorf("device %s client is not healthy after refresh", deviceID)
	}
	
	// ... rest of the sending logic remains the same
}

// IMPORTANT: Update all references from:
// - wcm := whatsapp.GetWorkerClientManager()
// - cm := whatsapp.GetClientManager()
// To:
// - dm := multidevice.GetDeviceManager()

// This ensures:
// 1. ONE manager for all 3000+ devices
// 2. NO duplicate clients
// 3. Self-healing on every message send
// 4. Works with 5-worker load balancer
// 5. Both campaigns and sequences use same system