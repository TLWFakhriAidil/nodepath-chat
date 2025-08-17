// Add this to your REST API (app.go or campaign endpoints)

// ManualTriggerCampaign - manually trigger a campaign
func (rest *App) ManualTriggerCampaign(c *fiber.Ctx) error {
    campaignID := c.Params("id")
    
    // Get user from context
    userID, ok := middleware.GetUserFromContext(c)
    if !ok {
        return c.JSON(utils.ResponseData{
            Status:  401,
            Code:    "UNAUTHORIZED",
            Message: "User not authenticated",
        })
    }
    
    // Get campaign and verify ownership
    campaignRepo := repository.GetCampaignRepository()
    campaign, err := campaignRepo.GetByID(campaignID)
    if err != nil {
        return c.JSON(utils.ResponseData{
            Status:  404,
            Code:    "NOT_FOUND",
            Message: "Campaign not found",
        })
    }
    
    if campaign.UserID != userID {
        return c.JSON(utils.ResponseData{
            Status:  403,
            Code:    "FORBIDDEN",
            Message: "Access denied",
        })
    }
    
    // Check if already sent
    if campaign.Status == "sent" {
        return c.JSON(utils.ResponseData{
            Status:  400,
            Code:    "ALREADY_SENT",
            Message: "Campaign already sent",
        })
    }
    
    // Trigger the campaign
    triggerService := usecase.NewCampaignTriggerService()
    go triggerService.ExecuteCampaign(&campaign)
    
    // Update status to processing
    campaign.Status = "processing"
    err = campaignRepo.Update(&campaign)
    if err != nil {
        logrus.Errorf("Failed to update campaign status: %v", err)
    }
    
    return c.JSON(utils.ResponseData{
        Status:  200,
        Code:    "SUCCESS",
        Message: "Campaign triggered successfully",
        Results: map[string]interface{}{
            "campaign_id": campaign.ID,
            "title": campaign.Title,
            "message": "Campaign is now being processed. Check Worker Status for progress.",
        },
    })
}

// Add this route to your API
// app.Post("/api/campaigns/:id/trigger", rest.ManualTriggerCampaign)