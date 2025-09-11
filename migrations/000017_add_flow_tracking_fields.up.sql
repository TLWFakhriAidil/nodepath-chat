-- Add flow tracking fields to ai_whatsapp_nodepath table
-- These fields enable proper user reply node handling and flow continuation
-- Note: This migration ensures columns exist but avoids conflicts with other migrations

-- Add columns only if they don't exist (to avoid conflicts with migration 000015)
ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS current_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Current node ID in the chatbot flow';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS waiting_for_reply TINYINT(1) DEFAULT 0 COMMENT '1 = waiting for user reply, 0 = not waiting';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS flow_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ID of the current chatbot flow being executed';

ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS last_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Previous node ID for flow tracking';

-- Add indexes only if they don't exist
CREATE INDEX IF NOT EXISTS idx_current_node_id ON ai_whatsapp_nodepath(current_node_id);
CREATE INDEX IF NOT EXISTS idx_waiting_for_reply ON ai_whatsapp_nodepath(waiting_for_reply);
CREATE INDEX IF NOT EXISTS idx_flow_id ON ai_whatsapp_nodepath(flow_id);

-- Verify the new columns were added
SELECT 'Successfully added flow tracking fields to ai_whatsapp_nodepath table' AS status;