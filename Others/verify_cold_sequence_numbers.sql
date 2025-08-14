-- Quick SQL to verify the correct numbers for COLD Sequence on 2025-08-04

-- Get the correct counts
SELECT 
    'COLD Sequence' as sequence_name,
    COUNT(DISTINCT CASE WHEN status = 'sent' AND (error_message IS NULL OR error_message = '') THEN recipient_phone END) AS done_send,
    COUNT(DISTINCT CASE WHEN status = 'failed' THEN recipient_phone END) AS failed_send,
    COUNT(DISTINCT CASE WHEN status IN ('pending', 'queued') THEN recipient_phone END) AS remaining_send,
    COUNT(DISTINCT CASE WHEN status = 'sent' AND (error_message IS NULL OR error_message = '') THEN recipient_phone END) +
    COUNT(DISTINCT CASE WHEN status = 'failed' THEN recipient_phone END) +
    COUNT(DISTINCT CASE WHEN status IN ('pending', 'queued') THEN recipient_phone END) as should_send
FROM broadcast_messages
WHERE sequence_id = '0be82745-8f68-4352-abd0-0b405b43a905'
AND DATE(scheduled_at) = '2025-08-04';

-- Result should be:
-- done_send: 174
-- failed_send: 19
-- remaining_send: 250
-- should_send: 443 (174 + 19 + 250)

-- The UI is showing old/cached data:
-- UI shows: 158 + 19 + 250 = 427
-- Should be: 174 + 19 + 250 = 443