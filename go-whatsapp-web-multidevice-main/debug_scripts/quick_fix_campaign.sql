-- Quick fix to trigger your campaign
-- Run this SQL directly in your database

-- 1. First, let's see the current state
SELECT id, title, status, campaign_date, scheduled_time, niche, target_status
FROM campaigns 
WHERE title = 'test';

-- 2. Update the campaign to pending status so it gets picked up
UPDATE campaigns 
SET status = 'pending',
    updated_at = NOW()
WHERE title = 'test' 
AND status = 'scheduled';

-- 3. Verify the update
SELECT id, title, status, updated_at 
FROM campaigns 
WHERE title = 'test';

-- 4. Check if you have any matching leads
SELECT COUNT(*) as lead_count
FROM leads
WHERE niche LIKE '%VITAC%'
AND target_status = 'customer';

-- If lead_count is 0, that's why campaign won't send!
-- You need leads with:
-- - niche containing "VITAC" 
-- - target_status = "customer"