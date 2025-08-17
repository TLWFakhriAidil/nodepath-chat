package usecase

import (
	"context"
	"fmt"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

// UltraFastAICampaignProcessor - NO DELAYS, MAXIMUM SPEED
type UltraFastAICampaignProcessor struct {
	broadcastRepo *repository.BroadcastRepository
	leadAIRepo    repository.LeadAIRepository
	userRepo      *repository.UserRepository
	campaignRepo  repository.CampaignRepository
	redisClient   *redis.Client
}

func NewUltraFastAICampaignProcessor(
	broadcastRepo *repository.BroadcastRepository,
	leadAIRepo repository.LeadAIRepository,
	userRepo *repository.UserRepository,
	campaignRepo repository.CampaignRepository,
	redisClient *redis.Client,
) *UltraFastAICampaignProcessor {
	return &UltraFastAICampaignProcessor{
		broadcastRepo: broadcastRepo,
		leadAIRepo:    leadAIRepo,
		userRepo:      userRepo,
		campaignRepo:  campaignRepo,
		redisClient:   redisClient,
	}
}

func (p *UltraFastAICampaignProcessor) ProcessAICampaign(ctx context.Context, campaignID int) error {
	logrus.Infof("🚀 Starting ULTRA FAST AI Campaign processing for campaign ID: %d", campaignID)
	
	// 1. Get campaign details
	campaign, err := p.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return fmt.Errorf("failed to get campaign: %w", err)
	}
	
	// Verify this is an AI campaign
	if campaign.AI == nil || *campaign.AI != "ai" {
		return fmt.Errorf("campaign %d is not an AI campaign", campaignID)
	}
	
	logrus.Infof("🚀 ULTRA FAST Campaign: %s, Device Limit: %d (NO SPEED LIMITS)", campaign.Title, campaign.Limit)
	
	// 2. Get all pending AI leads
	var leads []models.LeadAI
	if campaign.TargetStatus == "all" {
		leads, err = p.leadAIRepo.GetLeadAIByNiche(campaign.UserID, campaign.Niche)
	} else {
		leads, err = p.leadAIRepo.GetLeadAIByNicheAndStatus(campaign.UserID, campaign.Niche, campaign.TargetStatus)
	}
	
	if err != nil {
		return fmt.Errorf("failed to get AI leads: %w", err)
	}
	
	// Filter only pending leads
	var pendingLeads []models.LeadAI
	for _, lead := range leads {
		if lead.Status == "pending" {
			pendingLeads = append(pendingLeads, lead)
		}
	}
	
	logrus.Infof("🚀 Found %d pending leads - SENDING AT MAXIMUM SPEED", len(pendingLeads))
	
	if len(pendingLeads) == 0 {
		logrus.Info("No pending leads found")
		p.campaignRepo.UpdateCampaignStatus(campaignID, "completed")
		return nil
	}
	
	// 3. Get all devices (don't care if online or not - we'll force them online)
	devices, err := p.userRepo.GetUserDevices(campaign.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user devices: %w", err)
	}
	
	// Use ALL devices - we'll force them online
	var availableDevices []*models.UserDevice
	for _, device := range devices {
		availableDevices = append(availableDevices, device)
		logrus.Infof("🚀 Device %s added to ULTRA FAST pool (will be forced online)", device.ID)
	}
	
	if len(availableDevices) == 0 {
		logrus.Error("No devices available")
		p.campaignRepo.UpdateCampaignStatus(campaignID, "failed")
		return fmt.Errorf("no devices available")
	}
	
	logrus.Infof("🚀 Using %d devices for MAXIMUM SPEED", len(availableDevices))
	
	// 4. Initialize device tracking
	deviceStatus := make(map[string]*DeviceTracker)
	for _, device := range availableDevices {
		deviceStatus[device.ID] = &DeviceTracker{
			DeviceID: device.ID,
			Sent:     0,
			Failed:   0,
			Limit:    campaign.Limit,
			Status:   "active",
		}
		
		// Initialize campaign progress
		progress := &models.AICampaignProgress{
			CampaignID: campaignID,
			DeviceID:   device.ID,
			LeadsSent:  0,
			LeadsFailed: 0,
			Status:     "active",
		}
		p.leadAIRepo.UpdateCampaignProgress(progress)
	}
	
	// 5. SEND AT MAXIMUM SPEED - NO DELAYS
	deviceIndex := 0
	successCount := 0
	failedCount := 0
	startTime := time.Now()
	
	logrus.Infof("🚀🚀🚀 SENDING %d MESSAGES AT MAXIMUM SPEED - NO DELAYS! 🚀🚀🚀", len(pendingLeads))
	
	// Process ALL leads as fast as possible
	for i, lead := range pendingLeads {
		// Round-robin through devices
		device := availableDevices[deviceIndex % len(availableDevices)]
		tracker := deviceStatus[device.ID]
		deviceIndex++
		
		// Skip if device reached limit
		if tracker.Sent >= tracker.Limit {
			// Try next device
			attempts := 0
			for attempts < len(availableDevices) {
				device = availableDevices[deviceIndex % len(availableDevices)]
				tracker = deviceStatus[device.ID]
				deviceIndex++
				attempts++
				
				if tracker.Sent < tracker.Limit {
					break
				}
			}
			
			// If all devices at limit, just send anyway (ignore limits)
			if tracker.Sent >= tracker.Limit {
				logrus.Warnf("🚀 All devices at limit - IGNORING LIMIT and continuing with device %s", device.ID)
			}
		}
		
		// SEND IMMEDIATELY - NO DELAY
		err := p.sendLeadMessageFast(ctx, lead, device, campaign)
		if err != nil {
			logrus.Errorf("❌ Failed to send to %s via device %s: %v", lead.Phone, device.ID, err)
			tracker.Failed++
			failedCount++
			
			// Update device progress
			p.updateDeviceProgress(campaignID, device.ID, tracker)
			
			// Don't mark device as failed - keep trying
		} else {
			// Success!
			tracker.Sent++
			successCount++
			
			// Update lead with device assignment and status
			p.leadAIRepo.AssignDevice(lead.ID, device.ID)
			p.leadAIRepo.UpdateStatus(lead.ID, "sent")
			
			// Update progress
			p.updateDeviceProgress(campaignID, device.ID, tracker)
			
			// Log speed
			elapsed := time.Since(startTime)
			msgsPerSec := float64(successCount) / elapsed.Seconds()
			logrus.Infof("✅ [%d/%d] Sent to %s via device %s | Speed: %.2f msgs/sec", 
				i+1, len(pendingLeads), lead.Phone, device.ID, msgsPerSec)
		}
		
		// NO DELAY - MAXIMUM SPEED!
	}
	
	// 6. Update campaign status
	totalTime := time.Since(startTime)
	avgSpeed := float64(successCount) / totalTime.Seconds()
	
	var campaignStatus string
	if successCount == len(pendingLeads) {
		campaignStatus = "completed"
		logrus.Infof("🎯 ULTRA FAST Campaign completed: %d messages in %v (%.2f msgs/sec)", 
			successCount, totalTime, avgSpeed)
	} else if successCount > 0 {
		campaignStatus = "completed_with_errors"
		logrus.Infof("⚡ ULTRA FAST Campaign finished: %d sent, %d failed in %v (%.2f msgs/sec)", 
			successCount, failedCount, totalTime, avgSpeed)
	} else {
		campaignStatus = "failed"
		logrus.Errorf("❌ ULTRA FAST Campaign failed: No messages sent")
	}
	
	p.campaignRepo.UpdateCampaignStatus(campaignID, campaignStatus)
	
	// Log final statistics
	logrus.Infof("🚀 ULTRA FAST Campaign %d Results:", campaignID)
	logrus.Infof("🚀 - Total time: %v", totalTime)
	logrus.Infof("🚀 - Messages sent: %d", successCount)
	logrus.Infof("🚀 - Messages failed: %d", failedCount)
	logrus.Infof("🚀 - Average speed: %.2f messages/second", avgSpeed)
	logrus.Infof("🚀 - Device statistics:")
	for _, device := range availableDevices {
		tracker := deviceStatus[device.ID]
		logrus.Infof("  🚀 Device %s: Sent=%d, Failed=%d", device.ID, tracker.Sent, tracker.Failed)
	}
	
	return nil
}

