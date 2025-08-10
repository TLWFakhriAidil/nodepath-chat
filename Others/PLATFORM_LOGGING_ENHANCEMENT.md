# Platform API Logging Enhancement - January 20, 2025

## Overview
Enhanced logging for platform API sending (Wablas and Whacenter) to provide detailed visibility into message sending process and errors.

## Log Levels and Prefixes

### Log Prefixes Used:
- `[PLATFORM]` - General platform operations
- `[PLATFORM-SEND]` - WhatsApp message sender platform routing
- `[WABLAS]` - Wablas general operations
- `[WABLAS-TEXT]` - Wablas text message sending
- `[WABLAS-IMAGE]` - Wablas image message sending
- `[WHACENTER]` - Whacenter operations
- `[PLATFORM-ANTISPAM]` - Anti-spam processing

### Log Levels:
- **INFO** - Important operations and results
- **DEBUG** - Detailed request/response data
- **ERROR** - Failures and error details

## What's Logged

### 1. Platform Selection (WhatsApp Message Sender)
```
[PLATFORM-SEND] 📤 Sending message via Wablas platform for device Sales Team (device-uuid)
[PLATFORM-SEND] Recipient: 60123456789, Message type: text
[PLATFORM-SEND] Using instance/token: wablas-token-here
[PLATFORM-SEND] ✅ Successfully sent message via Wablas platform to 60123456789 (took 1.2s)
```

### 2. Platform Sender Main
```
[PLATFORM] Starting send via Wablas - Phone: 60123456789, Device: device-uuid, Has Image: false
[PLATFORM] ✅ SUCCESS sending via Wablas to 60123456789 (took 1.2s)
```

### 3. Wablas Text Messages
```
[WABLAS] Sending to 60123456789 with token: wablas-tok...
[WABLAS-TEXT] Preparing request to https://my.wablas.com/api/send-message for phone: 60123456789
[WABLAS-TEXT] Request data: phone=60123456789, message_length=250
[WABLAS-TEXT] Sending POST request with Authorization token
[WABLAS-TEXT] Response Status: 200
[WABLAS-TEXT] Response Body: {"status":true,"message":"Message sent successfully"}
[WABLAS-TEXT] ✅ Message sent successfully to 60123456789
```

### 4. Wablas Image Messages
```
[WABLAS-IMAGE] Preparing request to https://my.wablas.com/api/send-image for phone: 60123456789
[WABLAS-IMAGE] Image URL: https://example.com/image.jpg
[WABLAS-IMAGE] Request data: phone=60123456789, image=https://example.com/image.jpg..., caption_length=50
[WABLAS-IMAGE] Response Status: 200
[WABLAS-IMAGE] Response Body: {"status":true,"message":"Image sent successfully"}
[WABLAS-IMAGE] ✅ Image sent successfully to 60123456789
```

### 5. Whacenter Messages
```
[WHACENTER] Preparing request to https://api.whacenter.com/api/send for phone: 60123456789
[WHACENTER] Request payload: device=whacenter-device-id, number=60123456789, message_length=250, has_file=false
[WHACENTER] JSON payload: {"device_id":"whacenter-device-id","message":"Hello...","number":"60123456789"}
[WHACENTER] Response Status: 200
[WHACENTER] Response Body: {"success":true,"message":"Message queued"}
[WHACENTER] ✅ Message sent successfully to 60123456789
```

### 6. Anti-Spam Processing
```
[PLATFORM-ANTISPAM] Starting anti-spam for 60123456789
[PLATFORM-ANTISPAM] Original message: Special promotion for gym membership...
[PLATFORM-ANTISPAM] After greeting: Hi Cik, apa khabar\n\nSpecial promotion for gym membership...
[PLATFORM-ANTISPAM] ✅ Anti-spam applied for 60123456789: greeting added, message randomized
[PLATFORM-ANTISPAM] Final message: Hi Cik, apa khabar\n\nS̲p̲e̲c̲i̲a̲l̲ p̲r̲o̲m̲o̲t̲i̲o̲n̲...
```

### 7. Error Logging
```
[PLATFORM-SEND] ❌ Failed to send via Wablas platform: platform send failed: wablas API error: status 401, body: {"error":"Invalid token"} (took 0.5s)
[WABLAS-TEXT] ❌ API Error - Status: 401, Body: {"error":"Invalid token"}
[WHACENTER] ❌ API Error - Status: 400, Body: {"error":"Invalid device_id"}
```

## How to Enable Debug Logging

To see detailed DEBUG level logs, start the application with debug mode:

```bash
# Option 1: Command line flag
./whatsapp.exe --debug=true

# Option 2: Environment variable
set APP_DEBUG=true
./whatsapp.exe
```

## Common Issues to Watch For

### 1. Authentication Errors
```
[WABLAS-TEXT] ❌ API Error - Status: 401, Body: {"error":"Invalid token"}
```
**Solution**: Check that device JID contains valid Wablas token

### 2. Invalid Phone Number
```
[WHACENTER] ❌ API Error - Status: 400, Body: {"error":"Invalid phone number format"}
```
**Solution**: Ensure phone numbers include country code without + (e.g., 60123456789)

### 3. Network Timeouts
```
[WABLAS-TEXT] HTTP request failed: Post "https://my.wablas.com/api/send-message": context deadline exceeded
```
**Solution**: Check internet connection and API endpoint availability

### 4. Rate Limiting
```
[PLATFORM] ❌ FAILED sending via Wablas to 60123456789 - Error: wablas API error: status 429, body: {"error":"Rate limit exceeded"}
```
**Solution**: Implement rate limiting in broadcast processor

## Performance Monitoring

Each log includes timing information:
- `(took 1.2s)` - Shows API response time
- Use this to identify slow API endpoints
- Normal response times: 0.5s - 2s

## Debugging Tips

1. **Enable debug mode** to see full request/response details
2. **Check log prefixes** to trace the flow
3. **Look for ❌ symbols** to quickly find errors
4. **Monitor response times** for performance issues
5. **Save logs** when issues occur for troubleshooting

## Log Rotation

Consider implementing log rotation for production:
- Platform API logs can be verbose
- Rotate daily or when size exceeds 100MB
- Keep last 7 days of logs for debugging
