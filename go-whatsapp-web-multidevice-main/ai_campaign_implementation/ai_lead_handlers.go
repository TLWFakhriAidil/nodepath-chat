// API Handlers for AI Lead Management
// Add these functions to ui/rest/app.go

// CreateLeadAI creates a new AI lead (without device assignment)
func (handler *App) CreateLeadAI(c *fiber.Ctx) error {
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
	
	var request struct {
		Name         string `json:"name"`
		Phone        string `json:"phone"`
		Email        string `json:"email"`
		Niche        string `json:"niche"`
		TargetStatus string `json:"target_status"`
		Notes        string `json:"notes"`
	}
	
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	// Validate required fields
	if request.Name == "" || request.Phone == "" {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Name and phone are required",
		})
	}
	
	// Set default target status if not provided
	if request.TargetStatus == "" {
		request.TargetStatus = "prospect"
	}
	
	leadAIRepo := repository.GetLeadAIRepository()
	lead := &models.LeadAI{
		UserID:       session.UserID,
		Name:         request.Name,
		Phone:        request.Phone,
		Email:        request.Email,
		Niche:        request.Niche,
		Source:       "ai_manual",
		Status:       "pending",
		TargetStatus: request.TargetStatus,
		Notes:        request.Notes,
	}
	
	err = leadAIRepo.CreateLeadAI(lead)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "CREATE_FAILED",
			Message: "Failed to create AI lead",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "AI lead created successfully",
		Results: lead,
	})
}
// GetLeadsAI retrieves all AI leads for the user
func (handler *App) GetLeadsAI(c *fiber.Ctx) error {
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
	
	leadAIRepo := repository.GetLeadAIRepository()
	leads, err := leadAIRepo.GetLeadAIByUser(session.UserID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "FETCH_FAILED",
			Message: "Failed to fetch AI leads",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "AI leads fetched successfully",
		Results: leads,
	})
}
// UpdateLeadAI updates an existing AI lead
func (handler *App) UpdateLeadAI(c *fiber.Ctx) error {
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
	
	leadID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid lead ID",
		})
	}
	
	var request struct {
		Name         string `json:"name"`
		Phone        string `json:"phone"`
		Email        string `json:"email"`
		Niche        string `json:"niche"`
		TargetStatus string `json:"target_status"`
		Notes        string `json:"notes"`
	}
	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid request body",
		})
	}
	
	leadAIRepo := repository.GetLeadAIRepository()
	
	// Verify lead belongs to user
	existingLead, err := leadAIRepo.GetLeadAIByID(leadID)
	if err != nil || existingLead.UserID != session.UserID {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "AI lead not found",
		})
	}
	
	// Update lead
	lead := &models.LeadAI{
		Name:         request.Name,
		Phone:        request.Phone,
		Email:        request.Email,
		Niche:        request.Niche,
		TargetStatus: request.TargetStatus,
		Notes:        request.Notes,
	}
	
	err = leadAIRepo.UpdateLeadAI(leadID, lead)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "UPDATE_FAILED",
			Message: "Failed to update AI lead",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "AI lead updated successfully",
	})
}
// DeleteLeadAI deletes an AI lead
func (handler *App) DeleteLeadAI(c *fiber.Ctx) error {
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
	
	leadID, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(400).JSON(utils.ResponseData{
			Status:  400,
			Code:    "BAD_REQUEST",
			Message: "Invalid lead ID",
		})
	}
	
	leadAIRepo := repository.GetLeadAIRepository()
	
	// Verify lead belongs to user
	existingLead, err := leadAIRepo.GetLeadAIByID(leadID)
	if err != nil || existingLead.UserID != session.UserID {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "AI lead not found",
		})
	}
	
	err = leadAIRepo.DeleteLeadAI(leadID)
	if err != nil {
		return c.Status(500).JSON(utils.ResponseData{
			Status:  500,
			Code:    "DELETE_FAILED",
			Message: "Failed to delete AI lead",
		})
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "AI lead deleted successfully",
	})
}