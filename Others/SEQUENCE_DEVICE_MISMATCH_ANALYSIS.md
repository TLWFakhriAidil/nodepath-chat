# Sequence Device Assignment Issue Analysis

## The Problem

When creating `broadcast_messages` records, the `device_id` doesn't match the `assigned_device_id` from `sequence_contacts`.

## Root Cause

1. **Sequence Contact Creation**: When a lead is enrolled, it copies the lead's `device_id` to `assigned_device_id`:
   ```sql
   INSERT INTO sequence_contacts (assigned_device_id) 
   VALUES (lead.device_id)
   ```

2. **Message Processing**: The system checks if the assigned device is available:
   ```go
   deviceID := s.selectDeviceForContact(job.preferredDevice.String, deviceLoads)
   ```
   
3. **Issue**: If the assigned device is offline/overloaded, it:
   - Returns empty string
   - Skips the contact entirely
   - Doesn't try other devices

## Why This Happens

The current logic is STRICT - it only uses the device that "owns" the lead. This is intentional to:
- Maintain conversation continuity
- Prevent cross-device message mixing
- Keep leads tied to their original device

## Solutions

### Option 1: Fix to Always Use Assigned Device (RECOMMENDED)
This ensures `broadcast_messages.device_id` always matches `sequence_contacts.assigned_device_id`:

```go
// In processContact function, replace the device selection with:
deviceID := job.preferredDevice.String
if deviceID == "" {
    logrus.Warnf("No assigned device for contact %s", job.phone)
    return false
}

// Don't check if device is available - let broadcast processor handle it
// This way, the message is queued even if device is offline
```

### Option 2: Allow Fallback to Other Devices
If you want messages to be sent even when assigned device is offline:

```go
func (s *SequenceTriggerProcessor) selectDeviceForContact(preferredDeviceID string, loads map[string]DeviceLoad) string {
    // Try preferred device first
    if preferredDeviceID != "" {
        if load, ok := loads[preferredDeviceID]; ok && load.CanAcceptMore() {
            return preferredDeviceID
        }
    }
    
    // NEW: Fallback to any available device
    for deviceID, load := range loads {
        if load.CanAcceptMore() {
            logrus.Warnf("Using fallback device %s instead of preferred %s", 
                deviceID, preferredDeviceID)
            return deviceID
        }
    }
    
    return ""
}
```

### Option 3: Update Assigned Device When Using Different One
Keep the current logic but update `assigned_device_id` when a different device is used:

```go
// After selecting device
if deviceID != job.preferredDevice.String {
    // Update assigned device
    s.db.Exec(`
        UPDATE sequence_contacts 
        SET assigned_device_id = $1 
        WHERE id = $2
    `, deviceID, job.contactID)
}
```

## SQL to Check Current State

```sql
-- See mismatches between assigned and actual device
SELECT 
    sc.contact_phone,
    sc.assigned_device_id,
    bm.device_id as broadcast_device_id,
    CASE 
        WHEN sc.assigned_device_id = bm.device_id THEN 'MATCH'
        ELSE 'MISMATCH'
    END as status
FROM sequence_contacts sc
JOIN broadcast_messages bm ON bm.recipient_phone = sc.contact_phone
WHERE bm.sequence_id = sc.sequence_id
  AND bm.created_at > NOW() - INTERVAL '1 hour'
ORDER BY bm.created_at DESC
LIMIT 20;

-- Check device availability
SELECT 
    d.id,
    d.device_name,
    d.status,
    COUNT(DISTINCT sc.contact_phone) as assigned_contacts,
    CASE 
        WHEN d.status = 'online' THEN 'Available'
        ELSE 'Not Available'
    END as availability
FROM user_devices d
LEFT JOIN sequence_contacts sc ON sc.assigned_device_id = d.id
GROUP BY d.id, d.device_name, d.status
ORDER BY assigned_contacts DESC;
```

## My Recommendation

Use **Option 1** - Always queue with assigned device. This:
- Maintains data integrity
- Keeps the relationship clear
- Lets broadcast processor handle offline devices
- Prevents confusion about which device sent what

The broadcast processor already handles offline devices by skipping them, so messages will wait until the device comes online.
