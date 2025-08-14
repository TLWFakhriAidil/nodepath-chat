-- Revert Database Changes Script
-- This script removes the migrations that were causing crashes
-- Run this if you want to completely remove the AI Campaign and other recent features

-- WARNING: This will DELETE data! Make sure to backup first!

-- 1. Remove AI Campaign tables and columns
DROP TABLE IF EXISTS ai_campaign_progress CASCADE;
DROP TABLE IF EXISTS leads_ai CASCADE;

-- Remove AI columns from campaigns table
ALTER TABLE campaigns DROP COLUMN IF EXISTS ai;
ALTER TABLE campaigns DROP COLUMN IF EXISTS "limit";

-- 2. Remove sequence progress tracking columns
ALTER TABLE sequences 
DROP COLUMN IF EXISTS total_contacts,
DROP COLUMN IF EXISTS active_contacts,
DROP COLUMN IF EXISTS completed_contacts,
DROP COLUMN IF EXISTS failed_contacts,
DROP COLUMN IF EXISTS progress_percentage,
DROP COLUMN IF EXISTS last_activity_at,
DROP COLUMN IF EXISTS estimated_completion_at;

-- 3. Drop indexes that were created
DROP INDEX IF EXISTS idx_sequences_progress;
DROP INDEX IF EXISTS idx_sequences_last_activity;

-- 4. Remove the whatsapp_messages table if it's causing issues
-- (Only if you don't need WhatsApp Web functionality)
-- DROP TABLE IF EXISTS whatsapp_messages CASCADE;

-- 5. Check what remains
SELECT 'Remaining tables:' as status;
SELECT table_name 
FROM information_schema.tables 
WHERE table_schema = 'public' 
ORDER BY table_name;

SELECT 'Campaigns table columns:' as status;
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'campaigns' 
ORDER BY ordinal_position;

SELECT 'Sequences table columns:' as status;
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'sequences' 
ORDER BY ordinal_position;
