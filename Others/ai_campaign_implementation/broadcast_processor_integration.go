// Modified broadcast processor logic to handle AI campaigns
// This should be integrated into the existing broadcast system

// In the main campaign trigger/processing function:
func ProcessCampaign(campaignID int) error {
    // Get campaign details
    campaign, err := campaignRepo.GetCampaignByID(campaignID)
    if err != nil {
        return err
    }
    
    // Check if this is an AI campaign
    if campaign.AI != nil && *campaign.AI == "ai" {
        // Use AI Campaign Processor
        leadAIRepo := repository.GetLeadAIRepository()
        deviceRepo := repository.GetDeviceRepository()
        broadcastRepo := repository.GetBroadcastRepository()
        redisClient := config.GetRedisClient()
        
        processor := NewAICampaignProcessor(
            broadcastRepo,
            leadAIRepo,
            deviceRepo,
            campaignRepo,
            redisClient,
        )
        
        return processor.ProcessAICampaign(context.Background(), campaignID)
    }
    
    // Otherwise, use regular campaign processing
    return processRegularCampaign(campaignID)
}