import psycopg2
import re

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

print("=== ANALYZING SEQUENCE TRIGGER ISSUES ===\n")

# Check the specific lead
print("1. Checking lead 601119667332:")
cursor.execute("""
    SELECT l.phone, l.trigger, l.niche, bm.sequence_id, s.name as sequence_name, 
           bm.sequence_stepid, ss.trigger as step_trigger, bm.content
    FROM leads l
    JOIN broadcast_messages bm ON l.phone = bm.recipient_phone
    LEFT JOIN sequences s ON bm.sequence_id = s.id
    LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
    WHERE l.phone = '601119667332'
    ORDER BY bm.created_at DESC
    LIMIT 5
""")
results = cursor.fetchall()
for r in results:
    print(f"  Lead Trigger: {r[1]}, Sequence: {r[4]}, Step Trigger: {r[6]}")

# Check VITAC sequences
print("\n2. VITAC Sequences and Entry Points:")
cursor.execute("""
    SELECT s.id, s.name, s.trigger, ss.id as step_id, ss.trigger as step_trigger, 
           ss.is_entry_point, ss.day_number
    FROM sequences s
    JOIN sequence_steps ss ON s.id = ss.sequence_id
    WHERE s.name LIKE '%VITAC%' AND ss.is_entry_point = true
    ORDER BY s.name
""")
vitac_sequences = cursor.fetchall()
for v in vitac_sequences:
    print(f"  {v[1]}: Sequence Trigger={v[2]}, Entry Step Trigger={v[4]}")

# Check for wrong sequence assignments
print("\n3. Checking for VITAC leads using EXSTART sequences:")
cursor.execute("""
    SELECT COUNT(*), l.trigger, s.name
    FROM leads l
    JOIN broadcast_messages bm ON l.phone = bm.recipient_phone
    JOIN sequences s ON bm.sequence_id = s.id
    WHERE l.trigger LIKE '%VITAC%' AND s.name LIKE '%EXSTART%'
    GROUP BY l.trigger, s.name
""")
wrong_assignments = cursor.fetchall()
if wrong_assignments:
    print("  FOUND WRONG ASSIGNMENTS:")
    for w in wrong_assignments:
        print(f"    {w[0]} leads with trigger '{w[1]}' using sequence '{w[2]}'")

# Check sequence enrollment logic
print("\n4. Analyzing sequence enrollment patterns:")
cursor.execute("""
    SELECT DISTINCT l.trigger, s.name, COUNT(*) as count
    FROM leads l
    JOIN broadcast_messages bm ON l.phone = bm.recipient_phone
    JOIN sequences s ON bm.sequence_id = s.id
    WHERE l.trigger IN ('COLDVITAC', 'WARMVITAC', 'HOTVITAC', 'COLDEXSTART', 'WARMEXSTART', 'HOTEXSTART')
    GROUP BY l.trigger, s.name
    ORDER BY l.trigger, count DESC
""")
patterns = cursor.fetchall()
for p in patterns:
    print(f"  Trigger '{p[0]}' → Sequence '{p[1]}' ({p[2]} messages)")

conn.close()
print("\nAnalysis complete!")
