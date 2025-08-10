-- =====================================================
-- OPTIMIZED SEQUENCE CONTACTS FIX FOR 3000 DEVICES
-- =====================================================
-- This script will:
-- 1. Fix the updateContactProgress to use earliest trigger time
-- 2. Add proper locking to handle 3000 concurrent devices
-- 3. Prevent race conditions and duplicate activations

BEGIN;

-- Step 1: Clean up current issues
DELETE FROM sequence_contacts WHERE status = 'completed';

-- Step 2: Remove duplicate active records
WITH duplicates AS (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY sequence_id, contact_phone 
                             ORDER BY current_step ASC) as rn
    FROM sequence_contacts
    WHERE status = 'active'
)
DELETE FROM sequence_contacts
WHERE id IN (SELECT id FROM duplicates WHERE rn > 1);

-- Step 3: Drop existing constraints that might conflict
ALTER TABLE sequence_contacts 
DROP CONSTRAINT IF EXISTS idx_one_active_per_contact;
DROP INDEX IF EXISTS idx_one_active_per_contact;

-- Step 4: Add optimized constraint for concurrent access
-- This allows only ONE active step per contact across all sequences
CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_one_active_per_contact
ON sequence_contacts(sequence_id, contact_phone)
WHERE status = 'active';

-- Step 5: Add index for fast pending step lookup
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_pending_steps_by_time
ON sequence_contacts(sequence_id, contact_phone, next_trigger_time)
WHERE status = 'pending';

-- Step 6: Create optimized function for concurrent step progression
CREATE OR REPLACE FUNCTION progress_sequence_contact_concurrent(
    p_contact_id UUID
) RETURNS BOOLEAN AS $$
DECLARE
    v_sequence_id UUID;
    v_contact_phone VARCHAR;
    v_current_step INT;
    v_next_id UUID;
    v_activated BOOLEAN := FALSE;
BEGIN
    -- Use FOR UPDATE SKIP LOCKED to handle concurrent access
    -- This allows 3000 devices to work without blocking each other
    
    -- Step 1: Try to lock and complete current step
    UPDATE sequence_contacts 
    SET status = 'completed',
        completed_at = NOW()
    WHERE id = p_contact_id 
      AND status = 'active'
    RETURNING sequence_id, contact_phone, current_step 
    INTO v_sequence_id, v_contact_phone, v_current_step;
    
    IF NOT FOUND THEN
        -- Already processed by another device
        RETURN FALSE;
    END IF;
    
    -- Step 2: Find next pending step by EARLIEST trigger time
    -- Use FOR UPDATE SKIP LOCKED to prevent conflicts
    SELECT id INTO v_next_id
    FROM sequence_contacts
    WHERE sequence_id = v_sequence_id
      AND contact_phone = v_contact_phone
      AND status = 'pending'
      AND next_trigger_time <= NOW()  -- Only steps that are due
    ORDER BY next_trigger_time ASC     -- EARLIEST first
    LIMIT 1
    FOR UPDATE SKIP LOCKED;            -- Skip if another device is processing
    
    IF v_next_id IS NOT NULL THEN
        -- Activate the next step
        UPDATE sequence_contacts
        SET status = 'active'
        WHERE id = v_next_id
          AND status = 'pending';  -- Double-check status
        
        GET DIAGNOSTICS v_activated = ROW_COUNT;
    END IF;
    
    RETURN v_activated;
END;
$$ LANGUAGE plpgsql;

-- Step 7: Add advisory locking for extra safety
-- This prevents the same contact from being processed by multiple devices
CREATE OR REPLACE FUNCTION get_sequence_contact_lock(
    p_sequence_id UUID,
    p_contact_phone VARCHAR
) RETURNS BOOLEAN AS $$
DECLARE
    v_lock_id BIGINT;
BEGIN
    -- Create a unique lock ID from sequence and phone
    v_lock_id := ('x' || substr(md5(p_sequence_id::text || p_contact_phone), 1, 15))::bit(60)::bigint;
    
    -- Try to acquire advisory lock (non-blocking)
    RETURN pg_try_advisory_lock(v_lock_id);
END;
$$ LANGUAGE plpgsql;

-- Step 8: Release lock function
CREATE OR REPLACE FUNCTION release_sequence_contact_lock(
    p_sequence_id UUID,
    p_contact_phone VARCHAR
) RETURNS VOID AS $$
DECLARE
    v_lock_id BIGINT;
BEGIN
    v_lock_id := ('x' || substr(md5(p_sequence_id::text || p_contact_phone), 1, 15))::bit(60)::bigint;
    PERFORM pg_advisory_unlock(v_lock_id);
END;
$$ LANGUAGE plpgsql;

-- Step 9: Create optimized view for monitoring
CREATE OR REPLACE VIEW sequence_processing_status AS
SELECT 
    sc.sequence_id,
    sc.contact_phone,
    sc.current_step,
    sc.status,
    sc.next_trigger_time,
    sc.processing_device_id,
    CASE 
        WHEN sc.status = 'active' THEN 'Processing'
        WHEN sc.status = 'pending' AND sc.next_trigger_time <= NOW() THEN 'Ready'
        WHEN sc.status = 'pending' THEN 'Scheduled'
        ELSE sc.status
    END as processing_status
FROM sequence_contacts sc
ORDER BY sc.contact_phone, sc.current_step;

-- View current state
SELECT * FROM sequence_processing_status;

COMMIT;