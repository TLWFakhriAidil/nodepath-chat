# Direct Broadcast Sequences Implementation

## Overview
This implementation allows sequences to work like campaigns - creating all messages upfront in `broadcast_messages` with proper scheduling, completely bypassing the `sequence_contacts` table.

## Key Features

### 1. Direct Enrollment
- Leads with triggers are enrolled directly into `broadcast_messages`
- No intermediate `sequence_contacts` table needed
- All messages created upfront with calculated `scheduled_at` times

### 2. Sequence Linking
- Automatic progression: COLD → WARM → HOT
- Only checks active status for initial enrollment
- Linked sequences are always followed regardless of active status

### 3. UUID Handling
- Uses repository pattern `QueueMessage()` which handles NULL values properly
- Validates that device_id and user_id are not empty strings
- Prevents "invalid input syntax for type uuid" errors

## Implementation Details

### Files Created:
1. `src/usecase/direct_broadcast_processor.go` - Main processor for direct enrollment
2. Added `ProcessDirectBroadcast()` method to sequence_trigger_processor.go

### How It Works:

1. **Find Eligible Leads**:
```sql
SELECT leads with triggers that match active sequence entry points
WHERE device_id IS NOT NULL AND user_id IS NOT NULL
AND NOT already enrolled (no pending/sent messages)
```

2. **Process Sequence Chain**:
- Start with matched sequence (must be active)
- Get all steps and create messages
- Follow next_trigger to linked sequences (don't check active)
- Continue until no more links

3. **Create Messages**:
- First message: NOW + 5 minutes
- Subsequent messages: Based on trigger_delay_hours
- Uses `broadcastRepo.QueueMessage()` for proper NULL handling

### Usage:

To use Direct Broadcast instead of the old sequence_contacts approach:

1. Call the processor manually:
```go
processor := usecase.NewDirectBroadcastProcessor(db)
enrolledCount, err := processor.ProcessDirectEnrollments()
```

2. Or set up a cron job:
```go
gocron.Every(5).Minutes().Do(func() {
    processor := usecase.NewSequenceTriggerProcessor(db, nil)
    processor.ProcessDirectBroadcast()
})
```

### Benefits:
- Simpler architecture (no sequence_contacts table)
- Better performance (one write operation)
- Predictable scheduling (like campaigns)
- No complex state management
- Automatic sequence linking

### Migration Note:
The old `processSequenceContacts()` method still exists for backward compatibility. You can run both systems in parallel during migration, then disable the old one.

## Testing

1. Ensure leads have valid device_id and user_id (not empty strings)
2. Set up sequences with proper triggers and linking
3. Add triggers to leads
4. Run the processor
5. Check broadcast_messages table for created messages

The system now handles sequences exactly like campaigns - pre-scheduled messages that device workers will send at the right time!
