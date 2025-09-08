-- Insert test data into ai_whatsapp_nodepath for testing analytics
-- This will create sample conversations for the device FakhriAidilTLW-001

-- First, make sure the device exists in device_setting_nodepath with a user_id
UPDATE device_setting_nodepath 
SET user_id = 1 
WHERE id_device = 'FakhriAidilTLW-001';

-- If device doesn't exist, create it
INSERT INTO device_setting_nodepath (id_device, provider, user_id, api_key_option, created_at, updated_at)
SELECT 'FakhriAidilTLW-001', 'waha', 1, 'openai/gpt-4.1', NOW(), NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM device_setting_nodepath WHERE id_device = 'FakhriAidilTLW-001'
);

-- Now insert sample conversations for testing
INSERT INTO ai_whatsapp_nodepath (
    id_device,
    prospect_num,
    human,
    stage,
    niche,
    date_order,
    created_at,
    updated_at
) VALUES
    ('FakhriAidilTLW-001', '601137508001', 0, 'lead', 'ecommerce', DATE_SUB(NOW(), INTERVAL 1 DAY), NOW(), NOW()),
    ('FakhriAidilTLW-001', '601137508002', 0, 'prospect', 'services', DATE_SUB(NOW(), INTERVAL 2 DAY), NOW(), NOW()),
    ('FakhriAidilTLW-001', '601137508003', 1, 'customer', 'retail', DATE_SUB(NOW(), INTERVAL 3 DAY), NOW(), NOW()),
    ('FakhriAidilTLW-001', '601137508004', 0, 'inquiry', 'technology', DATE_SUB(NOW(), INTERVAL 4 DAY), NOW(), NOW()),
    ('FakhriAidilTLW-001', '601137508005', 0, 'lead', 'ecommerce', DATE_SUB(NOW(), INTERVAL 5 DAY), NOW(), NOW()),
    ('FakhriAidilTLW-001', '601137508006', 1, 'prospect', 'services', DATE_SUB(NOW(), INTERVAL 6 DAY), NOW(), NOW()),
    ('FakhriAidilTLW-001', '601137508007', 0, 'customer', 'retail', DATE_SUB(NOW(), INTERVAL 7 DAY), NOW(), NOW()),
    ('FakhriAidilTLW-001', '601137508008', 0, 'inquiry', 'technology', DATE_SUB(NOW(), INTERVAL 8 DAY), NOW(), NOW()),
    ('FakhriAidilTLW-001', '601137508009', 1, 'lead', 'ecommerce', DATE_SUB(NOW(), INTERVAL 9 DAY), NOW(), NOW()),
    ('FakhriAidilTLW-001', '601137508010', 0, 'prospect', 'services', DATE_SUB(NOW(), INTERVAL 10 DAY), NOW(), NOW());

-- Verify the data was inserted
SELECT 
    COUNT(*) as total_conversations,
    COUNT(CASE WHEN human = 0 THEN 1 END) as ai_active,
    COUNT(CASE WHEN human = 1 THEN 1 END) as human_takeover,
    COUNT(DISTINCT id_device) as unique_devices,
    COUNT(DISTINCT niche) as unique_niches,
    COUNT(DISTINCT stage) as unique_stages
FROM ai_whatsapp_nodepath
WHERE id_device = 'FakhriAidilTLW-001';
