-- Migration to update API key options to use OpenRouter model names
-- This script updates the enum values and existing data

-- First, add new enum values to the existing enum
ALTER TABLE device_setting_nodepath 
MODIFY COLUMN api_key_option ENUM(
    'chat_gpt_so', 'chat_gpt_5_mini', 'chat_gpt_4o', 'chat_gpt_4_1_new', 'gemini_pro_25', 'gemini_pro_15',
    'openai/gpt-5-chat', 'openai/gpt-5-mini', 'openai/chatgpt-4o-latest', 'openai/gpt-4.1', 'google/gemini-2.5-pro', 'google/gemini-pro-1.5'
) DEFAULT 'openai/gpt-4.1';

-- Update existing records to use new values
UPDATE device_setting_nodepath SET api_key_option = 'openai/gpt-5-chat' WHERE api_key_option = 'chat_gpt_so';
UPDATE device_setting_nodepath SET api_key_option = 'openai/gpt-5-mini' WHERE api_key_option = 'chat_gpt_5_mini';
UPDATE device_setting_nodepath SET api_key_option = 'openai/chatgpt-4o-latest' WHERE api_key_option = 'chat_gpt_4o';
UPDATE device_setting_nodepath SET api_key_option = 'openai/gpt-4.1' WHERE api_key_option = 'chat_gpt_4_1_new';
UPDATE device_setting_nodepath SET api_key_option = 'google/gemini-2.5-pro' WHERE api_key_option = 'gemini_pro_25';
UPDATE device_setting_nodepath SET api_key_option = 'google/gemini-pro-1.5' WHERE api_key_option = 'gemini_pro_15';

-- Remove old enum values (this will recreate the enum with only new values)
ALTER TABLE device_setting_nodepath 
MODIFY COLUMN api_key_option ENUM(
    'openai/gpt-5-chat', 'openai/gpt-5-mini', 'openai/chatgpt-4o-latest', 'openai/gpt-4.1', 'google/gemini-2.5-pro', 'google/gemini-pro-1.5'
) DEFAULT 'openai/gpt-4.1';