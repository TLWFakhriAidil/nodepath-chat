import psycopg2

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

print("=== INVESTIGATING NICHE VS TRIGGER ENROLLMENT ===")

# Check how many leads have VITAC niche without triggers
cursor.execute("""
    SELECT COUNT(*) 
    FROM leads 
    WHERE niche = 'VITAC' 
    AND (trigger IS NULL OR trigger = '')
""")
vitac_no_trigger = cursor.fetchone()[0]

cursor.execute("""
    SELECT COUNT(*) 
    FROM leads 
    WHERE niche = 'VITAC' 
    AND trigger IS NOT NULL AND trigger != ''
""")
vitac_with_trigger = cursor.fetchone()[0]

print(f"\nVITAC Leads:")
print(f"- Without trigger: {vitac_no_trigger}")
print(f"- With trigger: {vitac_with_trigger}")

# Check recent enrollment patterns
print("\n=== RECENT ENROLLMENTS ===")
cursor.execute("""
    SELECT 
        l.phone,
        l.niche,
        l.trigger,
        s.name as sequence_name,
        MIN(bm.created_at) as enrolled_at
    FROM leads l
    JOIN broadcast_messages bm ON bm.recipient_phone = l.phone
    JOIN sequences s ON bm.sequence_id = s.id
    WHERE l.niche = 'VITAC'
    GROUP BY l.phone, l.niche, l.trigger, s.name
    ORDER BY MIN(bm.created_at) DESC
    LIMIT 10
""")

enrollments = cursor.fetchall()
for e in enrollments:
    print(f"\nPhone: {e[0]}")
    print(f"  Niche: {e[1]}, Trigger: {e[2]}")
    print(f"  Enrolled in: {e[3]}")
    print(f"  At: {e[4]}")

# Check if there's a different enrollment mechanism
print("\n=== CHECKING ENROLLMENT LOGIC ===")

# Check sequence_contacts table (if it exists)
cursor.execute("""
    SELECT COUNT(*) 
    FROM information_schema.tables 
    WHERE table_name = 'sequence_contacts'
""")
if cursor.fetchone()[0] > 0:
    cursor.execute("""
        SELECT COUNT(*) FROM sequence_contacts
    """)
    seq_contacts = cursor.fetchone()[0]
    print(f"sequence_contacts table exists with {seq_contacts} records")
else:
    print("sequence_contacts table does not exist (using direct broadcast)")

conn.close()
