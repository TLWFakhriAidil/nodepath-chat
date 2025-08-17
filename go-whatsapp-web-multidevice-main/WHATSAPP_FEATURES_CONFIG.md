# WhatsApp Multi-Device Configuration Guide

## Environment Variables for Enhanced Features

### 1. Enable Chat Storage
To save all WhatsApp chats and messages for viewing in WhatsApp Web:

```env
WHATSAPP_CHAT_STORAGE=true
```

This will:
- Store all incoming/outgoing messages in the database
- Allow viewing chat history in the WhatsApp Web interface
- Keep messages even when device is offline
- Tables used: `whatsapp_chats`, `whatsapp_messages`

### 2. Enable Webhooks
To receive real-time events:

```env
WHATSAPP_WEBHOOK=https://your-webhook-url.com/webhook
WHATSAPP_WEBHOOK_SECRET=your-secret-key
```

Multiple webhooks (comma-separated):
```env
WHATSAPP_WEBHOOK=https://webhook1.com,https://webhook2.com
```

### 3. Auto-Reply
Set automatic responses:

```env
WHATSAPP_AUTO_REPLY="Thank you for your message. We'll get back to you soon!"
```

### 4. Complete Railway Configuration

Add these to your Railway environment variables:

```env
# Database (already set)
DB_URI=postgresql://...

# Application
APP_PORT=3000
APP_DEBUG=false
APP_OS=YourBusinessName
APP_BASIC_AUTH=admin:changeme123

# WhatsApp Features
WHATSAPP_CHAT_STORAGE=true
WHATSAPP_ACCOUNT_VALIDATION=true
WHATSAPP_AUTO_REPLY="Thanks for contacting us!"
WHATSAPP_WEBHOOK=https://your-webhook.com/whatsapp
WHATSAPP_WEBHOOK_SECRET=your-webhook-secret
APP_CHAT_FLUSH_INTERVAL=30

# Optional
NODE_ENV=production
```

## Webhook Events

When enabled, webhooks will receive:

### Message Events
```json
{
  "event": "message",
  "data": {
    "deviceId": "device-uuid",
    "from": "+1234567890",
    "message": "Hello!",
    "timestamp": "2025-06-27T10:00:00Z",
    "isFromMe": false
  }
}
```

### Status Events
```json
{
  "event": "status",
  "data": {
    "deviceId": "device-uuid",
    "status": "online|offline",
    "timestamp": "2025-06-27T10:00:00Z"
  }
}
```

## Benefits for Your Broadcast System

1. **Chat Storage Benefits**:
   - Review sent broadcasts
   - Check delivery status
   - Analyze responses
   - Export chat data
   - Compliance/audit trail

2. **Webhook Benefits**:
   - Real-time delivery notifications
   - Integrate with CRM systems
   - Trigger automated workflows
   - Track engagement metrics
   - Build custom dashboards

## Implementation Steps

1. **Update Railway Environment**:
   - Go to Railway Dashboard
   - Add the environment variables above
   - Railway will auto-redeploy

2. **Verify Chat Storage**:
   - Connect a device
   - Send/receive messages
   - Check WhatsApp Web view
   - Messages should persist

3. **Test Webhooks**:
   - Use webhook.site for testing
   - Send a message
   - Verify webhook receives data

## Database Tables Created

When `WHATSAPP_CHAT_STORAGE=true`:

```sql
-- Stores chat metadata
whatsapp_chats (
  id, device_id, chat_jid, chat_name, 
  is_group, last_message_text, 
  last_message_time, unread_count
)

-- Stores all messages
whatsapp_messages (
  id, device_id, chat_jid, message_id,
  sender_jid, message_text, message_type,
  media_url, timestamp
)
```

This enables full chat history and message tracking!
