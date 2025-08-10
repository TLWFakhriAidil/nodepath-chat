# CORRECTED Sequence Flow Example

## When Lead is Enrolled (at 10:00 AM)

### All 5 Steps Created Immediately:

```
Current Time: 2025-01-19 10:00:00

Step 1: Status = ACTIVE,  next_trigger = 10:05:00 (NOW + 5 min) ⚡
        trigger_delay_hours = 24

Step 2: Status = pending, next_trigger = 2025-01-20 10:05:00 (Step 1 + 24h)
        trigger_delay_hours = 48

Step 3: Status = pending, next_trigger = 2025-01-22 10:05:00 (Step 2 + 48h)
        trigger_delay_hours = 72

Step 4: Status = pending, next_trigger = 2025-01-25 10:05:00 (Step 3 + 72h)
        trigger_delay_hours = 24

Step 5: Status = pending, next_trigger = 2025-01-26 10:05:00 (Step 4 + 24h)
        trigger_delay_hours = 0
```

## Processing Timeline:

### At 10:05 AM (5 minutes later):
- Processor runs: `WHERE status = 'active' AND next_trigger_time <= NOW()`
- Finds Step 1 (already active, time has come)
- Message queued and sent ✅
- Step 1: active → **completed**
- No other steps are activated yet (their time hasn't come)

### Day 2 at 10:05 AM:
- First: `UPDATE WHERE status = 'pending' AND next_trigger_time <= NOW()`
- Step 2: pending → **active** ⚡
- Then: Process active contacts
- Message sent ✅
- Step 2: active → **completed**

### Day 4 at 10:05 AM:
- Step 3: pending → **active** → message sent → **completed**

### Day 7 at 10:05 AM:
- Step 4: pending → **active** → message sent → **completed**

### Day 8 at 10:05 AM:
- Step 5: pending → **active** → message sent → **completed**
- All steps completed → Remove trigger from lead 🎉

## Key Differences:

1. **First Step = ACTIVE** immediately (will send in 5 minutes)
2. **Other Steps = PENDING** (waiting for their time)
3. **Processing only checks**: `status = 'active' AND next_trigger_time <= NOW()`
4. **Activation happens**: When pending steps reach their time

## SQL Query Flow:

```sql
-- Main processing query (every 15 seconds)
SELECT * FROM sequence_contacts 
WHERE status = 'active' 
AND next_trigger_time <= NOW()
AND processing_device_id IS NULL;

-- Activation query (also every 15 seconds)
UPDATE sequence_contacts 
SET status = 'active'
WHERE status = 'pending' 
AND next_trigger_time <= NOW();
```

## Visual Status Flow:

```
Step 1: ACTIVE → [send] → COMPLETED
Step 2: pending → [wait] → ACTIVE → [send] → COMPLETED
Step 3: pending → [wait] → ACTIVE → [send] → COMPLETED
Step 4: pending → [wait] → ACTIVE → [send] → COMPLETED
Step 5: pending → [wait] → ACTIVE → [send] → COMPLETED
```
