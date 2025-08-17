-- Check AI Campaign ID 51 details
SELECT id, user_id, title, niche, target_status, ai, device_limit, status
FROM campaigns  
WHERE id = 51;

-- Check if campaign user matches lead user
SELECT 
    c.id as campaign_id,
    c.user_id as campaign_user,
    c.niche as campaign_niche,
    c.target_status as campaign_target_status,
    COUNT(la.id) as matching_leads
FROM campaigns c
LEFT JOIN lead_ai la ON la.user_id = c.user_id 
    AND la.niche = c.niche 
    AND la.status = 'pending'
WHERE c.id = 51
GROUP BY c.id, c.user_id, c.niche, c.target_status;

-- Show what niches are available for pending AI leads
SELECT DISTINCT user_id, niche, COUNT(*) as count
FROM lead_ai
WHERE status = 'pending'
GROUP BY user_id, niche;

-- Debug: Show campaign niche vs lead niche
SELECT 
    'Campaign' as type,
    (SELECT niche FROM campaigns WHERE id = 51) as niche,
    (SELECT user_id FROM campaigns WHERE id = 51) as user_id
UNION ALL
SELECT 
    'AI Leads' as type,
    niche,
    user_id
FROM lead_ai
WHERE status = 'pending'
GROUP BY niche, user_id;