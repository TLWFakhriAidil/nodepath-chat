-- Check analytics data in the database
-- Run this SQL to understand the data structure

-- 1. Check device_setting_nodepath table
SELECT 'Device Settings:' as info;
SELECT id_device, provider, user_id, instance 
FROM device_setting_nodepath 
LIMIT 10;

-- 2. Check ai_whatsapp_nodepath table  
SELECT 'AI WhatsApp Conversations:' as info;
SELECT id_device, prospect_num, human, stage, date_order 
FROM ai_whatsapp_nodepath 
LIMIT 10;

-- 3. Check for linked data
SELECT 'Linked Data Count:' as info;
SELECT COUNT(*) as total_linked_conversations
FROM ai_whatsapp_nodepath a
JOIN device_setting_nodepath d ON a.id_device = d.id_device
WHERE d.user_id IS NOT NULL;

-- 4. Check devices by user
SELECT 'Devices by User:' as info;
SELECT user_id, COUNT(*) as device_count
FROM device_setting_nodepath
WHERE user_id IS NOT NULL
GROUP BY user_id;

-- 5. Sample analytics for user_id = 1
SELECT 'Analytics for user_id=1 (all time):' as info;
SELECT 
    COUNT(*) as total_conversations,
    COUNT(CASE WHEN a.human = 0 THEN 1 END) as ai_active,
    COUNT(CASE WHEN a.human = 1 THEN 1 END) as human_takeover
FROM ai_whatsapp_nodepath a
JOIN device_setting_nodepath d ON a.id_device = d.id_device
WHERE d.user_id = 1;

-- 6. Check if id_device values match between tables
SELECT 'Device ID Matching:' as info;
SELECT 
    (SELECT COUNT(DISTINCT id_device) FROM device_setting_nodepath) as devices_in_settings,
    (SELECT COUNT(DISTINCT id_device) FROM ai_whatsapp_nodepath) as devices_in_ai_whatsapp,
    (SELECT COUNT(DISTINCT a.id_device) 
     FROM ai_whatsapp_nodepath a 
     JOIN device_setting_nodepath d ON a.id_device = d.id_device) as matching_devices;

-- 7. Sample unmatched devices
SELECT 'Unmatched Devices in ai_whatsapp_nodepath:' as info;
SELECT DISTINCT a.id_device
FROM ai_whatsapp_nodepath a
LEFT JOIN device_setting_nodepath d ON a.id_device = d.id_device
WHERE d.id_device IS NULL
LIMIT 10;
