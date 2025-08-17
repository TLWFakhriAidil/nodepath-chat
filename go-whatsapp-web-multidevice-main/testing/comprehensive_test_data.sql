-- Comprehensive Test Data for WhatsApp Multi-Device System
-- Testing: Campaigns, AI Campaigns, and 7-day Sequences with 3000 devices

-- First, let's check what we have
SELECT 'Current Database Status:' as info;
SELECT 'Users:', COUNT(*) FROM users;
SELECT 'Devices:', COUNT(*) FROM user_devices;
SELECT 'Leads:', COUNT(*) FROM leads;
SELECT 'Campaigns:', COUNT(*) FROM campaigns;
SELECT 'Sequences:', COUNT(*) FROM sequences;
SELECT 'AI Campaigns:', COUNT(*) FROM ai_campaigns;

-- Clean up any existing test data
DELETE FROM broadcast_messages WHERE device_id IN (SELECT id FROM user_devices WHERE device_name LIKE 'TestDevice%');
DELETE FROM sequence_contacts WHERE sequence_id IN (SELECT id FROM sequences WHERE name LIKE 'Test%');
DELETE FROM sequence_steps WHERE sequence_id IN (SELECT id FROM sequences WHERE name LIKE 'Test%');
DELETE FROM ai_campaign_leads WHERE campaign_id IN (SELECT id FROM ai_campaigns WHERE campaign_name LIKE 'Test%');
DELETE FROM sequences WHERE name LIKE 'Test%';
DELETE FROM ai_campaigns WHERE campaign_name LIKE 'Test%';
DELETE FROM campaigns WHERE name LIKE 'Test%';
DELETE FROM leads WHERE name LIKE 'TestLead%';
DELETE FROM user_devices WHERE device_name LIKE 'TestDevice%';
DELETE FROM users WHERE email = 'test@whatsapp.com';

-- Create test user
INSERT INTO users (id, email, full_name, password_hash, is_active, created_at, updated_at)
VALUES (
    'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
    'test@whatsapp.com',
    'Test User',
    'dGVzdDEyMw==',
    true,
    NOW(),
    NOW()
);

-- Create exactly 3000 devices (90% online, 10% offline)
DO $$
DECLARE
    i INT;
    device_id UUID;
BEGIN
    FOR i IN 1..3000 LOOP
        device_id := gen_random_uuid();
        
        INSERT INTO user_devices (
            id, user_id, device_name, phone, jid, status, 
            min_delay_seconds, max_delay_seconds,
            created_at, updated_at, last_seen
        ) VALUES (
            device_id,
            'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
            'TestDevice' || LPAD(i::TEXT, 4, '0'),
            '60' || LPAD(FLOOR(RANDOM() * 999999999 + 100000000)::TEXT, 9, '0'),
            '60' || LPAD(FLOOR(RANDOM() * 999999999 + 100000000)::TEXT, 9, '0') || ':' || FLOOR(RANDOM() * 99 + 1) || '@s.whatsapp.net',
            CASE WHEN i <= 2700 THEN 'online' ELSE 'offline' END,
            5, 15,
            NOW() - INTERVAL '30 days' * RANDOM(),
            NOW(),
            NOW() - INTERVAL '1 hour' * RANDOM()
        );
        
        IF i % 100 = 0 THEN
            RAISE NOTICE 'Created % devices...', i;
        END IF;
    END LOOP;
END $$;

-- Create 100,000 leads distributed across devices
DO $$
DECLARE
    i INT;
    device_ids UUID[];
    triggers TEXT[] := ARRAY[
        'fitness_week1', 'crypto_intro', 'business_growth', 
        'health_journey', 'education_start', 'tech_basics'
    ];
    statuses TEXT[] := ARRAY['Active', 'Inactive', 'Pending'];
