package whatsapp

import (
    "go-whatsapp-web-multidevice/src/infrastructure/database"
    "go-whatsapp-web-multidevice/src/ui/websocket"
    "log"
)

// Enhanced logout handler that properly cleans up WhatsApp session
func HandleDeviceLogout(deviceID string) error {
    log.Printf("Handling device logout for: %s", deviceID)
    
    // First, clear all WhatsApp session data from database
    err := ClearWhatsAppSession(deviceID)
    if err != nil {
        log.Printf("Error clearing WhatsApp session: %v", err)
        // Don't return error, continue with logout
    }
    
    // Get client
    client := GetClient(deviceID)
    if client != nil {
        // Disconnect properly
        client.Disconnect()
    }
    
    // Remove from client manager
    cm := GetClientManager()
    cm.RemoveClient(deviceID)
    
    // Clear QR channel
    ClearDeviceQRChannel(deviceID)
    
    // Update device status in database
    db := database.DBConn
    _, err = db.Exec(`
        UPDATE devices 
        SET status = 'offline', 
            phone = NULL, 
            jid = NULL, 
            updated_at = CURRENT_TIMESTAMP 
        WHERE id = $1
    `, deviceID)
    
    if err != nil {
        log.Printf("Error updating device status: %v", err)
    }
    
    // Broadcast logout via WebSocket
    websocket.Broadcast <- websocket.BroadcastMessage{
        Code:    "DEVICE_LOGGED_OUT",
        Message: "Device logged out",
        Result: map[string]interface{}{
            "deviceId": deviceID,
        },
    }
    
    log.Printf("Device %s logged out successfully", deviceID)
    return nil
}

// Clear all WhatsApp session data for a device
func ClearWhatsAppSession(deviceID string) error {
    db := database.DBConn
    
    // Start transaction
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    // Clear tables in correct order to avoid foreign key constraints
    tables := []string{
        "whatsmeow_app_state_mutation_macs",
        "whatsmeow_app_state_sync_keys", 
        "whatsmeow_app_state_version",
        "whatsmeow_chat_settings",
        "whatsmeow_contacts",
        "whatsmeow_disappearing_timers",
        "whatsmeow_group_participants",
        "whatsmeow_groups",
        "whatsmeow_history_syncs",
        "whatsmeow_media_backfill_requests",
        "whatsmeow_message_secrets",
        "whatsmeow_portal_backfill",
        "whatsmeow_portal_backfill_queue", 
        "whatsmeow_portal_message",
        "whatsmeow_portal_message_part",
        "whatsmeow_portal_reaction",
        "whatsmeow_portal",
        "whatsmeow_privacy_settings",
        "whatsmeow_sender_keys",
        "whatsmeow_sessions",
        "whatsmeow_device",
    }
    
    for _, table := range tables {
        _, err = tx.Exec(`DELETE FROM ` + table + ` WHERE jid = $1`, deviceID)
        if err != nil {
            log.Printf("Error clearing %s: %v", table, err)
            // Continue with other tables
        }
    }
    
    // Commit transaction
    return tx.Commit()
}
