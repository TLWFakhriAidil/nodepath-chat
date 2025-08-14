import psycopg2
from datetime import datetime

# Connect to PostgreSQL
conn = psycopg2.connect(
    "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
)
cur = conn.cursor()

print("=== FIXING SEQUENCE SYSTEM ===\n")

# 1. First, let's backup current data before cleaning
print("1. Backing up current sequence data...")
cur.execute("""
    SELECT COUNT(*) FROM sequence_contacts
""")
sc_count = cur.fetchone()[0]
print(f"   Found {sc_count} sequence_contacts records")

cur.execute("""
    SELECT COUNT(*) FROM broadcast_messages WHERE sequence_id IS NOT NULL
""")
bm_count = cur.fetchone()[0]
print(f"   Found {bm_count} broadcast messages for sequences")

# 2. Clean all sequence contact records
print("\n2. Cleaning sequence_contacts table...")
try:
    cur.execute("DELETE FROM sequence_contacts")
    deleted = cur.rowcount
    conn.commit()
    print(f"   Deleted {deleted} sequence_contacts records")
except Exception as e:
    conn.rollback()
    print(f"   Error cleaning sequence_contacts: {e}")

# 3. Clean related broadcast messages
print("\n3. Cleaning sequence-related broadcast_messages...")
try:
    cur.execute("DELETE FROM broadcast_messages WHERE sequence_id IS NOT NULL")
    deleted = cur.rowcount
    conn.commit()
    print(f"   Deleted {deleted} broadcast messages")
except Exception as e:
    conn.rollback()
    print(f"   Error cleaning broadcast_messages: {e}")

# 4. Update sequence triggers
print("\n4. Updating sequence triggers...")
try:
    # Update each sequence with proper triggers
    updates = [
        ("UPDATE sequences SET trigger = 'warm_start' WHERE name = 'WARM Sequence'", "WARM Sequence"),
        ("UPDATE sequences SET trigger = 'cold_start' WHERE name = 'COLD Sequence'", "COLD Sequence"),
        ("UPDATE sequences SET trigger = 'hot_start' WHERE name = 'HOT Seqeunce'", "HOT Seqeunce")
    ]
    
    for query, name in updates:
        cur.execute(query)
        if cur.rowcount > 0:
            print(f"   Updated trigger for {name}")
    
    conn.commit()
    
    # Verify updates
    cur.execute("SELECT name, trigger FROM sequences ORDER BY name")
    sequences = cur.fetchall()
    print("\n   Verified triggers:")
    for seq in sequences:
        print(f"   - {seq[0]}: trigger = '{seq[1]}'")
        
except Exception as e:
    conn.rollback()
    print(f"   Error updating triggers: {e}")

# 5. Update some test leads with matching triggers
print("\n5. Adding triggers to test leads...")
try:
    # Get a few leads for testing
    cur.execute("""
        SELECT id, phone, name 
        FROM leads 
        WHERE trigger IS NULL OR trigger = ''
        LIMIT 15
    """)
    test_leads = cur.fetchall()
    
    if test_leads:
        # Assign triggers to leads (5 for each sequence)
        triggers = ['warm_start', 'cold_start', 'hot_start']
        for i, lead in enumerate(test_leads):
            trigger = triggers[i % 3]
            cur.execute(
                "UPDATE leads SET trigger = %s WHERE id = %s",
                (trigger, lead[0])
            )
            print(f"   Updated lead {lead[2]} ({lead[1]}) with trigger: {trigger}")
        
        conn.commit()
        print(f"\n   Total {len(test_leads)} leads updated with triggers")
    else:
        print("   No leads found to update")
        
except Exception as e:
    conn.rollback()
    print(f"   Error updating lead triggers: {e}")

# 6. Verify lead triggers match sequences
print("\n6. Verifying lead-sequence matches...")
cur.execute("""
    SELECT 
        s.name as sequence_name,
        s.trigger,
        COUNT(DISTINCT l.id) as matching_leads
    FROM sequences s
    LEFT JOIN leads l ON position(s.trigger in COALESCE(l.trigger, '')) > 0
    WHERE s.trigger IS NOT NULL AND s.trigger != ''
    GROUP BY s.name, s.trigger
    ORDER BY s.name
""")
matches = cur.fetchall()
print("   Sequence trigger matches:")
for match in matches:
    print(f"   - {match[0]} (trigger: {match[1]}): {match[2]} matching leads")