func (p *UltraFastAICampaignProcessor) sendLeadMessageFast(ctx context.Context, lead models.LeadAI, device *models.UserDevice, campaign *models.Campaign) error {
	// Create broadcast message with NO DELAYS
	message := domainBroadcast.BroadcastMessage{
		UserID:         campaign.UserID,
		DeviceID:       device.ID,
		CampaignID:     &campaign.ID,
		RecipientPhone: lead.Phone,
		RecipientName:  lead.Name,
		RecipientJID:   lead.Phone + "@s.whatsapp.net",
		Type:           "text",
		Content:        campaign.Message,
		Message:        campaign.Message,
		MediaURL:       campaign.ImageURL,
		ImageURL:       campaign.ImageURL,
		Status:         "queued",
		CreatedAt:      time.Now(),
		MinDelay:       0, // NO DELAY
		MaxDelay:       0, // NO DELAY
	}
	
	// If image URL is provided, set type to image
	if campaign.ImageURL != "" {
		message.Type = "image"
	}
	
	// Queue message to database
	err := p.broadcastRepo.QueueMessage(message)
	if err != nil {
		return fmt.Errorf("failed to queue broadcast message: %w", err)
	}
	
	return nil
}

func (p *UltraFastAICampaignProcessor) updateDeviceProgress(campaignID int, deviceID string, tracker *DeviceTracker) {
	progress := &models.AICampaignProgress{
		CampaignID:  campaignID,
		DeviceID:    deviceID,
		LeadsSent:   tracker.Sent,
		LeadsFailed: tracker.Failed,
		Status:      tracker.Status,
		LastActivity: time.Now(),
	}
	
	err := p.leadAIRepo.UpdateCampaignProgress(progress)
	if err != nil {
		logrus.Errorf("Failed to update campaign progress for device %s: %v", deviceID, err)
	}
}
