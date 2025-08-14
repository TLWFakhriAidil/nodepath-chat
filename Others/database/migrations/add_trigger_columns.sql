-- Add trigger columns to sequences and sequence_steps tables
-- Date: 2025-07-01

-- Add start_trigger and end_trigger to sequences table
ALTER TABLE sequences 
ADD COLUMN IF NOT EXISTS start_trigger VARCHAR(255),
ADD COLUMN IF NOT EXISTS end_trigger VARCHAR(255);

-- Add trigger to sequence_steps table
ALTER TABLE sequence_steps
ADD COLUMN IF NOT EXISTS trigger VARCHAR(255);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_sequences_start_trigger ON sequences(start_trigger);
CREATE INDEX IF NOT EXISTS idx_sequences_end_trigger ON sequences(end_trigger);
CREATE INDEX IF NOT EXISTS idx_sequence_steps_trigger ON sequence_steps(trigger);
