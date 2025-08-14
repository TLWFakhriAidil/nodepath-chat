-- Fix for GetSequenceSummary PostgreSQL to MySQL conversion
-- This fixes the UUID casting issue in the sequence summary query

-- Original PostgreSQL query (lines 1871-1876):
-- SELECT COUNT(*) 
-- FROM sequence_steps ss
-- INNER JOIN sequences s ON s.id = ss.sequence_id
-- WHERE s.user_id = $1::uuid

-- MySQL version (no UUID casting needed):
-- SELECT COUNT(*) 
-- FROM sequence_steps ss
-- INNER JOIN sequences s ON s.id = ss.sequence_id
-- WHERE s.user_id = ?
