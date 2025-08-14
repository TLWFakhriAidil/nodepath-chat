# Device Connection Improvements

## Current Issues with Device Going Offline

### Problem
Devices are going offline unexpectedly even when they're not banned or manually logged out. This happens because:

1. **Network Fluctuations**: Temporary network issues cause `IsConnected()` to return false
2. **WhatsApp Server Issues**: WhatsApp servers sometimes disconnect clients temporarily
3. **Health Monitor Aggressiveness**: The health monitor immediately marks devices as offline on first disconnection

### Current Behavior
```go
if !client.IsConnected() {
    // Immediately marks as disconnected/offline
    userRepo.UpdateDeviceStatus(deviceID, "disconnected", "", "")
}
```

### Recommended Improvements

1. **Add Retry Logic Before Marking Offline**
```go
// Try reconnecting 3 times before marking offline
reconnectAttempts := 3
reconnectDelay := 5 * time.Second

for i := 0; i < reconnectAttempts; i++ {
    if client.IsConnected() {
        return // Still connected, was temporary
    }
    
    // Try to reconnect
    client.Connect()
    time.Sleep(reconnectDelay)
}

// Only mark offline after all attempts fail
userRepo.UpdateDeviceStatus(deviceID, "offline", "", "")
```

2. **Implement Connection State Machine**
- `online` - Fully connected and logged in
- `reconnecting` - Temporarily disconnected, attempting reconnection
- `offline` - Failed to reconnect after multiple attempts
- `banned` - Account is banned
- `logged_out` - User manually logged out

3. **Add Connection Resilience**
```go
// Keep-alive ping every 30 seconds
ticker := time.NewTicker(30 * time.Second)
go func() {
    for range ticker.C {
        if client.IsConnected() {
            // Send presence update to keep connection alive
            client.SendPresence(types.PresenceAvailable)
        }
    }
}()
```

4. **Better Error Handling**
- Distinguish between temporary network errors and permanent issues
- Don't mark offline for temporary errors like:
  - Network timeout
  - DNS resolution failures
  - Temporary server errors (5xx)

5. **Connection Events Buffer**
- Don't react immediately to disconnection events
- Wait 30 seconds before marking offline
- If reconnection happens within that time, ignore the disconnect

## Implementation Priority

1. **Quick Fix** (Done in this update): Add retry logic to health monitor
2. **Medium Term**: Implement connection state machine
3. **Long Term**: Full connection resilience system with event buffering
