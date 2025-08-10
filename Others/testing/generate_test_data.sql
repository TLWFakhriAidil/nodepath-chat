-- Testing Script: Generate 3000 Devices and Test Data
-- This script creates fake data for testing without sending real WhatsApp messages

-- Clear existing test data (optional - comment out if you want to keep existing data)
-- DELETE FROM broadcast_messages WHERE device_name LIKE 'TestDevice%';
-- DELETE FROM leads WHERE name LIKE 'TestLead%';
-- DELETE FROM user_devices WHERE device_name LIKE 'TestDevice%';

-- Function to generate random phone numbers
CREATE OR REPLACE FUNCTION generate_phone() RETURNS TEXT AS $$
BEGIN
    RETURN '60' || LPAD(FLOOR(RANDOM() * 999999999 + 100000000)::TEXT, 9, '0');
END;
$$ LANGUAGE plpgsql;

-- Generate 3000 test devices for testing user
DO $$
DECLARE
    test_user_id UUID;
    device_count INT := 3000;
    i INT;
BEGIN
    -- Get or create test user
    SELECT id INTO test_user_id FROM users WHERE email = 'test@whatsapp.com';
    
    IF test_user_id IS NULL THEN
        INSERT INTO users (id, email, full_name, password_hash, is_active, created_at, updated_at)
        VALUES (
            gen_random_uuid(),
            'test@whatsapp.com',
            'Test User',
            'dGVzdDEyMw==', -- base64 encoded 'test123'
            true,
            NOW(),
            NOW()
        ) RETURNING id INTO test_user_id;
    END IF;
    
    -- Generate devices
    FOR i IN 1..device_count LOOP
        INSERT INTO user_devices (
            id, 
            user_id, 
            device_name, 
            phone, 
            jid,
            status, 
            min_delay_seconds,
            max_delay_seconds,
            created_at, 
            updated_at,
            last_seen
        ) VALUES (
            gen_random_uuid(),
            test_user_id,
            'TestDevice' || LPAD(i::TEXT, 4, '0'),
            generate_phone(),
            generate_phone() || ':' || FLOOR(RANDOM() * 99 + 1) || '@s.whatsapp.net',
            CASE 
                WHEN i <= 2700 THEN 'online'  -- 90% online
                ELSE 'offline'                 -- 10% offline
            END,
            5,
            15,
            NOW() - INTERVAL '30 days' * RANDOM(),
            NOW(),
            NOW() - INTERVAL '1 hour' * RANDOM()
        );
        
        -- Show progress every 100 devices
        IF i % 100 = 0 THEN
            RAISE NOTICE 'Generated % devices...', i;
        END IF;
    END LOOP;
    
    RAISE NOTICE 'Successfully generated % test devices', device_count;
END $$;

-- Generate 50,000 test leads with various triggers
DO $$
DECLARE
    test_user_id UUID;
    lead_count INT := 50000;
    triggers TEXT[] := ARRAY[
        'fitness_start', 
        'crypto_welcome', 
        'business_growth', 
        'health_tips',
        'education_start',
        'tech_updates',
        'finance_basics',
        'marketing_101'
    ];
    statuses TEXT[] := ARRAY['Active', 'Inactive', 'Pending', 'Converted'];
    i INT;
    device_ids UUID[];
BEGIN
    -- Get test user
    SELECT id INTO test_user_id FROM users WHERE email = 'test@whatsapp.com';
    
    -- Get all device IDs for random assignment
    SELECT ARRAY_AGG(id) INTO device_ids 
    FROM user_devices 
    WHERE user_id = test_user_id;
    
    -- Generate leads
    FOR i IN 1..lead_count LOOP
        INSERT INTO leads (
            id,
            user_id,
            device_id,
            name,
            phone,
            email,
            address,
            status,
            created_at,
            updated_at,
            trigger,
            source
        ) VALUES (
            gen_random_uuid(),
            test_user_id,
            device_ids[1 + FLOOR(RANDOM() * array_length(device_ids, 1))],
            'TestLead' || i,
            generate_phone(),
            'testlead' || i || '@example.com',
            'Test Address ' || i || ', Test City',
            statuses[1 + FLOOR(RANDOM() * array_length(statuses, 1))],
            NOW() - INTERVAL '90 days' * RANDOM(),
            NOW(),
            triggers[1 + FLOOR(RANDOM() * array_length(triggers, 1))],
            CASE FLOOR(RANDOM() * 4)
                WHEN 0 THEN 'Facebook'
                WHEN 1 THEN 'Instagram'
                WHEN 2 THEN 'Website'
                ELSE 'Manual'
            END
        );
        
        -- Show progress every 1000 leads
        IF i % 1000 = 0 THEN
            RAISE NOTICE 'Generated % leads...', i;
        END IF;
    END LOOP;
    
    RAISE NOTICE 'Successfully generated % test leads', lead_count;
END $$;

