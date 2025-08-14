# PLATFORM DEVICE MESSAGE ROUTING - January 19, 2025

## How Platform Devices Work for Campaigns & Sequences

Both campaigns and sequences use the same message sending infrastructure that automatically routes based on device platform.

### Message Flow:

1. **Campaign/Sequence creates message** → Queued to `broadcast_messages` table
2. **Broadcast Worker picks up message** → Calls `messageSender.SendMessage(deviceID, msg)`
3. **WhatsAppMessageSender checks device**:

```go
// Get device details
device := GetDeviceByID(deviceID)

if device.Platform != "" {
    // Route to external API
    platformSender.SendMessage(
        device.Platform,     // "Wablas" or "Whacenter"
        device.JID,         // API token/instance
        phone,
        message...
    )
} else {
    // Use WhatsApp Web
    sendViaWhatsApp(deviceID, msg)
}
```

### Platform API Details:

#### **Wablas**:
- Text endpoint: `https://my.wablas.com/api/send-message`
- Image endpoint: `https://my.wablas.com/api/send-image`
- Authentication: Authorization header with token from device.JID
- Format: Form-encoded (application/x-www-form-urlencoded)

#### **Whacenter**:
- Single endpoint: `https://api.whacenter.com/api/send`
- Authentication: device_id parameter with value from device.JID
- Format: JSON payload

### Important Notes:

1. **Anti-Spam Applied**: Platform messages also get Malaysian greetings and randomization
2. **Same Queue**: Both WhatsApp Web and Platform devices use same broadcast_messages queue
3. **Device JID Usage**: For platform devices, JID column stores API credentials
4. **No Status Checks**: Platform devices skip WhatsApp connection checks
5. **Error Handling**: API errors are logged and message marked as failed

### Database Configuration:

To set a device as platform device:
```sql
-- For Wablas
UPDATE user_devices 
SET platform = 'Wablas',
    jid = 'your-wablas-token'
WHERE id = 'device-uuid';

-- For Whacenter
UPDATE user_devices 
SET platform = 'Whacenter',
    jid = 'your-whacenter-device-id'
WHERE id = 'device-uuid';
```

### Benefits:
- Seamless integration - campaigns/sequences don't need to know about platforms
- Automatic routing based on device configuration
- Same anti-spam protection for all messages
- Unified queue and processing system
- Easy to add new platforms (just add case in platform_sender.go)

The system is already fully implemented and working for both campaigns and sequences!
