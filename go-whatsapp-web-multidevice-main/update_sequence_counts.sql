-- Add this to fix sequence summary to show actual contact counts
-- Update the sequence summary logic to:

-- 1. Get total contacts per sequence from sequence_contacts
UPDATE sequences s
SET total_contacts = (
    SELECT COUNT(*) 
    FROM sequence_contacts sc 
    WHERE sc.sequence_id = s.id
);

-- 2. Get active contacts per sequence
UPDATE sequences s
SET active_contacts = (
    SELECT COUNT(*) 
    FROM sequence_contacts sc 
    WHERE sc.sequence_id = s.id 
    AND sc.status = 'active'
);

-- 3. Get completed contacts per sequence
UPDATE sequences s
SET completed_contacts = (
    SELECT COUNT(*) 
    FROM sequence_contacts sc 
    WHERE sc.sequence_id = s.id 
    AND sc.status = 'completed'
);

-- 4. Get failed contacts per sequence
UPDATE sequences s
SET failed_contacts = (
    SELECT COUNT(*) 
    FROM sequence_contacts sc 
    WHERE sc.sequence_id = s.id 
    AND sc.status = 'failed'
);

-- 5. Update progress percentage
UPDATE sequences s
SET progress_percentage = CASE 
    WHEN total_contacts > 0 THEN 
        (completed_contacts * 100.0 / total_contacts)
    ELSE 0 
END;
