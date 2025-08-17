-- COMPREHENSIVE DUPLICATE CHECK

-- 1. Check if unique constraints exist on broadcast_messages
SELECT 
    CONSTRAINT_NAME,
    COLUMN_NAME
FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
WHERE TABLE_NAME = 'broadcast_messages'
    AND TABLE_SCHEMA = 'admin_railway'
    AND CONSTRAINT_NAME LIKE '%unique%'
ORDER BY CONSTRAINT_NAME;

-- 2. Check for any messages with similar content pattern
SELECT 
    id,
    recipient_phone,
    LEFT(content, 100) as message_preview,
    status,
    DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
    DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
    sequence_stepid,
    processing_worker_id
FROM broadcast_messages
WHERE recipient_phone LIKE '%128198574%'
    AND (content LIKE '%Pagi%' 
        OR content LIKE '%Daddy%' 
        OR content LIKE '%Dassler%'
        OR content LIKE '%90%'
        OR content LIKE '%nutrisi%')
ORDER BY created_at DESC;

-- 3. Check sequence_steps table for the original message
SELECT 
    id,
    sequence_id,
    day_number,
    LEFT(message_text, 100) as message_preview,
    LEFT(content, 100) as content_preview
FROM sequence_steps
WHERE (message_text LIKE '%Pagi Daddy%' 
    OR content LIKE '%Pagi Daddy%'
    OR message_text LIKE '%90% anak%'
    OR content LIKE '%90% anak%');

-- 4. Check if there are multiple active sequences for this contact
SELECT 
    sc.sequence_id,
    s.name as sequence_name,
    sc.status,
    sc.current_step,
    sc.contact_phone,
    DATE_FORMAT(sc.created_at, '%Y-%m-%d %H:%i:%s') as enrolled_at
FROM sequence_contacts sc
JOIN sequences s ON sc.sequence_id = s.id
WHERE sc.contact_phone LIKE '%128198574%'
ORDER BY sc.created_at DESC;

-- 5. Check for any orphaned messages (no sequence_id but have sequence_stepid)
SELECT 
    COUNT(*) as orphaned_count,
    GROUP_CONCAT(id) as message_ids
FROM broadcast_messages
WHERE sequence_id IS NULL 
    AND sequence_stepid IS NOT NULL
    AND recipient_phone LIKE '%128198574%';

-- 6. Check processing logs for this phone number
SELECT 
    id,
    recipient_phone,
    processing_worker_id,
    DATE_FORMAT(processing_started_at, '%Y-%m-%d %H:%i:%s') as processing_started,
    status,
    error_message
FROM broadcast_messages
WHERE recipient_phone LIKE '%128198574%'
    AND processing_worker_id IS NOT NULL
ORDER BY processing_started_at DESC;
