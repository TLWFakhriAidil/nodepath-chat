-- Fix analytics data issue by ensuring proper linking between tables
-- This script will:
-- 1. Update device_setting_nodepath to have proper user_id values
-- 2. Insert test data into ai_whatsapp_nodepath if needed
-- 3. Ensure proper linking between tables

-- First, let's check current state
SELECT 'Current Device Settings:' as info;
SELECT id_device, user_id FROM device_setting_nodepath;

-- Update device_setting_nodepath to assign user_id = 1 to test devices
-- This assumes you have a test user with ID = 1
UPDATE device_setting_nodepath 
SET user_id = 1 
WHERE id_device IN ('FakhriAidilTLW-001', 'SCHQ-S94', 'SCHQ-S12')
AND (user_id IS NULL OR user_id = 0);

-- Insert sample data into ai_whatsapp_nodepath if it doesn't exist
-- This creates test conversations for analytics
INSERT INTO ai_whatsapp_nodepath (
    id_device, 
    prospect_num, 
    human, 
    stage, 
    date_order, 
    niche,
    created_at,
    updated_at
) 
SELECT 
    'FakhriAidilTLW-001',
    CONCAT('60113750', LPAD(FLOOR(RAND() * 10000), 4, '0')),
    FLOOR(RAND() * 2), -- Random 0 or 1 for human flag
    CASE FLOOR(RAND() * 4) 
        WHEN 0 THEN 'lead'
        WHEN 1 THEN 'prospect'
        WHEN 2 THEN 'customer'
        ELSE 'inquiry'
    END,
    DATE_SUB(NOW(), INTERVAL FLOOR(RAND() * 30) DAY), -- Random date in last 30 days
    CASE FLOOR(RAND() * 3)
        WHEN 0 THEN 'ecommerce'
        WHEN 1 THEN 'services'
        ELSE 'retail'
    END,
    NOW(),
    NOW()
FROM 
    (SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5) t1,
    (SELECT 1 UNION SELECT 2 UNION SELECT 3 UNION SELECT 4 UNION SELECT 5) t2
WHERE NOT EXISTS (
    SELECT 1 FROM ai_whatsapp_nodepath WHERE id_device = 'FakhriAidilTLW-001'
)
LIMIT 10;

-- Verify the fix
SELECT 'After Fix - Analytics Data for user_id=1:' as info;
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

-- Show daily breakdown for current month
SELECT 'Daily Breakdown (Current Month):' as info;
SELECT 
    DATE(a.date_order) as date,
    COUNT(*) as conversations
FROM ai_whatsapp_nodepath a
JOIN device_setting_nodepath d ON a.id_device = d.id_device
WHERE d.user_id = 1 
AND a.date_order >= DATE_FORMAT(NOW(), '%Y-%m-01')
GROUP BY DATE(a.date_order)
ORDER BY DATE(a.date_order);
