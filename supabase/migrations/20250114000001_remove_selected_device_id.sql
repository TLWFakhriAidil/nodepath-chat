-- Remove selected_device_id column from chatbot_flows_nodepath table
-- Keep only the niche field as requested

ALTER TABLE chatbot_flows_nodepath 
DROP COLUMN IF EXISTS selected_device_id;