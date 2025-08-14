# 🔍 ROOT CAUSE ANALYSIS: Why GRR Campaigns Fail

## THE BUG FOUND:

The campaign trigger code in `optimized_campaign_trigger.go` is trying to insert `MinDelay` and `MaxDelay` into the `broadcast_messages` table, but **these columns don't exist there!**

### What the code does (line 160-165):
```go
msg := domainBroadcast.BroadcastMessage{
    UserID:         campaign.UserID,
    DeviceID:       lead.DeviceID,
    // ...
    MinDelay:       campaign.MinDelaySeconds,  // ❌ PROBLEM!
    MaxDelay:       campaign.MaxDelaySeconds,  // ❌ PROBLEM!
}
```

### What the database has:
- `campaigns` table HAS `min_delay_seconds` and `max_delay_seconds` ✅
- `broadcast_messages` table DOES NOT have these columns ❌

## WHY SEQUENCES WORK:

Looking at the code, sequences handle delays differently:
1. They get delays from the `sequences` or `sequence_steps` tables
2. They don't try to store delays in `broadcast_messages`
3. The delay is applied at processing time by reading from the source table

## THE EVIDENCE:

1. **Campaign 59 (GRR)**: 
   - Found 1 lead ✅
   - Tried to create broadcast message ✅
   - FAILED with error: "device not connected" ❌
   - But the real issue might be the column mismatch

2. **Campaign 60 (GRR)**:
   - Found 1 lead ✅
   - Can't create message because of time issue + column error ❌
   - Stuck in infinite loop

3. **Sequences**:
   - Successfully created 36 messages in last 7 days ✅
   - 6 sent, 30 pending
   - Working because they don't have this bug

## THE FIX NEEDED:

The `BroadcastMessage` struct should NOT have `MinDelay` and `MaxDelay` fields that map to database columns. Instead, delays should be:
1. Stored in the campaign/sequence tables
2. Retrieved when processing messages
3. Applied at send time

## SUMMARY:

**Campaigns fail because the code tries to insert into non-existent database columns!**
**Sequences work because they handle delays differently!**
