-- Check why device IDs don't match

-- 1. See the mismatch
SELECT 
    sc.contact_phone,
    sc.contact_name,
    sc.assigned_device_id as sequence_device,
    l.device_id as lead_device,
    CASE 
        WHEN sc.assigned_device_id IS NULL THEN 'NULL - will use lead device!'
        WHEN sc.assigned_device_id = l.device_id THEN 'MATCH'
        ELSE 'MISMATCH!'
    END as status
FROM sequence_contacts sc
LEFT JOIN leads l ON l.phone = sc.contact_phone
WHERE sc.status IN ('active', 'pending');

-- 2. Fix: Update sequence_contacts to use lead's device_id where NULL
UPDATE sequence_contacts sc
SET assigned_device_id = l.device_id
FROM leads l
WHERE sc.contact_phone = l.phone
  AND sc.assigned_device_id IS NULL
  AND l.device_id IS NOT NULL;

-- 3. Fix: Update sequence_contacts to match lead's device (if you want them to match)
-- WARNING: Only run this if you want sequence to use the same device as lead
UPDATE sequence_contacts sc
SET assigned_device_id = l.device_id
FROM leads l
WHERE sc.contact_phone = l.phone
  AND l.device_id IS NOT NULL;

-- 4. Verify fix
SELECT 
    COUNT(*) as total,
    COUNT(CASE WHEN sc.assigned_device_id IS NULL THEN 1 END) as null_devices,
    COUNT(CASE WHEN sc.assigned_device_id != l.device_id THEN 1 END) as mismatches
FROM sequence_contacts sc
LEFT JOIN leads l ON l.phone = sc.contact_phone;
