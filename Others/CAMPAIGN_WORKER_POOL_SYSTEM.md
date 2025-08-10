# Campaign Scheduling and Worker Pool System

## YES - Campaigns Use the Same System as Sequences!

### 1. **Timezone Handling (8-Hour Offset)**
Both campaigns and sequences use the **SAME timezone adjustment**:

```sql
-- In broadcast_repository.go
WHERE scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
```

**What this means:**
- Server time is UTC-8 (US West Coast)
- Malaysia time is UTC+8 
- 8-hour offset added to handle the 16-hour difference
- **10-minute window**: Messages scheduled within next 10 minutes are picked up

### 2. **Worker Pool System (5 Workers per Device)**
Both campaigns and sequences use the **SAME worker pool**:

```go
// From broadcast_worker_processor.go
// Get the Ultra Scale Broadcast Manager (Worker Pool System)
broadcastManager := broadcast.GetBroadcastManager()

// Queue messages to Worker Pool
err := broadcastManager.QueueMessage(&msg)
```

**How it works:**
1. Background worker runs every 5 seconds
2. Finds all devices with pending messages
3. For each device, gets up to 100 pending messages
4. **Queues to worker pool** (not direct send)
5. Worker pool handles:
   - Status updates (pending → queued → sent)
   - Anti-spam delays (random between min/max)
   - Sequential sending with mutex lock
   - No duplicate sends (channel-based)

### 3. **Message Flow (Same for Both)**

```
Campaign/Sequence Created
    ↓
broadcast_messages created (status: pending)
    ↓
Worker checks every 5 seconds
    ↓
Finds messages where:
  - scheduled_at <= NOW() + 8 hours
  - scheduled_at >= NOW() + 8 hours - 10 minutes
    ↓
Updates to status: processing (locks them)
    ↓
Queues to Worker Pool (5 workers/device)
    ↓
Worker sends with anti-spam delay
    ↓
Updates to status: sent/failed
```

### 4. **Key Points**

**Scheduling:**
- ✅ Both use 8-hour timezone offset
- ✅ Both use 10-minute processing window
- ✅ Both check every 5 seconds

**Worker Pool:**
- ✅ Both use same worker pool system
- ✅ 5 workers per device
- ✅ Channel-based queuing (no duplicates)
- ✅ Mutex lock for sequential sending

**Anti-Spam:**
- ✅ Both respect min/max delay settings
- ✅ Random delay between messages
- ✅ Device-level rate limiting

### 5. **Configuration Examples**

**Campaign:**
```sql
campaigns table:
- scheduled_at: 2025-08-06 10:00:00 (Malaysia time)
- min_delay_seconds: 5
- max_delay_seconds: 15
```

**Sequence:**
```sql
sequence_steps table:
- send_time: "10:00" (daily time)
- min_delay_seconds: 10
- max_delay_seconds: 30
```

Both will be processed by the same system with timezone adjustment!

## Summary

**YES - Campaigns and Sequences use:**
- ✅ Same timezone handling (8-hour offset)
- ✅ Same 10-minute processing window
- ✅ Same worker pool system (5 workers/device)
- ✅ Same anti-spam delays
- ✅ Same broadcast_messages table
- ✅ Same status progression

The only difference is:
- **Campaigns**: One-time send to many
- **Sequences**: Multi-step send over time

But the underlying sending mechanism is **IDENTICAL**!