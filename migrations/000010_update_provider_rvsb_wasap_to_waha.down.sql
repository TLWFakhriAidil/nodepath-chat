-- Rollback migration: Update provider from 'waha' back to 'rvsb_wasap' in device_setting_nodepath table
-- This migration reverts WAHA provider back to RVSB WASAP provider

-- Update existing records that have 'waha' provider back to 'rvsb_wasap'
UPDATE device_setting_nodepath 
SET provider = 'rvsb_wasap' 
WHERE provider = 'waha';

-- Revert the comment to reflect the original provider options
ALTER TABLE device_setting_nodepath 
MODIFY COLUMN provider VARCHAR(255) NOT NULL COMMENT 'whacenter, wablas, rvsb_wasap';