-- Quick fix for immediate sequence steps visibility
-- Run this SQL while your app is running

-- Check current data
SELECT 'Before fix:' as status;
SELECT COUNT(*) as total_steps FROM sequence_steps WHERE sequence_id = '394d567f-e5bd-476d-ae7c-c39f74819d70';

-- Fix the specific step data
UPDATE sequence_steps 
SET 
    day = 1,
    day_number = 1,
    message_type = 'text',
    send_time = '10:00',
    time_schedule = '10:00',
    trigger = 'asdaxx_day1',
    created_at = CURRENT_TIMESTAMP,
    updated_at = CURRENT_TIMESTAMP
WHERE sequence_id = '394d567f-e5bd-476d-ae7c-c39f74819d70';

-- Update sequence counts
UPDATE sequences 
SET 
    step_count = 1,
    total_steps = 1
WHERE id = '394d567f-e5bd-476d-ae7c-c39f74819d70';

-- Verify fix
SELECT 'After fix:' as status;
SELECT s.name, s.step_count, ss.day_number, ss.content, ss.send_time, ss.message_type
FROM sequences s
LEFT JOIN sequence_steps ss ON ss.sequence_id = s.id
WHERE s.id = '394d567f-e5bd-476d-ae7c-c39f74819d70';
