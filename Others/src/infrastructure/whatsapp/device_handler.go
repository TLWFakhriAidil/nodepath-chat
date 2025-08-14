package whatsapp

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"
	
	"github.com/sirupsen/logrus"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/repository"
	"github.com/aldinokemal/go-whatsapp-web-multidevice/infrastructure/whatsapp/multidevice"
	websocket "github.com/aldinokemal/go-whatsapp-web-multidevice/ui/websocket"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types/events"
)

var (
	deviceQRChannels   = make(map[string]<-chan whatsmeow.QRChannelItem) // deviceID -> QR channel
	qrMutex            sync.RWMutex
)

// SetDeviceQRChannel stores QR channel for a device
func SetDeviceQRChannel(deviceID string, qrChan <-chan whatsmeow.QRChannelItem) {
	qrMutex.Lock()
	defer qrMutex.Unlock()
	deviceQRChannels[deviceID] = qrChan
	
	// Start goroutine to handle QR updates
	go func() {
		for qrItem := range qrChan {
			// Broadcast QR update via websocket
			websocket.Broadcast <- websocket.BroadcastMessage{
				Code:    "QR_UPDATE",
				Message: "QR code updated",
				Result: map[string]interface{}{
					"deviceId": deviceID,
					"qr":       qrItem.Code,
					"timeout":  qrItem.Timeout,
				},
			}
		}
		// Channel closed, remove it
		ClearDeviceQRChannel(deviceID)
	}()
}

// GetDeviceQR gets the current QR from channel
func GetDeviceQR(deviceID string) (string, error) {
	qrMutex.RLock()
	qrChan, exists := deviceQRChannels[deviceID]
	qrMutex.RUnlock()
	
	if !exists {
		return "", fmt.Errorf("no QR channel for device %s", deviceID)
	}
	
	// Try to get QR with timeout
	select {
	case qrItem, ok := <-qrChan:
		if !ok {
			return "", fmt.Errorf("QR channel closed")
		}
		return qrItem.Code, nil
	case <-time.After(1 * time.Second):
		return "", fmt.Errorf("no QR available")
	}
}

// ClearDeviceQRChannel removes QR channel for a device
func ClearDeviceQRChannel(deviceID string) {
	qrMutex.Lock()
	defer qrMutex.Unlock()
	delete(deviceQRChannels, deviceID)
}

// HandleDeviceEvent handles WhatsApp events for a specific device
func HandleDeviceEvent(ctx context.Context, deviceID string, rawEvt interface{}) {
	switch evt := rawEvt.(type) {
	case *events.PairSuccess:
		handleDevicePairSuccess(ctx, deviceID, evt)
	case *events.Connected:
		handleDeviceConnected(ctx, deviceID)
	case *events.PushNameSetting:
		handleDeviceConnected(ctx, deviceID)
	case *events.LoggedOut:
		handleDeviceLoggedOut(ctx, deviceID)
	case *events.Message:
		// Skip - already handled in main handler (init.go)
		// This prevents duplicate message storage
		logrus.Debugf("Message event for device %s handled by main handler", deviceID)
	case *events.HistorySync:
		// Process history sync to get recent messages
		HandleHistorySyncForWebView(deviceID, evt)
		
		// Also update chat list
		cm := GetClientManager()
		if client, err := cm.GetClient(deviceID); err == nil {
			HandleHistorySyncForChats(deviceID, client, evt)
		}
	}
}

// handleDevicePairSuccess handles successful QR pairing for a device
func handleDevicePairSuccess(ctx context.Context, deviceID string, evt *events.PairSuccess) {
	logrus.Infof("Device %s paired successfully with %s", deviceID, evt.ID.String())
	
	// Get device connection
	dm := multidevice.GetDeviceManager()
	conn, err := dm.GetDeviceConnection(deviceID)
	if err != nil {
		logrus.Errorf("Failed to get device connection: %v", err)
		return
	}
	
	// Update connection info
	conn.Phone = evt.ID.User
	
	// Broadcast pairing success
	websocket.Broadcast <- websocket.BroadcastMessage{
		Code:    "DEVICE_PAIRED",
		Message: fmt.Sprintf("Device paired with %s", evt.ID.String()),
		Result: map[string]interface{}{
			"deviceId": deviceID,
			"phone":    evt.ID.User,
		},
	}
}

