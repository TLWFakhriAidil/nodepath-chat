-- Add progress tracking fields to sequences table
ALTER TABLE sequences 
ADD COLUMN IF NOT EXISTS total_contacts INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS active_contacts INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS completed_contacts INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS failed_contacts INTEGER DEFAULT 0,
ADD COLUMN IF NOT EXISTS progress_percentage DECIMAL(5,2) DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS last_activity_at TIMESTAMP,
ADD COLUMN IF NOT EXISTS estimated_completion_at TIMESTAMP;

-- Add index for better performance
CREATE INDEX IF NOT EXISTS idx_sequences_progress ON sequences(progress_percentage);
CREATE INDEX IF NOT EXISTS idx_sequences_last_activity ON sequences(last_activity_at);

-- Create a function to update sequence progress
CREATE OR REPLACE FUNCTION update_sequence_progress(seq_id UUID)
RETURNS void AS $$
DECLARE
    total_count INTEGER;
    active_count INTEGER;
    completed_count INTEGER;
    failed_count INTEGER;
    progress DECIMAL(5,2);
BEGIN
    -- Get counts
    SELECT 
        COUNT(*) FILTER (WHERE TRUE),
        COUNT(*) FILTER (WHERE is_completed = FALSE AND last_message_at IS NOT NULL),
        COUNT(*) FILTER (WHERE is_completed = TRUE),
        COUNT(*) FILTER (WHERE status = 'failed')
    INTO total_count, active_count, completed_count, failed_count
    FROM sequence_contacts
    WHERE sequence_id = seq_id;
    
    -- Calculate progress
    IF total_count > 0 THEN
        progress := ((completed_count + failed_count)::DECIMAL / total_count::DECIMAL) * 100;
    ELSE
        progress := 0;
    END IF;
    
    -- Update sequence
    UPDATE sequences
    SET 
        total_contacts = total_count,
        active_contacts = active_count,
        completed_contacts = completed_count,
        failed_contacts = failed_count,
        progress_percentage = progress,
        last_activity_at = NOW()
    WHERE id = seq_id;
END;
$$ LANGUAGE plpgsql;

-- Add status column to sequence_contacts if missing
ALTER TABLE sequence_contacts 
ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'active';
