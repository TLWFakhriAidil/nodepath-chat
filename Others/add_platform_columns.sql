-- Add platform column to user_devices table
ALTER TABLE user_devices 
ADD COLUMN IF NOT EXISTS platform VARCHAR(255);

-- Add platform column to leads table  
ALTER TABLE leads
ADD COLUMN IF NOT EXISTS platform VARCHAR(255);

-- Update existing records with default value if needed
UPDATE user_devices SET platform = 'WhatsApp' WHERE platform IS NULL;
UPDATE leads SET platform = 'WhatsApp' WHERE platform IS NULL;
