# Duplicate Prevention in Self-Healing Architecture

## How GetOrRefreshClient Prevents Duplicates

### 1. **Initial Check - Returns Existing Client**
```go
// FIRST - Always check for existing healthy client
client, err := cm.GetClient(deviceID)
if err == nil && client != nil && client.IsConnected() && client.IsLoggedIn() {
    return client, nil  // ← NO REFRESH, NO DUPLICATE!
}
```

### 2. **Per-Device Mutex**
```go
// Each device has its own mutex
deviceMutex := wcm.refreshMutex[deviceID]
deviceMutex.Lock()
defer deviceMutex.Unlock()
```
This ensures only ONE refresh can happen per device at a time.

### 3. **Refresh-in-Progress Check**
```go
if wcm.refreshing[deviceID] {
    // Another worker is already refreshing
    time.Sleep(2 * time.Second)
    // Try to get client again (it might be ready now)
    client, err := cm.GetClient(deviceID)
    if err == nil && client != nil && client.IsConnected() {
        return client, nil
    }
    return nil, fmt.Errorf("refresh in progress")
}
```

## Flow Diagram

```
Worker 1 calls GetOrRefreshClient("device123")
    ↓
Check existing client → Healthy? → Return it (NO DUPLICATE)
    ↓ Not healthy
Lock mutex for device123
    ↓
Check if already refreshing → Yes → Wait & retry
    ↓ No
Mark as refreshing
    ↓
Perform refresh (create new client)
    ↓
Register new client (replaces old one)
    ↓
Unlock mutex

Meanwhile, Worker 2 calls GetOrRefreshClient("device123")
    ↓
Check existing client → Not healthy
    ↓
Try to lock mutex → BLOCKED (Worker 1 has it)
    ↓
Wait...
    ↓
Worker 1 finishes → Mutex unlocked
    ↓
Worker 2 locks mutex
    ↓
Check existing client → Now healthy! → Return it (NO DUPLICATE)
```

## Key Points

1. **Always checks existing client first** - If healthy, returns immediately
2. **Per-device mutex** - Only one refresh per device at a time
3. **Refresh tracking** - Knows which devices are being refreshed
4. **Double-check after wait** - If another worker refreshed it, use that client

## Result
- ✅ No duplicate clients
- ✅ No race conditions
- ✅ Efficient - reuses healthy clients
- ✅ Thread-safe - mutex protection
