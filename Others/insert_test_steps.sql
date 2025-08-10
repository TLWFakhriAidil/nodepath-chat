-- Insert test steps for your sequences to verify the system is working

-- First, check what we have
SELECT id, name FROM sequences 
WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f';

-- Insert a test step for sequence '394d567f-e5bd-476d-ae7c-c39f74819d70'
INSERT INTO sequence_steps (
    id,
    sequence_id,
    day_number,
    trigger,
    next_trigger,
    trigger_delay_hours,
    is_entry_point,
    message_type,
    content,
    media_url,
    caption,
    send_time,
    time_schedule,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    1,
    'asdaxx',
    'asdaxx_day2',
    24,
    true,
    'text',
    'Test message for day 1',
    '',
    '',
    '',
    '10:00',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Insert a test step for sequence 'b7319119-4d21-43cc-97e2-9644ded0608c'
INSERT INTO sequence_steps (
    id,
    sequence_id,
    day_number,
    trigger,
    next_trigger,
    trigger_delay_hours,
    is_entry_point,
    message_type,
    content,
    media_url,
    caption,
    send_time,
    time_schedule,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    'b7319119-4d21-43cc-97e2-9644ded0608c',
    1,
    'FUNP',
    'FUNP_day2',
    24,
    true,
    'text',
    'Follow up message for day 1',
    '',
    '',
    '',
    '09:00',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Verify the inserts
SELECT 
    s.id as sequence_id,
    s.name as sequence_name,
    COUNT(ss.id) as step_count
FROM sequences s
LEFT JOIN sequence_steps ss ON s.id = ss.sequence_id
WHERE s.user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
GROUP BY s.id, s.name;

-- Show the steps
SELECT 
    ss.sequence_id,
    ss.day_number,
    ss.trigger,
    ss.content
FROM sequence_steps ss
WHERE ss.sequence_id IN (
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    'b7319119-4d21-43cc-97e2-9644ded0608c'
)
ORDER BY ss.sequence_id, ss.day_number;

-- Update sequence step counts
UPDATE sequences s
SET 
    step_count = (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = s.id),
    total_steps = (SELECT COUNT(*) FROM sequence_steps WHERE sequence_id = s.id)
WHERE s.id IN (
    '394d567f-e5bd-476d-ae7c-c39f74819d70',
    'b7319119-4d21-43cc-97e2-9644ded0608c'
);
