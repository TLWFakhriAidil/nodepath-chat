# IMPLEMENTATION GUIDE: Pending-First Sequence Approach

## Quick Summary
Instead of creating one ACTIVE step and rest PENDING, we create ALL steps as PENDING. The worker finds the earliest pending step and decides what to do based on time.

## Changes Required

### 1. In `enrollContactInSequence` function (Line ~290)
```go
// CHANGE THIS:
if i == 0 {
    status = "active"
} else {
    status = "pending"
}

// TO THIS:
status = "pending"  // ALL steps start as pending
```

### 2. Replace entire `processSequenceContacts` function (Line ~380)
Replace the entire function with the new version that:
- Queries for earliest PENDING (not ACTIVE)
- Uses CTE to get one pending step per contact
- Calls new `processContactWithNewLogic` instead of `processContact`

### 3. Add new `processContactWithNewLogic` function
This function:
- Checks if current time >= next_trigger_time
- If NO: Updates status to 'active' (marks it as next in line)
- If YES: Sends message and marks as 'completed'

### 4. Update contactJob struct
Add these fields:
```go
sequenceStepID   string
nextTriggerTime  time.Time
```

### 5. Remove or simplify `updateContactProgress`
No longer needed since we don't chain activations.

## Testing Steps

1. Delete all existing sequence data (already done)
2. Create a test sequence with 3 steps
3. Add a lead with matching trigger
4. Watch the logs:
   - Should see "Step 1: PENDING - will trigger at..."
   - After 5 minutes: "Time reached for X step 1"
   - Message should appear in broadcast_messages
   - Only ONE message at a time

## Benefits
- No more cascade of all pending creating messages
- Clear state machine: pending → active → completed
- Time-based processing is explicit
- Easier to debug and understand

## Rollback Plan
If issues occur, you have:
- sequence_trigger_processor.go.backup
- Git history with the working version