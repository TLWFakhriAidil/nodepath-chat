// Fix for WhatsApp Multi-Device Logout and Auto-Reconnection Issues

## Issues to Fix:

### Issue #3: Manual logout button clearing phone/JID
The manual logout button in dashboard.html is clearing phone and JID, preventing reconnection.

### Issue #4: Railway restart auto-disconnect
When Railway restarts, devices disconnect but can't reconnect because session data is lost.

### Issue #5: Auto-reconnect on startup
When Railway restarts and linked devices are still active on WhatsApp, they should auto-reconnect.

## Solutions:

### 1. Fix Manual Logout Button (dashboard.html)
Don't clear phone/JID when logging out manually:
```javascript
// Update device status
const device = devices.find(d => d.id === deviceId);
if (device) {
    device.status = 'offline';
    // Keep phone and JID for reconnection
    device.lastSeen = new Date().toISOString();
}
```

### 2. Fix Logout Endpoint (app.go)
Update the logout endpoint to preserve phone/JID:
```go
// Get current device info before updating
var phone, jid string
err = userRepo.DB().QueryRow("SELECT phone, jid FROM user_devices WHERE id = $1", deviceId).Scan(&phone, &jid)

// Update device status but keep phone and JID
err = userRepo.UpdateDeviceStatus(deviceId, "disconnected", phone, jid)
```

### 3. Auto-Reconnect on Startup
Add auto-reconnect logic when the server starts:
- Check all devices with status "online" or have valid JID
- Attempt to reconnect them automatically
- If WhatsApp session is still valid, it will reconnect without QR

### 4. Handle Railway Restarts Gracefully
- Preserve device sessions in database
- On startup, attempt to restore sessions
- Use device JID to reconnect without QR if possible
