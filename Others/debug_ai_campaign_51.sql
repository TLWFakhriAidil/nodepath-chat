-- Debug AI Campaign ID 51
-- Check the exact values
SELECT 
    id,
    user_id,
    title,
    niche,
    target_status,
    ai,
    device_limit as limit,
    status
FROM campaigns
WHERE id = 51;

-- Check if there are matching leads
SELECT COUNT(*) as matching_pending_leads
FROM leads_ai
WHERE user_id = (SELECT user_id FROM campaigns WHERE id = 51)
  AND niche = (SELECT niche FROM campaigns WHERE id = 51)
  AND status = 'pending';

-- Show the exact comparison
SELECT 
    'Campaign User' as source,
    (SELECT user_id FROM campaigns WHERE id = 51) as user_id,
    (SELECT niche FROM campaigns WHERE id = 51) as niche,
    (SELECT target_status FROM campaigns WHERE id = 51) as target_status
UNION ALL
SELECT 
    'AI Leads User' as source,
    user_id,
    niche,
    target_status
FROM leads_ai
WHERE status = 'pending'
LIMIT 5;