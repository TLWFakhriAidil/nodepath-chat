-- Update balas field from INT to VARCHAR for timestamp tracking
-- This matches the PHP implementation for time throttling

ALTER TABLE ai_whatsapp_nodepath 
MODIFY COLUMN balas VARCHAR(255) DEFAULT NULL 
COMMENT 'Timestamp for last response - used for throttling';

-- Add index for better performance
CREATE INDEX IF NOT EXISTS idx_balas ON ai_whatsapp_nodepath(balas);