BEGIN
    -- Get all device IDs
    SELECT ARRAY_AGG(id) INTO device_ids 
    FROM user_devices 
    WHERE user_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';
    
    FOR i IN 1..100000 LOOP
        INSERT INTO leads (
            id, user_id, device_id, name, phone, email,
            status, trigger, source, created_at, updated_at
        ) VALUES (
            gen_random_uuid(),
            'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
            device_ids[1 + FLOOR(RANDOM() * array_length(device_ids, 1))],
            'TestLead' || i,
            '60' || LPAD(FLOOR(RANDOM() * 999999999 + 100000000)::TEXT, 9, '0'),
            'testlead' || i || '@example.com',
            statuses[1 + FLOOR(RANDOM() * array_length(statuses, 1))],
            triggers[1 + FLOOR(RANDOM() * array_length(triggers, 1))],
            CASE FLOOR(RANDOM() * 3)
                WHEN 0 THEN 'Facebook'
                WHEN 1 THEN 'Instagram'
                ELSE 'Website'
            END,
            NOW() - INTERVAL '60 days' * RANDOM(),
            NOW()
        );
        
        IF i % 10000 = 0 THEN
            RAISE NOTICE 'Created % leads...', i;
        END IF;
    END LOOP;
END $$;

-- Create test campaigns
INSERT INTO campaigns (id, user_id, name, message, status, created_at, updated_at, campaign_date, time_schedule, target_status)
VALUES
    -- Active campaign targeting Active leads
    ('b0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 
     'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
     'Test Campaign 1 - Active Leads Only',
     'Hello {name}, this is campaign 1 for active leads. Your status is {status}.',
     'active', NOW(), NOW(), CURRENT_DATE, '09:00-18:00', 'Active'),
    
    -- Campaign for all leads
    ('c0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 
     'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
     'Test Campaign 2 - All Leads',
     'Hi {name}, this is a broadcast to all leads regardless of status.',
     'active', NOW(), NOW(), CURRENT_DATE, '10:00-20:00', NULL),
    
    -- Scheduled campaign
    ('d0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11', 
     'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
     'Test Campaign 3 - Scheduled',
     'Good morning {name}, this is a scheduled campaign message.',
     'scheduled', NOW(), NOW(), CURRENT_DATE + INTERVAL '1 day', '08:00-12:00', 'Pending');

-- Create 7-day sequences for different triggers
DO $$
DECLARE
    seq_id UUID;
    trigger_name TEXT;
    triggers TEXT[] := ARRAY['fitness_week1', 'crypto_intro', 'business_growth'];
    i INT;
    j INT;
BEGIN
    FOR i IN 1..array_length(triggers, 1) LOOP
        trigger_name := triggers[i];
        seq_id := gen_random_uuid();
        
        -- Create sequence
        INSERT INTO sequences (id, user_id, name, niche, trigger, status, created_at, updated_at)
        VALUES (
            seq_id,
            'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
            'Test 7-Day Sequence - ' || trigger_name,
            SPLIT_PART(trigger_name, '_', 1),
            trigger_name,
            'active',
            NOW(),
            NOW()
        );
        
        -- Create 7 days of messages
        FOR j IN 1..7 LOOP
            INSERT INTO sequence_steps (
                id, sequence_id, day_number, content,
                trigger, next_trigger, trigger_delay_hours,
                is_entry_point, created_at, updated_at
            ) VALUES (
                gen_random_uuid(),
                seq_id,
                j,
                'Day ' || j || ' of ' || trigger_name || '. Hello {name}, ' ||
                CASE j
                    WHEN 1 THEN 'Welcome to your journey!'
                    WHEN 2 THEN 'Hope you are doing great today.'
                    WHEN 3 THEN 'Midweek check-in - keep going!'
                    WHEN 4 THEN 'You are making great progress.'
                    WHEN 5 THEN 'Almost at the end of week 1!'
                    WHEN 6 THEN 'Weekend vibes - stay motivated.'
                    WHEN 7 THEN 'Congratulations on completing week 1!'
                END,
                trigger_name || '_day' || j,
                CASE WHEN j < 7 THEN trigger_name || '_day' || (j + 1) ELSE NULL END,
                24, -- 24 hours between messages
                CASE WHEN j = 1 THEN true ELSE false END,
                NOW(),
                NOW()
            );
        END LOOP;
    END LOOP;
    
    RAISE NOTICE 'Created 3 sequences with 7 days each';
