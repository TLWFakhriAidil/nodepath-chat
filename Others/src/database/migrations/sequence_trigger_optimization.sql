-- Sequence System Optimization Migrations
-- Trigger-based system for 3000 devices with multiple sequences support

-- ============================================
-- PHASE 1: Simple Trigger System
-- ============================================

-- Add trigger column to leads table (comma-separated like niche)
ALTER TABLE leads ADD COLUMN IF NOT EXISTS trigger VARCHAR(500);

-- Example: trigger = "FITNESS_START,CRYPTO_START,PROMO_2024"
-- This allows one lead to be in multiple sequences

-- Create index for trigger searches
CREATE INDEX IF NOT EXISTS idx_leads_trigger ON leads(trigger) WHERE trigger IS NOT NULL AND trigger != '';

-- ============================================
-- PHASE 2: Enhance Sequence Steps with Better Triggers
-- ============================================

-- Ensure sequence_steps has proper trigger column
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS next_trigger VARCHAR(255);
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS trigger_delay_hours INT DEFAULT 24;
ALTER TABLE sequence_steps ADD COLUMN IF NOT EXISTS is_final_step BOOLEAN DEFAULT false;

-- Create trigger lookup index
CREATE INDEX IF NOT EXISTS idx_sequence_steps_trigger ON sequence_steps(trigger);

-- Update sequence_steps to mark final steps
UPDATE sequence_steps ss
SET is_final_step = true
WHERE NOT EXISTS (
    SELECT 1 FROM sequence_steps ss2 
    WHERE ss2.sequence_id = ss.sequence_id 
    AND ss2.day_number > ss.day_number
);

-- ============================================
-- PHASE 3: Sequence Contacts Optimization
-- ============================================

-- Add processing fields to sequence_contacts for better tracking
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS last_trigger_sent VARCHAR(255);
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS next_trigger_time TIMESTAMP;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS processing_device_id UUID;
ALTER TABLE sequence_contacts ADD COLUMN IF NOT EXISTS processing_started_at TIMESTAMP;

-- Index for fast processing queries
CREATE INDEX IF NOT EXISTS idx_sequence_contacts_processing 
ON sequence_contacts(status, next_trigger_time) 
WHERE status = 'active' AND next_trigger_time IS NOT NULL;

-- ============================================
-- PHASE 4: Performance Views
-- ============================================

-- View to find leads that need sequence enrollment
CREATE OR REPLACE VIEW leads_pending_enrollment AS
SELECT 
    l.id,
    l.phone,
    l.trigger,
    l.niche,
    l.device_id,
    s.id as sequence_id,
    s.trigger_prefix,
    s.auto_enroll
FROM leads l
CROSS JOIN sequences s
WHERE l.trigger IS NOT NULL 
    AND l.trigger != ''
    AND s.is_active = true
    AND (
        -- Check if lead's trigger contains sequence trigger
        l.trigger LIKE '%' || s.trigger_prefix || '%'
        -- OR auto-enroll by niche match
        OR (s.auto_enroll = true AND l.niche = s.niche)
    )
    AND NOT EXISTS (
        -- Not already in this sequence
        SELECT 1 FROM sequence_contacts sc 
        WHERE sc.sequence_id = s.id 
        AND sc.contact_phone = l.phone
    );

-- View for active sequence contacts ready for next message
CREATE OR REPLACE VIEW sequence_contacts_ready AS
SELECT 
    sc.*,
    ss.message_text,
    ss.message_type,
    ss.media_url,
    ss.caption,
    ss.next_trigger,
    ss.trigger_delay_hours,
    ss.is_final_step,
    s.name as sequence_name,
    l.device_id as lead_device_id
FROM sequence_contacts sc
JOIN sequences s ON s.id = sc.sequence_id
JOIN sequence_steps ss ON ss.sequence_id = sc.sequence_id 
    AND ss.day_number = (sc.current_day + 1)
LEFT JOIN leads l ON l.phone = sc.contact_phone
WHERE sc.status = 'active'
    AND s.is_active = true
    AND (sc.last_message_at IS NULL 
         OR sc.last_message_at < NOW() - INTERVAL '24 hours')
ORDER BY s.priority DESC, sc.last_message_at ASC;

-- ============================================
-- PHASE 5: Trigger Management Functions
-- ============================================

