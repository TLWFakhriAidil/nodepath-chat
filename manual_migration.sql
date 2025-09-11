-- Manual Migration Script for Railway Production Database
-- Execute this script directly in Railway MySQL console or phpMyAdmin

USE admin_railway;

-- Check current table structure
DESCRIBE ai_whatsapp_nodepath;

-- Add missing columns safely (only new schema columns)
ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS prospect_name VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Prospect name field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS intro VARCHAR(255) DEFAULT NULL COMMENT 'Introduction field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS date_order DATETIME DEFAULT NULL COMMENT 'Order date field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS balas VARCHAR(255) DEFAULT NULL COMMENT 'Reply field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS keywordiklan VARCHAR(255) DEFAULT NULL COMMENT 'Advertisement keyword field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS marketer VARCHAR(255) DEFAULT NULL COMMENT 'Marketer field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS update_today DATETIME DEFAULT NULL COMMENT 'Update today timestamp';

-- Add new schema columns
ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS flow_reference VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Reference to chatbot flow being executed';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS execution_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Unique execution identifier';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS execution_status ENUM('active','completed','failed') COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Flow execution status';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS flow_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ID of the current chatbot flow being executed';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS current_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Current node ID in the chatbot flow';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS last_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Previous node ID for flow tracking';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS waiting_for_reply TINYINT(1) DEFAULT 0 COMMENT '1 = waiting for user reply, 0 = not waiting';

-- Fix data types for existing columns
ALTER TABLE ai_whatsapp_nodepath 
MODIFY COLUMN id_prospect INT DEFAULT NULL COMMENT 'Prospect ID as integer';

-- Remove deprecated columns if they exist
ALTER TABLE ai_whatsapp_nodepath 
DROP COLUMN IF EXISTS jam,
DROP COLUMN IF EXISTS conv_stage,
DROP COLUMN IF EXISTS variables,
DROP COLUMN IF EXISTS catatan_staff,
DROP COLUMN IF EXISTS data_image,
DROP COLUMN IF EXISTS current_node,
DROP COLUMN IF EXISTS bot_balas;

-- Verify the updated table structure
DESCRIBE ai_whatsapp_nodepath;

-- Show sample data to verify columns exist
SELECT * FROM ai_whatsapp_nodepath LIMIT 1;

-- Success message
SELECT 'Migration completed successfully! All columns added.' AS status;