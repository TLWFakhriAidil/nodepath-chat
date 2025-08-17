-- Fix sequence_contacts table by removing any triggers that might reference updated_at
-- and ensure the table structure is correct

-- Drop the existing trigger if it exists
DROP TRIGGER IF EXISTS enforce_step_sequence ON sequence_contacts;

-- Drop the function if we need to recreate it
DROP FUNCTION IF EXISTS check_step_sequence();

-- Recreate the function without any updated_at references
CREATE OR REPLACE FUNCTION check_step_sequence() RETURNS TRIGGER AS $$
BEGIN
    -- When activating a step, ensure all previous steps are completed
    IF NEW.status = 'active' AND OLD.status = 'pending' THEN
        -- Check if there are any incomplete previous steps
        IF EXISTS (
            SELECT 1 
            FROM sequence_contacts 
            WHERE sequence_id = NEW.sequence_id 
              AND contact_phone = NEW.contact_phone
              AND current_step < NEW.current_step
              AND status NOT IN ('completed', 'sent', 'failed')
        ) THEN
            RAISE EXCEPTION 'Cannot activate step % before completing previous steps', NEW.current_step;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Recreate the trigger
CREATE TRIGGER enforce_step_sequence 
    BEFORE UPDATE ON sequence_contacts
    FOR EACH ROW 
    EXECUTE FUNCTION check_step_sequence();

-- Ensure there's no updated_at column
ALTER TABLE sequence_contacts DROP COLUMN IF EXISTS updated_at;

-- Add any missing columns that might be needed
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT NOW();

COMMENT ON TABLE sequence_contacts IS 'Fixed to remove updated_at references - uses status column instead';
