-- Manual Migration Script for Railway Production Database
-- Execute this script directly in Railway MySQL console or phpMyAdmin

USE admin_railway;

-- Check current table structure
DESCRIBE ai_whatsapp_nodepath;

-- Add missing columns safely
ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS jam VARCHAR(255) DEFAULT NULL COMMENT 'Jam field for AI WhatsApp conversations';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS intro VARCHAR(255) DEFAULT NULL COMMENT 'Introduction field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS date_order DATETIME DEFAULT NULL COMMENT 'Order date field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS balas VARCHAR(255) DEFAULT NULL COMMENT 'Reply field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS data_image TEXT DEFAULT NULL COMMENT 'Image data field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS conv_stage VARCHAR(100) DEFAULT NULL COMMENT 'Conversation stage field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS keywordiklan VARCHAR(255) DEFAULT NULL COMMENT 'Advertisement keyword field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS marketer VARCHAR(255) DEFAULT NULL COMMENT 'Marketer field';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS update_today TINYINT(1) DEFAULT 0 COMMENT 'Update today flag';

-- Fix data types for existing columns
ALTER TABLE ai_whatsapp_nodepath 
MODIFY COLUMN id_prospect INT DEFAULT NULL COMMENT 'Prospect ID as integer';

ALTER TABLE ai_whatsapp_nodepath 
MODIFY COLUMN bot_balas TIMESTAMP NULL DEFAULT NULL COMMENT 'Bot reply timestamp';

-- Verify the updated table structure
DESCRIBE ai_whatsapp_nodepath;

-- Show sample data to verify columns exist
SELECT * FROM ai_whatsapp_nodepath LIMIT 1;

-- Success message
SELECT 'Migration completed successfully! All columns added.' AS status;