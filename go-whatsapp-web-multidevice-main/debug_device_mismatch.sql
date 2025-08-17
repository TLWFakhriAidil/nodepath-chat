-- Debug sequence device mismatch issue

-- 1. Check for mismatches between assigned device and broadcast device
SELECT 
    sc.sequence_id,
    sc.contact_phone,
    sc.contact_name,
    sc.current_step,
    sc.assigned_device_id,
    bm.device_id as broadcast_device_id,
    CASE 
        WHEN sc.assigned_device_id = bm.device_id THEN '✓ MATCH'
        WHEN sc.assigned_device_id IS NULL THEN '⚠ No assigned device'
        ELSE '❌ MISMATCH!'
    END as status,
    bm.created_at,
    bm.status as message_status
FROM sequence_contacts sc
INNER JOIN broadcast_messages bm 
    ON bm.recipient_phone = sc.contact_phone 
    AND bm.sequence_id = sc.sequence_id
WHERE bm.created_at > NOW() - INTERVAL '24 hours'
ORDER BY bm.created_at DESC
LIMIT 50;

-- 2. Check what devices are in sequence_contacts vs leads
SELECT 
    sc.contact_phone,
    sc.contact_name,
    sc.assigned_device_id as seq_assigned_device,
    l.device_id as lead_device_id,
    CASE 
        WHEN sc.assigned_device_id = l.device_id THEN '✓ MATCH'
        WHEN sc.assigned_device_id IS NULL AND l.device_id IS NOT NULL THEN '⚠ Using lead device'
        WHEN sc.assigned_device_id IS NOT NULL AND l.device_id IS NULL THEN '⚠ Lead missing device'
        ELSE '❌ DIFFERENT!'
    END as device_match,
    ud1.device_name as seq_device_name,
    ud2.device_name as lead_device_name
FROM sequence_contacts sc
LEFT JOIN leads l ON l.phone = sc.contact_phone
LEFT JOIN user_devices ud1 ON ud1.id = sc.assigned_device_id
LEFT JOIN user_devices ud2 ON ud2.id = l.device_id
WHERE sc.status IN ('active', 'pending')
ORDER BY device_match DESC, sc.contact_phone
LIMIT 50;

-- 3. Show the SQL query that sequence processor uses
-- This is what COALESCE does:
SELECT 
    sc.contact_phone,
    sc.assigned_device_id,
    l.device_id as lead_device_id,
    COALESCE(sc.assigned_device_id, l.device_id) as preferred_device_id,
    CASE 
        WHEN sc.assigned_device_id IS NOT NULL THEN 'Using sequence assigned device'
        WHEN l.device_id IS NOT NULL THEN 'Using lead device (fallback)'
        ELSE 'No device available!'
    END as device_source
FROM sequence_contacts sc
LEFT JOIN leads l ON l.phone = sc.contact_phone
WHERE sc.status = 'active'
LIMIT 20;

-- 4. Check if assigned_device_id is NULL in sequence_contacts
SELECT 
    COUNT(*) as total_contacts,
    COUNT(assigned_device_id) as with_assigned_device,
    COUNT(*) - COUNT(assigned_device_id) as missing_assigned_device,
    ROUND((COUNT(*) - COUNT(assigned_device_id))::numeric / COUNT(*) * 100, 2) as missing_percentage
FROM sequence_contacts
WHERE status IN ('active', 'pending');

-- 5. Find recent broadcast messages to see what device they're using
SELECT 
    bm.id,
    bm.recipient_phone,
    bm.device_id,
    ud.device_name,
    bm.sequence_id,
    bm.created_at,
    bm.status
FROM broadcast_messages bm
LEFT JOIN user_devices ud ON ud.id = bm.device_id
WHERE bm.sequence_id IS NOT NULL
  AND bm.created_at > NOW() - INTERVAL '1 hour'
ORDER BY bm.created_at DESC
LIMIT 20;

-- 6. Fix: Update sequence_contacts to have assigned_device_id from leads
-- Run this if assigned_device_id is NULL but leads have device_id
UPDATE sequence_contacts sc
SET assigned_device_id = l.device_id
FROM leads l
WHERE sc.contact_phone = l.phone
  AND sc.assigned_device_id IS NULL
  AND l.device_id IS NOT NULL;
