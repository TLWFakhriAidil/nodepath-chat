-- Drop unused columns from ai_whatsapp_nodepath table
-- This migration removes columns that are no longer needed

ALTER TABLE ai_whatsapp_nodepath 
DROP COLUMN IF EXISTS jam,
DROP COLUMN IF EXISTS catatan_staff,
DROP COLUMN IF EXISTS data_image,
DROP COLUMN IF EXISTS conv_stage,
DROP COLUMN IF EXISTS variables,
DROP COLUMN IF EXISTS bot_balas,
DROP COLUMN IF EXISTS current_node;

-- Add any missing columns from the new schema if they don't exist
ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS prospect_name VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL AFTER niche,
ADD COLUMN IF NOT EXISTS flow_reference VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Reference to chatbot flow being executed' AFTER id_prospect,
ADD COLUMN IF NOT EXISTS execution_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Unique execution identifier' AFTER flow_reference,
ADD COLUMN IF NOT EXISTS execution_status ENUM('active','completed','failed') COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Flow execution status' AFTER conv_current,
ADD COLUMN IF NOT EXISTS flow_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ID of the current chatbot flow being executed' AFTER execution_status,
ADD COLUMN IF NOT EXISTS current_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Current node ID in the chatbot flow' AFTER flow_id,
ADD COLUMN IF NOT EXISTS last_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Previous node ID for flow tracking' AFTER current_node_id,
ADD COLUMN IF NOT EXISTS waiting_for_reply TINYINT(1) DEFAULT 0 COMMENT '1 = waiting for user reply, 0 = not waiting' AFTER last_node_id;

-- Add indexes if they don't exist
ALTER TABLE ai_whatsapp_nodepath 
ADD INDEX IF NOT EXISTS idx_flow_id (flow_id),
ADD INDEX IF NOT EXISTS idx_current_node_id (current_node_id),
ADD INDEX IF NOT EXISTS idx_id_device (id_device),
ADD INDEX IF NOT EXISTS idx_prospect_num (prospect_num),
ADD UNIQUE INDEX IF NOT EXISTS uniq_execution_id (execution_id);