END $$;

-- Create AI campaigns
INSERT INTO ai_campaigns (
    id, user_id, campaign_name, lead_source, lead_status,
    min_delay, max_delay, device_limit_per_device,
    start_date, end_date, daily_limit, status,
    created_at, updated_at
)
VALUES
    -- Facebook leads AI campaign
    ('e0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
     'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
     'Test AI Campaign - Facebook Leads',
     'Facebook', 'Active',
     5, 15, 80, -- 80 messages per device per hour
     CURRENT_DATE, CURRENT_DATE + INTERVAL '7 days',
     10000, -- 10k messages per day
     'active', NOW(), NOW()),
     
    -- Instagram leads AI campaign
    ('f0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
     'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11',
     'Test AI Campaign - Instagram Leads',
     'Instagram', 'Pending',
     10, 30, 50,
     CURRENT_DATE, CURRENT_DATE + INTERVAL '14 days',
     5000, -- 5k messages per day
     'active', NOW(), NOW());

-- Summary of test data created
SELECT '===========================================' as separator;
SELECT 'Test Data Creation Complete!' as status;
SELECT '===========================================' as separator;

SELECT 'Devices Summary:' as info
UNION ALL
SELECT '  Total: ' || COUNT(*)::TEXT || ' devices' FROM user_devices WHERE device_name LIKE 'TestDevice%'
UNION ALL
SELECT '  Online: ' || COUNT(*)::TEXT || ' (90%)' FROM user_devices WHERE device_name LIKE 'TestDevice%' AND status = 'online'
UNION ALL
SELECT '  Offline: ' || COUNT(*)::TEXT || ' (10%)' FROM user_devices WHERE device_name LIKE 'TestDevice%' AND status = 'offline';

SELECT '' as blank;

SELECT 'Leads Summary:' as info
UNION ALL
SELECT '  Total: ' || COUNT(*)::TEXT || ' leads' FROM leads WHERE name LIKE 'TestLead%'
UNION ALL
SELECT '  By Status:' as info
UNION ALL
SELECT '    Active: ' || COUNT(*)::TEXT FROM leads WHERE name LIKE 'TestLead%' AND status = 'Active'
UNION ALL
SELECT '    Pending: ' || COUNT(*)::TEXT FROM leads WHERE name LIKE 'TestLead%' AND status = 'Pending'
UNION ALL
SELECT '    Inactive: ' || COUNT(*)::TEXT FROM leads WHERE name LIKE 'TestLead%' AND status = 'Inactive';

SELECT '' as blank;

SELECT 'Campaigns:' as info
UNION ALL
SELECT '  ' || name || ' (Status: ' || status || ', Target: ' || COALESCE(target_status, 'All') || ')' 
FROM campaigns 
WHERE user_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';

SELECT '' as blank;

SELECT 'Sequences (7-day):' as info
UNION ALL
SELECT '  ' || s.name || ' (Trigger: ' || s.trigger || ', Steps: ' || COUNT(ss.id) || ')'
FROM sequences s
LEFT JOIN sequence_steps ss ON ss.sequence_id = s.id
WHERE s.user_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11'
GROUP BY s.id, s.name, s.trigger;

SELECT '' as blank;

SELECT 'AI Campaigns:' as info
UNION ALL
SELECT '  ' || campaign_name || ' (Source: ' || lead_source || ', Daily Limit: ' || daily_limit || ')'
FROM ai_campaigns
WHERE user_id = 'a0eebc99-9c0b-4ef8-bb6d-6bb9bd380a11';

SELECT '' as blank;
SELECT 'Ready to test with 3000 devices and 100k leads!' as status;
