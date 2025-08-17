-- Sequence System Optimization Migrations
-- For 3000 device simultaneous operation with trigger-based flow

-- ============================================
-- PHASE 1: Add Simple Trigger to Leads Table
-- ============================================

-- Add trigger column to leads table (comma-separated for multiple sequences)
ALTER TABLE leads ADD COLUMN IF NOT EXISTS trigger VARCHAR(1000); -- e.g., "fitness_start,crypto_start,realestate_start"

-- Create index for trigger searches
CREATE INDEX IF NOT EXISTS idx_leads_trigger ON leads(trigger) WHERE trigger IS NOT NULL;

-- ============================================
-- PHASE 2: Enhanced Sequence Tracking in sequence_contacts
-- ============================================

-- Add fields for better trigger-based processing
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS current_trigger VARCHAR(255);
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS next_trigger_time TIMESTAMP;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS processing_device_id UUID;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS processing_started_at TIMESTAMP;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS last_error TEXT;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS retry_count INT DEFAULT 0;

-- Create indexes for fast processing
CREATE INDEX IF NOT EXISTS idx_seq_contacts_trigger ON sequence_contacts(current_trigger, next_trigger_time) 
WHERE status = 'active' AND current_trigger IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_seq_contacts_processing ON sequence_contacts(processing_device_id, processing_started_at)
WHERE processing_device_id IS NOT NULL;

-- ============================================
-- PHASE 3: Link Sequence Steps with Triggers
-- ============================================

-- Ensure sequence_steps has proper trigger flow
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS next_trigger VARCHAR(255);
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger_delay_hours INT DEFAULT 24;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS is_entry_point BOOLEAN DEFAULT false;

-- Create unique constraint on trigger
CREATE UNIQUE INDEX IF NOT EXISTS idx_sequence_steps_unique_trigger ON sequence_steps(trigger) WHERE trigger IS NOT NULL;

-- Mark entry points for each sequence
UPDATE sequence_steps SET is_entry_point = true WHERE day_number = 1;

-- ============================================
-- PHASE 4: Create Processing Tables
-- ============================================

-- Track trigger processing for monitoring
CREATE TABLE IF NOT EXISTS trigger_process_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sequence_contact_id UUID NOT NULL,
    lead_phone VARCHAR(50) NOT NULL,
    device_id UUID NOT NULL,
    trigger_name VARCHAR(255) NOT NULL,
    status VARCHAR(50), -- 'success', 'failed', 'skipped'
    error_message TEXT,
    processed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    INDEX idx_process_log_contact (sequence_contact_id, processed_at DESC),
    INDEX idx_process_log_device (device_id, processed_at DESC)
);

