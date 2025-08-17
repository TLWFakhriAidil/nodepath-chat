# Campaign Not Triggering Issue Analysis

## Problem
Campaigns 59 and 60 are repeatedly finding matching leads but not creating broadcast messages or updating their status to "triggered".

## Root Cause
Looking at the logs:
```
2025/07/27 13:16:03 Campaign 60 - Matching leads:
2025/07/27 13:16:03 - Phone: 60108924904, Device: d409cadc-75e2-4004-a789-c2bad0b31393, Niche: GRR, TargetStatus: prospect
2025/07/27 13:16:03 Campaign 60 - UserID: de078f16-3266-4ab3-8153-a248b015228f, Niche: GRR, TargetStatus: prospect, ShouldSend: 1
```

The campaign is finding 1 matching lead but the messages are not being queued. The campaign keeps repeating because its status remains "pending" instead of being updated to "triggered".

## Possible Issues

1. **Campaign Time Issue**: The campaign might be set to trigger at a specific time that has already passed, causing it to be repeatedly selected but not actually executed.

2. **QueueMessage Failure**: The `broadcastRepo.QueueMessage(msg)` call might be failing silently or the error logging is not showing up.

3. **Connection Issue**: The logs show "connection unregistered" and "connection registered" repeatedly, which might indicate the device connection is unstable.

## Solution

To fix this issue, you need to:

1. **Check the campaign date/time settings** - Make sure they're not set to a past time that keeps getting selected.

2. **Force update the campaign status** to stop the loop:
```sql
UPDATE campaigns 
SET status = 'triggered' 
WHERE id IN (59, 60);
```

3. **Check for any constraint violations** in the broadcast_messages table that might prevent insertion.

4. **Enable more verbose logging** to see if QueueMessage is actually being called and what errors it might return.

5. **Verify the device connection** is stable before processing campaigns.
