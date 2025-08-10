package whatsapp

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp/multidevice"
	"github.com/sirupsen/logrus"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
)

// ClientManager manages multiple WhatsApp client instances
// Now it's just a wrapper around DeviceManager for backward compatibility
type ClientManager struct {
	mutex sync.RWMutex
}

var (
	clientManager *ClientManager
	once          sync.Once
)

// GetClientManager returns the singleton client manager
func GetClientManager() *ClientManager {
	once.Do(func() {
		clientManager = &ClientManager{}
	})
	return clientManager
}

// AddClient adds a WhatsApp client for a device
func (cm *ClientManager) AddClient(deviceID string, client *whatsmeow.Client) {
	// Just delegate to DeviceManager - don't maintain separate storage
	dm := multidevice.GetDeviceManager()
	
	// Check if device already exists in DeviceManager
	if _, err := dm.GetDeviceConnection(deviceID); err != nil {
		// Device doesn't exist, we need to get user info
		userRepo := repository.GetUserRepository()
		device, err := userRepo.GetDeviceByID(deviceID)
		if err != nil {
			logrus.Errorf("Failed to get device info for %s: %v", deviceID, err)
			return
		}
		
		// Register with DeviceManager
		dm.RegisterDevice(deviceID, device.UserID, device.Phone, client)
	}
	
	// DISABLED - Using self-healing instead
	// km := GetKeepaliveManager()
	// km.StartKeepalive(deviceID, client)
	
	logrus.Infof("Added WhatsApp client for device: %s", deviceID)
}

// GetClient retrieves a WhatsApp client for a device
func (cm *ClientManager) GetClient(deviceID string) (*whatsmeow.Client, error) {
	// Use DeviceManager's GetOrRefreshClient for self-healing
	dm := multidevice.GetDeviceManager()
	return dm.GetOrRefreshClient(deviceID)
}

// RemoveClient removes a WhatsApp client for a device
func (cm *ClientManager) RemoveClient(deviceID string) {
	// DISABLED - Using self-healing instead
	// km := GetKeepaliveManager()
	// km.StopKeepalive(deviceID)
	
	// Remove from DeviceManager
	dm := multidevice.GetDeviceManager()
	dm.UnregisterDevice(deviceID)
	
	logrus.Infof("Removed WhatsApp client for device: %s", deviceID)
}

// GetAllClients returns all registered clients (for debugging)
func (cm *ClientManager) GetAllClients() map[string]*whatsmeow.Client {
	dm := multidevice.GetDeviceManager()
	connections := dm.GetAllDeviceConnections()
	
	clients := make(map[string]*whatsmeow.Client)
	for deviceID, conn := range connections {
		if conn.Client != nil && conn.Client.IsConnected() {
			clients[deviceID] = conn.Client
		}
	}
	return clients
}

