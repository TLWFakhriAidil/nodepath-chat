package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	domainBroadcast "github.com/aldinokemal/go-whatsapp-web-multidevice/domains/broadcast"
	"github.com/go-redis/redis/v8"
	"github.com/sirupsen/logrus"
)

type DeviceTracker struct {
	DeviceID string
	Sent     int
	Failed   int
	Limit    int
	Status   string // active, limit_reached, failed
}

type AICampaignProcessor struct {
	broadcastRepo *repository.BroadcastRepository
	leadAIRepo    repository.LeadAIRepository
	userRepo      *repository.UserRepository
	campaignRepo  repository.CampaignRepository
	redisClient   *redis.Client
}

func NewAICampaignProcessor(
	broadcastRepo *repository.BroadcastRepository,
	leadAIRepo repository.LeadAIRepository,
	userRepo *repository.UserRepository,
	campaignRepo repository.CampaignRepository,
	redisClient *redis.Client,
) *AICampaignProcessor {
	return &AICampaignProcessor{
		broadcastRepo: broadcastRepo,
		leadAIRepo:    leadAIRepo,
		userRepo:      userRepo,
		campaignRepo:  campaignRepo,
		redisClient:   redisClient,
	}
}

func (p *AICampaignProcessor) ProcessAICampaign(ctx context.Context, campaignID int) error {
	logrus.Infof("Starting AI Campaign processing for campaign ID: %d", campaignID)
	
	// 1. Get campaign details
	campaign, err := p.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return fmt.Errorf("failed to get campaign: %w", err)
	}
	
	// Verify this is an AI campaign
	if campaign.AI == nil || *campaign.AI != "ai" {
		return fmt.Errorf("campaign %d is not an AI campaign", campaignID)
	}
	
	logrus.Infof("AI Campaign: %s, Device Limit: %d", campaign.Title, campaign.Limit)
	logrus.Infof("Looking for leads with UserID: %s, Niche: %s, TargetStatus: %s", 
		campaign.UserID, campaign.Niche, campaign.TargetStatus)
	
	// 2. Get all pending AI leads based on campaign criteria
	var leads []models.LeadAI
	if campaign.TargetStatus == "all" {
		logrus.Infof("Getting AI leads with niche '%s' (all statuses)", campaign.Niche)
		leads, err = p.leadAIRepo.GetLeadAIByNiche(campaign.UserID, campaign.Niche)
	} else {
		logrus.Infof("Getting AI leads with niche '%s' and target_status '%s'", campaign.Niche, campaign.TargetStatus)
		leads, err = p.leadAIRepo.GetLeadAIByNicheAndStatus(campaign.UserID, campaign.Niche, campaign.TargetStatus)
	}
	
	if err != nil {
		return fmt.Errorf("failed to get AI leads: %w", err)
	}
	
	logrus.Infof("Found %d total AI leads before filtering", len(leads))
	
	// Filter only pending leads
	var pendingLeads []models.LeadAI
	for _, lead := range leads {
		if lead.Status == "pending" {
			pendingLeads = append(pendingLeads, lead)
		}
	}
	
	logrus.Infof("Found %d pending leads after filtering", len(pendingLeads))
	
	if len(pendingLeads) == 0 {
		logrus.Info("No pending leads found for AI campaign")
		p.campaignRepo.UpdateCampaignStatus(campaignID, "completed")
		return nil
	}
	
	logrus.Infof("Found %d pending leads to process", len(pendingLeads))
	
	// 3. Get all connected devices for the user
	devices, err := p.userRepo.GetUserDevices(campaign.UserID)
	if err != nil {
		return fmt.Errorf("failed to get user devices: %w", err)
	}
	
	// Filter only connected devices (check multiple status variations)
	var connectedDevices []*models.UserDevice
	for _, device := range devices {
		// Platform devices are always treated as online
		if device.Platform != "" {
			connectedDevices = append(connectedDevices, device)
			continue
		}
		
		if device.Status == "online" || device.Status == "Online" || 
		   device.Status == "connected" || device.Status == "Connected" {
			connectedDevices = append(connectedDevices, device)
		}
	}
	
	if len(connectedDevices) == 0 {
		logrus.Error("No connected devices available for AI campaign")
		p.campaignRepo.UpdateCampaignStatus(campaignID, "failed")
		return fmt.Errorf("no connected devices available")
	}
	
	logrus.Infof("Found %d connected devices", len(connectedDevices))
	
	// Check total capacity
	totalCapacity := len(connectedDevices) * campaign.Limit
	if len(pendingLeads) > totalCapacity {
		logrus.Warnf("Warning: %d leads but only %d total capacity across all devices", 
			len(pendingLeads), totalCapacity)
	}
	
	// 4. Initialize device tracking
	deviceStatus := make(map[string]*DeviceTracker)
	for _, device := range connectedDevices {
		deviceStatus[device.ID] = &DeviceTracker{
			DeviceID: device.ID,
			Sent:     0,
			Failed:   0,
			Limit:    campaign.Limit,
			Status:   "active",
		}
		
		// Initialize campaign progress in database
		progress := &models.AICampaignProgress{
			CampaignID: campaignID,
			DeviceID:   device.ID,
			LeadsSent:  0,
			LeadsFailed: 0,
			Status:     "active",
		}
		p.leadAIRepo.UpdateCampaignProgress(progress)
	}
	
	// 5. SEQUENTIAL Round-Robin Assignment (ONE BY ONE)
	deviceIndex := 0
	successCount := 0
	pendingCount := 0
	
	// Process each lead SEQUENTIALLY
	for i, lead := range pendingLeads {
		logrus.Debugf("Processing lead %d/%d: %s", i+1, len(pendingLeads), lead.Phone)
		
		assigned := false
		attempts := 0
		
		// Try each device in round-robin until we find one that can send
		for attempts < len(connectedDevices) && !assigned {
			device := connectedDevices[deviceIndex % len(connectedDevices)]
			tracker := deviceStatus[device.ID]
			deviceIndex++ // Move to next device for next iteration
			
			// Skip if device is failed or reached limit
			if tracker.Status == "failed" || tracker.Sent >= tracker.Limit {
				attempts++
				continue
			}
			
			// Try to send with this device
			logrus.Debugf("Attempting to send lead %s via device %s", lead.Phone, device.ID)
			
			err := p.sendLeadMessage(ctx, lead, device, campaign)
			if err != nil {
				// Device failed (probably banned) - mark it and don't retry
				logrus.Errorf("Device %s failed to send (possibly banned): %v", device.ID, err)
				tracker.Status = "failed"
				tracker.Failed++
				
				// Update device progress in database
				p.updateDeviceProgress(campaignID, device.ID, tracker)
			} else {
				// Success!
				tracker.Sent++
				successCount++
				assigned = true
				
				// Update lead with device assignment and status
				p.leadAIRepo.AssignDevice(lead.ID, device.ID)
				p.leadAIRepo.UpdateStatus(lead.ID, "sent")
				
				// Check if device reached limit
				if tracker.Sent >= tracker.Limit {
					tracker.Status = "limit_reached"
					logrus.Infof("Device %s reached its limit of %d messages", device.ID, tracker.Limit)
				}
				
				// Update progress in database
				p.updateDeviceProgress(campaignID, device.ID, tracker)
				
				logrus.Infof("Successfully sent lead %s via device %s (%d/%d)", 
					lead.Phone, device.ID, tracker.Sent, tracker.Limit)
			}
			
			attempts++
		}
		
		// If not assigned, it stays pending (don't mark as failed)
		if !assigned {
			pendingCount++
			logrus.Infof("Lead %s remains pending - no available devices", lead.Phone)
			// Status remains "pending" - we do NOT update to "failed"
		}
		
		// Add human-like delay between messages (even if failed)
		if i < len(pendingLeads)-1 { // Don't delay after last message
			delay := p.getRandomDelay(campaign.MinDelaySeconds, campaign.MaxDelaySeconds)
			logrus.Debugf("Waiting %d seconds before next message", delay)
			time.Sleep(time.Duration(delay) * time.Second)
		}
	}
	
	// 6. Update campaign status
	var campaignStatus string
	if successCount == len(pendingLeads) {
		campaignStatus = "completed"
		logrus.Infof("AI Campaign completed successfully: %d/%d messages sent", successCount, len(pendingLeads))
	} else if successCount > 0 {
		campaignStatus = "completed_with_pending"
		logrus.Infof("AI Campaign completed with pending: %d sent, %d pending", successCount, pendingCount)
	} else {
		campaignStatus = "failed"
		logrus.Errorf("AI Campaign failed: No messages could be sent")
	}
	
	p.campaignRepo.UpdateCampaignStatus(campaignID, campaignStatus)
	
	// Log final statistics
	logrus.Infof("AI Campaign %d finished:", campaignID)
	logrus.Infof("- Total leads processed: %d", len(pendingLeads))
	logrus.Infof("- Successfully sent: %d", successCount)
	logrus.Infof("- Remaining pending: %d", pendingCount)
	logrus.Infof("- Device statistics:")
	for _, device := range connectedDevices {
		tracker := deviceStatus[device.ID]
		logrus.Infof("  - Device %s: Sent=%d, Status=%s", device.ID, tracker.Sent, tracker.Status)
	}
	
	return nil
}

func (p *AICampaignProcessor) sendLeadMessage(ctx context.Context, lead models.LeadAI, device *models.UserDevice, campaign *models.Campaign) error {
	// Create broadcast message
	message := domainBroadcast.BroadcastMessage{
		// ID will be auto-generated by database
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
		MinDelay:       campaign.MinDelaySeconds,
		MaxDelay:       campaign.MaxDelaySeconds,
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
	
	// Message queued successfully
	return nil
}

func (p *AICampaignProcessor) updateDeviceProgress(campaignID int, deviceID string, tracker *DeviceTracker) {
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

func (p *AICampaignProcessor) getRandomDelay(minSeconds, maxSeconds int) int {
	if minSeconds < 5 {
		minSeconds = 5
	}
	if maxSeconds < minSeconds {
		maxSeconds = minSeconds + 10
	}
	
	return rand.Intn(maxSeconds-minSeconds+1) + minSeconds
}
