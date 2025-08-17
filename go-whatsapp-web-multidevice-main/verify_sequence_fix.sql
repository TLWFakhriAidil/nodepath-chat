-- Verify the sequence duplicates fix

-- 1. Check if duplicates are gone
SELECT 'Checking for duplicates:' as status;
SELECT sequence_id, contact_phone, current_step, COUNT(*) as count
FROM sequence_contacts
GROUP BY sequence_id, contact_phone, current_step
HAVING COUNT(*) > 1;

-- 2. Check constraints
SELECT 'Checking constraints:' as status;
SELECT constraint_name, constraint_type
FROM information_schema.table_constraints
WHERE table_name = 'sequence_contacts'
  AND constraint_type IN ('UNIQUE', 'PRIMARY KEY');

-- 3. View current sequence contacts (show first 20)
SELECT 'Current sequence contacts:' as status;
SELECT 
    substring(sequence_id::text, 1, 8) as seq_id,
    contact_phone,
    contact_name,
    current_step,
    status,
    next_trigger_time,
    completed_at
FROM sequence_contacts
ORDER BY contact_name, current_step
LIMIT 20;

-- 4. Check for any stuck active records
SELECT 'Checking stuck active records:' as status;
SELECT 
    contact_name,
    contact_phone,
    current_step,
    status,
    processing_started_at,
    NOW() - processing_started_at as stuck_duration
FROM sequence_contacts
WHERE status = 'active'
  AND processing_started_at IS NOT NULL
  AND processing_started_at < NOW() - INTERVAL '30 minutes';

-- 5. Summary by status
SELECT 'Summary by status:' as status;
SELECT status, COUNT(*) as count
FROM sequence_contacts
GROUP BY status
ORDER BY status;
