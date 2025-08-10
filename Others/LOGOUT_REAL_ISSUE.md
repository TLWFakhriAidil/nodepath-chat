## The Real Problem with Logout

After analyzing the code and logs, here's what's happening:

### When "Clear All Sessions" works:
1. Clears all WhatsApp session tables (whatsmeow_*)
2. Sets all devices to "offline" status
3. Removes clients from ClientManager
4. Fresh state - QR code scanning works

### When device-specific logout fails:
1. Removes client from ClientManager ✓
2. Sets status to "disconnected" (not "offline") ✗
3. Does NOT clear WhatsApp session tables ✗
4. Old session data remains in database
5. When trying to scan QR again, it conflicts with existing session data

### The Fix Needed:

The backend logout endpoint needs to:
1. Clear the WhatsApp session data for that specific device
2. Set status to "offline" (not "disconnected")
3. Remove from ClientManager (already does this)

### Backend Fix Required (in app.go LogoutDevice function):

```go
// After removing from client manager, clear WhatsApp session data
if client != nil && client.Store != nil && client.Store.ID != nil {
    jid := client.Store.ID.String()
    
    // Clear WhatsApp session tables for this JID
    db := userRepo.DB()
    tables := []string{
        "whatsmeow_device",
        "whatsmeow_identity_keys",
        "whatsmeow_pre_keys",
        "whatsmeow_sessions",
        "whatsmeow_sender_keys",
        "whatsmeow_app_state_sync_keys",
        "whatsmeow_app_state_version",
        "whatsmeow_message_secrets",
        "whatsmeow_privacy_tokens",
        "whatsmeow_chat_settings",
        "whatsmeow_contacts",
    }
    
    for _, table := range tables {
        query := fmt.Sprintf("DELETE FROM %s WHERE jid = $1", table)
        db.Exec(query, jid)
    }
}

// Update device status to "offline" not "disconnected"
err = userRepo.UpdateDeviceStatus(deviceId, "offline", "", "")
```

### Temporary Frontend Workaround:

For now, the best workaround is to use "Clear All Sessions" since it properly clears the database.