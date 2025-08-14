# SEQUENCE TRIGGER PROCESSOR - PENDING-FIRST APPROACH

## Overview
This patch changes the sequence processing logic to:
1. Create ALL steps as PENDING (no active steps initially)
2. Worker finds the earliest pending step for each contact
3. If current time >= next_trigger_time → Send message and mark COMPLETED
4. If current time < next_trigger_time → Mark as ACTIVE (to track it's next)

## Benefits
- Simpler logic - no complex activation chains
- No duplicate messages - only processes when time arrives
- Clear status tracking: pending → active → completed
- Prevents race conditions

## Key Changes

### 1. enrollContactInSequence - ALL steps start as PENDING
```go
// Change from:
if i == 0 {
    status = "active"  // First step active
} else {
    status = "pending" // Others pending
}

// To:
status = "pending"  // ALL steps start as pending
```

### 2. processContactsReadyForMessages - New query logic
```go
// Old: Query for active contacts
WHERE sc.status = 'active' AND next_trigger_time <= NOW()

// New: Query for earliest pending per contact
WITH earliest_pending AS (
    SELECT DISTINCT ON (sequence_id, contact_phone)
    WHERE status = 'pending'
    ORDER BY sequence_id, contact_phone, next_trigger_time ASC
)
```

### 3. Time-based processing
```go
if triggerTime.After(now) {
    // Not ready - mark as active (tracking)
    UPDATE status = 'active'
} else {
    // Ready - send message and mark completed
    Queue to broadcast_messages
    UPDATE status = 'completed'
}
```

### 4. Remove updateContactProgress function
No longer needed since we don't activate next steps. The worker will naturally find the next pending step.

## Implementation Steps

1. Backup current file:
   ```bash
   cp sequence_trigger_processor.go sequence_trigger_processor.go.backup
   ```

2. Apply the changes to enrollContactInSequence function
3. Replace processContactsReadyForMessages with new logic
4. Remove or simplify updateContactProgress function
5. Test with a small sequence first

## Database Impact
- No schema changes needed
- Existing sequences will continue to work
- New enrollments will use pending-first approach

## Expected Behavior

### Example: 3-step sequence
```
Time 0: Lead enrolls
- Step 1: pending, triggers at 0+5min
- Step 2: pending, triggers at 0+5min+24hr  
- Step 3: pending, triggers at 0+5min+48hr

Time 5min: Worker runs
- Finds Step 1 (earliest pending, time reached)
- Sends message, marks completed
- Step 2 & 3 remain pending

Time 24hr+5min: Worker runs
- Finds Step 2 (earliest pending, time reached)
- Sends message, marks completed
- Step 3 remains pending

Time 48hr+5min: Worker runs
- Finds Step 3 (earliest pending, time reached)
- Sends message, marks completed
- Sequence complete!
```

This approach is cleaner and prevents the issue of all pending steps creating broadcast messages!