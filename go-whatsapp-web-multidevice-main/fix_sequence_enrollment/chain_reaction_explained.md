# Chain Reaction Flow - How It Works

## Initial State (When Lead Enrolled at 10:00 AM)

```sql
Step 1: status = 'active',  next_trigger = 10:05 AM  ← Ready to process in 5 min
Step 2: status = 'pending', next_trigger = Day 2 10:05 AM
Step 3: status = 'pending', next_trigger = Day 4 10:05 AM  
Step 4: status = 'pending', next_trigger = Day 7 10:05 AM
Step 5: status = 'pending', next_trigger = Day 8 10:05 AM
```

## Processing Flow:

### 10:05 AM - Worker Run #1
1. **Finds**: Step 1 (active + time reached)
2. **Sends**: Message to customer
3. **Updates**: 
   - Step 1 → `completed`
   - Step 2 → `active` ⚡ (but next_trigger is tomorrow)
4. **Result**:
   ```sql
   Step 1: status = 'completed' ✓
   Step 2: status = 'active',  next_trigger = Day 2 10:05 AM ← Active but not ready
   Step 3: status = 'pending', next_trigger = Day 4 10:05 AM
   Step 4: status = 'pending', next_trigger = Day 7 10:05 AM
   Step 5: status = 'pending', next_trigger = Day 8 10:05 AM
   ```

### 10:20 AM - Worker Run #2
- **Finds**: Nothing! (Step 2 is active but next_trigger > NOW)
- **No messages sent**

### Day 2 at 10:05 AM - Worker Run #3
1. **Finds**: Step 2 (active + time reached)
2. **Sends**: Message
3. **Updates**:
   - Step 2 → `completed`
   - Step 3 → `active` ⚡
4. **Result**:
   ```sql
   Step 1: status = 'completed' ✓
   Step 2: status = 'completed' ✓
   Step 3: status = 'active',  next_trigger = Day 4 10:05 AM ← Active but waiting
   Step 4: status = 'pending', next_trigger = Day 7 10:05 AM
   Step 5: status = 'pending', next_trigger = Day 8 10:05 AM
   ```

### Day 4 at 10:05 AM - Worker Run #4
1. **Finds**: Step 3 (active + time reached)
2. **Chain continues**...

## Key Points:

1. **Only ONE step is active at a time** (except the very first moment)
2. **Worker query is simple**: `WHERE status = 'active' AND next_trigger_time <= NOW()`
3. **No separate activation needed** - it's part of the completion process
4. **Steps wait patiently** - Active but future-dated steps are ignored

## The Beauty of This Approach:

- **No complex timing logic** in the processor
- **Natural flow** - complete one, activate next
- **Respects delays** - Step 2 waits even though it's active
- **Single source of truth** - The `next_trigger_time` controls everything

## SQL Flow in updateContactProgress:

```sql
-- In one transaction:
BEGIN;

-- 1. Complete current step
UPDATE sequence_contacts SET status = 'completed' WHERE id = $1;

-- 2. Activate next pending step (lowest step number)
UPDATE sequence_contacts 
SET status = 'active'
WHERE sequence_id = $X AND contact_phone = $Y AND status = 'pending'
AND current_step = (SELECT MIN(current_step) FROM ... WHERE status = 'pending');

COMMIT;
```

## Why This Works Better:

1. **Predictable**: Always know which step is next
2. **Atomic**: Complete + Activate in one transaction
3. **No Race Conditions**: Only one active step per contact
4. **Simple Query**: Worker just looks for active + time reached
5. **Failsafe**: If worker crashes, active step remains active
