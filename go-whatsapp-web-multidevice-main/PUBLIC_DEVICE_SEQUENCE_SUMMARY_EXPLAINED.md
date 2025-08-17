# Sequence Summary Implementation for Public Device View

## How It Works:

### 1. **URL Structure**
The public device view uses URL parameters to filter data:
- Pattern: `/public/device/{device_name}/sequences`
- Example: `/public/device/Device-123/sequences`

### 2. **The Query Logic**

From the Go code in `GetPublicDeviceSequences` function:

```sql
SELECT DISTINCT
    s.id,
    s.name,
    s.trigger,
    (SELECT COUNT(DISTINCT ss.id) FROM sequence_steps ss WHERE ss.sequence_id = s.id) as total_flows,
    
    -- Total unique contacts (using sequence_stepid + phone + device combination)
    COUNT(DISTINCT CONCAT(bm.sequence_stepid, '|', bm.recipient_phone, '|', bm.device_id)) as total_contacts,
    
    -- Successfully sent messages
    COUNT(DISTINCT CASE WHEN bm.status = 'sent' AND (bm.error_message IS NULL OR bm.error_message = '') 
        THEN CONCAT(bm.sequence_stepid, '|', bm.recipient_phone, '|', bm.device_id) END) as contacts_done,
    
    -- Failed messages
    COUNT(DISTINCT CASE WHEN bm.status IN ('failed', 'error') OR (bm.status = 'sent' AND bm.error_message IS NOT NULL AND bm.error_message != '') 
        THEN CONCAT(bm.sequence_stepid, '|', bm.recipient_phone, '|', bm.device_id) END) as contacts_failed
        
FROM sequences s
INNER JOIN broadcast_messages bm ON bm.sequence_id = s.id
WHERE bm.device_id = ?  -- This is the deviceID parameter
GROUP BY s.id, s.name, s.trigger
ORDER BY s.created_at DESC
```

### 3. **Key Components Explained**

#### Device Filtering:
```go
// First get device ID from device name
var deviceID string
err = db.QueryRow("SELECT id FROM user_devices WHERE device_name = ?", deviceName).Scan(&deviceID)

// Then use deviceID in the main query
WHERE bm.device_id = ?
```

#### Grouping by Sequence:
- `GROUP BY s.id` - Groups all broadcast messages by sequence
- Shows only sequences that have messages for this specific device

#### Counting Logic:
- Uses `CONCAT(bm.sequence_stepid, '|', bm.recipient_phone, '|', bm.device_id)` to create unique identifier
- This ensures each step message to each recipient on each device is counted separately
- `COUNT(DISTINCT ...)` prevents counting duplicates

### 4. **What Each Count Means**

1. **total_contacts**: Total unique step messages that should be sent
   - Includes all statuses (pending, sent, failed)
   - Each sequence step for each recipient counts as 1

2. **contacts_done**: Successfully sent messages
   - `status = 'sent'` AND no error message
   - These are confirmed delivered

3. **contacts_failed**: Failed messages
   - `status IN ('failed', 'error')` OR
   - `status = 'sent'` but has error message

4. **contacts_remaining** (calculated in frontend):
   - `total_contacts - contacts_done - contacts_failed`

### 5. **Success Rate Calculation**
```go
successRate := 0.0
if seq.TotalContacts > 0 {
    successRate = float64(seq.ContactsDone) / float64(seq.TotalContacts) * 100
}
```

### 6. **Example Output**
```json
{
  "status": 200,
  "code": "SUCCESS",
  "message": "Sequences retrieved successfully",
  "results": [
    {
      "id": "seq-123",
      "name": "Welcome Sequence",
      "trigger": "welcome",
      "total_flows": 7,          // 7 steps in this sequence
      "total_contacts": 350,     // 50 recipients × 7 steps = 350 total messages
      "contacts_done": 300,      // 300 successfully sent
      "contacts_failed": 20,     // 20 failed
      "success_rate": "85.7"     // 300/350 = 85.7%
    }
  ]
}
```

### 7. **Why This Approach?**

1. **Accurate Counting**: Each sequence step is counted separately
2. **Device Isolation**: Only shows data for the specific device
3. **Performance**: Single query with efficient GROUP BY
4. **Real-time**: Based on actual broadcast_messages table

### 8. **The sequence_stepid Column**
- Links each broadcast message to a specific sequence step
- Allows tracking which step of the sequence each message belongs to
- Essential for accurate progress tracking

## Summary

The public device sequence summary:
1. Filters by device_id from URL parameter
2. Groups broadcast messages by sequence_id
3. Counts distinct combinations of (sequence_stepid + phone + device)
4. Provides real-time status of sequence execution for that specific device

This gives an accurate view of:
- How many sequence messages should be sent
- How many have been successfully sent
- How many failed
- Success rate per sequence

All filtered to show only data relevant to the specific device being viewed.