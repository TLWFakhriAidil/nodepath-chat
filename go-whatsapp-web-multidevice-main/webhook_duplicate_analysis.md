# Webhook Lead Duplicate Prevention Analysis

## Current Situation

### 1. **Database Level**
- ❌ **No UNIQUE constraint** on (user_id, phone)
- ❌ **No UNIQUE constraint** on (device_id, phone)
- ❌ **No UNIQUE constraint** on (device_id, user_id, phone, niche)
- Only PRIMARY KEY on ID (auto-generated)

### 2. **Application Level (webhook_lead.go)**

The webhook has some duplicate prevention:

```go
// Check if lead with same device_id, user_id, phone AND niche already exists
existingLead, err := leadRepo.GetLeadByDeviceUserPhoneNiche(device.ID, request.UserID, request.Phone, request.Niche)
if err == nil && existingLead != nil {
    // Lead already exists - skip creation
    return "DUPLICATE_SKIPPED"
}
```

**This checks for duplicates based on 4 fields:**
- device_id
- user_id  
- phone
- niche

### 3. **Potential Issues**

1. **Race Condition**: If webhook is called twice simultaneously:
   - Both requests check for existing lead (both find none)
   - Both proceed to create lead
   - Result: Duplicate leads created

2. **Different Niches**: Same phone can be added multiple times with different niches:
   - Phone: 60123456789, Niche: "Property" ✅
   - Phone: 60123456789, Niche: "Finance" ✅
   - Both will be created (intended behavior?)

3. **Different Devices**: Same phone can be added to different devices:
   - Device A: 60123456789 ✅
   - Device B: 60123456789 ✅
   - Both will be created

4. **No Database Constraint**: The application-level check can be bypassed:
   - Direct database inserts
   - Concurrent requests
   - Other API endpoints

### 4. **Observed Behavior**

From our duplicate cleanup earlier:
- Found 8,069 duplicate leads based on device_id + phone
- Some phones had up to 39 duplicates
- This confirms duplicates ARE being created

## Recommendations

### Option 1: Add Database Constraint (Recommended)
```sql
-- Prevent duplicate phone per device
ALTER TABLE leads 
ADD CONSTRAINT unique_device_phone 
UNIQUE (device_id, phone);
```

### Option 2: Add More Comprehensive Constraint
```sql
-- Prevent duplicate phone+niche per device
ALTER TABLE leads 
ADD CONSTRAINT unique_device_phone_niche 
UNIQUE (device_id, phone, niche);
```

### Option 3: Application-Level Improvements
1. Add a mutex/lock for the dedupe key
2. Use INSERT ... ON CONFLICT for PostgreSQL
3. Add retry logic with exponential backoff

## Questions for You:

1. **Should the same phone be allowed multiple times with different niches?**
   - Current: YES (by design)
   - Alternative: One phone per device regardless of niche

2. **Should the same phone be allowed on different devices?**
   - Current: YES
   - Alternative: One phone per user across all devices

3. **What's the expected behavior for webhook retries?**
   - Current: Returns "DUPLICATE_SKIPPED" 
   - Alternative: Update existing lead with new data

4. **Are webhooks called concurrently?**
   - If yes, we need stronger duplicate prevention

Let me know your preferences and I can implement the appropriate solution!