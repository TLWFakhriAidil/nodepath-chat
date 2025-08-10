// LogoutDevice logs out from WhatsApp
func (handler *App) LogoutDevice(c *fiber.Ctx) error {
	deviceId := c.Query("deviceId")
	
	// Get session from cookie
	sessionToken := c.Cookies("session_token")
	if sessionToken == "" {
		return c.Status(401).JSON(utils.ResponseData{
			Status:  401,
			Code:    "UNAUTHORIZED",
			Message: "No session token",
		})
	}
	
	// Verify device ownership
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
	
	// Get device
	device, err := userRepo.GetDevice(user.ID, deviceId)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Get WhatsApp client
	cm := whatsapp.GetClientManager()
	client, err := cm.GetClient(device.ID)
	if err == nil && client != nil {
		// Logout from WhatsApp
		err = client.Logout()
		if err != nil {
			log.Printf("Error logging out device %s: %v", device.ID, err)
		}
		
		// Remove from client manager
		cm.RemoveClient(device.ID)
	}
	
	// Update device status in database
	err = userRepo.UpdateDeviceStatus(device.ID, "offline", "", "")
	if err != nil {
		log.Printf("Error updating device status: %v", err)
	}
	
	// Call the actual logout service to clean up files
	err = handler.Service.Logout(c.UserContext())
	if err != nil {
		log.Printf("Error calling logout service: %v", err)
	}
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device logged out successfully",
		Results: map[string]interface{}{
			"deviceId": deviceId,
			"status":   "offline",
		},
	})
}
