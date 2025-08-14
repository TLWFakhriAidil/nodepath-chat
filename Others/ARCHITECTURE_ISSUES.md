## WhatsApp Multi-Device Architecture Issues

### Current Problems:
1. **Single Global Client**: The system uses one WhatsApp client (`var cli *whatsmeow.Client`) for all devices
2. **No Per-Device Isolation**: When Device A connects, then Device B connects, Device B replaces Device A
3. **Contact Mixing**: Shows contacts from whatever device connected last
4. **No Chat Isolation**: Can't separate chats between different devices

### Root Cause:
In `src/infrastructure/whatsapp/init.go`:
```go
var cli *whatsmeow.Client  // SINGLE GLOBAL CLIENT - This is the problem!
```

### Solution Required:
1. Create separate WhatsApp client for each device
2. Store clients in ClientManager mapped by device ID
3. Each device needs its own:
   - WhatsApp connection
   - Contact list
   - Chat history
   - Message store

### Temporary Workaround:
For now, the system can only handle ONE WhatsApp connection at a time.
- If you need multiple devices, you'll need to run multiple instances of the application
- Or disconnect one device before connecting another

### Files That Need Major Changes:
1. `infrastructure/whatsapp/init.go` - Remove global client, create per-device clients
2. `usecase/app.go` - Pass device ID to all WhatsApp operations
3. `infrastructure/whatsapp/client_manager.go` - Already supports multiple clients but not used properly
