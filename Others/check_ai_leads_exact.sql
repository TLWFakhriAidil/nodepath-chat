-- Direct test of the exact query used by AI Campaign
SELECT COUNT(*) as should_find_leads
FROM leads_ai
WHERE user_id = 'c57309ce-0e4c-4c26-b6cd-092cc69f3806'
  AND niche = 'TITAN,B1'
  AND target_status = 'customer'
  AND status = 'pending';

-- Check if there's any case sensitivity or spacing issue
SELECT DISTINCT 
    user_id,
    niche,
    LENGTH(niche) as niche_length,
    target_status,
    status
FROM leads_ai
WHERE user_id = 'c57309ce-0e4c-4c26-b6cd-092cc69f3806'
  AND status = 'pending';

-- Check exact character codes to detect hidden characters
SELECT 
    id,
    niche,
    ENCODE(niche::bytea, 'hex') as niche_hex,
    target_status,
    ENCODE(target_status::bytea, 'hex') as target_status_hex
FROM leads_ai
WHERE user_id = 'c57309ce-0e4c-4c26-b6cd-092cc69f3806'
  AND status = 'pending'
LIMIT 5;