-- Quick fix for analytics data issue
-- This script ensures proper data linking for the analytics sidebar

-- 1. First, let's see what user_id values we have in device_setting_nodepath
SELECT 'Current device_setting_nodepath user_id values:' as info;
SELECT DISTINCT user_id FROM device_setting_nodepath WHERE user_id IS NOT NULL;

-- 2. Update test devices to have user_id = 1 (assuming this is a test user)
UPDATE device_setting_nodepath 
SET user_id = 1 
WHERE id_device IN ('FakhriAidilTLW-001', 'SCHQ-S94', 'SCHQ-S12')
AND (user_id IS NULL OR user_id = 0);

-- 3. Make sure we have at least one device with user_id = 1
INSERT INTO device_setting_nodepath (
    id_device, 
    provider, 
    user_id, 
    api_key_option, 
    created_at, 
    updated_at
)
VALUES (
    'FakhriAidilTLW-001', 
    'waha', 
    1, 
    'openai/gpt-4.1', 
    NOW(), 
    NOW()
)
ON DUPLICATE KEY UPDATE 
    user_id = 1,
    updated_at = NOW();

-- 4. Create sample ai_whatsapp_nodepath data for testing
INSERT INTO ai_whatsapp_nodepath (
    id_device,
    prospect_num,
    human,
    stage,
    niche,
    date_order,
    created_at,
    updated_at
)
SELECT 
    'FakhriAidilTLW-001' as id_device,
    CONCAT('60113750', LPAD(n, 4, '0')) as prospect_num,
    FLOOR(RAND() * 2) as human,
    ELT(1 + FLOOR(RAND() * 4), 'lead', 'prospect', 'customer', 'inquiry') as stage,
    ELT(1 + FLOOR(RAND() * 3), 'ecommerce', 'services', 'retail') as niche,
    DATE_SUB(NOW(), INTERVAL FLOOR(RAND() * 30) DAY) as date_order,
    NOW() as created_at,
    NOW() as updated_at
FROM (
    SELECT @row := @row + 1 as n 
    FROM (SELECT 0 UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4) t1,
         (SELECT 0 UNION ALL SELECT 1 UNION ALL SELECT 2 UNION ALL SELECT 3 UNION ALL SELECT 4) t2,
         (SELECT @row := 0) t3
    LIMIT 20
) numbers
WHERE NOT EXISTS (
    SELECT 1 FROM ai_whatsapp_nodepath WHERE id_device = 'FakhriAidilTLW-001'
    LIMIT 1
);

-- 5. Verify the fix worked
SELECT 'Analytics data after fix:' as info;
SELECT 
    COUNT(*) as total_conversations,
    COUNT(CASE WHEN a.human = 0 THEN 1 END) as ai_active,
    COUNT(CASE WHEN a.human = 1 THEN 1 END) as human_takeover,
    COUNT(DISTINCT a.id_device) as unique_devices,
    COUNT(DISTINCT a.niche) as unique_niches,
    COUNT(DISTINCT a.stage) as unique_stages
FROM ai_whatsapp_nodepath a
JOIN device_setting_nodepath d ON a.id_device = d.id_device
WHERE d.user_id = 1;

-- 6. Show daily breakdown for current month
SELECT 'Daily breakdown for current month:' as info;
SELECT 
    DATE(a.date_order) as date,
    COUNT(*) as conversations
FROM ai_whatsapp_nodepath a
JOIN device_setting_nodepath d ON a.id_device = d.id_device
WHERE d.user_id = 1 
    AND MONTH(a.date_order) = MONTH(NOW())
    AND YEAR(a.date_order) = YEAR(NOW())
GROUP BY DATE(a.date_order)
ORDER BY date DESC
LIMIT 10;