// GetChatsForDevice fetches and saves chats for a specific device (personal chats only)
func GetChatsForDevice(deviceID string) ([]repository.WhatsAppChat, error) {
	fmt.Printf("GetChatsForDevice called for device: %s\n", deviceID)
	
	cm := GetClientManager()
	client, err := cm.GetClient(deviceID)
	
	// Always try to get from database first
	repo := repository.GetWhatsAppRepository()
	savedChats, _ := repo.GetChats(deviceID)
	fmt.Printf("Found %d saved chats in database for device %s\n", len(savedChats), deviceID)
	
	// Filter out groups - only keep personal chats
	var personalChats []repository.WhatsAppChat
	for _, chat := range savedChats {
		if !chat.IsGroup && !strings.Contains(chat.ChatJID, "@g.us") && !strings.Contains(chat.ChatJID, "@broadcast") {
			personalChats = append(personalChats, chat)
		}
	}
	fmt.Printf("Filtered to %d personal chats for device %s\n", len(personalChats), deviceID)
	
	if err != nil {
		// If client not connected, return saved personal chats from database
		fmt.Printf("Client not connected for device %s: %v\n", deviceID, err)
		return personalChats, nil
	}
	
	fmt.Printf("Client connected for device %s, syncing contacts...\n", deviceID)
	
	// Client is connected, try to sync latest data
	// Get all contacts first to have proper names
	contacts, err := client.Store.Contacts.GetAllContacts(context.Background())
	if err != nil {
		fmt.Printf("Failed to get contacts: %v\n", err)
	}
	
	// Create a map for quick updates
	chatMap := make(map[string]*repository.WhatsAppChat)
	for i := range personalChats {
		chatMap[personalChats[i].ChatJID] = &personalChats[i]
	}
	
	// Process each contact as a potential chat
	for jid, contact := range contacts {
		// Skip groups, broadcasts, and status
		if jid.Server != types.DefaultUserServer || jid.User == "status" {
			continue
		}
		
		chatJID := jid.String()
		
		// Get or create chat entry
		chat, exists := chatMap[chatJID]
		if !exists {
			chat = &repository.WhatsAppChat{
				DeviceID: deviceID,
				ChatJID:  chatJID,
				IsGroup:  false, // Always false for personal chats
			}
			chatMap[chatJID] = chat
		}
		
		// Get contact name
		if contact.PushName != "" {
			chat.ChatName = contact.PushName
		} else if contact.BusinessName != "" {
			chat.ChatName = contact.BusinessName
		} else if contact.FullName != "" {
			chat.ChatName = contact.FullName
		} else {
			// Format phone number nicely
			chat.ChatName = formatPhoneNumber(jid.User)
		}
		
		// Set default values if new
		if !exists {
			chat.LastMessageTime = time.Now().Add(-365 * 24 * time.Hour) // Old date
			chat.LastMessageText = ""
			chat.UnreadCount = 0
		}
		
		// Save to database
		if err := repo.SaveOrUpdateChat(chat); err != nil {
			fmt.Printf("Error saving chat %s: %v\n", chatJID, err)
		}
	}
	
	// Convert map back to slice
	var updatedChats []repository.WhatsAppChat
	for _, chat := range chatMap {
		// Final filter to ensure no groups
		if !chat.IsGroup && !strings.Contains(chat.ChatJID, "@g.us") {
			updatedChats = append(updatedChats, *chat)
		}
	}
	
	// Sort by last message time (newest first)
	sort.Slice(updatedChats, func(i, j int) bool {
		return updatedChats[i].LastMessageTime.After(updatedChats[j].LastMessageTime)
	})
	
	fmt.Printf("Found %d personal chats for device %s\n", len(updatedChats), deviceID)
	
	return updatedChats, nil
}

// formatPhoneNumber formats a phone number for display
func formatPhoneNumber(phone string) string {
	if len(phone) > 10 {
		return fmt.Sprintf("+%s", phone)
	}
	return phone
}

// GetMessagesForChat fetches and saves messages for a specific chat
func GetMessagesForChat(deviceID, chatJID string, limit int) ([]repository.WhatsAppMessage, error) {
	// Get the repository
	repo := repository.GetWhatsAppRepository()
	
	// For now, we only return messages from database
	// whatsmeow doesn't provide direct access to message history
	// Messages are captured through real-time events
	return repo.GetMessages(deviceID, chatJID, limit)
}

// RegisterDeviceClient registers a WhatsApp client when a device connects
func RegisterDeviceClient(deviceID string, client *whatsmeow.Client) {
	cm := GetClientManager()
	cm.AddClient(deviceID, client)
	
	// Start syncing chats in background
	go func() {
		time.Sleep(2 * time.Second) // Wait for connection to stabilize
		_, err := GetChatsForDevice(deviceID)
		if err != nil {
			fmt.Printf("Error syncing chats for device %s: %v\n", deviceID, err)
		}
	}()
}

// UnregisterDeviceClient removes a WhatsApp client when a device disconnects
func UnregisterDeviceClient(deviceID string) {
	cm := GetClientManager()
	cm.RemoveClient(deviceID)
}