// handleDeviceConnected handles full connection for a device
func handleDeviceConnected(ctx context.Context, deviceID string) {
	logrus.Infof("Device %s fully connected", deviceID)
	
	// Get device connection
	dm := multidevice.GetDeviceManager()
	conn, err := dm.GetDeviceConnection(deviceID)
	if err != nil {
		logrus.Errorf("Failed to get device connection: %v", err)
		return
	}
	
	if conn.Client == nil || !conn.Client.IsLoggedIn() {
		logrus.Warnf("Device %s connected event but client not logged in", deviceID)
		return
	}
	
	// Get WhatsApp info
	var phoneNumber, jid string
	if conn.Client.Store.ID != nil {
		jid = conn.Client.Store.ID.String()
		phoneNumber = conn.Client.Store.ID.User
		logrus.Infof("Device %s connected as: %s (Phone: %s)", deviceID, jid, phoneNumber)
	}
	
	// Update device in database
	userRepo := repository.GetUserRepository()
	err = userRepo.UpdateDeviceStatus(deviceID, "online", phoneNumber, jid)
	if err != nil {
		logrus.Errorf("Failed to update device status: %v", err)
	} else {
		logrus.Infof("Successfully updated device %s to online status", deviceID)
	}
	
	// Update device manager
	dm.UpdateDeviceStatus(deviceID, true, phoneNumber)
	
	// Register with client manager for broadcasts
	cm := GetClientManager()
	cm.AddClient(deviceID, conn.Client)
	logrus.Infof("Registered device %s with client manager", deviceID)
	
	// Clear QR channel
	ClearDeviceQRChannel(deviceID)
	
	// Broadcast connection success
	websocket.Broadcast <- websocket.BroadcastMessage{
		Code:    "DEVICE_CONNECTED",
		Message: "WhatsApp fully connected and logged in",
		Result: map[string]interface{}{
			"deviceId": deviceID,
			"phone":    phoneNumber,
			"jid":      jid,
		},
	}
	
	// trigger initial sync after connection
	go func() {
		time.Sleep(3 * time.Second)
		chats, err := GetChatsForDevice(deviceID)
		if err != nil {
			logrus.Errorf("Failed to sync chats for device %s: %v", deviceID, err)
		} else {
			logrus.Infof("Successfully synced %d chats for device %s", len(chats), deviceID)
		}
	}()
}

// handleDeviceLoggedOut handles device logout
func handleDeviceLoggedOut(ctx context.Context, deviceID string) {
	logrus.Infof("Device %s logged out", deviceID)
	
	// Get phone number and JID before updating
	phoneNumber := ""
	jidStr := ""
	cm := GetClientManager()
	client, err := cm.GetClient(deviceID)
	if err == nil && client != nil && client.Store != nil && client.Store.ID != nil {
		phoneNumber = client.Store.ID.User
		jidStr = client.Store.ID.String()
	}
	
	// If we couldn't get from client, try to get from database
	if phoneNumber == "" || jidStr == "" {
		userRepo := repository.GetUserRepository()
		var dbPhone, dbJID sql.NullString
		err = userRepo.DB().QueryRow("SELECT phone, jid FROM user_devices WHERE id = ?", deviceID).Scan(&dbPhone, &dbJID)
		if err == nil {
			if phoneNumber == "" && dbPhone.Valid {
				phoneNumber = dbPhone.String
			}
			if jidStr == "" && dbJID.Valid {
				jidStr = dbJID.String
			}
		}
	}
	
	// Update device status - KEEP JID AND PHONE for easier reconnection
	userRepo := repository.GetUserRepository()
	err = userRepo.UpdateDeviceStatus(deviceID, "offline", phoneNumber, jidStr)
	if err != nil {
		logrus.Errorf("Failed to update device status: %v", err)
	}
	
	// Update device manager
	dm := multidevice.GetDeviceManager()
	dm.UpdateDeviceStatus(deviceID, false, phoneNumber)
	
	// Remove from client manager
	cm.RemoveClient(deviceID)
	
	// Clear QR channel
	ClearDeviceQRChannel(deviceID)
	
	// Broadcast logout with phone number (like QR scan does)
	websocket.Broadcast <- websocket.BroadcastMessage{
		Code:    "DEVICE_LOGGED_OUT",
		Message: "Device logged out",
		Result: map[string]interface{}{
			"deviceId": deviceID,
			"phone":    phoneNumber,
		},
	}
}

