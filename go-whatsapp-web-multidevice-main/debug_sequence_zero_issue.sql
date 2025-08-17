-- DEBUG: Why Sequence Summary Shows 0
-- Run these queries to diagnose the issue

-- 1. First, check if there are ANY sequence messages in broadcast_messages
SELECT 
    COUNT(*) as total_sequence_messages,
    COUNT(DISTINCT