#!/usr/bin/env python3
"""
PostgreSQL Table Cleanup SQL Generator
Generates SQL commands to clean up PostgreSQL tables
"""

from datetime import datetime

tables_to_clear = [
    'leads',
    'leads_ai', 
    'sequences',
    'sequence_contacts',
    'broadcast_messages',
    'campaigns'
]

print("-- PostgreSQL Cleanup Commands")
print("-- Run these commands in your PostgreSQL database to free up disk space")
print("-- WARNING: This will DELETE ALL DATA from these tables!")
print()

print("-- Step 1: Check table sizes before cleanup")
print("SELECT")
print("    schemaname,")
print("    tablename,")
print("    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size")
print("FROM pg_tables")
print("WHERE tablename IN ({})".format(', '.join(f"'{t}'" for t in tables_to_clear)))
print("ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;")
print()

print("-- Step 2: Count records in each table")
for table in tables_to_clear:
    print(f"SELECT '{table}' as table_name, COUNT(*) as record_count FROM {table};")
print()

print("-- Step 3: Clear the tables (CASCADE will handle foreign key constraints)")
for table in tables_to_clear:
    print(f"TRUNCATE TABLE {table} CASCADE;")
print()

print("-- Step 4: Run VACUUM to reclaim disk space")
print("VACUUM FULL;")
print()

print("-- Step 5: Check database size after cleanup")
print("SELECT pg_size_pretty(pg_database_size(current_database())) as database_size;")
print()

print("-- Step 6: Verify tables are empty")
for table in tables_to_clear:
    print(f"SELECT '{table}' as table_name, COUNT(*) as record_count FROM {table};")
print()

print("-- Alternative: If you want to keep some recent data")
print("-- Example: Delete records older than 30 days")
print()
print("-- For broadcast_messages:")
print("DELETE FROM broadcast_messages WHERE created_at < NOW() - INTERVAL '30 days';")
print()
print("-- For sequences:")
print("DELETE FROM sequences WHERE created_at < NOW() - INTERVAL '30 days';")
print()

# Save to file
with open('postgresql_cleanup_commands.sql', 'w') as f:
    f.write("-- PostgreSQL Cleanup Commands\n")
    f.write("-- Generated on: " + str(datetime.now()) + "\n")
    f.write("-- Run these commands in your PostgreSQL database\n\n")
    
    f.write("-- Check sizes\n")
    f.write("SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size\n")
    f.write("FROM pg_tables WHERE tablename IN ({}) ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;\n\n".format(', '.join(f"'{t}'" for t in tables_to_clear)))
    
    f.write("-- Clear tables\n")
    for table in tables_to_clear:
        f.write(f"TRUNCATE TABLE {table} CASCADE;\n")
    
    f.write("\n-- Reclaim space\n")
    f.write("VACUUM FULL;\n")

print("\nSQL commands saved to: postgresql_cleanup_commands.sql")
