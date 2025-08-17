-- Add missing columns to sequences table
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS device_id UUID;
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS total_days INTEGER DEFAULT 0;
ALTER TABLE sequences ADD COLUMN IF NOT EXISTS is_active BOOLEAN DEFAULT true;

-- Add missing columns to sequence_steps table
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS day INTEGER;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS send_time VARCHAR(10);

-- Add missing columns to sequence_contacts table  
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS current_day INTEGER DEFAULT 0;

-- Create sequence_logs table if not exists
CREATE TABLE IF NOT EXISTS sequence_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_id UUID NOT NULL REFERENCES sequences(id) ON DELETE CASCADE,
    contact_id UUID NOT NULL,
    step_id UUID NOT NULL,
    day INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL,
    message_id VARCHAR(255),
    error_message TEXT,
    sent_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for sequence_logs
CREATE INDEX IF NOT EXISTS idx_sequence_logs_sequence_id ON sequence_logs(sequence_id);
CREATE INDEX IF NOT EXISTS idx_sequence_logs_contact_id ON sequence_logs(contact_id);
CREATE INDEX IF NOT EXISTS idx_sequence_logs_sent_at ON sequence_logs(sent_at);
