-- PostgreSQL Cleanup Commands
-- Generated on: 2025-07-30 12:35:45.929013
-- Run these commands in your PostgreSQL database

-- Check sizes
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size
FROM pg_tables WHERE tablename IN ('leads', 'leads_ai', 'sequences', 'sequence_contacts', 'broadcast_messages', 'campaigns') ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Clear tables
TRUNCATE TABLE leads CASCADE;
TRUNCATE TABLE leads_ai CASCADE;
TRUNCATE TABLE sequences CASCADE;
TRUNCATE TABLE sequence_contacts CASCADE;
TRUNCATE TABLE broadcast_messages CASCADE;
TRUNCATE TABLE campaigns CASCADE;

-- Reclaim space
VACUUM FULL;
