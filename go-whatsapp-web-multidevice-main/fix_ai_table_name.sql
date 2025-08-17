-- Fix AI Campaign table name issue
-- Rename lead_ai to leads_ai to match the code

ALTER TABLE lead_ai RENAME TO leads_ai;

-- Verify the change
SELECT COUNT(*) as total_leads FROM leads_ai WHERE status = 'pending';