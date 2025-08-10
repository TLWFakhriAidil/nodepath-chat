import psycopg2
import pandas as pd
from datetime import datetime

# Connect to PostgreSQL
conn = psycopg2.connect(
    "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
)
cur = conn.cursor()

print("=== SEQUENCE DATA ANALYSIS ===\n")

# 1. Check sequence progress overview
print("1. SEQUENCE PROGRESS OVERVIEW:")
cur.execute("""
    SELECT 
        s.name as sequence_name,
        s.trigger,
        COUNT(DISTINCT sc.contact_phone) as total_contacts,
        COUNT(DISTINCT CASE WHEN sc.status = 'active' THEN sc.contact_phone END) as active_contacts,
        COUNT(DISTINCT CASE WHEN sc.status = 'pending' THEN sc.contact_phone END) as pending_contacts,
        COUNT(DISTINCT CASE WHEN sc.status = 'completed' THEN sc.contact_phone END) as completed_contacts,
        COUNT(DISTINCT CASE WHEN sc.status = 'failed' THEN sc.contact_phone END) as failed_contacts,
        COUNT(*) as total_steps
    FROM sequence_contacts sc
    JOIN sequences s ON s.id = sc.sequence_id
    GROUP BY s.id, s.name, s.trigger
""")
results = cur.fetchall()
if results:
    for row in results:
        print(f"\nSequence: {row[0]} (Trigger: {row[1]})")
        print(f"  Total contacts enrolled: {row[2]}")
        print(f"  Active: {row[3]}, Pending: {row[4]}, Completed: {row[5]}, Failed: {row[6]}")
        print(f"  Total steps across all contacts: {row[7]}")
else:
    print("  No sequence contacts found!")

# 2. Check leads that SHOULD be in sequences
print("\n\n2. LEADS THAT SHOULD BE ENROLLED:")
cur.execute("""
    SELECT 
        COUNT(*) as total_leads,
        COUNT(CASE WHEN l.trigger IS NOT NULL AND l.trigger != '' THEN 1 END) as leads_with_triggers
    FROM leads l
""")
lead_counts = cur.fetchone()
print(f"  Total leads: {lead_counts[0]}")
print(f"  Leads with triggers: {lead_counts[1]}")

# Check which triggers match sequences
cur.execute("""
    SELECT 
        s.name as sequence_name,
        s.trigger as sequence_trigger,
        COUNT(DISTINCT l.phone) as matching_leads
    FROM sequences s
    CROSS JOIN leads l
    WHERE l.trigger IS NOT NULL 
    AND l.trigger != ''
    AND position(s.trigger in l.trigger) > 0
    GROUP BY s.name, s.trigger
""")
trigger_matches = cur.fetchall()
print("\n  Trigger matches:")
for row in trigger_matches:
    print(f"    {row[0]} (trigger: {row[1]}): {row[2]} matching leads")

# 3. Check broadcast messages for sequences
print("\n\n3. BROADCAST MESSAGES FOR SEQUENCES:")
cur.execute("""
    SELECT 
        s.name as sequence_name,
        COUNT(*) as total_messages,
        COUNT(CASE WHEN bm.status = 'pending' THEN 1 END) as pending,
        COUNT(CASE WHEN bm.status = 'sent' THEN 1 END) as sent,
        COUNT(CASE WHEN bm.status = 'failed' THEN 1 END) as failed
    FROM broadcast_messages bm
    JOIN sequences s ON s.id = bm.sequence_id
    WHERE bm.sequence_id IS NOT NULL
    GROUP BY s.name
""")
broadcast_results = cur.fetchall()
if broadcast_results:
    for row in broadcast_results:
        print(f"\n  {row[0]}:")
        print(f"    Total: {row[1]}, Pending: {row[2]}, Sent: {row[3]}, Failed: {row[4]}")
else:
    print("  No broadcast messages found for sequences!")

# 4. Check recent sequence activity
print("\n\n4. RECENT SEQUENCE ACTIVITY (Last 10):")
cur.execute("""
    SELECT 
        sc.contact_phone,
        s.name as sequence_name,
        sc.current_step,
        sc.status,
        sc.next_trigger_time,
        sc.completed_at,
        CASE 
            WHEN sc.next_trigger_time > NOW() THEN 'Scheduled'
            WHEN sc.next_trigger_time <= NOW() AND sc.status = 'active' THEN 'Ready to send'
            WHEN sc.status = 'completed' THEN 'Done'
            ELSE sc.status
        END as current_state
    FROM sequence_contacts sc
    JOIN sequences s ON s.id = sc.sequence_id
    ORDER BY COALESCE(sc.completed_at, sc.next_trigger_time) DESC
    LIMIT 10
""")
recent_activity = cur.fetchall()
if recent_activity:
    for row in recent_activity:
        print(f"\n  {row[0]} - {row[1]} Step {row[2]}")
        print(f"    Status: {row[3]}, State: {row[6]}")
        print(f"    Next trigger: {row[4]}")
        if row[5]:
            print(f"    Completed at: {row[5]}")
else:
    print("  No recent activity found!")

# 5. Check for issues
print("\n\n5. POTENTIAL ISSUES:")

# Check for stuck active contacts
cur.execute("""
    SELECT COUNT(*) 
    FROM sequence_contacts 
    WHERE status = 'active' 
    AND next_trigger_time < NOW() - INTERVAL '1 hour'
""")
stuck_count = cur.fetchone()[0]
if stuck_count > 0:
    print(f"  ⚠️ {stuck_count} contacts stuck in active state (trigger time passed > 1 hour ago)")

# Check for orphaned broadcast messages
cur.execute("""
    SELECT COUNT(*)
    FROM broadcast_messages bm
    WHERE bm.sequence_id IS NOT NULL
    AND bm.status = 'pending'
    AND bm.created_at < NOW() - INTERVAL '1 hour'
""")
orphaned_count = cur.fetchone()[0]
if orphaned_count > 0:
    print(f"  ⚠️ {orphaned_count} broadcast messages stuck in pending state")

# Check sequence steps without contacts
cur.execute("""
    SELECT 
        s.name,
        COUNT(DISTINCT ss.id) as total_steps,
        COUNT(DISTINCT sc.sequence_stepid) as used_steps
    FROM sequences s
    JOIN sequence_steps ss ON ss.sequence_id = s.id
    LEFT JOIN sequence_contacts sc ON sc.sequence_stepid = ss.id
    GROUP BY s.name
    HAVING COUNT(DISTINCT ss.id) > COUNT(DISTINCT sc.sequence_stepid)
""")
unused_steps = cur.fetchall()
if unused_steps:
    print("\n  Sequences with unused steps:")
    for row in unused_steps:
        print(f"    {row[0]}: {row[1]} total steps, {row[2]} used")

print("\n\n=== ANALYSIS COMPLETE ===")

# Close connection
cur.close()
conn.close()
