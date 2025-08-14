-- Migration: Remove sequence step enforcement trigger
-- Date: 2025-01-20
-- Purpose: Allow pending-first approach for sequence processing

-- Drop the trigger that prevents activating steps out of order
DROP TRIGGER IF EXISTS enforce_step_sequence ON sequence_contacts;

-- Drop the associated function
DROP FUNCTION IF EXISTS check_step_sequence() CASCADE;

-- Add comment explaining the change
COMMENT ON TABLE sequence_contacts IS 'Stores sequence contact progression. Uses pending-first approach - all steps start as pending, no activation required.';
