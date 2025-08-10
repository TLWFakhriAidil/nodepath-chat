-- SEQUENCE DEBUGGING QUERIES FOR PENDING-FIRST APPROACH

-- 1. View all sequence contacts with their status
SELECT 
    sc.contact_phone,
    sc.current_step,
    sc.status,
    sc.next_trigger_time,
    CASE 
        WHEN sc.next_trigger_time <= NOW() THEN 'READY TO SEND'
        ELSE 'WAITING (' || (sc.next_trigger_time - NOW())::text || ')'
    END as time_status,
    s.name as sequence_name
FROM sequence_contacts sc
JOIN sequences s ON s.id = sc.sequence_id
ORDER BY sc.contact_phone, sc.current_step;

-- 2. Find earliest pending step for each contact (what worker will process)
WITH earliest_pending AS (
    SELECT DISTINCT ON (sc.sequence_id, sc.contact_phone)
        sc.contact_phone,
        sc.current_step,
        sc.status,
        sc.next_trigger_time,
        s.name as sequence_name
    FROM sequence_contacts sc
    JOIN sequences s ON s.id = sc.sequence_id
    WHERE sc.status = 'pending'
    ORDER BY sc.sequence_id, sc.contact_phone, sc.next_trigger_time ASC
)
SELECT * FROM earliest_pending
ORDER BY next_trigger_time ASC;

-- 3. Count by status
SELECT 
    status,
    COUNT(*) as count,
    MIN(next_trigger_time) as earliest_trigger,
    MAX(next_trigger_time) as latest_trigger
FROM sequence_contacts
GROUP BY status;

-- 4. Show what will happen in next processor run
SELECT 
    contact_phone,
    current_step,
    next_trigger_time,
    CASE 
        WHEN next_trigger_time <= NOW() THEN 'WILL SEND MESSAGE'
        ELSE 'WILL MARK ACTIVE'
    END as action
FROM (
    SELECT DISTINCT ON (sequence_id, contact_phone)
        contact_phone,
        current_step,
        next_trigger_time
    FROM sequence_contacts
    WHERE status = 'pending'
    ORDER BY sequence_id, contact_phone, next_trigger_time ASC
) earliest
ORDER BY next_trigger_time ASC
LIMIT 10;

-- 5. View broadcast messages created from sequences
SELECT 
    bm.recipient_phone,
    bm.status as message_status,
    bm.created_at,
    sc.current_step,
    sc.status as contact_status
FROM broadcast_messages bm
LEFT JOIN sequence_contacts sc ON sc.sequence_stepid::text = bm.sequence_stepid
WHERE bm.sequence_id IS NOT NULL
ORDER BY bm.created_at DESC
LIMIT 20;