-- Function to add trigger to lead (handles comma separation)
CREATE OR REPLACE FUNCTION add_trigger_to_lead(
    p_lead_id UUID,
    p_new_trigger VARCHAR(255)
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE leads
    SET trigger = CASE
        WHEN trigger IS NULL OR trigger = '' THEN p_new_trigger
        WHEN trigger LIKE '%' || p_new_trigger || '%' THEN trigger -- Already has it
        ELSE trigger || ',' || p_new_trigger
    END
    WHERE id = p_lead_id;
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- Function to remove trigger from lead
CREATE OR REPLACE FUNCTION remove_trigger_from_lead(
    p_lead_id UUID,
    p_trigger_to_remove VARCHAR(255)
) RETURNS BOOLEAN AS $$
BEGIN
    UPDATE leads
    SET trigger = TRIM(BOTH ',' FROM 
        REPLACE(
            REPLACE(
                REPLACE(',' || trigger || ',', ',' || p_trigger_to_remove || ',', ','),
                ',,', ','
            ),
            ',', ', '
        )
    )
    WHERE id = p_lead_id;
    
    -- Clean up empty triggers
    UPDATE leads SET trigger = NULL WHERE trigger = '' AND id = p_lead_id;
    
    RETURN true;
END;
$$ LANGUAGE plpgsql;

-- Function to check if lead has specific trigger
CREATE OR REPLACE FUNCTION lead_has_trigger(
    p_lead_trigger TEXT,
    p_check_trigger VARCHAR(255)
) RETURNS BOOLEAN AS $$
BEGIN
    RETURN p_lead_trigger IS NOT NULL 
        AND (',' || p_lead_trigger || ',') LIKE '%,' || p_check_trigger || ',%';
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- PHASE 6: Device Workload Tracking
-- ============================================

CREATE TABLE IF NOT EXISTS device_message_queue (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL,
    sequence_contact_id UUID,
    phone VARCHAR(20) NOT NULL,
    message_type VARCHAR(50),
    message_text TEXT,
    media_url TEXT,
    status VARCHAR(50) DEFAULT 'pending', -- pending, processing, sent, failed
    priority INT DEFAULT 5,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    processed_at TIMESTAMP,
    error_message TEXT,
    retry_count INT DEFAULT 0,
    
    INDEX idx_queue_device_status (device_id, status, priority DESC, created_at ASC)
);

-- Device performance tracking
CREATE TABLE IF NOT EXISTS device_performance (
    device_id UUID PRIMARY KEY,
    messages_sent_hour INT DEFAULT 0,
    messages_sent_today INT DEFAULT 0,
    last_hour_reset TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    last_day_reset TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    avg_send_time_ms INT,
    success_rate_percent INT DEFAULT 100,
    is_healthy BOOLEAN DEFAULT true,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================
-- PHASE 7: Optimized Processing Functions
-- ============================================

-- Main function to process sequences (called every minute)
CREATE OR REPLACE FUNCTION process_sequence_batch(
    p_batch_size INT DEFAULT 1000
) RETURNS INT AS $$
DECLARE
    v_processed INT := 0;
    v_contact RECORD;
    v_device_id UUID;
BEGIN
    -- Process ready contacts
    FOR v_contact IN 
        SELECT * FROM sequence_contacts_ready 
        LIMIT p_batch_size
    LOOP
        -- Get least loaded device
        SELECT device_id INTO v_device_id
        FROM device_performance
        WHERE is_healthy = true
            AND messages_sent_hour < 80  -- WhatsApp limit
            AND messages_sent_today < 800
        ORDER BY messages_sent_hour ASC
        LIMIT 1;
        
        IF v_device_id IS NULL THEN
            -- Use lead's device as fallback
            v_device_id := v_contact.lead_device_id;
        END IF;
        
        -- Queue message
        INSERT INTO device_message_queue (
            device_id,
            sequence_contact_id,
            phone,
            message_type,
            message_text,
            media_url,
            priority
        ) VALUES (
            v_device_id,
            v_contact.id,
            v_contact.contact_phone,
            v_contact.message_type,
            v_contact.message_text,
            v_contact.media_url,
            5
        );
        
        -- Update sequence contact
        UPDATE sequence_contacts
        SET 
            current_day = current_day + 1,
            last_message_at = CURRENT_TIMESTAMP,
            last_trigger_sent = v_contact.trigger,
            next_trigger_time = CURRENT_TIMESTAMP + (v_contact.trigger_delay_hours || ' hours')::INTERVAL,
            status = CASE 
                WHEN v_contact.is_final_step THEN 'completed'
                ELSE 'active'
            END,
            completed_at = CASE
                WHEN v_contact.is_final_step THEN CURRENT_TIMESTAMP
                ELSE NULL
            END
        WHERE id = v_contact.id;
        
        -- Remove trigger from lead if sequence completed
        IF v_contact.is_final_step AND v_contact.trigger IS NOT NULL THEN
            PERFORM remove_trigger_from_lead(
                (SELECT id FROM leads WHERE phone = v_contact.contact_phone LIMIT 1),
                v_contact.trigger
            );
        END IF;
        
        v_processed := v_processed + 1;
    END LOOP;
    
    RETURN v_processed;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- PHASE 8: Monitoring & Analytics
-- ============================================

CREATE OR REPLACE VIEW sequence_performance_dashboard AS
SELECT 
    s.name as sequence_name,
    s.trigger_prefix,
    COUNT(DISTINCT sc.id) as total_contacts,
    COUNT(DISTINCT CASE WHEN sc.status = 'active' THEN sc.id END) as active_contacts,
    COUNT(DISTINCT CASE WHEN sc.status = 'completed' THEN sc.id END) as completed_contacts,
    AVG(sc.current_day) as avg_progress_day,
    COUNT(DISTINCT l.id) as potential_leads
FROM sequences s
LEFT JOIN sequence_contacts sc ON sc.sequence_id = s.id
LEFT JOIN leads l ON l.trigger LIKE '%' || s.trigger_prefix || '%'
WHERE s.is_active = true
GROUP BY s.id, s.name, s.trigger_prefix;

-- Reset hourly/daily counters
CREATE OR REPLACE FUNCTION reset_device_counters() RETURNS void AS $$
BEGIN
    -- Reset hourly counters
    UPDATE device_performance
    SET messages_sent_hour = 0, last_hour_reset = CURRENT_TIMESTAMP
    WHERE last_hour_reset < CURRENT_TIMESTAMP - INTERVAL '1 hour';
    
    -- Reset daily counters
    UPDATE device_performance
    SET messages_sent_today = 0, last_day_reset = CURRENT_TIMESTAMP
    WHERE last_day_reset < CURRENT_TIMESTAMP - INTERVAL '1 day';
END;
$$ LANGUAGE plpgsql;