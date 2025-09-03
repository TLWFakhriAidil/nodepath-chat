-- Update balas column from INT to VARCHAR(255) for timestamp tracking
-- This migration changes the balas field to store timestamp strings

ALTER TABLE ai_whatsapp_nodepath 
MODIFY COLUMN balas VARCHAR(255) DEFAULT NULL 
COMMENT 'Timestamp for last response - used for throttling (format: YYYY-MM-DD HH:MM:SS)';

-- Optional: If you want to preserve existing INT values as timestamps
-- UPDATE ai_whatsapp_nodepath 
-- SET balas = FROM_UNIXTIME(balas) 
-- WHERE balas IS NOT NULL AND balas > 0;