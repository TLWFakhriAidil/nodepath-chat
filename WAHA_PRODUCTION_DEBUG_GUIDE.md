# WAHA Production Debugging Guide

## Overview
This guide helps debug WAHA webhook integration issues in Railway production environment. The system now includes comprehensive debugging capabilities to identify and resolve payload structure mismatches.

## Current Issue Analysis

### Railway Logs Error
```
2025-09-02T02:47:41.752810256Z [err] time="2025-09-02T02:47:35Z" level=warning msg="⚠️ WEBHOOK: Missing required fields (from or message)" from= id_device=FakhriAidilTLW-001 message=
```

### Root Cause
- Real WAHA webhook payload structure differs from test data
- Production payload may not follow expected `payload._data` structure
- Field extraction failing due to unknown payload format

## Enhanced Debugging Features

### 1. Production Debug Endpoint
**URL**: `POST /api/ai-whatsapp/debug/waha/{device_id}`

**Purpose**: 
- Logs complete webhook request details
- Analyzes payload structure without processing
- Returns comprehensive debug information
- Safe to use in production (no side effects)

**Usage**:
```bash
# Replace YOUR_RAILWAY_URL with actual Railway app URL
curl -X POST https://YOUR_RAILWAY_URL/api/ai-whatsapp/debug/waha/FakhriAidilTLW-001 \
  -H "Content-Type: application/json" \
  -d '{"your":"actual_waha_payload"}'
```

### 2. Enhanced Logging
- **Error Level Logs**: All WAHA debugging uses ERROR level for visibility
- **Complete Headers**: Logs all HTTP headers from WAHA
- **Payload Analysis**: Deep structure analysis of incoming data
- **Fallback Tracking**: Shows which extraction method succeeded

### 3. Multiple Extraction Fallbacks

#### Primary Method (Expected WAHA Format)
```json
{
  "payload": {
    "_data": {
      "from": "phone@c.us",
      "body": "message text",
      "fromMe": false,
      "info": {
        "pushName": "Sender Name",
        "IsGroup": false
      }
    }
  }
}
```

#### Fallback 1: Direct Top-Level Fields
```json
{
  "from": "phone@c.us",
  "body": "message text",
  "message": "message text",
  "text": "message text"
}
```

#### Fallback 2: Data Field (No Underscore)
```json
{
  "data": {
    "from": "phone@c.us",
    "body": "message text"
  }
}
```

#### Fallback 3: Alternative Message Fields
- `content`, `msg`, `messageContent`, `textContent`

#### Fallback 4: Alternative Phone Fields
- `phone`, `number`, `phoneNumber`, `sender`, `contact`

## Production Debugging Steps

### Step 1: Configure WAHA to Use Debug Endpoint
1. Temporarily change WAHA webhook URL to debug endpoint:
   ```
   https://your-railway-app.railway.app/api/ai-whatsapp/debug/waha/FakhriAidilTLW-001
   ```

2. Send a test message through WAHA

3. Check Railway logs for detailed payload structure:
   ```bash
   railway logs --follow
   ```

### Step 2: Analyze Payload Structure
Look for these log entries:
- `🚨 WAHA DEBUG ENDPOINT: Complete webhook request details`
- `🚨 WAHA PRODUCTION DEBUG: Complete payload structure analysis`
- `🚨 WAHA DEBUG: Complete payload analysis`

### Step 3: Identify Missing Fields
Check the logs for:
- `payload_keys`: Shows top-level keys in the payload
- `payload_analysis`: Shows structure depth and types
- `extracted_data`: Shows what was successfully extracted
- `extraction_success`: Boolean indicating if required fields found

### Step 4: Apply Fixes
Based on the payload structure found:

1. **If using different field names**: Add new fallback extraction methods
2. **If using different nesting**: Modify the payload traversal logic
3. **If completely different format**: Create new extraction branch

### Step 5: Switch Back to Production Endpoint
Once payload structure is understood and fixes applied:
```
https://your-railway-app.railway.app/api/ai-whatsapp/webhook/waha/FakhriAidilTLW-001
```

## Common WAHA Payload Variations

### Variation 1: Direct Structure
```json
{
  "from": "601137508067@c.us",
  "body": "Hello world",
  "fromMe": false,
  "isGroup": false,
  "pushName": "User Name"
}
```

### Variation 2: Event-Based Structure
```json
{
  "event": "message",
  "data": {
    "from": "601137508067@c.us",
    "text": "Hello world",
    "fromMe": false
  }
}
```

### Variation 3: Nested Payload Structure (Current Expected)
```json
{
  "payload": {
    "_data": {
      "from": "601137508067@c.us",
      "body": "Hello world",
      "fromMe": false,
      "info": {
        "pushName": "User Name",
        "IsGroup": false
      }
    }
  }
}
```

## Log Analysis Examples

### Successful Extraction
```
level=error msg="🚨 WAHA PRODUCTION: Final extraction results" 
extraction_success=true 
is_from_me=false 
is_group=false 
message="Hello world" 
sender_name="User Name" 
sender_phone="601137508067@c.us"
```

### Failed Extraction
```
level=error msg="🚨 WAHA PRODUCTION CRITICAL: All extraction methods failed - payload structure unknown" 
missing_message=true 
missing_sender_phone=true 
all_payload_keys="[event,timestamp,session]" 
payload_structure="map[event:map[type:string] timestamp:map[type:float64]]"
```

### Fallback Success
```
level=error msg="🚨 FALLBACK 1: Extracted sender_phone from top-level 'from'"
level=error msg="🚨 FALLBACK 1: Extracted message from top-level 'body'"
```

## Testing Commands

### Local Testing
```bash
# Test current extraction logic
Invoke-RestMethod -Uri "http://localhost:8080/api/ai-whatsapp/debug/waha/FakhriAidilTLW-001" -Method POST -ContentType "application/json" -Body '{"payload":{"_data":{"from":"601137508067@c.us","body":"Test message","fromMe":false,"info":{"pushName":"Test User","IsGroup":false}}}}'

# Test fallback extraction
Invoke-RestMethod -Uri "http://localhost:8080/api/ai-whatsapp/debug/waha/FakhriAidilTLW-001" -Method POST -ContentType "application/json" -Body '{"from":"601137508067@c.us","body":"Test message","fromMe":false}'
```

### Production Testing
```bash
# Replace with your Railway URL
Invoke-RestMethod -Uri "https://your-app.railway.app/api/ai-whatsapp/debug/waha/FakhriAidilTLW-001" -Method POST -ContentType "application/json" -Body '{"test":"payload"}'
```

## Next Steps

1. **Immediate**: Use debug endpoint to capture real WAHA payload structure
2. **Analysis**: Compare real payload with expected structure
3. **Fix**: Add appropriate extraction methods for the real structure
4. **Test**: Verify extraction works with real payload
5. **Deploy**: Switch back to production webhook endpoint

## Monitoring

### Key Metrics to Watch
- `extraction_success=true` rate
- Absence of "Missing required fields" warnings
- Successful AI conversation processing

### Alert Conditions
- `extraction_success=false` for multiple consecutive requests
- "All extraction methods failed" errors
- High rate of "Missing required fields" warnings

---

**Note**: All debugging logs use ERROR level to ensure visibility in Railway logs. This is intentional for production debugging and should be changed back to INFO/DEBUG level once issues are resolved.