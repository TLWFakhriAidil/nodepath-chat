# 📊 Checking Redis and Worker Status

## 1. 🔍 Check Redis Status

### Via API Endpoint:
```
GET https://your-app.up.railway.app/api/system/redis-check
```

### Response Example:
```json
{
  "status": 200,
  "code": "SUCCESS",
  "message": "Redis status check",
  "results": {
    "manager_type": "Ultra Scale Redis Manager (3000+ devices)",
    "is_redis_enabled": true,
    "message": "✅ Ultra Scale Redis is properly configured and running! Ready for 3000+ devices!",
    "environment_vars": {
      "REDIS_URL": "redis://****@redis.railway.internal:6379",
      "config.RedisURL": "redis://****@redis.railway.internal:6379"
    },
    "validation_checks": {
      "not_empty": true,
      "no_template_vars": true,
      "not_localhost": true,
      "has_redis_scheme": true
    }
  }
}
```

## 2. 👷 Check Worker Status

### Check Specific Device Worker:
```
GET https://your-app.up.railway.app/api/workers/device/{deviceId}
```

### Response Example (Worker Active):
```json
{
  "status": 200,
  "code": "SUCCESS",
  "message": "Worker status",
  "results": {
    "device_id": "uuid-123",
    "worker_exists": true,
    "status": "active",
    "queue_size": 45,
    "processed_count": 1250,
    "failed_count": 3,
    "last_activity": "2025-06-27T14:30:00Z",
    "is_active": true,
    "message": "✅ Worker is active and processing messages"
  }
}
```

### Response Example (No Worker):
```json
{
  "status": 200,
  "code": "SUCCESS",
  "message": "Worker status",
  "results": {
    "device_id": "uuid-456",
    "worker_exists": false,
    "status": "no_worker",
    "message": "No worker running for this device. Worker will start automatically when messages are queued."
  }
}
```

### Check All Workers Status:
```
GET https://your-app.up.railway.app/api/workers/status
```

### Response Example:
```json
{
  "status": 200,
  "code": "SUCCESS",
  "message": "All workers status",
  "results": {
    "summary": {
      "total_workers": 150,
      "active_workers": 120,
      "idle_workers": 28,
      "error_workers": 2,
      "total_queued": 4580,
      "total_processed": 125000,
      "total_failed": 45
    },
    "workers": [
      {
        "DeviceID": "uuid-123",
        "Status": "active",
        "QueueSize": 45,
        "ProcessedCount": 1250,
        "FailedCount": 3,
        "LastActivity": "2025-06-27T14:30:00Z"
      }
      // ... more workers
    ],
    "message": "✅ 120 workers active out of 150 total"
  }
}
```

## 3. 💬 Updated Message Sending Logic

### Single Message Approach (NEW)
The system now sends messages as a single unit instead of separate image + text:

#### Message Types:

1. **Text Only**:
   - Sends a single text message
   - No image attached

2. **Image with Caption**:
   - Sends image with text as caption
   - Single message delivery
   - No 3-second delay between image and text

3. **Image Only**:
   - Sends image without any text
   - No caption

### Old vs New Behavior:

| Scenario | Old Behavior | New Behavior |
|----------|--------------|--------------|
| Text only | Send text message | Send text message |
| Image + Text | 1. Send image<br>2. Wait 3 seconds<br>3. Send text | Send single image message with caption |
| Image only | Send image | Send image without caption |

### Benefits:
- 🚀 **Faster delivery**: No 3-second delay for image+text
- 📱 **Better UX**: Recipients see image and text together
- 💾 **Less bandwidth**: Single message instead of two
- 📊 **Higher throughput**: Process more messages per minute

### Example Message Flow (Updated):
```
User A (15 connected devices) creates campaign:

Lead 1 (Image + Text):
→ Device A3 sends image with caption to +60123456789
→ Wait 10-20 seconds (random delay based on device's min/max settings)

Lead 2 (Text only):
→ Device A7 sends text to +60987654321
→ Wait 15-25 seconds (different device might have different settings)

Lead 3 (Image only):
→ Device A11 sends image to +60111222333
→ Wait 10-30 seconds (device specific random delay)

... continues distributing across all 15 devices
```

**Important Notes**:
- Each device has its own `min_delay_seconds` and `max_delay_seconds`
- The delay between messages is always random within this range
- No more fixed 3-second delays for grouped messages
- This natural randomization helps avoid WhatsApp detection

## 4. 🎯 Quick Status Check Commands

### Using cURL:
```bash
# Check Redis
curl https://your-app.up.railway.app/api/system/redis-check

# Check specific device worker
curl https://your-app.up.railway.app/api/workers/device/YOUR_DEVICE_ID

# Check all workers
curl https://your-app.up.railway.app/api/workers/status
```

### From Dashboard:
1. Go to Dashboard
2. Click "System Status" → "Redis Check"
3. Click "Worker Status" → View all workers
4. Click on any device → "Check Worker"

## 5. 🔧 Troubleshooting

### Redis Not Working?
1. Check `/api/system/redis-check`
2. Verify REDIS_URL in Railway environment
3. Ensure Redis addon is active
4. Check Railway logs for connection errors

### Worker Not Starting?
1. Check device authentication status
2. Verify messages are queued
3. Check worker limit (3000 max)
4. Look for errors in `/api/workers/device/{id}`

### Messages Not Sending?
1. Verify worker is active
2. Check queue size
3. Monitor failed count
4. Review device connection status

## Message Delay Configuration

Delays are configured at the **campaign or sequence level**, not per device. All devices follow the same delay settings from the campaign/sequence.

### Campaign Delays:
When creating a campaign, you set:
- `min_delay_seconds`: Minimum delay between messages
- `max_delay_seconds`: Maximum delay between messages

Example:
```json
{
  "title": "Promo Merdeka",
  "min_delay_seconds": 10,
  "max_delay_seconds": 30,
  "message": "Special discount today!"
}
```

### Sequence Delays:
Sequences can have:
1. **Sequence-level delays**: Applied to all steps
2. **Step-level delays**: Override for specific days

Example:
```json
{
  "name": "7-Day Welcome Series",
  "min_delay_seconds": 15,
  "max_delay_seconds": 45,
  "steps": [
    {
      "day": 1,
      "content": "Welcome!",
      "min_delay_seconds": 5,  // Override for Day 1
      "max_delay_seconds": 10
    },
    {
      "day": 2,
      "content": "Check our products"
      // Uses sequence-level delays (15-45)
    }
  ]
}
```

### How It Works:
1. **Campaign sends message** → Uses campaign's min/max delays
2. **Sequence sends message** → Uses step delays (if set) OR sequence delays
3. **All devices** sending the same campaign/sequence use the same delays
4. **Random calculation**: Each message waits random(min, max) seconds

### Example:
Campaign with min=10, max=30 being sent by 15 devices:
- Device A1: Sends to Lead 1, waits 17 seconds, sends to Lead 2
- Device A2: Sends to Lead 3, waits 24 seconds, sends to Lead 4  
- Device A3: Sends to Lead 5, waits 11 seconds, sends to Lead 6
- All devices use same 10-30 range, but random within that range

## ✅ Implementation Status

**Fixed** (as of June 27, 2025):
- ✅ Delays now properly come from campaigns/sequences
- ✅ All devices use the same delay range when processing the same campaign/sequence
- ✅ Campaign messages include min/max delays from the campaign
- ✅ Sequence messages will include delays from sequence/step settings
- ✅ Worker uses message-specific delays with fallback to defaults
