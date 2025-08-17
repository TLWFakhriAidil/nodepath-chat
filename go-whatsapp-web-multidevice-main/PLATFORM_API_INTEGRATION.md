# Platform API Integration for Campaign & Sequence

## Overview
When a device has a platform value ("Wablas" or "Whacenter"), messages are sent via external APIs instead of WhatsApp Web.

## How It Works

### 1. **Platform Detection**
- Check device's `platform` column
- If empty/null → Use normal WhatsApp Web
- If "Wablas" → Send via Wablas API
- If "Whacenter" → Send via Whacenter API

### 2. **Instance/Token Mapping**
- The device's `jid` column is used as the instance/token
- For Wablas: Used as Authorization header
- For Whacenter: Used as device_id parameter

### 3. **Message Flow**
```
Campaign/Sequence → Broadcast Queue → Check Platform → Route to API
                                                    ↓
                                          Platform? → External API
                                                    ↓
                                          No Platform → WhatsApp Web
```

## API Details

### Wablas API
**Text Message:**
- URL: `https://my.wablas.com/api/send-message`
- Method: POST
- Headers: 
  - `Authorization: {device.jid}`
  - `Content-Type: application/x-www-form-urlencoded`
- Body:
  - `phone`: Recipient number
  - `message`: Text content

**Image Message:**
- URL: `https://my.wablas.com/api/send-image`
- Method: POST
- Headers: Same as text
- Body:
  - `phone`: Recipient number
  - `image`: Image URL
  - `caption`: Caption text

### Whacenter API
**Both Text & Image:**
- URL: `https://api.whacenter.com/api/send`
- Method: POST
- Headers: `Content-Type: application/json`
- Body (JSON):
  - `device_id`: {device.jid}
  - `number`: Recipient number
  - `message`: Text content
  - `file`: Image URL (if image)

## Setup Instructions

### 1. Set Device Platform
```sql
-- Set device to use Wablas
UPDATE user_devices 
SET platform = 'Wablas',
    jid = 'your-wablas-token'
WHERE id = 'device-uuid';

-- Set device to use Whacenter
UPDATE user_devices 
SET platform = 'Whacenter',
    jid = 'your-whacenter-device-id'
WHERE id = 'device-uuid';
```

### 2. How Messages Are Sent

**For Campaigns:**
- Platform devices are always included (treated as online)
- Messages sent via appropriate API based on platform
- Failed API calls mark message as failed

**For Sequences:**
- Same as campaigns - platform devices always included
- Each sequence step sent via the platform API

### 3. Error Handling
- API errors are logged with full response
- Messages marked as "failed" in database
- Error details stored for debugging

## Example Scenarios

### Scenario 1: Mixed Devices
- Device A: No platform → Uses WhatsApp Web
- Device B: Platform = "Wablas" → Uses Wablas API
- Device C: Platform = "Whacenter" → Uses Whacenter API

### Scenario 2: Campaign with Image
1. Campaign has image URL and text
2. For Wablas: Sends to `/api/send-image` with caption
3. For Whacenter: Sends to `/api/send` with file parameter

### Scenario 3: Sequence with Multiple Steps
1. Each step processed independently
2. Platform API used for each message
3. Delays respected between messages

## Testing

### Test Wablas Integration:
```sql
-- Set up test device
UPDATE user_devices 
SET platform = 'Wablas', 
    jid = 'test-token',
    status = 'offline'
WHERE device_name = 'Test Device';

-- Create campaign targeting this device
-- It will use Wablas API despite being offline
```

### Test Whacenter Integration:
```sql
-- Set up test device
UPDATE user_devices 
SET platform = 'Whacenter',
    jid = 'test-device-id',
    status = 'offline'  
WHERE device_name = 'Test Device 2';

-- Create campaign/sequence
-- It will use Whacenter API
```

## Important Notes

1. **No Status Checking**: Platform devices skip all status checks
2. **Always Online**: Treated as online for campaigns/sequences
3. **No Manual Operations**: Cannot logout or refresh platform devices
4. **API Response Logging**: All API responses logged for debugging
5. **Instance Storage**: Currently using `jid` column for API credentials

## Future Enhancements

Consider adding:
- Separate column for API credentials
- API response parsing for better error messages
- Retry logic for failed API calls
- Webhook support for delivery reports
