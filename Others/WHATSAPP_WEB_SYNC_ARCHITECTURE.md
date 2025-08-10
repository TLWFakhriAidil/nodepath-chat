# WhatsApp Web Chat Sync Architecture Study

## Overview
The WhatsApp Web implementation in this system automatically syncs and stores chat data using WhatsApp's built-in history sync mechanism. Here's a comprehensive breakdown of how it works:

## 1. Architecture Components

### Database Tables
1. **whatsapp_chats** - Stores chat metadata
   - `device_id` - WhatsApp device identifier
   - `chat_jid` - WhatsApp chat ID (phone@s.whatsapp.net)
   - `chat_name` - Contact/chat display name
   - `last_message_time` - Timestamp of last message
   - `is_group`, `is_muted`, `unread_count`, etc.

2. **whatsapp_messages** - Stores message history
   - `device_id` - Device that received the message
   - `chat_jid` - Chat the message belongs to
   - `message_id` - Unique WhatsApp message ID
   - `sender_jid` - Who sent the message
   - `message_text` - Message content
   - `message_type` - text/image/video/audio/document
   - `timestamp` - Message timestamp (Unix seconds)

### Key Components
1. **History Sync Manager** (`sync/history_sync.go`)
2. **Chat Store** (`chat_store.go`)
3. **Message Store** (`message_store.go`)
4. **WhatsApp Web Helpers** (`whatsapp_web_helpers.go`)

## 2. How Sync Works

### Automatic Sync Process

#### A. Initial Connection
When a device connects to WhatsApp:
1. WhatsApp automatically sends a **HistorySync** event
2. This contains recent conversations and messages
3. No manual sync request needed - it's automatic!

#### B. History Sync Event Processing
```go
func handleHistorySync(evt *events.HistorySync) {
    // WhatsApp sends these automatically:
    // - Type: INITIAL, RECENT, FULL
    // - Progress: 0-100%
    // - Conversations with messages
}
```

#### C. Data Flow
1. **HistorySync Event Received** → 
2. **Process Conversations** → 
3. **Extract Chat Info** → 
4. **Store in whatsapp_chats** → 
5. **Extract Messages** → 
6. **Store in whatsapp_messages**

### Real-time Message Sync

#### A. New Message Arrives
```go
func HandleMessageForWebView(deviceID string, evt *events.Message) {
    // 1. Filter personal chats only (no groups)
    // 2. Extract message content
    // 3. Store in whatsapp_messages
    // 4. Update chat's last message time
}
```

#### B. Message Processing
1. **Check chat type** - Only personal chats (DefaultUserServer)
2. **Extract content** - Text, captions, media indicators
3. **Store with timestamp** - Validates and fixes timestamps
4. **Update chat list** - Updates last message time

## 3. Storage Mechanisms

### Chat Storage
```go
func StoreChat(deviceID, chatJID, name string, lastMessageTime time.Time) {
    // UPSERT operation - insert or update
    // Updates name and last message time
    // Maintains unique constraint on (device_id, chat_jid)
}
```

### Message Storage
```go
func StoreWhatsAppMessageWithTimestamp(...) {
    // 1. Validates timestamp (fixes milliseconds)
    // 2. Checks for future dates
    // 3. Stores with proper indexing
    // 4. Trigger limits to 20 messages per chat
}
```

### Timestamp Validation
- Converts milliseconds to seconds (÷ 1000)
- Fixes future timestamps (> 1 year ahead)
- Uses current time for invalid timestamps

## 4. Retrieval Process

### Getting Chats
```go
func GetChatsFromDatabase(deviceID string) {
    // 1. Joins whatsapp_chats with latest message
    // 2. Orders by most recent activity
    // 3. Formats timestamps (Today, Yesterday, etc.)
    // 4. Returns structured chat list
}
```

### Getting Messages
```go
func GetWhatsAppWebMessages(deviceID, chatJID string, limit int) {
    // 1. Queries messages for specific chat
    // 2. Orders by timestamp DESC
    // 3. Limits to requested count (max 100)
    // 4. Reverses for chronological order
}
```

## 5. Key Features

### Automatic Features
1. **Auto History Sync** - WhatsApp sends automatically
2. **Message Limit** - Keeps only 20 recent messages per chat
3. **Timestamp Fixing** - Handles milliseconds and future dates
4. **Real-time Updates** - New messages stored immediately

### Filtering
1. **Personal Chats Only** - Filters out groups/broadcasts
2. **Skip Status Updates** - Ignores status@broadcast
3. **Valid Messages Only** - Skips empty or invalid messages

### Message Types Supported
- Text messages
- Extended text (with formatting)
- Image messages (with captions)
- Video messages (with captions)
- Audio/Voice messages
- Documents (with filenames)
- Stickers
- Location
- Contacts
- Polls

## 6. Sync Triggers

### Automatic Triggers
1. **Device Connection** - Initial sync on connect
2. **New Messages** - Real-time as received
3. **App Restart** - Re-syncs on reconnection

### Manual Triggers
- **Not needed!** WhatsApp handles everything automatically
- The "Sync" button in UI just returns success

## 7. Performance Optimizations

### Database
1. **Indexes** on device_id, chat_jid, timestamp
2. **Message Limit** - Max 20 per chat
3. **Efficient Queries** - Uses LATERAL joins

### Processing
1. **Batch Processing** - Handles multiple messages at once
2. **Async Storage** - Non-blocking database writes
3. **Skip Invalid** - Continues on errors

## 8. Important Notes

### What Works
- ✅ Automatic history sync on connection
- ✅ Real-time message capture
- ✅ Personal chat support
- ✅ Multiple message types
- ✅ Timestamp validation
- ✅ Contact name resolution

### Limitations
- ❌ No manual sync control (WhatsApp decides)
- ❌ Groups not supported (by design)
- ❌ Limited to recent messages (WhatsApp's limit)
- ❌ No media download (only indicators)

### Best Practices
1. Let WhatsApp handle sync timing
2. Don't force manual syncs
3. Handle timestamp issues gracefully
4. Keep message storage minimal
5. Use database indexes efficiently

## 9. Code Flow Example

```
User Opens WhatsApp Web → 
GetWhatsAppWebChats(deviceID) →
  Query whatsapp_chats table →
  Join with latest message →
  Format and return

User Opens Chat →
GetWhatsAppWebMessages(deviceID, chatJID) →
  Query whatsapp_messages →
  Order by timestamp →
  Format and return

New Message Arrives →
HandleMessageForWebView() →
  Extract content →
  Store in database →
  Update chat list
```

This architecture ensures efficient, automatic synchronization of WhatsApp data without manual intervention!