// GetAllPersonalChats attempts to get ALL personal chats, including those without recent messages
func GetAllPersonalChats(deviceID string) ([]repository.WhatsAppChat, error) {
	cm := GetClientManager()
	client, err := cm.GetClient(deviceID)
	if err != nil {
		return nil, fmt.Errorf("device not connected")
	}
	
	var allPersonalChats []repository.WhatsAppChat
	chatMap := make(map[string]*repository.WhatsAppChat)
	
	// Method 1: Get all contacts from the contact store
	contacts, err := client.Store.Contacts.GetAllContacts(context.Background())
	if err == nil {
		fmt.Printf("Found %d contacts in store\n", len(contacts))
		for jid, contact := range contacts {
			// Only process personal contacts
			if jid.Server != types.DefaultUserServer || jid.User == "status" {
				continue
			}
			
			chatJID := jid.String()
			chatName := ""
			
			// Get best available name
			if contact.PushName != "" {
				chatName = contact.PushName
			} else if contact.BusinessName != "" {
				chatName = contact.BusinessName
			} else if contact.FullName != "" {
				chatName = contact.FullName
			} else if contact.FirstName != "" {
				firstName := contact.FirstName
				chatName = strings.TrimSpace(firstName)
			}
			
			if chatName == "" {
				chatName = formatPhoneNumber(jid.User)
			}
			
			chat := &repository.WhatsAppChat{
				DeviceID:        deviceID,
				ChatJID:         chatJID,
				ChatName:        chatName,
				IsGroup:         false,
				IsMuted:         false,
				LastMessageText: "",
				LastMessageTime: time.Now().Add(-365 * 24 * time.Hour),
				UnreadCount:     0,
			}
			
			chatMap[chatJID] = chat
		}
	}
	
	// Save all to database and build result slice
	repo := repository.GetWhatsAppRepository()
	for _, chat := range chatMap {
		// Try to get existing chat from DB to preserve message info
		existingChat, err := repo.GetChatByJID(deviceID, chat.ChatJID)
		if err == nil && existingChat != nil {
			// Update only necessary fields
			existingChat.ChatName = chat.ChatName
			existingChat.IsMuted = chat.IsMuted
			chat = existingChat
		}
		
		// Save to database
		if err := repo.SaveOrUpdateChat(chat); err != nil {
			fmt.Printf("Error saving chat %s: %v\n", chat.ChatJID, err)
		}
		
		allPersonalChats = append(allPersonalChats, *chat)
	}
	
	// Sort by last message time (newest first)
	sort.Slice(allPersonalChats, func(i, j int) bool {
		return allPersonalChats[i].LastMessageTime.After(allPersonalChats[j].LastMessageTime)
	})
	
	fmt.Printf("Total personal chats found: %d\n", len(allPersonalChats))
	
	return allPersonalChats, nil
}
// Enhanced ClientManager methods for better device registration

// RegisterDeviceOnConnection ensures proper device registration when connecting
func (cm *ClientManager) RegisterDeviceOnConnection(deviceID string, client *whatsmeow.Client) error {
	if deviceID == "" {
		return fmt.Errorf("deviceID cannot be empty")
	}
	
	if client == nil {
		return fmt.Errorf("client cannot be nil")
	}
	
	// Verify client is connected
	if !client.IsConnected() {
		return fmt.Errorf("client is not connected")
	}
	
	// Add to manager
	cm.AddClient(deviceID, client)
	
	// Update device status in database
	userRepo := repository.GetUserRepository()
	if client.Store.ID != nil {
		phoneNumber := client.Store.ID.User
		jid := client.Store.ID.String()
		
		err := userRepo.UpdateDeviceStatus(deviceID, "online", phoneNumber, jid)
		if err != nil {
			logrus.Errorf("Failed to update device status: %v", err)
		} else {
			logrus.Infof("Device %s registered successfully - Phone: %s, JID: %s", deviceID, phoneNumber, jid)
		}
	}
	
	// Start syncing chats in background
	go func() {
		time.Sleep(2 * time.Second)
		_, err := GetChatsForDevice(deviceID)
		if err != nil {
			logrus.Errorf("Error syncing chats for device %s: %v", deviceID, err)
		}
	}()
	
	return nil
}

// GetConnectedDeviceCount returns number of connected devices
func (cm *ClientManager) GetConnectedDeviceCount() int {
	dm := multidevice.GetDeviceManager()
	connections := dm.GetAllDeviceConnections()
	
	count := 0
	for _, conn := range connections {
		if conn.Client != nil && conn.Client.IsConnected() {
			count++
		}
	}
	return count
}

// GetDeviceStatus returns detailed status of a device
func (cm *ClientManager) GetDeviceStatus(deviceID string) (status string, connected bool) {
	dm := multidevice.GetDeviceManager()
	conn, err := dm.GetDeviceConnection(deviceID)
	
	if err != nil || conn == nil {
		return "not_registered", false
	}
	
	if conn.Client == nil {
		return "no_client", false
	}
	
	if !conn.Client.IsConnected() {
		return "disconnected", false
	}
	
	if !conn.Client.IsLoggedIn() {
		return "not_logged_in", false
	}
	
	return "online", true
}

// CleanupDisconnectedClients removes disconnected clients from manager
func (cm *ClientManager) CleanupDisconnectedClients() int {
	// This is now handled by DeviceManager
	// Just return 0 for backward compatibility
	return 0
}

// GetClientCount returns the number of registered clients
func (cm *ClientManager) GetClientCount() int {
	dm := multidevice.GetDeviceManager()
	connections := dm.GetAllDeviceConnections()
	return len(connections)
}