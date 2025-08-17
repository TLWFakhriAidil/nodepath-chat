package usecase

import (
	"fmt"
	"time"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/database"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/broadcast"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/sirupsen/logrus"
)

// UltraOptimizedBroadcastProcessor uses broadcast-specific worker pools
type UltraOptimizedBroadcastProcessor struct {
	manager *broadcast.UltraScaleBroadcastManager
	ticker  *time.Ticker
}

// StartUltraOptimizedBroadcastProcessor starts the ultra-optimized processor
func StartUltraOptimizedBroadcastProcessor() {
	processor := &UltraOptimizedBroadcastProcessor{
		manager: broadcast.GetUltraScaleBroadcastManager(),
		ticker:  time.NewTicker(5 * time.Second), // Check every 5 seconds
	}
	
	logrus.Info("Ultra-optimized broadcast processor started (MySQL 5.7 compatible with worker locking)")
	
	// Process immediately on start
	processor.processMessages()
	
	// Then process periodically
	for range processor.ticker.C {
		processor.processMessages()
	}
}

func (p *UltraOptimizedBroadcastProcessor) processMessages() {
	// Get repository instance
	broadcastRepo := repository.GetBroadcastRepository()
	userRepo := repository.GetUserRepository()
	
	// Get all devices with pending messages
	devices, err := broadcastRepo.GetDevicesWithPendingMessages()
	if err != nil {
		logrus.Errorf("Failed to get devices with pending messages: %v", err)
		return
	}
	
	if len(devices) == 0 {
		return
	}
	
	// logrus.Infof("Found %d devices with pending messages", len(devices))
	
	messageCount := 0
	campaignPools := make(map[int]bool)
	sequencePools := make(map[string]bool)
	
	// Process each device using GetPendingMessagesAndLock for atomic locking
	for _, deviceID := range devices {
		// Use GetPendingMessagesAndLock to atomically claim messages
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
		
		// Check if device is online (platform devices always considered online)
		if device.Platform == "" && device.Status != "connected" && device.Status != "online" {
			// Skip this WhatsApp Web device - mark messages as skipped
			db := database.GetDB()
			db.Exec(`UPDATE broadcast_messages SET status = 'skipped', error_message = 'Device offline' 
					 WHERE device_id = ? AND status = 'processing'`, deviceID)
			continue
		}
		
		// Process each message
		for _, msg := range messages {
			// Create pool if needed
			if msg.CampaignID != nil && !campaignPools[*msg.CampaignID] {
				_, err := p.manager.StartBroadcastPool("campaign", fmt.Sprintf("%d", *msg.CampaignID))
				if err != nil {
					logrus.Errorf("Failed to start campaign pool: %v", err)
					continue
				}
				campaignPools[*msg.CampaignID] = true
				
				// Update campaign status to processing
				db := database.GetDB()
				db.Exec(`UPDATE campaigns SET status = 'processing', 
						 updated_at = NOW() 
						 WHERE id = ?`, *msg.CampaignID)
			}
			
			if msg.SequenceID != nil && !sequencePools[*msg.SequenceID] {
				_, err := p.manager.StartBroadcastPool("sequence", *msg.SequenceID)
				if err != nil {
					logrus.Errorf("Failed to start sequence pool: %v", err)
					continue
				}
				sequencePools[*msg.SequenceID] = true
			}
			
			// Queue message to appropriate pool
			var broadcastType, broadcastID string
			if msg.CampaignID != nil {
				broadcastType = "campaign"
				broadcastID = fmt.Sprintf("%d", *msg.CampaignID)
			} else if msg.SequenceID != nil {
				broadcastType = "sequence"
				broadcastID = *msg.SequenceID
			}
			
			if broadcastType != "" {
				err = p.manager.QueueMessageToBroadcast(broadcastType, broadcastID, &msg)
				if err != nil {
					logrus.Errorf("Failed to queue message: %v", err)
					// Update to failed
					db := database.GetDB()
					db.Exec(`UPDATE broadcast_messages SET status = 'failed', error_message = ? WHERE id = ?`, 
						err.Error(), msg.ID)
				} else {
					messageCount++
				}
			}
		}
	}
	
	if messageCount > 0 {
		logrus.Infof("Queued %d messages to broadcast pools (with worker locking)", messageCount)
	}
}
