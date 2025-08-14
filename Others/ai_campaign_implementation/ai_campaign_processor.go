package usecase

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
	
	"go-whatsapp-web-multidevice/models"
	"go-whatsapp-web-multidevice/repository"
	"go-whatsapp-web-multidevice/utils"
	"github.com/redis/go-redis/v9"
)

type DeviceTracker struct {
	DeviceID string
	Sent     int
	Failed   int
	Limit    int
	Status   string // active, limit_reached, failed
}

type AICampaignProcessor struct {
	broadcastRepo repository.BroadcastRepository
	leadAIRepo    repository.LeadAIRepository
	deviceRepo    repository.DeviceRepository
	campaignRepo  repository.CampaignRepository
	redisClient   *redis.Client
	log           utils.Logger
}

func NewAICampaignProcessor(
	broadcastRepo repository.BroadcastRepository,
	leadAIRepo repository.LeadAIRepository,
	deviceRepo repository.DeviceRepository,
	campaignRepo repository.CampaignRepository,
	redisClient *redis.Client,
) *AICampaignProcessor {
	return &AICampaignProcessor{
		broadcastRepo: broadcastRepo,
		leadAIRepo:    leadAIRepo,
		deviceRepo:    deviceRepo,
		campaignRepo:  campaignRepo,
		redisClient:   redisClient,
		log:           utils.GetLogger(),
	}
}
func (p *AICampaignProcessor) ProcessAICampaign(ctx context.Context, campaignID int) error {
	p.log.Infof("Starting AI campaign processing for campaign ID: %d", campaignID)
	
	// 1. Get campaign details
	campaign, err := p.campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		p.log.Errorf("Failed to get campaign: %v", err)
		return fmt.Errorf("failed to get campaign: %w", err)
	}
	
	// Verify this is an AI campaign
	if campaign.AI == nil || *campaign.AI != "ai" {
		return fmt.Errorf("campaign %d is not an AI campaign", campaignID)
	}
	
	// 2. Get all pending AI leads based on campaign criteria
	var leads []models.LeadAI
	if campaign.TargetStatus == "all" {
		leads, err = p.leadAIRepo.GetLeadAIByNiche(campaign.UserID, campaign.Niche)
	} else {
		leads, err = p.leadAIRepo.GetLeadAIByNicheAndStatus(campaign.UserID, campaign.Niche, campaign.TargetStatus)
	}
	
	if err != nil {
		p.log.Errorf("Failed to get AI leads: %v", err)
		return fmt.Errorf("failed to get AI leads: %w", err)
	}
	
	// Filter only pending leads
	var pendingLeads []models.LeadAI
	for _, lead := range leads {
		if lead.Status == "pending" {
			pendingLeads = append(pendingLeads, lead)
		}
	}
	
	p.log.Infof("Found %d pending AI leads for campaign %d", len(pendingLeads), campaignID)
	
	if len(pendingLeads) == 0 {
		p.log.Warnf("No pending leads found for AI campaign %d", campaignID)
		return nil
	}
	// 3. Get all connected devices for the user
	devices, err := p.deviceRepo.GetConnectedDevices(campaign.UserID)
	if err != nil {
		p.log.Errorf("Failed to get connected devices: %v", err)
		return fmt.Errorf("failed to get connected devices: %w", err)
	}
	
	if len(devices) == 0 {
		p.log.Errorf("No connected devices found for AI campaign %d", campaignID)
		// Mark campaign as failed
		p.campaignRepo.UpdateCampaignStatus(campaignID, "failed")
		return fmt.Errorf("no connected devices available")
	}
	
	p.log.Infof("Found %d connected devices for campaign %d", len(devices), campaignID)
	
	// 4. Initialize device tracking
	deviceStatus := make(map[string]*DeviceTracker)
	for _, device := range devices {
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
	// 5. Start round-robin assignment
	deviceIndex := 0
	successCount := 0
	failCount := 0
	
	// Create a wait group for concurrent processing
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 5) // Limit concurrent sends
	
	for _, lead := range pendingLeads {
		assigned := false
		attempts := 0
		
		// Try to assign to a device
		for attempts < len(devices) && !assigned {
			device := devices[deviceIndex%len(devices)]
			tracker := deviceStatus[device.ID]
			
			// Check if device can accept more leads
			if tracker.Status == "active" && tracker.Sent < tracker.Limit {
				// Assign lead to device
				wg.Add(1)
				semaphore <- struct{}{} // Acquire semaphore
				
				go func(l models.LeadAI, d models.Device, t *DeviceTracker) {
					defer wg.Done()
					defer func() { <-semaphore }() // Release semaphore
					
					err := p.assignAndSendLead(ctx, l, d, campaign)
					if err != nil {
						p.log.Errorf("Failed to send lead %d to device %s: %v", l.ID, d.ID, err)
						t.Failed++
						failCount++
						
						// Mark device as failed after 3 consecutive failures
						if t.Failed > 3 {
							t.Status = "failed"
							p.updateDeviceProgress(campaignID, d.ID, t)
						}
						
						// Update lead status to failed
						p.leadAIRepo.UpdateStatus(l.ID, "failed")
					} else {
						t.Sent++
						successCount++
						assigned = true
						
						// Update lead with device assignment
						p.leadAIRepo.AssignDevice(l.ID, d.ID)
						p.leadAIRepo.UpdateStatus(l.ID, "sent")
						
						// Check if device reached limit
						if t.Sent >= t.Limit {
							t.Status = "limit_reached"
						}
						
						// Update progress in database
						p.updateDeviceProgress(campaignID, d.ID, t)
					}
				}(lead, *device, tracker)
				
				assigned = true // Mark as assigned to prevent multiple attempts
			}
			
			deviceIndex++
			attempts++
		}
		// If not assigned after all attempts, mark as failed
		if !assigned {
			p.leadAIRepo.UpdateStatus(lead.ID, "failed")
			failCount++
			p.log.Warnf("Could not assign lead %d to any device - all devices at limit or failed", lead.ID)
		}
		
		// Add a small delay between lead assignments
		time.Sleep(time.Duration(rand.Intn(500)+500) * time.Millisecond)
	}
	
	// Wait for all goroutines to complete
	wg.Wait()
	
	// 6. Update campaign status
	var campaignStatus string
	if successCount > 0 {
		if failCount > 0 {
			campaignStatus = "completed_with_errors"
		} else {
			campaignStatus = "completed"
		}
	} else {
		campaignStatus = "failed"
	}
	
	p.campaignRepo.UpdateCampaignStatus(campaignID, campaignStatus)
	
	p.log.Infof("AI Campaign %d completed. Success: %d, Failed: %d, Status: %s", 
		campaignID, successCount, failCount, campaignStatus)
	
	// 7. Generate summary report
	p.generateCampaignReport(campaignID)
	
	return nil
}
func (p *AICampaignProcessor) assignAndSendLead(ctx context.Context, lead models.LeadAI, device models.Device, campaign *models.Campaign) error {
	// Create broadcast message
	message := &models.BroadcastMessage{
		CampaignID: campaign.ID,
		DeviceID:   device.ID,
		RecipientJID: lead.Phone + "@s.whatsapp.net",
		Message:    campaign.Message,
		ImageURL:   campaign.ImageURL,
		Status:     "queued",
		CreatedAt:  time.Now(),
	}
	
	// Save message to database
	err := p.broadcastRepo.CreateMessage(message)
	if err != nil {
		return fmt.Errorf("failed to create broadcast message: %w", err)
	}
	
	// Queue message for sending via Redis
	queueKey := fmt.Sprintf("broadcast:device:%s:queue", device.ID)
	messageData := fmt.Sprintf("%d|%s|%s|%s", 
		message.ID, 
		message.RecipientJID, 
		message.Message,
		message.ImageURL,
	)
	
	err = p.redisClient.RPush(ctx, queueKey, messageData).Err()
	if err != nil {
		// Update message status to failed
		p.broadcastRepo.UpdateMessageStatus(message.ID, "failed")
		return fmt.Errorf("failed to queue message: %w", err)
	}
	
	// Add delay between messages (human-like behavior)
	minDelay := campaign.MinDelaySeconds
	maxDelay := campaign.MaxDelaySeconds
	if minDelay < 5 {
		minDelay = 5
	}
	if maxDelay < minDelay {
		maxDelay = minDelay + 10
	}
	
	delay := rand.Intn(maxDelay-minDelay) + minDelay
	time.Sleep(time.Duration(delay) * time.Second)
	
	return nil
}
func (p *AICampaignProcessor) updateDeviceProgress(campaignID int, deviceID string, tracker *DeviceTracker) {
	progress := &models.AICampaignProgress{
		CampaignID:  campaignID,
		DeviceID:    deviceID,
		LeadsSent:   tracker.Sent,
		LeadsFailed: tracker.Failed,
		Status:      tracker.Status,
	}
	
	err := p.leadAIRepo.UpdateCampaignProgress(progress)
	if err != nil {
		p.log.Errorf("Failed to update campaign progress for device %s: %v", deviceID, err)
	}
}

func (p *AICampaignProcessor) generateCampaignReport(campaignID int) {
	// Get campaign progress for all devices
	progresses, err := p.leadAIRepo.GetCampaignProgress(campaignID)
	if err != nil {
		p.log.Errorf("Failed to get campaign progress: %v", err)
		return
	}
	
	totalSent := 0
	totalFailed := 0
	
	p.log.Infof("=== AI Campaign %d Report ===", campaignID)
	for _, progress := range progresses {
		totalSent += progress.LeadsSent
		totalFailed += progress.LeadsFailed
		p.log.Infof("Device %s: Sent=%d, Failed=%d, Status=%s", 
			progress.DeviceID, progress.LeadsSent, progress.LeadsFailed, progress.Status)
	}
	
	p.log.Infof("Total: Sent=%d, Failed=%d", totalSent, totalFailed)
	p.log.Infof("=== End of Report ===")
}