-- Check all WhatsApp-related tables to find any stored sessions
-- This will help us identify where the old session data is stored

-- Check if there are any whatsmeow tables (WhatsApp session storage)
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND table_name LIKE '%whatsmeow%'
ORDER BY table_name;

-- Also check for any device-related tables
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
AND (table_name LIKE '%device%' OR table_name LIKE '%whatsapp%')
ORDER BY table_name;

-- If whatsmeow_device table exists, check its contents
-- This is where WhatsApp sessions are stored
SELECT * FROM whatsmeow_device;
