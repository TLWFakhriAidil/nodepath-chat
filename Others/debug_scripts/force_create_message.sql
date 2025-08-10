-- Force trigger the campaign manually

-- 1. First ensure campaign is in correct state
UPDATE campaigns 
SET status = 'pending',
    updated_at = NOW()
WHERE title = 'tsst send'
AND status != 'sent';

-- 2. Manually create broadcast message to test
DO $$
DECLARE
    v_campaign_id INTEGER;
    v_user_id UUID;
    v_device_id UUID;
    v_lead_phone VARCHAR;
    v_message TEXT;
    v_image_url TEXT;
BEGIN
    -- Get campaign details
    SELECT id, user_id, message, image_url
    INTO v_campaign_id, v_user_id, v_message, v_image_url
    FROM campaigns
    WHERE title = 'tsst send'
    LIMIT 1;
    
    -- Get device and lead details
    SELECT l.phone, l.device_id
    INTO v_lead_phone, v_device_id
    FROM leads l
    WHERE l.phone = '60108924904'
    LIMIT 1;
    
    -- Create broadcast message
    IF v_campaign_id IS NOT NULL AND v_device_id IS NOT NULL THEN
        INSERT INTO broadcast_messages (
            user_id,
            device_id,
            campaign_id,
            recipient_phone,
            type,
            content,
            media_url,
            status,
            scheduled_at,
            created_at,
            updated_at
        ) VALUES (
            v_user_id,
            v_device_id,
            v_campaign_id,
            v_lead_phone,
            'text',
            v_message,
            v_image_url,
            'pending',
            NOW(),
            NOW(),
            NOW()
        );
        
        RAISE NOTICE 'Broadcast message created successfully!';
    ELSE
        RAISE NOTICE 'Failed to create message - missing campaign or device';
    END IF;
END $$;

-- 3. Check if message was created
SELECT * FROM broadcast_messages 
WHERE recipient_phone = '60108924904'
ORDER BY created_at DESC
LIMIT 1;