-- Create test campaigns
INSERT INTO campaigns (id, user_id, name, message, status, created_at, updated_at, campaign_date, time_schedule, target_status)
VALUES
    (gen_random_uuid(), (SELECT id FROM users WHERE email = 'test@whatsapp.com'), 
     'Test Campaign 1 - Active Leads', 
     'Hello {name}, this is a test campaign message for active leads!', 
     'active', NOW(), NOW(), CURRENT_DATE, '10:00-18:00', 'Active'),
    
    (gen_random_uuid(), (SELECT id FROM users WHERE email = 'test@whatsapp.com'), 
     'Test Campaign 2 - All Leads', 
     'Hi {name}, testing broadcast to all leads. Your phone: {phone}', 
     'active', NOW(), NOW(), CURRENT_DATE, '09:00-20:00', NULL),
    
    (gen_random_uuid(), (SELECT id FROM users WHERE email = 'test@whatsapp.com'), 
     'Test Campaign 3 - Scheduled', 
     'Good morning {name}, this is a scheduled test message!', 
     'scheduled', NOW(), NOW(), CURRENT_DATE + INTERVAL '1 day', '08:00-12:00', 'Pending');

-- Create test sequences with multiple steps
DO $$
DECLARE
    test_user_id UUID;
    seq_id UUID;
    step_triggers TEXT[] := ARRAY[
        'fitness_start', 'crypto_welcome', 'business_growth', 'health_tips'
    ];
    trigger_name TEXT;
BEGIN
    SELECT id INTO test_user_id FROM users WHERE email = 'test@whatsapp.com';
    
    FOREACH trigger_name IN ARRAY step_triggers LOOP
        -- Create sequence
        INSERT INTO sequences (id, user_id, name, niche, trigger, status, created_at, updated_at)
        VALUES (
            gen_random_uuid(),
            test_user_id,
            'Test Sequence - ' || trigger_name,
            SPLIT_PART(trigger_name, '_', 1),
            trigger_name,
            'active',
            NOW(),
            NOW()
        ) RETURNING id INTO seq_id;
        
        -- Create 30 steps for each sequence
        FOR i IN 1..30 LOOP
            INSERT INTO sequence_steps (
                id,
                sequence_id,
                day_number,
                content,
                trigger,
                next_trigger,
                trigger_delay_hours,
                is_entry_point,
                created_at,
                updated_at
            ) VALUES (
                gen_random_uuid(),
                seq_id,
                i,
                'Day ' || i || ' message for ' || trigger_name || '. Hello {name}, this is step ' || i || ' of your journey!',
                trigger_name || '_day' || i,
                CASE 
                    WHEN i < 30 THEN trigger_name || '_day' || (i + 1)
                    ELSE NULL
                END,
                24, -- 24 hours between messages
                CASE WHEN i = 1 THEN true ELSE false END,
                NOW(),
                NOW()
            );
        END LOOP;
    END LOOP;
    
    RAISE NOTICE 'Successfully created test sequences with steps';
END $$;

-- Create test AI campaigns
INSERT INTO ai_campaigns (
    id, user_id, campaign_name, lead_source, lead_status, min_delay, max_delay, 
    device_limit_per_device, start_date, end_date, daily_limit, status, 
    created_at, updated_at
)
VALUES
    (gen_random_uuid(), 
     (SELECT id FROM users WHERE email = 'test@whatsapp.com'),
     'Test AI Campaign 1',
     'Facebook',
     'Active',
     5, 15, 100,
     CURRENT_DATE,
     CURRENT_DATE + INTERVAL '7 days',
     1000,
     'active',
     NOW(), NOW()),
     
    (gen_random_uuid(), 
     (SELECT id FROM users WHERE email = 'test@whatsapp.com'),
     'Test AI Campaign 2',
     'Instagram',
     'Pending',
     10, 30, 50,
     CURRENT_DATE,
     CURRENT_DATE + INTERVAL '14 days',
     500,
     'active',
     NOW(), NOW());

-- Summary of generated data
SELECT 'Test Data Summary:' as info
UNION ALL
SELECT 'Users: ' || COUNT(*)::TEXT FROM users WHERE email = 'test@whatsapp.com'
UNION ALL
SELECT 'Devices: ' || COUNT(*)::TEXT FROM user_devices WHERE device_name LIKE 'TestDevice%'
UNION ALL
SELECT 'Online Devices: ' || COUNT(*)::TEXT FROM user_devices WHERE device_name LIKE 'TestDevice%' AND status = 'online'
UNION ALL
SELECT 'Leads: ' || COUNT(*)::TEXT FROM leads WHERE name LIKE 'TestLead%'
UNION ALL
SELECT 'Campaigns: ' || COUNT(*)::TEXT FROM campaigns WHERE user_id = (SELECT id FROM users WHERE email = 'test@whatsapp.com')
UNION ALL
SELECT 'Sequences: ' || COUNT(*)::TEXT FROM sequences WHERE user_id = (SELECT id FROM users WHERE email = 'test@whatsapp.com')
UNION ALL
SELECT 'AI Campaigns: ' || COUNT(*)::TEXT FROM ai_campaigns WHERE user_id = (SELECT id FROM users WHERE email = 'test@whatsapp.com');

-- Clean up function
DROP FUNCTION IF EXISTS generate_phone();
