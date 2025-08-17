-- Remove the database trigger that prevents pending-first approach
-- This trigger was blocking Step 2 from being marked active before Step 1 completed

-- Drop the trigger
DROP TRIGGER IF EXISTS enforce_step_sequence ON sequence_contacts;

-- Drop the function if it exists
DROP FUNCTION IF EXISTS check_step_sequence();

-- Verify triggers are removed
SELECT 
    t.tgname AS trigger_name,
    t.tgenabled AS enabled,
    pg_get_triggerdef(t.oid) AS definition
FROM pg_trigger t
WHERE t.tgrelid = 'sequence_contacts'::regclass
AND NOT t.tgisinternal;
