-- Emergency fix for sequence steps not showing
-- Run this SQL to check and fix the specific sequence data

-- First, let's see the current state
SELECT 'Current sequence data:' as info;
SELECT id, name, trigger, status, step_count, total_steps 
FROM sequences 
WHERE name = 'zxczxcxx';

SELECT 'Current sequence steps data:' as info;
SELECT id, sequence_id, day_number, content, send_time, trigger, message_type, created_at
FROM sequence_steps 
WHERE sequence_id = '394d567f-e5bd-476d-ae7c-c39f74819d70';

-- Fix the step data to ensure it's properly linked
UPDATE sequence_steps 
SET 
    day = COALESCE(day, day_number, 1),
    day_number = COALESCE(day_number, day, 1),
    message_type = COALESCE(message_type, 'text'),
    send_time = CASE 
        WHEN send_time IS NULL OR send_time = '' OR send_time = 'Invalid Date' 
        THEN '10:00' 
        ELSE send_time 
    END,
    time_schedule = CASE 
        WHEN time_schedule IS NULL OR time_schedule = '' OR time_schedule = 'Invalid Date' 
        THEN COALESCE(send_time, '10:00')
        ELSE time_schedule 
    END,
    trigger = CASE 
        WHEN trigger IS NULL OR trigger = '' 
        THEN 'asdaxx_day1'
        ELSE trigger 
    END,
    created_at = COALESCE(created_at, CURRENT_TIMESTAMP),
    updated_at = COALESCE(updated_at, CURRENT_TIMESTAMP)
WHERE sequence_id = '394d567f-e5bd-476d-ae7c-c39f74819d70';

-- Update the sequence step count
UPDATE sequences 
SET 
    step_count = (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = sequences.id),
    total_steps = (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = sequences.id),
    status = COALESCE(status, 'inactive')
WHERE id = '394d567f-e5bd-476d-ae7c-c39f74819d70';

-- Verify the fix
SELECT 'After fix - sequence data:' as info;
SELECT id, name, trigger, status, step_count, total_steps 
FROM sequences 
WHERE name = 'zxczxcxx';

SELECT 'After fix - sequence steps:' as info;
SELECT id, sequence_id, day_number, content, send_time, trigger, message_type, 
       CASE WHEN created_at IS NULL THEN 'NULL' ELSE 'SET' END as created_at_status,
       CASE WHEN updated_at IS NULL THEN 'NULL' ELSE 'SET' END as updated_at_status
FROM sequence_steps 
WHERE sequence_id = '394d567f-e5bd-476d-ae7c-c39f74819d70';
