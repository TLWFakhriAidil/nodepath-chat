// TriggerAICampaign - Handler to manually trigger an AI campaign
func (handler *App) TriggerAICampaign(c *fiber.Ctx) error {
	// Get session from cookie
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token",
		})
	}
	
	userRepo := repository.GetUserRepository()
	session, err := userRepo.GetSession(sessionToken)
	if err != nil {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "Invalid session",
		})
	}
	
	campaignID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid campaign ID",
		})
	}
	
	campaignRepo := repository.GetCampaignRepository()
	campaign, err := campaignRepo.GetCampaignByID(campaignID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Campaign not found",
		})
	}
	// Verify campaign belongs to user
	if campaign.UserID != session.UserID {
		return c.Status(403).JSON(utils.ResponseData{
			Status:  403,
			Code:    "FORBIDDEN",
			Message: "You don't have permission to trigger this campaign",
		})
	}
	
	// Verify this is an AI campaign
	if campaign.AI == nil || *campaign.AI != "ai" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "NOT_AI_CAMPAIGN",
			Message: "This is not an AI campaign",
		})
	}
	
	// Check if campaign is already running
	if campaign.Status != "pending" && campaign.Status != "failed" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "CAMPAIGN_RUNNING",
			Message: "Campaign is already running or completed",
		})
	}
	
	// Update campaign status to triggered
	err = campaignRepo.UpdateCampaignStatus(campaignID, "triggered")
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "UPDATE_FAILED",
			Message: "Failed to update campaign status",
		})
	}
	// Initialize AI Campaign Processor
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
	
	// Process campaign in background
	go func() {
		ctx := context.Background()
		err := processor.ProcessAICampaign(ctx, campaignID)
		if err != nil {
			handler.logger.Errorf("Failed to process AI campaign %d: %v", campaignID, err)
			// Update status to failed if processing fails
			campaignRepo.UpdateCampaignStatus(campaignID, "failed")
		}
	}()
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "AI campaign triggered successfully",
		Results: map[string]interface{}{
			"campaign_id": campaignID,
			"status":      "triggered",
		},
	})
}