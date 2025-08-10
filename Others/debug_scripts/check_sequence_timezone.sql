-- Check and fix sequence tables for timezone issues

-- 1. Check sequence structure
SELECT 
    s.id,
    s.name,
    s.niche,
    s.target_status,
    s.is_active,
    s.created_at,
    COUNT(sc.id) as contact_count
FROM sequences s
LEFT JOIN sequence_contacts sc ON s.id = sc.sequence_id
GROUP BY s.id, s.name, s.niche, s.target_status, s.is_active, s.created_at;

-- 2. Check sequence steps with time fields
SELECT 
    ss.id,
    ss.sequence_id,
    ss.day,
    ss.send_time,
    ss.message_type,
    s.name as sequence_name
FROM sequence_steps ss
JOIN sequences s ON ss.sequence_id = s.id
ORDER BY s.name, ss.day;

-- 3. Check sequence contacts and their progress
SELECT 
    sc.id,
    sc.contact_phone,
    sc.current_day,
    sc.status,
    sc.added_at,
    sc.last_message_at,
    s.name as sequence_name
FROM sequence_contacts sc
JOIN sequences s ON sc.sequence_id = s.id
WHERE sc.status = 'active'
ORDER BY sc.added_at DESC;

-- 4. Fix any time format issues in sequence_steps
UPDATE sequence_steps
SET send_time = '09:00:00'
WHERE send_time IS NULL 
   OR LENGTH(send_time::text) > 8;

-- 5. Check for sequences that should be processing today
SELECT 
    s.name,
    sc.contact_phone,
    sc.current_day + 1 as next_day,
    ss.send_time,
    NOW() as current_server_time,
    CASE 
        WHEN sc.last_message_at IS NULL THEN 'Never sent'
        WHEN sc.last_message_at < NOW() - INTERVAL '24 hours' THEN 'Ready to send'
        ELSE 'Too soon'
    END as send_status
FROM sequences s
JOIN sequence_contacts sc ON s.id = sc.sequence_id
JOIN sequence_steps ss ON s.id = ss.sequence_id AND ss.day = sc.current_day + 1
WHERE s.is_active = true
  AND sc.status = 'active';