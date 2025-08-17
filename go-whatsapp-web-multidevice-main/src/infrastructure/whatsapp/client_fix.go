package whatsapp

import (
	"fmt"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/models"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"go.mau.fi/whatsmeow"
)

// SetGlobalClient sets the global WhatsApp client instance
func SetGlobalClient(client *whatsmeow.Client) {
	cli = client
}

// GetGlobalClient returns the global WhatsApp client instance
func GetGlobalClient() *whatsmeow.Client {
	return cli
}

// RegisterDeviceWithGlobalClient registers a device with the global client if available
func RegisterDeviceWithGlobalClient(deviceID string) error {
	if cli == nil {
		return fmt.Errorf("no global WhatsApp client available")
	}
	
	if !cli.IsConnected() {
		return fmt.Errorf("global WhatsApp client is not connected")
	}
	
	// Register with client manager
	cm := GetClientManager()
	cm.AddClient(deviceID, cli)
	
	fmt.Printf("Successfully registered device %s with global client\n", deviceID)
	return nil
}

// DiagnoseClients provides diagnostic information about all registered clients
func DiagnoseClients() map[string]interface{} {
	cm := GetClientManager()
	allClients := cm.GetAllClients()
	
	diagnostics := map[string]interface{}{
		"total_clients": len(allClients),
		"clients": make(map[string]interface{}),
		"global_client": nil,
	}
	
	// Check global client
	if cli != nil {
		diagnostics["global_client"] = map[string]interface{}{
			"connected": cli.IsConnected(),
			"logged_in": cli.IsLoggedIn(),
		}
		
		if cli.Store != nil && cli.Store.ID != nil {
			diagnostics["global_client"].(map[string]interface{})["jid"] = cli.Store.ID.String()
			diagnostics["global_client"].(map[string]interface{})["phone"] = cli.Store.ID.User
		}
	}
	
	// Check all registered clients
	for deviceID, client := range allClients {
		clientInfo := map[string]interface{}{
			"device_id": deviceID,
			"connected": false,
			"logged_in": false,
		}
		
		if client != nil {
			clientInfo["connected"] = client.IsConnected()
			clientInfo["logged_in"] = client.IsLoggedIn()
			
			if client.Store != nil && client.Store.ID != nil {
				clientInfo["jid"] = client.Store.ID.String()
				clientInfo["phone"] = client.Store.ID.User
			}
		}
		
		diagnostics["clients"].(map[string]interface{})[deviceID] = clientInfo
	}
	
	return diagnostics
}

// TryRegisterDeviceFromDatabase attempts to register a device using the global client
func TryRegisterDeviceFromDatabase(deviceID string) error {
	// Check if device exists in database
	userRepo := repository.GetUserRepository()
	
	// Get all users to find the device
	users, err := userRepo.GetAllUsers()
	if err != nil {
		return fmt.Errorf("failed to get users: %v", err)
	}
	
	var foundDevice *models.UserDevice
	for _, user := range users {
		devices, err := userRepo.GetUserDevices(user.ID)
		if err != nil {
			continue
		}
		
		for _, device := range devices {
			if device.ID == deviceID {
				foundDevice = device
				break
			}
		}
		
		if foundDevice != nil {
			break
		}
	}
	
	if foundDevice == nil {
		return fmt.Errorf("device %s not found in database", deviceID)
	}
	
	// Try to register with global client
	if cli != nil && cli.IsConnected() && foundDevice.Status == "online" {
		cm := GetClientManager()
		cm.AddClient(deviceID, cli)
		
		fmt.Printf("Successfully registered device %s (JID: %s) with global client\n", deviceID, foundDevice.JID)
		return nil
	}
	
	return fmt.Errorf("unable to register device %s - global client not available or device not online", deviceID)
}
