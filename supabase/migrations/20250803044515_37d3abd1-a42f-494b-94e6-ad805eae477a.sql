-- Add new columns to ai_settings_nodepath table for OpenRouter integration
ALTER TABLE public.ai_settings_nodepath 
ADD COLUMN IF NOT EXISTS open_model TEXT DEFAULT 'openai/gpt-4.1',
ADD COLUMN IF NOT EXISTS open_router_key TEXT;