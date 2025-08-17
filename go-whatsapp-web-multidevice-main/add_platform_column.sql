-- Add platform column to user_devices table if it doesn't exist
ALTER TABLE user_devices 
ADD COLUMN IF NOT EXISTS platform VARCHAR(50);

-- Create index on platform column for better query performance
CREATE INDEX IF NOT EXISTS idx_user_devices_platform 
ON user_devices(platform) 
WHERE platform IS NOT NULL AND platform != '';

-- Update the getDeviceWorkloads query to exclude devices with platform
-- This is used in sequence_trigger_processor.go
COMMENT ON COLUMN user_devices.platform IS 'Platform identifier - if set, device is skipped from auto status checks';
