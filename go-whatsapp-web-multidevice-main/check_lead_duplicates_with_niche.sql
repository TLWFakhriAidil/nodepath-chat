-- Check leads that would be duplicates with new niche-based logic
-- This shows leads with same device_id, user_id, phone, AND niche

-- 1. Find duplicate leads (same device, user, phone, AND niche)
SELECT 
    device_id,
    user_id,
    phone,
    niche,
    COUNT(*) as duplicate_count,
    STRING_AGG(id::text, ', ') as lead_ids,
    STRING_AGG(name, ', ') as names,
    STRING_AGG(created_at::text, ', ') as created_dates
FROM leads
WHERE niche IS NOT NULL AND niche != ''
GROUP BY device_id, user_id, phone, niche
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC;

-- 2. Show leads with same phone but different niches (these are NOT duplicates anymore)
SELECT 
    device_id,
    user_id,
    phone,
    COUNT(DISTINCT niche) as niche_count,
    STRING_AGG(DISTINCT niche, ', ') as niches,
    COUNT(*) as total_leads
FROM leads
WHERE niche IS NOT NULL AND niche != ''
GROUP BY device_id, user_id, phone
HAVING COUNT(DISTINCT niche) > 1
ORDER BY niche_count DESC;

-- 3. Example: Show all leads for a specific phone number to see niche variations
-- Replace '60123456789' with an actual phone number from your data
/*
SELECT 
    id,
    name,
    phone,
    niche,
    device_id,
    created_at
FROM leads
WHERE phone = '60123456789'
ORDER BY created_at DESC;
*/