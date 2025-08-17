-- IMMEDIATE FIX: Set campaign to trigger right away

-- Option 1: Set scheduled_time to NULL (triggers immediately)
UPDATE campaigns 
SET scheduled_time = NULL,
    status = 'pending',
    updated_at = NOW()
WHERE title = 'tsst send';

-- Option 2: Create a fresh test campaign that will trigger
INSERT INTO campaigns (
    user_id, 
    title, 
    niche, 
    target_status, 
    message, 
    campaign_date, 
    scheduled_time,  -- NULL means run immediately
    min_delay_seconds, 
    max_delay_seconds, 
    status, 
    created_at, 
    updated_at
) VALUES (
    'de078f16-3266-4ab3-8153-a248b015228f',  -- Your user ID
    'test_immediate_' || NOW()::time,
    'VITAC',
    'customer',
    'URGENT TEST: Campaign trigger test!',
    CURRENT_DATE,  -- Today's date
    NULL,          -- No scheduled time = immediate
    5,
    10,
    'pending',
    NOW(),
    NOW()
);

-- Check results
SELECT id, title, campaign_date, scheduled_time, status 
FROM campaigns 
WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
ORDER BY created_at DESC;