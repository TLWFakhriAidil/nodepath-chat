package usecase

import (
	"time"
	"math/rand"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
)

// StartBroadcastWorkerProcessor starts the background worker that processes queued messages
func StartBroadcastWorkerProcessor() {
	logrus.Info("Starting broadcast worker processor...")
	
	// Get repositories
	broadcastRepo := repository.GetBroadcastRepository()
	userRepo := repository.GetUserRepository()
	
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
			
			logrus.Infof("Found %d devices with pending messages", len(devices))
			
			// Process each device
			for _, deviceID := range devices {
				// Get pending messages for this device
				messages, err := broadcastRepo.GetPendingMessages(deviceID, 100) // Increased from 10 to 100 for better throughput
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
					logrus.Infof("Processing %d messages for platform device %s (%s)", 
						len(messages), deviceID, device.Platform)
				} else {
					// WhatsApp Web device - check if online
					if device.Status != "online" && device.Status != "Online" && 
					   device.Status != "connected" && device.Status != "Connected" {
						logrus.Debugf("Device %s is not online (status: %s), skipping", deviceID, device.Status)
						continue
					}
					logrus.Infof("Processing %d messages for WhatsApp device %s", len(messages), deviceID)
				}
				
				// Process messages directly
				// Create self-healing message sender
				messageSender := broadcast.NewWhatsAppMessageSender()
				
				for i, msg := range messages {
					// Add delay BEFORE sending (except for first message)
					if i > 0 {
						if msg.MinDelay > 0 && msg.MaxDelay > 0 {
							// Calculate random delay between min and max
							minDelay := msg.MinDelay
							maxDelay := msg.MaxDelay
							
							// Random delay between min and max
							randomDelay := minDelay
							if maxDelay > minDelay {
								// Use proper random number generation
								delayRange := maxDelay - minDelay
								randomDelay = minDelay + rand.Intn(delayRange+1)
							}
							
							delay := time.Duration(randomDelay) * time.Second
							logrus.Infof("Applying delay of %d seconds before sending to %s", randomDelay, msg.RecipientPhone)
							time.Sleep(delay)
						} else {
							// Default delay
							logrus.Info("Applying default 5 second delay")
							time.Sleep(5 * time.Second)
						}
					} else {
						logrus.Infof("First message in batch - no delay for %s", msg.RecipientPhone)
					}
					
					// Update status to processing
					err := broadcastRepo.UpdateMessageStatus(msg.ID, "processing", "")
					if err != nil {
						logrus.Errorf("Failed to update message status: %v", err)
						continue
					}
					
					// SELF-HEALING: Use WhatsAppMessageSender instead of manager
					// This ensures WhatsApp devices go through connection refresh
					err = messageSender.SendMessage(msg.DeviceID, &msg)
					if err != nil {
						logrus.Errorf("Failed to send message %s: %v", msg.ID, err)
						// Update status to failed
						broadcastRepo.UpdateMessageStatus(msg.ID, "failed", err.Error())
					} else {
						logrus.Infof("Successfully sent message %s to %s", msg.ID, msg.RecipientPhone)
						// Update status to sent
						broadcastRepo.UpdateMessageStatus(msg.ID, "sent", "")
					}
				}
			}
		}
	}
}
