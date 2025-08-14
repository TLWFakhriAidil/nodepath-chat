-- Check sequences and their steps
-- Run this query to see if sequence data is being saved properly

-- 1. View all sequences
SELECT 
    id,
    name,
    niche,
    description,
    status,
    total_days,
    created_at
FROM sequences
ORDER BY created_at DESC;

-- 2. View sequence steps for a specific sequence
-- Replace 'YOUR_SEQUENCE_ID' with an actual sequence ID from the first query
SELECT 
    s.name as sequence_name,
    ss.day,
    ss.day_number,
    ss.message_type,
    ss.content,
    ss.media_url,
    ss.send_time,
    ss.min_delay_seconds,
    ss.max_delay_seconds
FROM sequence_steps ss
JOIN sequences s ON s.id = ss.sequence_id
-- WHERE s.id = 'YOUR_SEQUENCE_ID'
ORDER BY s.name, ss.day;

-- 3. Check if steps are being created at all
SELECT 
    COUNT(*) as total_steps,
    sequence_id
FROM sequence_steps
GROUP BY sequence_id;

-- 4. Check the actual column names in sequence_steps
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'sequence_steps'
ORDER BY ordinal_position;
