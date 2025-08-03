-- Remove toolkit mode and flow prompt from chatbot_flows table
ALTER TABLE public.chatbot_flows 
DROP COLUMN IF EXISTS toolkit_mode,
DROP COLUMN IF EXISTS flow_prompt;

-- Remove the check constraint
ALTER TABLE public.chatbot_flows 
DROP CONSTRAINT IF EXISTS chatbot_flows_toolkit_mode_check;