-- Create columns for AI prompt node data in chatbot_flows_nodepath table
-- First check if the table exists and add missing columns
ALTER TABLE chatbot_flows_nodepath 
ADD COLUMN IF NOT EXISTS system_prompt TEXT,
ADD COLUMN IF NOT EXISTS instance VARCHAR(255),
ADD COLUMN IF NOT EXISTS apiprovider VARCHAR(255);

-- Ensure conv_last and conv_current columns exist in chatbot_executions_nodepath
ALTER TABLE chatbot_executions_nodepath 
ADD COLUMN IF NOT EXISTS conv_last JSON,
ADD COLUMN IF NOT EXISTS conv_current TEXT;