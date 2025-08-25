-- Rollback: Remove flow tracking fields from ai_whatsapp_nodepath table
-- This migration reverts the flow tracking field additions

-- Drop the indexes first
DROP INDEX IF EXISTS idx_current_node_id ON ai_whatsapp_nodepath;
DROP INDEX IF EXISTS idx_waiting_for_reply ON ai_whatsapp_nodepath;
DROP INDEX IF EXISTS idx_flow_id ON ai_whatsapp_nodepath;

-- Remove the columns
ALTER TABLE ai_whatsapp_nodepath 
DROP COLUMN IF EXISTS current_node_id,
DROP COLUMN IF EXISTS waiting_for_reply,
DROP COLUMN IF EXISTS flow_id,
DROP COLUMN IF EXISTS last_node_id;

-- Verify the rollback
SELECT 'Successfully removed flow tracking fields from ai_whatsapp_nodepath table' AS status;