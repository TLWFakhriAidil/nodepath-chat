import psycopg2

# Connect to PostgreSQL
conn = psycopg2.connect(
    "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
)
cur = conn.cursor()

print("=== SEQUENCE FIXES ===\n")

# 1. Check current sequence triggers
print("1. Current sequence triggers:")
cur.execute("SELECT id, name, trigger FROM sequences ORDER BY name")
sequences = cur.fetchall()
for seq in sequences:
    print(f"   {seq[1]}: trigger = '{seq[2]}'")

# 2. Create monitoring view
print("\n2. Creating sequence progress monitoring view...")
try:
    cur.execute("""
        CREATE OR REPLACE VIEW sequence_progress_monitor AS
        SELECT 
            s.name as sequence_name,
            sc.contact_phone,
            sc.current_step,
            sc.status as step_status,
            bm.status as message_status,
            sc.next_trigger_time,
            sc.completed_at,
            bm.sent_at,
            bm.error_message,
            CASE 
                WHEN sc.status = 'completed' THEN 'Step completed'
                WHEN sc.status = 'sent' THEN 'Message sent'
                WHEN sc.status = 'failed' THEN 'Message failed'
                WHEN sc.status = 'active' AND sc.next_trigger_time <= NOW() THEN 'Ready to send'
                WHEN sc.status = 'active' AND sc.next_trigger_time > NOW() THEN 'Scheduled'
                WHEN sc.status = 'pending' THEN 'Waiting'
                ELSE sc.status
            END as current_state
        FROM sequence_contacts sc
        JOIN sequences s ON s.id = sc.sequence_id
        LEFT JOIN broadcast_messages bm ON 
            bm.sequence_id = sc.sequence_id 
            AND bm.recipient_phone = sc.contact_phone
            AND bm.sequence_stepid = sc.sequence_stepid
        ORDER BY s.name, sc.contact_phone, sc.current_step
    """)
    conn.commit()
    print("   Successfully created monitoring view!")
except Exception as e:
    conn.rollback()
    print(f"   Error creating view: {e}")

# 3. Check orphaned broadcast messages
print("\n3. Checking orphaned sequence messages...")
cur.execute("""
    SELECT 
        COUNT(*) as total,
        COUNT(CASE WHEN sequence_stepid IS NULL THEN 1 END) as missing_stepid
    FROM broadcast_messages
    WHERE sequence_id IS NOT NULL
""")
result = cur.fetchone()
print(f"   Total sequence messages: {result[0]}")
print(f"   Messages without stepid: {result[1]}")

# 4. Sample data from monitoring view
print("\n4. Sample from monitoring view (5 records):")
try:
    cur.execute("SELECT * FROM sequence_progress_monitor LIMIT 5")
    results = cur.fetchall()
    if results:
        for row in results:
            print(f"   {row[0]} - {row[1]} Step {row[2]}: {row[9]}")
    else:
        print("   No data in monitoring view")
except Exception as e:
    print(f"   Error querying view: {e}")

# 5. Fix suggestion for empty triggers
print("\n5. RECOMMENDATION: Update sequence triggers")
print("   The sequences have empty triggers. To fix enrollment, run:")
print("   UPDATE sequences SET trigger = 'warm_start' WHERE name = 'WARM Sequence';")
print("   UPDATE sequences SET trigger = 'cold_start' WHERE name = 'COLD Sequence';") 
print("   UPDATE sequences SET trigger = 'hot_start' WHERE name = 'HOT Seqeunce';")
print("\n   Then update your leads to have matching triggers in their 'trigger' field")

cur.close()
conn.close()
print("\n=== FIXES COMPLETE ===")