-- Device load balancing
CREATE TABLE IF NOT EXISTS device_load_balance (
    device_id UUID PRIMARY KEY,
    messages_hour INT DEFAULT 0,
    messages_today INT DEFAULT 0,
    last_reset_hour TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_reset_day TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    is_available BOOLEAN DEFAULT true,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- PHASE 5: Functions for Trigger Management
-- ============================================

-- Function to parse comma-separated triggers
CREATE OR REPLACE FUNCTION parse_lead_triggers(p_triggers VARCHAR(1000))
RETURNS TABLE(trigger_name VARCHAR(255)) AS $$
BEGIN
    RETURN QUERY
    SELECT TRIM(unnest(string_to_array(p_triggers, ','))) AS trigger_name
    WHERE p_triggers IS NOT NULL AND p_triggers != '';
END;
$$ LANGUAGE plpgsql;

-- Function to add trigger to lead
CREATE OR REPLACE FUNCTION add_trigger_to_lead(
    p_lead_id UUID,
    p_new_trigger VARCHAR(255)
) RETURNS VOID AS $$
DECLARE
    v_current_triggers VARCHAR(1000);
BEGIN
    -- Get current triggers
    SELECT trigger INTO v_current_triggers FROM leads WHERE id = p_lead_id;
    
    -- Add new trigger if not already present
    IF v_current_triggers IS NULL OR v_current_triggers = '' THEN
        UPDATE leads SET trigger = p_new_trigger WHERE id = p_lead_id;
    ELSIF position(p_new_trigger in v_current_triggers) = 0 THEN
        UPDATE leads SET trigger = v_current_triggers || ',' || p_new_trigger WHERE id = p_lead_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to remove trigger from lead
CREATE OR REPLACE FUNCTION remove_trigger_from_lead(
    p_lead_id UUID,
    p_trigger_to_remove VARCHAR(255)
) RETURNS VOID AS $$
DECLARE
    v_current_triggers VARCHAR(1000);
    v_new_triggers VARCHAR(1000);
BEGIN
    -- Get current triggers
    SELECT trigger INTO v_current_triggers FROM leads WHERE id = p_lead_id;
    
    IF v_current_triggers IS NOT NULL THEN
        -- Remove the trigger and clean up commas
        v_new_triggers := REGEXP_REPLACE(
            REGEXP_REPLACE(v_current_triggers, '(^|,)' || p_trigger_to_remove || '(,|$)', '\1\2', 'g'),
            '^,|,$|,,+', '', 'g'
        );
        
        -- Update lead
        UPDATE leads SET trigger = NULLIF(v_new_triggers, '') WHERE id = p_lead_id;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Function to process sequence triggers
CREATE OR REPLACE FUNCTION process_sequence_triggers()
RETURNS TABLE(
    contact_id UUID,
    lead_phone VARCHAR(50),
    sequence_id UUID,
    current_trigger VARCHAR(255),
    message_text TEXT,
    message_type VARCHAR(50),
    media_url TEXT,
    next_trigger VARCHAR(255)
) AS $$
BEGIN
    RETURN QUERY
    WITH active_sequences AS (
        -- Get all active sequences and their entry triggers
        SELECT s.id, s.trigger_prefix, ss.trigger as entry_trigger
        FROM sequences s
        JOIN sequence_steps ss ON ss.sequence_id = s.id
        WHERE s.is_active = true AND ss.is_entry_point = true
    ),
    leads_with_triggers AS (
        -- Get all leads with matching triggers
        SELECT DISTINCT l.id, l.phone, t.trigger_name, l.device_id
        FROM leads l
        CROSS JOIN LATERAL parse_lead_triggers(l.trigger) t
        WHERE t.trigger_name IN (SELECT entry_trigger FROM active_sequences)
    ),
    new_enrollments AS (
        -- Find leads that need to be enrolled in sequences
        SELECT 
            lwt.id as lead_id,
            lwt.phone,
            a.id as sequence_id,
            a.entry_trigger
        FROM leads_with_triggers lwt
        JOIN active_sequences a ON a.entry_trigger = lwt.trigger_name
        LEFT JOIN sequence_contacts sc ON sc.sequence_id = a.id AND sc.contact_phone = lwt.phone
        WHERE sc.id IS NULL -- Not already enrolled
    )
    -- Insert new enrollments
    INSERT INTO sequence_contacts (
        sequence_id, 
        contact_phone, 
        contact_name, 
        current_step,
        current_day,
        current_trigger,
        next_trigger_time,
        status,
        enrolled_at
    )
    SELECT 
        ne.sequence_id,
        ne.phone,
        l.name,
        1,
        1,
        ne.entry_trigger,
        CURRENT_TIMESTAMP, -- Process immediately
        'active',
        CURRENT_TIMESTAMP
    FROM new_enrollments ne
    JOIN leads l ON l.id = ne.lead_id
    ON CONFLICT (sequence_id, contact_phone) DO NOTHING;
    
    -- Return contacts ready for processing
    RETURN QUERY
    SELECT 
        sc.id,
        sc.contact_phone,
        sc.sequence_id,
        sc.current_trigger,
        ss.message_text,
        ss.message_type,
        ss.media_url,
        ss.next_trigger
    FROM sequence_contacts sc
    JOIN sequence_steps ss ON ss.trigger = sc.current_trigger
    JOIN sequences s ON s.id = sc.sequence_id
    WHERE sc.status = 'active'
        AND s.is_active = true
        AND sc.next_trigger_time <= CURRENT_TIMESTAMP
        AND sc.processing_device_id IS NULL
    ORDER BY s.priority DESC, sc.next_trigger_time ASC
    LIMIT 1000;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- PHASE 6: Optimized Batch Processing
-- ============================================

-- Function to claim contacts for device processing
CREATE OR REPLACE FUNCTION claim_contacts_for_device(
    p_device_id UUID,
    p_batch_size INT DEFAULT 100
) RETURNS INT AS $$
DECLARE
    v_claimed INT;
BEGIN
    -- Claim batch of contacts
    WITH claimed AS (
        UPDATE sequence_contacts sc
        SET 
            processing_device_id = p_device_id,
            processing_started_at = CURRENT_TIMESTAMP
        FROM (
            SELECT sc2.id
            FROM sequence_contacts sc2
            JOIN sequence_steps ss ON ss.trigger = sc2.current_trigger
            JOIN sequences s ON s.id = sc2.sequence_id
            WHERE sc2.status = 'active'
                AND s.is_active = true
                AND sc2.next_trigger_time <= CURRENT_TIMESTAMP
                AND sc2.processing_device_id IS NULL
            ORDER BY s.priority DESC, sc2.next_trigger_time ASC
            LIMIT p_batch_size
            FOR UPDATE SKIP LOCKED
        ) AS to_claim
        WHERE sc.id = to_claim.id
        RETURNING sc.id
    )
    SELECT COUNT(*) INTO v_claimed FROM claimed;
    
    RETURN v_claimed;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- PHASE 7: Add Monitoring Views
-- ============================================

-- View for sequence progress monitoring
CREATE OR REPLACE VIEW sequence_progress_view AS
SELECT 
    s.name as sequence_name,
    s.trigger_prefix,
    COUNT(DISTINCT sc.contact_phone) as total_contacts,
    COUNT(DISTINCT CASE WHEN sc.status = 'active' THEN sc.contact_phone END) as active_contacts,
    COUNT(DISTINCT CASE WHEN sc.status = 'completed' THEN sc.contact_phone END) as completed_contacts,
    COUNT(DISTINCT ss.trigger) as total_steps,
    AVG(sc.current_day) as avg_progress_day
FROM sequences s
LEFT JOIN sequence_contacts sc ON sc.sequence_id = s.id
LEFT JOIN sequence_steps ss ON ss.sequence_id = s.id
GROUP BY s.id, s.name, s.trigger_prefix;

-- View for device workload
CREATE OR REPLACE VIEW device_workload_view AS
SELECT 
    d.id as device_id,
    d.name as device_name,
    d.status,
    COUNT(DISTINCT sc.id) as processing_count,
    dlb.messages_hour,
    dlb.messages_today,
    dlb.is_available
FROM user_devices d
LEFT JOIN sequence_contacts sc ON sc.processing_device_id = d.id 
    AND sc.processing_started_at > CURRENT_TIMESTAMP - INTERVAL '5 minutes'
LEFT JOIN device_load_balance dlb ON dlb.device_id = d.id
GROUP BY d.id, d.name, d.status, dlb.messages_hour, dlb.messages_today, dlb.is_available;

-- ============================================
-- PHASE 8: Cleanup and Maintenance
-- ============================================

-- Function to clean up stuck processing
CREATE OR REPLACE FUNCTION cleanup_stuck_processing() RETURNS VOID AS $$
BEGIN
    -- Release contacts stuck in processing for more than 5 minutes
    UPDATE sequence_contacts
    SET 
        processing_device_id = NULL,
        processing_started_at = NULL,
        retry_count = retry_count + 1
    WHERE processing_device_id IS NOT NULL
        AND processing_started_at < CURRENT_TIMESTAMP - INTERVAL '5 minutes';
        
    -- Mark contacts as failed after 5 retries
    UPDATE sequence_contacts
    SET status = 'failed'
    WHERE retry_count >= 5 AND status = 'active';
END;
$$ LANGUAGE plpgsql;

-- Create unique constraint to prevent duplicate enrollments
ALTER TABLE sequence_contacts ADD CONSTRAINT uq_sequence_contact 
UNIQUE (sequence_id, contact_phone);

-- Final optimization indexes
CREATE INDEX IF NOT EXISTS idx_leads_phone ON leads(phone);
CREATE INDEX IF NOT EXISTS idx_seq_contacts_phone ON sequence_contacts(contact_phone);
CREATE INDEX IF NOT EXISTS idx_seq_steps_sequence ON sequence_steps(sequence_id, day_number);