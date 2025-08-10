// Modified CreateCampaign function to support AI campaigns
// This replaces the existing CreateCampaign function in app.go

func (handler *App) CreateCampaign(c *fiber.Ctx) error {
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
	
	user, err := userRepo.GetUserByID(session.UserID)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
	}
	var request struct {
		CampaignDate    string  `json:"campaign_date"`
		Title           string  `json:"title"`
		Niche           string  `json:"niche"`
		TargetStatus    string  `json:"target_status"`
		Message         string  `json:"message"`
		ImageURL        string  `json:"image_url"`
		TimeSchedule    string  `json:"time_schedule"`
		MinDelaySeconds int     `json:"min_delay_seconds"`
		MaxDelaySeconds int     `json:"max_delay_seconds"`
		AI              *string `json:"ai"`    // New field for AI campaigns
		Limit           int     `json:"limit"` // New field for device limit
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	// Validate required fields
	if request.CampaignDate == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Campaign date is required",
		})
	}
	// Parse scheduled time if provided
	var timeSchedule string
	if request.TimeSchedule != "" {
		timeSchedule = request.TimeSchedule
	} else {
		// Default to current time if not provided
		timeSchedule = time.Now().Format("15:04")
	}
	
	// Validate and set target_status
	targetStatus := request.TargetStatus
	if targetStatus != "prospect" && targetStatus != "customer" && targetStatus != "all" {
		targetStatus = "all" // Default to all if invalid
	}
	
	// Set default delays if not provided
	minDelay := request.MinDelaySeconds
	maxDelay := request.MaxDelaySeconds
	if minDelay <= 0 {
		minDelay = 10
	}
	if maxDelay <= 0 || maxDelay < minDelay {
		maxDelay = 30
	}
	
	// For AI campaigns, ensure limit is set
	if request.AI != nil && *request.AI == "ai" && request.Limit <= 0 {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Device limit must be greater than 0 for AI campaigns",
		})
	}
	campaignRepo := repository.GetCampaignRepository()
	
	// Check for existing campaign on the same date
	existingCampaign, _ := campaignRepo.GetCampaignByDate(user.ID, request.CampaignDate)
	if existingCampaign != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "DUPLICATE_CAMPAIGN",
			Message: "A campaign already exists for this date",
		})
	}
	
	// Create the campaign
	campaign := &models.Campaign{
		UserID:          user.ID,
		DeviceID:        "", // Not used for AI campaigns
		Title:           request.Title,
		Niche:           request.Niche,
		TargetStatus:    targetStatus,
		Message:         request.Message,
		ImageURL:        request.ImageURL,
		CampaignDate:    request.CampaignDate,
		ScheduledDate:   request.CampaignDate,
		TimeSchedule:    timeSchedule,
		MinDelaySeconds: minDelay,
		MaxDelaySeconds: maxDelay,
		Status:          "pending",
		AI:              request.AI,
		Limit:           request.Limit,
	}
	
	err = campaignRepo.CreateCampaign(campaign)
	if err != nil {
		handler.logger.Errorf("Failed to create campaign: %v", err)
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "CREATE_FAILED",
			Message: "Failed to create campaign",
		})
	}
	
	handler.logger.Infof("Campaign created successfully: %+v", campaign)
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Campaign created successfully",
		Results: campaign,
	})
}