# Sequence Device Assignment - Like Campaigns

## ✅ Changes Made:

### 1. **Added device_id to sequence_contacts during enrollment**

**Before:**
```sql
INSERT INTO sequence_contacts (
    sequence_id, contact_phone, contact_name, 
    current_step, status, completed_at, current_trigger,
    next_trigger_time, sequence_stepid
) VALUES (...)
```

**After:**
```sql
INSERT INTO sequence_contacts (
    sequence_id, contact_phone, contact_name, 
    current_step, status, completed_at, current_trigger,
    next_trigger_time, sequence_stepid, assigned_device_id
) VALUES (..., lead.DeviceID)
```

### 2. **Query uses assigned_device_id first**

```sql
COALESCE(sc.assigned_device_id, l.device_id) as preferred_device_id
```
- First tries to use `assigned_device_id` from sequence_contacts
- Falls back to `device_id` from leads table if not set

### 3. **Strict device matching (no fallback)**

```go
func selectDeviceForContact(preferredDeviceID string, loads map[string]DeviceLoad) string {
    // ONLY use the device that owns the lead
    if preferredDeviceID != "" {
        if load, ok := loads[preferredDeviceID]; ok && load.CanAcceptMore() {
            return preferredDeviceID
        }
        // Device not available - do NOT fall back
        return ""
    }
    return ""
}
```

## 🔄 How It Works Now (Exactly Like Campaigns):

### During Enrollment:
```
Lead (device_id: device-A) → Enroll in Sequence
→ Creates sequence_contacts with assigned_device_id = device-A
```

### During Processing:
```
Sequence Processor runs
→ Fetches contacts WHERE device is online
→ Device A processes ONLY its assigned contacts
→ Device B processes ONLY its assigned contacts
→ If Device A offline, its contacts wait
```

## 📊 Benefits:

1. **Consistent Sender**: Lead always receives messages from same device/number
2. **Data Isolation**: Devices only see their own leads
3. **Predictable Load**: Know exactly which device handles which contacts
4. **Same as Campaigns**: Unified behavior across system

## ⚠️ Important Notes:

1. **Device Assignment Required**: Every lead MUST have a device_id
2. **No Cross-Device Processing**: Strict isolation between devices
3. **Device Must Be Online**: Offline devices = their leads wait
4. **Load Balancing**: Distribute leads evenly when creating them

The sequence system now works EXACTLY like campaigns - strict device ownership!