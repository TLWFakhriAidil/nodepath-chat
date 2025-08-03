-- Add toolkit mode and flow prompt to chatbot_flows table
ALTER TABLE public.chatbot_flows 
ADD COLUMN toolkit_mode text NOT NULL DEFAULT 'manual',
ADD COLUMN flow_prompt text DEFAULT NULL;

-- Add check constraint for toolkit_mode values
ALTER TABLE public.chatbot_flows 
ADD CONSTRAINT chatbot_flows_toolkit_mode_check 
CHECK (toolkit_mode IN ('manual', 'prompt'));

-- Add expected_input and response_output to nodes data structure
-- Note: These will be stored in the existing nodes jsonb column as part of node data