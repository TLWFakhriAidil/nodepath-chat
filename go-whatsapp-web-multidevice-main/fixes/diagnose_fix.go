// DiagnoseDevice provides diagnostic information about a device's WhatsApp connection
func (handler *App) DiagnoseDevice(c *fiber.Ctx) error {
	deviceId := c.Params("id")
	
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
	
	// Get device info
	device, err := userRepo.GetDevice(user.ID, deviceId)
	if err != nil {
		return c.Status(404).JSON(utils.ResponseData{
			Status:  404,
			Code:    "NOT_FOUND",
			Message: "Device not found",
		})
	}
	
	// Get full client diagnostics
	allDiagnostics := whatsapp.DiagnoseClients()
	
	// Check WhatsApp client status
	cm := whatsapp.GetClientManager()
	client, clientErr := cm.GetClient(device.ID)
	
	diagnostics := map[string]interface{}{
		"device": map[string]interface{}{
			"id":     device.ID,
			"name":   device.DeviceName,
			"phone":  device.Phone,
			"status": device.Status,
			"jid":    device.JID,
		},
		"whatsapp_client": map[string]interface{}{
			"connected": clientErr == nil,
			"error":     "",
		},
		"database": map[string]interface{}{
			"chats_count":    0,
			"messages_count": 0,
		},
		"client_manager": allDiagnostics,
	}
	
	if clientErr != nil {
		diagnostics["whatsapp_client"].(map[string]interface{})["error"] = clientErr.Error()
		
		// Try to register from database if device is online
		if device.Status == "online" {
			log.Printf("Attempting to auto-register device %s from database", device.ID)
			if err := whatsapp.TryRegisterDeviceFromDatabase(device.ID); err == nil {
				diagnostics["whatsapp_client"].(map[string]interface{})["auto_registered"] = true
				// Retry getting client
				client, clientErr = cm.GetClient(device.ID)
				if clientErr == nil {
					diagnostics["whatsapp_client"].(map[string]interface{})["connected"] = true
					diagnostics["whatsapp_client"].(map[string]interface{})["error"] = ""
				}
			} else {
				log.Printf("Auto-registration failed: %v", err)
			}
		}
	} 
	
	if clientErr == nil && client != nil {
		// Check if client is logged in
		diagnostics["whatsapp_client"].(map[string]interface{})["logged_in"] = client.IsLoggedIn()
		diagnostics["whatsapp_client"].(map[string]interface{})["is_connected"] = client.IsConnected()
		
		// Try to get contacts count
		contacts, err := client.Store.Contacts.GetAllContacts(context.Background())
		if err == nil {
			diagnostics["whatsapp_client"].(map[string]interface{})["contacts_count"] = len(contacts)
		}
	}
	
	// Check database
	whatsappRepo := repository.GetWhatsAppRepository()
	chats, _ := whatsappRepo.GetChats(device.ID)
	diagnostics["database"].(map[string]interface{})["chats_count"] = len(chats)
	
	// Force a sync attempt
	go func() {
		log.Printf("Forcing sync for device %s", device.ID)
		chats, err := whatsapp.GetChatsForDevice(device.ID)
		if err != nil {
			log.Printf("Sync error for device %s: %v", device.ID, err)
		} else {
			log.Printf("Sync complete for device %s: %d chats", device.ID, len(chats))
		}
	}()
	
	return c.JSON(utils.ResponseData{
		Status:  200,
		Code:    "SUCCESS",
		Message: "Device diagnostics",
		Results: diagnostics,
	})
}