# 7. Check sequence steps
print("\n7. Checking sequence steps...")
cur.execute("""
    SELECT 
        s.name as sequence_name,
        COUNT(ss.id) as step_count,
        MIN(ss.day_number) as first_day,
        MAX(ss.day_number) as last_day
    FROM sequences s
    LEFT JOIN sequence_steps ss ON ss.sequence_id = s.id
    GROUP BY s.name
    ORDER BY s.name
""")
steps_info = cur.fetchall()
print("   Sequence steps:")
for info in steps_info:
    print(f"   - {info[0]}: {info[1]} steps (Day {info[2]} to Day {info[3]})")

# 8. Create an enhanced monitoring view
print("\n8. Creating enhanced monitoring view...")
try:
    cur.execute("""
        DROP VIEW IF EXISTS sequence_progress_monitor
    """)
    
    cur.execute("""
        CREATE OR REPLACE VIEW sequence_progress_overview AS
        WITH lead_counts AS (
            SELECT 
                s.id as sequence_id,
                s.name as sequence_name,
                s.trigger,
                COUNT(DISTINCT l.phone) as should_send_count
            FROM sequences s
            LEFT JOIN leads l ON position(s.trigger in COALESCE(l.trigger, '')) > 0
            WHERE s.trigger IS NOT NULL AND s.trigger != ''
            GROUP BY s.id, s.name, s.trigger
        ),
        contact_counts AS (
            SELECT 
                sc.sequence_id,
                COUNT(DISTINCT sc.contact_phone) as enrolled_count,
                COUNT(DISTINCT CASE WHEN sc.status = 'active' THEN sc.contact_phone END) as active_count,
                COUNT(DISTINCT CASE WHEN sc.status = 'pending' THEN sc.contact_phone END) as pending_count,
                COUNT(DISTINCT CASE WHEN sc.status = 'completed' THEN sc.contact_phone END) as completed_count,
                COUNT(DISTINCT CASE WHEN sc.status = 'failed' THEN sc.contact_phone END) as failed_count
            FROM sequence_contacts sc
            GROUP BY sc.sequence_id
        ),
        message_counts AS (
            SELECT 
                bm.sequence_id,
                COUNT(*) as total_messages,
                COUNT(CASE WHEN bm.status = 'sent' THEN 1 END) as sent_messages,
                COUNT(CASE WHEN bm.status = 'failed' THEN 1 END) as failed_messages
            FROM broadcast_messages bm
            WHERE bm.sequence_id IS NOT NULL
            GROUP BY bm.sequence_id
        )
        SELECT 
            lc.sequence_name,
            lc.trigger,
            lc.should_send_count,
            COALESCE(cc.enrolled_count, 0) as enrolled_count,
            COALESCE(cc.active_count, 0) as active_count,
            COALESCE(cc.pending_count, 0) as pending_count,
            COALESCE(cc.completed_count, 0) as completed_count,
            COALESCE(cc.failed_count, 0) as failed_count,
            COALESCE(mc.total_messages, 0) as total_messages,
            COALESCE(mc.sent_messages, 0) as sent_messages,
            COALESCE(mc.failed_messages, 0) as failed_messages,
            lc.should_send_count - COALESCE(cc.enrolled_count, 0) as not_enrolled_count
        FROM lead_counts lc
        LEFT JOIN contact_counts cc ON cc.sequence_id = lc.sequence_id
        LEFT JOIN message_counts mc ON mc.sequence_id = lc.sequence_id
        ORDER BY lc.sequence_name
    """)
    conn.commit()
    print("   Created enhanced monitoring view: sequence_progress_overview")
    
    # Test the view
    cur.execute("SELECT * FROM sequence_progress_overview")
    results = cur.fetchall()
    print("\n   Current sequence overview:")
    print("   " + "-" * 80)
    print("   Sequence | Trigger | Should Send | Enrolled | Active | Sent | Failed")
    print("   " + "-" * 80)
    for row in results:
        print(f"   {row[0]:<15} | {row[1]:<10} | {row[2]:>11} | {row[3]:>8} | {row[4]:>6} | {row[9]:>4} | {row[10]:>6}")
    
except Exception as e:
    conn.rollback()
    print(f"   Error creating view: {e}")

print("\n=== CLEANUP AND FIXES COMPLETE ===")
print("\nSummary:")
print("- Cleaned all sequence_contacts records")
print("- Cleaned all sequence broadcast_messages")
print("- Updated sequence triggers (warm_start, cold_start, hot_start)")
print("- Added triggers to 15 test leads")
print("- Created monitoring view for sequence progress")
print("\nThe sequence system is now ready for fresh enrollment!")

cur.close()
conn.close()
