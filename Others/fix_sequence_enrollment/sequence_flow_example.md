# Sequence Flow Example

## When Lead is Enrolled (at 10:00 AM)

### All 5 Steps Created Immediately:

```
Current Time: 2025-01-19 10:00:00

Step 1: Status = pending, next_trigger = 10:05:00 (NOW + 5 min)
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
- Processor runs: `UPDATE WHERE status = 'pending' AND next_trigger_time <= NOW()`
- Step 1: pending → **active**
- Message queued and sent
- Step 1: active → **completed**

### Day 2 at 10:05 AM:
- Step 2: pending → **active**
- Message sent
- Step 2: active → **completed**

### Day 4 at 10:05 AM (Step 2 + 48 hours):
- Step 3: pending → **active**
- Message sent
- Step 3: active → **completed**

### Day 7 at 10:05 AM (Step 3 + 72 hours):
- Step 4: pending → **active**
- Message sent
- Step 4: active → **completed**

### Day 8 at 10:05 AM (Step 4 + 24 hours):
- Step 5: pending → **active**
- Message sent
- Step 5: active → **completed**
- All steps completed → Remove trigger from lead

## Database Status Throughout:

```sql
-- After enrollment (10:00 AM)
SELECT current_step, status, next_trigger_time FROM sequence_contacts WHERE contact_phone = '60123456789';

1 | pending | 2025-01-19 10:05:00
2 | pending | 2025-01-20 10:05:00
3 | pending | 2025-01-22 10:05:00
4 | pending | 2025-01-25 10:05:00
5 | pending | 2025-01-26 10:05:00

-- After Step 1 sent (10:05 AM)
1 | completed | 2025-01-19 10:05:00
2 | pending   | 2025-01-20 10:05:00
3 | pending   | 2025-01-22 10:05:00
4 | pending   | 2025-01-25 10:05:00
5 | pending   | 2025-01-26 10:05:00

-- After all steps complete
1 | completed | 2025-01-19 10:05:00
2 | completed | 2025-01-20 10:05:00
3 | completed | 2025-01-22 10:05:00
4 | completed | 2025-01-25 10:05:00
5 | completed | 2025-01-26 10:05:00
```
