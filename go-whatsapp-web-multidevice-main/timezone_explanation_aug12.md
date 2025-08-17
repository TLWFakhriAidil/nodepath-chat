# Understanding the 125 Failed Messages - Scheduled Times Explained

## The Key Finding: Messages Were Scheduled for August 12 (NOT August 11 + 8 hours)

### What the Data Shows:

**Scheduled Times (in database)**:
- Earliest: 2025-08-12 00:05:11 
- Latest: 2025-08-12 04:05:11

**Failed Times (when they actually failed)**:
- Earliest: 2025-08-11 16:05:11
- Latest: 2025-08-11 20:05:11

### The Timeline (What Actually Happened):

```
Database Schedule:        Aug 12, 00:05 (midnight)
                                ↓
System adds +8 hours:     Aug 12, 08:05 (Malaysia time)
                                ↓
But compares to NOW():    Aug 11, 16:05 (current time)
                                ↓
Thinks it's time to send: Because 16:05 + 8 = 00:05 next day
                                ↓
Tries to send:            FAILS - devices disconnected
```

## The Answer to Your Question:

**The 125 failed messages were scheduled for August 12** in the database (NOT August 11 + 8 hours).

Here's what happened:
1. Messages were created with `scheduled_at = '2025-08-12 00:05:11'` (August 12)
2. The system's query adds 8 hours: `WHERE scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)`
3. On August 11 at 16:05 (4 PM), the system calculated: 16:05 + 8 hours = 00:05 (midnight)
4. It thought "These messages scheduled for Aug 12 00:05 should be sent NOW"
5. But this was WRONG - it was still August 11!

## Why This Is a Problem:

The system is comparing:
- `scheduled_at` (stored as Aug 12 in database)
- With `NOW() + 8 hours` (which on Aug 11 afternoon equals Aug 12 midnight)

This makes messages process 8 hours EARLY in real time.

## The Fix:

If your MySQL server is already in Malaysia timezone, you should:
1. Remove the `+ INTERVAL 8 HOUR` from all queries
2. Just use `WHERE scheduled_at <= NOW()`

Or if MySQL is in UTC:
1. Store all times in UTC
2. Convert only for display purposes

## To Verify:
Run this SQL to see the timezone confusion:
```sql
SELECT 
    NOW() as mysql_now,
    DATE_ADD(NOW(), INTERVAL 8 HOUR) as now_plus_8,
    '2025-08-12 00:05:11' as scheduled_time,
    CASE 
        WHEN '2025-08-12 00:05:11' <= DATE_ADD(NOW(), INTERVAL 8 HOUR) 
        THEN 'Would process now' 
        ELSE 'Would wait' 
    END as decision;
```

The messages were correctly scheduled for August 12, but the timezone logic made them process on August 11!
