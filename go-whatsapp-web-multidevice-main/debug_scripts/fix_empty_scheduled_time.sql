-- Debug: Check what's actually in the campaign

-- 1. See exact values
SELECT 
    id,
    title,
    campaign_date,
    scheduled_time,
    scheduled_time IS NULL as is_null,
    scheduled_time = '' as is_empty_string,
    LENGTH(COALESCE(scheduled_time::text, '')) as time_length,
    status
FROM campaigns
WHERE title = 'tsst send';

-- 2. Force both scheduled_time to empty string AND status to pending
UPDATE campaigns 
SET scheduled_time = '',  -- Empty string instead of NULL
    status = 'pending',
    updated_at = NOW()
WHERE title = 'tsst send';

-- 3. Alternative: Update campaign date to ensure it matches
UPDATE campaigns 
SET campaign_date = '2025-06-27',  -- Match server date
    scheduled_time = '',            -- Empty string
    status = 'pending'
WHERE title = 'tsst send';

-- 4. Nuclear option: Create message directly to test worker
INSERT INTO broadcast_messages (
    user_id, device_id, campaign_id, recipient_phone, 
    type, content, status, scheduled_at, created_at, updated_at
)
SELECT 
    c.user_id,
    '2de48db2-f1ab-4d81-8a26-58b01df75bdf',
    c.id,
    '60108924904',
    'text',
    c.message,
    'pending',
    NOW(), NOW(), NOW()
FROM campaigns c
WHERE c.title = 'tsst send'
LIMIT 1;