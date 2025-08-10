## How Sequence Processing SHOULD Work:

### 1. **Enrollment Phase** (when lead matches trigger):
When a lead with trigger "WARMEXAMA" is found, the system creates 4 separate records in sequence_contacts:

```
Lead: 60123456789 (trigger: WARMEXAMA)

Creates in sequence_contacts:
┌─────────────┬──────────┬─────────────────┬────────────────────┬─────────────────────┐
│ sequence_   │ current_ │ status          │ next_trigger_time  │ Notes               │
│ stepid      │ step     │                 │                    │                     │
├─────────────┼──────────┼─────────────────┼────────────────────┼─────────────────────┤
│ step-1-uuid │ 1        │ active          │ NOW()              │ Process immediately │
│ step-2-uuid │ 2        │ pending         │ NOW() + 12 hours   │ Wait 12 hours       │
│ step-3-uuid │ 3        │ pending         │ NOW() + 24 hours   │ Wait 24 hours       │
│ step-4-uuid │ 4        │ pending         │ NOW() + 36 hours   │ Wait 36 hours       │
└─────────────┴──────────┴─────────────────┴────────────────────┴─────────────────────┘
```

### 2. **Processing Phase** (every 15 seconds):
The processor queries:
```sql
SELECT * FROM sequence_contacts 
WHERE status = 'active' 
AND next_trigger_time <= NOW()
```

This should find ONLY Step 1 (because it's the only 'active' one).

### 3. **After Sending Message**:
When Step 1 message is sent successfully:
1. Update Step 1: status = 'completed'
2. NO NEED to find next record - it already exists!
3. Just update Step 2: status = 'active' (so it will be picked up when its time comes)

### 4. **The Flow**:
```
Hour 0:   Step 1 (active) → Send message → Mark completed
          Step 2 (pending) → Change to active
          Step 3 (pending) → Stay pending
          Step 4 (pending) → Stay pending

Hour 12:  Step 1 (completed) 
          Step 2 (active) → Send message → Mark completed
          Step 3 (pending) → Change to active
          Step 4 (pending) → Stay pending

Hour 24:  Step 1 (completed)
          Step 2 (completed)
          Step 3 (active) → Send message → Mark completed
          Step 4 (pending) → Change to active

Hour 36:  Step 1 (completed)
          Step 2 (completed)
          Step 3 (completed)
          Step 4 (active) → Send message → Mark completed
```

### WHAT'S ACTUALLY HAPPENING (The Bug):

1. **All records created correctly** ✓
2. **But the processor is:**
   - Not respecting the 'active' status properly
   - Not checking next_trigger_time properly
   - Processing ALL steps at once
   - Marking all as 'completed' immediately

### THE KEY CONCEPT:
- Each step is a SEPARATE record with its own timing
- The `next_trigger_time` on each record tells WHEN to process it
- Only 'active' status records should be processed
- After processing, activate the next pending record