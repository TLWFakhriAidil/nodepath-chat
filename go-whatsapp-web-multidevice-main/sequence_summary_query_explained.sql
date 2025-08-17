-- SEQUENCE SUMMARY QUERY EXPLANATION
-- For public device view with URL parameter filtering

-- Step 1: Get all sequences for a specific device
-- URL parameter: device_id (e.g., ?device_id=abc123)

-- Main query to get sequence summary data:
SELECT 
    s.id as sequence_id,
    s.name as sequence_name,
    s.trigger,
    s.niche,
    s.target_status,
    s.status as sequence_status,
    s.total_days,
    
    -- Count total unique contacts that should receive messages
    COUNT(DISTINCT CONCAT(bm.recipient_phone, '|', bm.device_id)) AS total_should_send,
    
    -- Count successfully sent messages (distinct by phone+device)
    COUNT(DISTINCT CASE 
        WHEN bm.status = 'sent' 
        AND (bm.error_message IS NULL OR bm.error_message = '')
        THEN CONCAT(bm.recipient_phone, '|', bm.device_id) 
    END) AS done_send,
    
    -- Count failed messages (distinct by phone+device)
    COUNT(DISTINCT CASE 
        WHEN bm.status = 'failed'
        THEN CONCAT(bm.recipient_phone, '|', bm.device_id) 
    END) AS failed_send,
    
    -- Count remaining (pending/queued) messages
    COUNT(DISTINCT CASE 
        WHEN bm.status IN ('pending', 'queued')
        THEN CONCAT(bm.recipient_phone, '|', bm.device_id) 
    END) AS remaining_send

FROM sequences s
INNER JOIN broadcast_messages bm ON bm.sequence_id = s.id
WHERE 
    -- Filter by device_id from URL parameter
    bm.device_id = ? -- This is the device_id parameter from URL
    -- Only show sequences that have messages
    AND bm.sequence_id IS NOT NULL
GROUP BY 
    s.id, s.name, s.trigger, s.niche, s.target_status, s.status, s.total_days
ORDER BY 
    s.created_at DESC;

-- EXPLANATION OF THE LOGIC:

-- 1. FILTERING BY DEVICE:
--    WHERE bm.device_id = ? filters to only show sequences for specific device
--    This parameter comes from the URL (e.g., /public/device/123/sequences)

-- 2. GROUPING BY SEQUENCE:
--    GROUP BY sequence_id gives us one row per sequence
--    Shows all sequences that have broadcast messages for this device

-- 3. COUNTING LOGIC:
--    - We use CONCAT(recipient_phone, '|', device_id) to create unique identifier
--    - COUNT(DISTINCT ...) ensures we count each recipient only once per sequence
--    - This handles cases where same phone might get multiple step messages

-- 4. STATUS BREAKDOWN:
--    - done_send: Messages successfully sent (status='sent' with no error)
--    - failed_send: Messages that failed (status='failed')
--    - remaining_send: Messages waiting to be sent (status='pending' or 'queued')
--    - total_should_send: All unique recipients for this sequence

-- EXAMPLE OUTPUT:
-- sequence_id | sequence_name | total_should_send | done_send | failed_send | remaining_send
-- ------------|---------------|-------------------|-----------|-------------|----------------
-- seq-123     | Welcome Flow  | 100              | 85        | 5           | 10
-- seq-456     | Follow Up     | 50               | 45        | 2           | 3

-- ADDITIONAL QUERIES FOR STEP-LEVEL DETAILS:

-- Get breakdown by sequence steps:
SELECT 
    bm.sequence_id,
    bm.sequence_stepid,
    ss.day_number,
    ss.trigger as step_trigger,
    COUNT(DISTINCT CONCAT(bm.recipient_phone, '|', bm.device_id)) as total_contacts,
    COUNT(DISTINCT CASE WHEN bm.status = 'sent' THEN CONCAT(bm.recipient_phone, '|', bm.device_id) END) as sent,
    COUNT(DISTINCT CASE WHEN bm.status = 'failed' THEN CONCAT(bm.recipient_phone, '|', bm.device_id) END) as failed
FROM broadcast_messages bm
LEFT JOIN sequence_steps ss ON ss.id = bm.sequence_stepid
WHERE 
    bm.device_id = ? -- Device filter
    AND bm.sequence_id = ? -- Specific sequence
GROUP BY 
    bm.sequence_id, bm.sequence_stepid, ss.day_number, ss.trigger
ORDER BY 
    ss.day_number;

-- VERIFICATION QUERY:
-- To verify the counts are correct:
SELECT 
    'Total Messages' as metric,
    COUNT(*) as count
FROM broadcast_messages
WHERE device_id = ? AND sequence_id IS NOT NULL
UNION ALL
SELECT 
    'Unique Recipients',
    COUNT(DISTINCT CONCAT(recipient_phone, '|', device_id))
FROM broadcast_messages  
WHERE device_id = ? AND sequence_id IS NOT NULL
UNION ALL
SELECT 
    'Total Sequences',
    COUNT(DISTINCT sequence_id)
FROM broadcast_messages
WHERE device_id = ? AND sequence_id IS NOT NULL;