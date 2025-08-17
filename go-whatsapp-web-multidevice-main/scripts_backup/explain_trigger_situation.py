import psycopg2
import sys
import io
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

phone = '60138847465'

print("="*100)
print(f"UNDERSTANDING TRIGGERS VS SEQUENCE ENROLLMENT FOR: {phone}")
print("="*100)

# Get lead info
cursor.execute("""
    SELECT 
        l.id,
        l.name,
        l.phone,
        l.niche,
        l.trigger,
        l.status,
        l.created_at,
        l.updated_at
    FROM leads l
    WHERE l.phone = %s
""", (phone,))

lead = cursor.fetchone()
if lead:
    print("\n1️⃣ LEAD TABLE DATA:")
    print(f"   - Name: {lead[1]}")
    print(f"   - Phone: {lead[2]}")
    print(f"   - Niche: {lead[3]}")
    print(f"   - Trigger: {lead[4] or 'NULL/Empty'} ⚠️")
    print(f"   - Status: {lead[5]}")
    print(f"   - Created: {lead[6]}")
    print(f"   - Updated: {lead[7]}")

# Check sequences and their triggers
print("\n\n2️⃣ SEQUENCE CONFIGURATION:")
cursor.execute("""
    SELECT 
        s.id,
        s.name,
        s.trigger,
        s.is_active
    FROM sequences s
    WHERE s.name LIKE '%VITAC%'
    ORDER BY s.name
""")
sequences = cursor.fetchall()
for seq in sequences:
    print(f"\n   Sequence: {seq[1]}")
    print(f"   - Sequence Trigger Field: {seq[2] or 'NULL/Empty'}")
    print(f"   - Active: {seq[3]}")

# Check sequence steps entry points
print("\n\n3️⃣ SEQUENCE ENTRY POINTS (from sequence_steps table):")
cursor.execute("""
    SELECT 
        s.name as sequence_name,
        ss.trigger as step_trigger,
        ss.is_entry_point,
        ss.day_number
    FROM sequence_steps ss
    JOIN sequences s ON s.id = ss.sequence_id
    WHERE s.name LIKE '%VITAC%'
    AND ss.is_entry_point = true
    ORDER BY s.name
""")
entry_points = cursor.fetchall()
for ep in entry_points:
    print(f"\n   {ep[0]}:")
    print(f"   - Entry Point Trigger: {ep[1]}")
    print(f"   - Day: {ep[3]}")

# Explain the situation
print("\n\n4️⃣ EXPLANATION:")
print("-"*100)
print("   🔍 The lead has NO TRIGGER in the leads table (trigger = NULL)")
print("   🔍 BUT the lead still has messages because:")
print("      1. Messages were created directly in broadcast_messages table")
print("      2. The sequence enrollment happened through some other process")
print("      3. Possibly enrolled before the trigger system was fully implemented")
print("\n   📌 NORMAL FLOW:")
print("      Lead has trigger (e.g., 'COLDVITAC') → System finds matching sequence → Creates messages")
print("\n   📌 THIS LEAD'S SITUATION:")
print("      Lead has NO trigger → But messages already exist in broadcast_messages")

# Check when messages were created
print("\n\n5️⃣ WHEN WERE MESSAGES CREATED?")
cursor.execute("""
    SELECT 
        MIN(created_at) as first_message_created,
        MAX(created_at) as last_message_created,
        COUNT(DISTINCT DATE(created_at)) as different_creation_days
    FROM broadcast_messages
    WHERE recipient_phone = %s
""", (phone,))
creation_info = cursor.fetchone()
if creation_info:
    print(f"   - First message created: {creation_info[0]}")
    print(f"   - Last message created: {creation_info[1]}")
    print(f"   - Created across {creation_info[2]} different days")

conn.close()
print("\n" + "="*100)
