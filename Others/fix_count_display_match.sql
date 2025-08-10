-- FIX COUNT MISMATCH: Ensure count matches displayed leads

-- Create a view that shows the EXACT same data for both count and display
DROP VIEW IF EXISTS v_sequence_pending_leads;

CREATE VIEW v_sequence_pending_leads AS
SELECT DISTINCT
    bm.device_id,
    bm.sequence_stepid,
    bm.recipient_phone,
    bm.recipient_name,
    bm.status,
    bm.scheduled_at,
    bm.created_at,
    ud.device_name,
    ss.day,
    ss.message_type,
    s.name as sequence_name
FROM broadcast_messages bm
JOIN user_devices ud ON ud.id = bm.device_id
LEFT JOIN sequence_steps ss ON ss.id = bm.sequence_stepid  
LEFT JOIN sequences s ON s.id = ss.sequence_id
WHERE bm.status = 'pending';

-- Example queries that MUST be used together:

-- For COUNT (shows "4 remaining"):
-- SELECT COUNT(DISTINCT recipient_phone) as remaining_count
-- FROM v_sequence_pending_leads
-- WHERE device_name = 'SCAS-S74' 
-- AND sequence_stepid = 'specific_step_id';

-- For DISPLAY (must show exactly 4 leads):
-- SELECT DISTINCT recipient_phone, recipient_name
-- FROM v_sequence_pending_leads
-- WHERE device_name = 'SCAS-S74'
-- AND sequence_stepid = 'specific_step_id'
-- ORDER BY recipient_phone;

-- The key is using the SAME view and SAME conditions for both!
