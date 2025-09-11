-- Add flow execution fields to ai_whatsapp_nodepath table
-- This migration adds new schema columns and removes deprecated ones

-- Drop deprecated columns if they exist
ALTER TABLE ai_whatsapp_nodepath 
DROP COLUMN IF EXISTS current_node,
DROP COLUMN IF EXISTS variables;

-- Add new schema columns
ALTER TABLE ai_whatsapp_nodepath 
ADD COLUMN IF NOT EXISTS flow_reference VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Reference to chatbot flow being executed',
ADD COLUMN IF NOT EXISTS execution_status ENUM('active', 'completed', 'failed') COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Flow execution status',
ADD COLUMN IF NOT EXISTS execution_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Unique execution identifier',
ADD COLUMN IF NOT EXISTS flow_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'ID of the current chatbot flow being executed',
ADD COLUMN IF NOT EXISTS current_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Current node ID in the chatbot flow',
ADD COLUMN IF NOT EXISTS last_node_id VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL COMMENT 'Previous node ID for flow tracking',
ADD COLUMN IF NOT EXISTS waiting_for_reply TINYINT(1) DEFAULT 0 COMMENT '1 = waiting for user reply, 0 = not waiting',
ADD COLUMN IF NOT EXISTS prospect_name VARCHAR(255) COLLATE utf8mb4_unicode_ci DEFAULT NULL;

-- Add indexes for better performance
CREATE INDEX IF NOT EXISTS idx_ai_whatsapp_flow_reference ON ai_whatsapp_nodepath(flow_reference);
CREATE INDEX IF NOT EXISTS idx_ai_whatsapp_execution_status ON ai_whatsapp_nodepath(execution_status);
CREATE INDEX IF NOT EXISTS idx_ai_whatsapp_execution_id ON ai_whatsapp_nodepath(execution_id);
CREATE INDEX IF NOT EXISTS idx_flow_id ON ai_whatsapp_nodepath(flow_id);
CREATE INDEX IF NOT EXISTS idx_current_node_id ON ai_whatsapp_nodepath(current_node_id);
CREATE INDEX IF NOT EXISTS idx_prospect_num ON ai_whatsapp_nodepath(prospect_num);

-- Drop deprecated indexes if they exist
DROP INDEX IF EXISTS idx_ai_whatsapp_current_node ON ai_whatsapp_nodepath;

-- Update id_staff to id_device for consistency (rename column)
ALTER TABLE ai_whatsapp_nodepath CHANGE COLUMN id_staff id_device VARCHAR(255) NOT NULL;

-- Update the index to use the new column name
DROP INDEX idx_id_staff ON ai_whatsapp_nodepath;
CREATE INDEX idx_id_device ON ai_whatsapp_nodepath(id_device);