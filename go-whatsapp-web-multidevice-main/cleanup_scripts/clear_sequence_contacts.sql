-- Clear all sequence contacts from the database
-- This will remove ALL records but keep the table structure

-- Option 1: DELETE (slower but logs each row deletion)
DELETE FROM sequence_contacts;

-- Option 2: TRUNCATE (faster, resets any auto-increment counters)
-- TRUNCATE TABLE sequence_contacts RESTART IDENTITY CASCADE;

-- Verify the cleanup
SELECT COUNT(*) as remaining_contacts FROM sequence_contacts;

-- Optional: Also clear related tables if needed
-- DELETE FROM sequence_logs;
-- DELETE FROM broadcast_messages WHERE sequence_id IS NOT NULL;