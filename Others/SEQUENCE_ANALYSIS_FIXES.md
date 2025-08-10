# Sequence System Analysis & Recommendations

## Current Findings from Database Analysis:

### 1. **Sequence Enrollment Issue**
- **Problem**: Sequences have empty triggers (`trigger: ""`)
- **Impact**: No leads are being enrolled because the matching query looks for trigger matches
- **Evidence**: 
  - WARM Sequence has 5 contacts enrolled (manually?)
  - COLD and HOT Sequences have 0 contacts despite 1742 matching leads
  - All sequences show `trigger: ` (empty) in the database

### 2. **Broadcast Messages Status**
- **WARM Sequence**: 1535 messages, 48 sent, 1487 failed
- **Success Rate**: Only 3% success rate
- **Issue**: High failure rate suggests device or messaging issues

### 3. **Sequence Progress Tracking**
- **Good**: All WARM Sequence contacts reached Step 4 and completed
- **Bad**: Status shows "completed" but next_trigger_time is in the future
- **Confusion**: Completed steps shouldn't have future trigger times

### 4. **Missing Step Linkage**
- **Issue**: broadcast_messages doesn't have sequence_stepid
- **Impact**: Can't track which specific step's message was sent/failed
- **Already Fixed**: Added column via ALTER TABLE

## Recommended Fixes:

### 1. **Fix Sequence Triggers**
```sql
-- Check current sequence triggers
SELECT id, name, trigger FROM sequences;

-- Update if empty (example)
UPDATE sequences SET trigger = 'warm_start' WHERE name = 'WARM Sequence';
UPDATE sequences SET trigger = 'cold_start' WHERE name = 'COLD Sequence';
UPDATE sequences SET trigger = 'hot_start' WHERE name = 'HOT Seqeunce';
```

### 2. **Fix the Enrollment Query**
The current query has issues:
- It's looking for `is_entry_point = true` in sequence_steps
- But also matching against empty sequence triggers

Updated query should be:
```sql
-- Better enrollment query
SELECT DISTINCT 
    l.id, l.phone, l.name, l.device_id, l.user_id, 
    s.id as sequence_id, s.trigger as entry_trigger
FROM leads l
JOIN sequences s ON s.is_active = true
WHERE l.trigger IS NOT NULL 
    AND l.trigger != ''
    AND s.trigger IS NOT NULL
    AND s.trigger != ''
    AND position(s.trigger in l.trigger) > 0
    AND NOT EXISTS (
        SELECT 1 FROM sequence_contacts sc
        WHERE sc.sequence_id = s.id 
        AND sc.contact_phone = l.phone
    )
```

### 3. **Update processContact to include sequence_stepid**
```go
// Add to contactJob struct
type contactJob struct {
    // ... existing fields ...
    sequenceStepID string  // Add this
}

// Update query in processSequenceContacts
query := `
    SELECT 
        sc.id, sc.sequence_id, sc.contact_phone, sc.contact_name,
        sc.current_trigger, sc.current_step,
        ss.content, ss.message_type, ss.media_url,
        ss.next_trigger, ss.trigger_delay_hours,
        COALESCE(sc.assigned_device_id, l.device_id) as preferred_device_id,
        COALESCE(ss.min_delay_seconds, 5) as min_delay_seconds,
        COALESCE(ss.max_delay_seconds, 15) as max_delay_seconds,
        l.user_id,
        sc.next_trigger_time,
        sc.sequence_stepid  -- Add this
    FROM sequence_contacts sc
    -- rest of query
`
```

### 4. **Fix Status Tracking**
```sql
-- Add new status for better tracking
ALTER TYPE sequence_contact_status ADD VALUE 'sent' AFTER 'completed';
ALTER TYPE sequence_contact_status ADD VALUE 'sequence_failed' AFTER 'failed';

-- Or if using VARCHAR, just use these values:
-- 'pending' - waiting to be processed
-- 'active' - ready to send
-- 'sent' - message sent successfully
-- 'completed' - step completed, next step activated
-- 'failed' - message send failed
-- 'sequence_failed' - entire sequence failed
```

### 5. **Create Monitoring View**
```sql
CREATE OR REPLACE VIEW sequence_progress_monitor AS
SELECT 
    s.name as sequence_name,
    sc.contact_phone,
    sc.current_step,
    sc.status as step_status,
    bm.status as message_status,
    sc.next_trigger_time,
    sc.completed_at,
    bm.sent_at,
    bm.error_message,
    CASE 
        WHEN sc.status = 'completed' THEN 'Step completed'
        WHEN sc.status = 'sent' THEN 'Message sent'
        WHEN sc.status = 'failed' THEN 'Message failed'
        WHEN sc.status = 'active' AND sc.next_trigger_time <= NOW() THEN 'Ready to send'
        WHEN sc.status = 'active' AND sc.next_trigger_time > NOW() THEN 'Scheduled'
        WHEN sc.status = 'pending' THEN 'Waiting'
        ELSE sc.status
    END as current_state
FROM sequence_contacts sc
JOIN sequences s ON s.id = sc.sequence_id
LEFT JOIN broadcast_messages bm ON 
    bm.sequence_id = sc.sequence_id 
    AND bm.recipient_phone = sc.contact_phone
    AND bm.sequence_stepid = sc.sequence_stepid
ORDER BY s.name, sc.contact_phone, sc.current_step;
```

## Summary of Issues:

1. **Empty sequence triggers** - preventing enrollment
2. **Missing sequence_stepid** in broadcast flow (now fixed)
3. **No status sync** between broadcast_messages and sequence_contacts
4. **High failure rate** (97%) in message sending
5. **Confusing status values** - completed with future trigger times

The core flow is working but needs these fixes to properly track progress and handle failures.
