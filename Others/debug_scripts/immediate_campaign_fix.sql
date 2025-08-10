-- IMMEDIATE FIX: Clean up scheduled_time and trigger campaign

-- 1. Check what's in scheduled_time
SELECT 
    id,
    scheduled_time,
    scheduled_time::text,
    '"' || scheduled_time::text || '"' as quoted_time
FROM campaigns
WHERE title = 'tsst send';

-- 2. Force clean update
UPDATE campaigns 
SET scheduled_time = '00:00:00'::time,  -- Force to time type
    campaign_date = '2025-06-28'::date,  -- Force to date type
    status = 'pending'
WHERE title = 'tsst send';

-- 3. Alternative: Set to NULL to trigger immediately
UPDATE campaigns 
SET scheduled_time = NULL,
    status = 'pending'
WHERE title = 'tsst send';

-- 4. Last resort: Create the campaign fresh
DELETE FROM campaigns WHERE title = 'tsst send';

INSERT INTO campaigns (
    user_id, title, niche, target_status, message, 
    campaign_date, scheduled_time, min_delay_seconds, max_delay_seconds, 
    status, created_at, updated_at
) VALUES (
    'de078f16-3266-4ab3-8153-a248b015228f',
    'test_auto_trigger',
    'VITAC',
    'customer',
    'PELUANG EMAS UNTUK MAK AYAH! DAPATKAN EXAMA VITAC @ MIN',
    CURRENT_DATE,
    NULL,  -- No scheduled time = run immediately
    10,
    30,
    'pending',
    NOW(),
    NOW()
);