// ClearWhatsAppSessionData clears all WhatsApp session data for a device
// This function is now in session_cleanup_enhanced.go with better handling
/*
func ClearWhatsAppSessionData(deviceID string) error {
	// Validate device ID format (should be UUID)
	if strings.HasPrefix(deviceID, "device_") {
		logrus.Warnf("Invalid device ID format: %s (expected UUID)", deviceID)
		return fmt.Errorf("invalid device ID format")
	}
	
	// Get repository to access DB
	userRepo := repository.GetUserRepository()
	db := userRepo.DB()
	
	// First, get the JID and phone from user_devices
	var jid sql.NullString
	var phone sql.NullString
	err := db.QueryRow("SELECT jid, phone from user_devices WHERE id = ?", deviceID).Scan(&jid, &phone)
	if err != nil {
		if err == sql.ErrNoRows {
			logrus.Warnf("Device %s not found in database", deviceID)
			return nil
		}
		logrus.Warnf("Failed to get device info for %s: %v", deviceID, err)
		return nil
	}
	
	// If no JID found, we can't clean WhatsApp sessions
	if !jid.Valid || jid.String == "" {
		logrus.Infof("No JID found for device %s, skipping WhatsApp session cleanup", deviceID)
		// Don't update status here - let the caller handle it
		return nil
	}
	
	logrus.Infof("Clearing WhatsApp session for device %s with JID %s", deviceID, jid.String)
	
	// Use separate transactions for each operation to avoid transaction abort issues
	jidsToCheck := []string{jid.String}
	
	// Also check if phone-based JID exists
	if phone.Valid && phone.String != "" {
		phoneJID := phone.String + "@s.whatsapp.net"
		jidsToCheck = append(jidsToCheck, phoneJID)
	}
	
	// Tables to clear in order (to avoid foreign key violations)
	clearOperations := []struct {
		name  string
		query string
	}{
		{"app_state_mutation_macs", "DELETE FROM whatsmeow_app_state_mutation_macs WHERE jid = ANY(?)"},
		{"app_state_sync_keys", "DELETE FROM whatsmeow_app_state_sync_keys WHERE jid = ANY(?)"},
		{"app_state_version", "DELETE FROM whatsmeow_app_state_version WHERE jid = ANY(?)"},
		{"chat_settings", "DELETE FROM whatsmeow_chat_settings WHERE jid = ANY(?)"},
		{"contacts", "DELETE FROM whatsmeow_contacts WHERE jid = ANY(?)"},
		{"disappearing_timers", "DELETE FROM whatsmeow_disappearing_timers WHERE jid = ANY(?)"},
		{"group_participants", "DELETE FROM whatsmeow_group_participants WHERE group_jid IN (SELECT jid FROM whatsmeow_groups WHERE jid = ANY(?))"},
		{"groups", "DELETE FROM whatsmeow_groups WHERE jid = ANY(?)"},
		{"history_syncs", "DELETE FROM whatsmeow_history_syncs WHERE device_jid = ANY(?)"},
		{"media_backfill_requests", "DELETE FROM whatsmeow_media_backfill_requests WHERE user_jid = ANY(?) OR portal_jid = ANY(?)"},
		{"message_secrets", "DELETE FROM whatsmeow_message_secrets WHERE chat_jid = ANY(?)"},
		{"portal_data", "DELETE FROM whatsmeow_portal_message_part WHERE message_id IN (SELECT id FROM whatsmeow_portal_message WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = ANY(?)))"},
		{"portal_messages", "DELETE FROM whatsmeow_portal_message WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = ANY(?))"},
		{"portal_reactions", "DELETE FROM whatsmeow_portal_reaction WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = ANY(?))"},
		{"portal_backfill", "DELETE FROM whatsmeow_portal_backfill WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = ANY(?))"},
		{"portal_backfill_queue", "DELETE FROM whatsmeow_portal_backfill_queue WHERE portal_jid IN (SELECT jid FROM whatsmeow_portal WHERE receiver = ANY(?))"},
		{"portals", "DELETE FROM whatsmeow_portal WHERE receiver = ANY(?)"},
		{"privacy_settings", "DELETE FROM whatsmeow_privacy_settings WHERE jid = ANY(?)"},
		{"sender_keys", "DELETE FROM whatsmeow_sender_keys WHERE our_jid = ANY(?)"},
		{"sessions", "DELETE FROM whatsmeow_sessions WHERE our_jid = ANY(?)"},
		{"pre_keys", "DELETE FROM whatsmeow_pre_keys WHERE jid = ANY(?)"},
		{"identity_keys", "DELETE FROM whatsmeow_identity_keys WHERE our_jid = ANY(?)"},
		{"device", "DELETE from whatsmeow_device WHERE jid = ANY(?)"},
	}
	
	// Execute each operation in its own transaction
	successCount := 0
	for _, op := range clearOperations {
		func() {
			tx, err := db.Begin()
			if err != nil {
				logrus.Debugf("Failed to begin transaction for %s: %v", op.name, err)
				return
			}
			defer tx.Rollback()
			
			_, err = tx.Exec(op.query, pq.Array(jidsToCheck))
			if err != nil {
				logrus.Debugf("Failed to clear %s: %v", op.name, err)
				return
			}
			
			if err = tx.Commit(); err != nil {
				logrus.Debugf("Failed to commit %s: %v", op.name, err)
				return
			}
			successCount++
		}()
	}
	
	logrus.Infof("Successfully cleared %d/%d tables for device %s", successCount, len(clearOperations), deviceID)
	
	// Don't update device status here - that should be handled by the caller
	// We're just clearing WhatsApp session data, not logging out the device
	
	return nil
}
*/
