-- Add missing columns to leads table to match Go model
ALTER TABLE leads ADD COLUMN IF NOT EXISTS target_status VARCHAR(50) DEFAULT 'prospect';

-- Update any NULL target_status values
UPDATE leads SET target_status = 'prospect' WHERE target_status IS NULL;

-- Ensure journey column exists (maps to Notes in Go model)
ALTER TABLE leads ALTER COLUMN journey TYPE TEXT;
