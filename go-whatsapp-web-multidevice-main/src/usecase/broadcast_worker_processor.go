package usecase

import (
	"time"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
)

// StartBroadcastWorkerProcessor starts the background worker that processes queued messages
// FIXED VERSION: Now uses Worker Pool System to prevent duplicate messages
func StartBroadcastWorkerProcessor() {
	logrus.Info("Starting broadcast worker processor with Worker Pool System...")
	
	// Get repositories
	broadcastRepo := repository.GetBroadcastRepository()
	userRepo := repository.GetUserRepository()
	
	// Get the Ultra Scale Broadcast Manager (Worker Pool System)
	broadcastManager := broadcast.GetBroadcastManager()
	
	// Process messages every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			// Get all devices with pending messages
			devices, err := broadcastRepo.GetDevicesWithPendingMessages()
			if err != nil {
				logrus.Errorf("Failed to get devices with pending messages: %v", err)
				continue
			}
			
			if len(devices) == 0 {
				continue // No pending messages
			}
			
			// logrus.Infof("Found %d devices with pending messages", len(devices))
			
			// Process each device
			for _, deviceID := range devices {
				// Get pending messages for this device
				messages, err := broadcastRepo.GetPendingMessagesAndLock(deviceID, 100)
				if err != nil {
					logrus.Errorf("Failed to get pending messages for device %s: %v", deviceID, err)
					continue
				}
				
				if len(messages) == 0 {
					continue
				}
				
				// Get device details
				device, err := userRepo.GetDeviceByID(deviceID)
				if err != nil {
					logrus.Errorf("Failed to get device %s: %v", deviceID, err)
					continue
				}
				
				// PLATFORM FIX: Skip status check for platform devices
				if device.Platform != "" {
					// Platform device (Wablas/Whacenter) - always available via API
					logrus.Infof("Processing %d messages for platform device %s (%s) using Worker Pool", 
						len(messages), deviceID, device.Platform)
				} else {
					// WhatsApp Web device - check if online
					if device.Status != "online" && device.Status != "Online" && 
					   device.Status != "connected" && device.Status != "Connected" {
						logrus.Debugf("Device %s is not online (status: %s), skipping", deviceID, device.Status)
						continue
					}
					logrus.Infof("Processing %d messages for WhatsApp device %s using Worker Pool", 
						len(messages), deviceID)
				}
				
				// FIXED: Queue messages to Worker Pool instead of direct processing
				// This prevents duplicate messages by using channel-based queuing
				for _, msg := range messages {
					// Worker Pool will handle:
					// 1. Status updates (pending -> queued -> sent)
					// 2. Anti-spam delays
					// 3. Sequential sending with mutex lock
					// 4. Automatic retries on failure
					
					err := broadcastManager.QueueMessage(&msg)
					if err != nil {
						logrus.Errorf("Failed to queue message %s to worker pool: %v", msg.ID, err)
						// Update status to failed if can't queue
						broadcastRepo.UpdateMessageStatus(msg.ID, "failed", err.Error())
					} else {
						logrus.Debugf("Message %s queued to worker pool for device %s", msg.ID, deviceID)
					}
				}
				
				logrus.Infof("Queued %d messages to worker pool for device %s", len(messages), deviceID)
			}
		}
	}
}
