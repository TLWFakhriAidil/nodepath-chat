-- SQL script to drop deprecated database tables
-- This script removes tables that are no longer used in the system
-- Execute this script on the production database to clean up deprecated tables

-- Drop chatbot_executions_nodepath table (functionality consolidated into ai_whatsapp_nodepath)
DROP TABLE IF EXISTS chatbot_executions_nodepath;

-- Drop conversation_log_nodepath_backup table (backup table no longer needed)
DROP TABLE IF EXISTS conversation_log_nodepath_backup;

-- Drop device_settings_nodepath table (if it exists, though no references found)
DROP TABLE IF EXISTS device_settings_nodepath;

-- Verify remaining tables
SELECT TABLE_NAME 
FROM INFORMATION_SCHEMA.TABLES 
WHERE TABLE_SCHEMA = DATABASE() 
AND TABLE_NAME LIKE '%_nodepath'
ORDER BY TABLE_NAME;

-- Expected remaining tables:
-- ai_whatsapp_nodepath
-- chatbot_flows_nodepath
-- conversation_log_nodepath
-- device_setting_nodepath