import psycopg2
import sys

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

print("=== INVESTIGATING SEQUENCE TRIGGER ISSUE ===")
print("Phone: 601119667332")
print()

# 1. Check the lead
cursor.execute("""
    SELECT id, name, phone, trigger, device_id, user_id, niche
    FROM leads
    WHERE phone = '601119667332'
""")
lead = cursor.fetchone()

if lead:
    print(f"Lead found:")
    print(f"  - ID: {lead[0]}")
    print(f"  - Name: {lead[1]}")
    print(f"  - Phone: {lead[2]}")
    print(f"  - Trigger: {lead[3]}")
    print(f"  - Device ID: {lead[4]}")
    print(f"  - User ID: {lead[5]}")
    print(f"  - Niche: {lead[6]}")
else:
    print("Lead not found!")
    exit()

# 2. Check broadcast messages
print("\n=== BROADCAST MESSAGES ===")
cursor.execute("""
    SELECT 
        bm.id, 
        bm.sequence_id, 
        bm.sequence_stepid,
        bm.content, 
        bm.status,
        bm.scheduled_at,
        s.name as sequence_name,
        ss.trigger as step_trigger
    FROM broadcast_messages bm
    LEFT JOIN sequences s ON bm.sequence_id = s.id
    LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
    WHERE bm.recipient_phone = '601119667332'
    ORDER BY bm.created_at DESC
    LIMIT 10
""")
messages = cursor.fetchall()

print(f"Found {len(messages)} messages:")
for msg in messages:
    print(f"\n- Message ID: {msg[0]}")
    print(f"  Sequence: {msg[6]}")
    print(f"  Step Trigger: {msg[7]}")
    print(f"  Status: {msg[4]}")
    print(f"  Scheduled: {msg[5]}")
    print(f"  Content: {msg[3][:50]}..." if msg[3] else "No content")

# 3. Check sequence enrollment logic
print("\n=== SEQUENCE ENROLLMENT ANALYSIS ===")

# Check what sequences have the VITAC triggers
cursor.execute("""
    SELECT 
        s.name,
        s.trigger,
        ss.trigger as step_trigger,
        ss.is_entry_point,
        ss.day_number
    FROM sequences s
    JOIN sequence_steps ss ON s.id = ss.sequence_id
    WHERE ss.is_entry_point = true
    AND (ss.trigger LIKE '%VITAC%' OR s.trigger LIKE '%VITAC%')
    ORDER BY s.name
""")
vitac_sequences = cursor.fetchall()

print("VITAC Sequences:")
for seq in vitac_sequences:
    print(f"- {seq[0]}: Sequence Trigger={seq[1]}, Step Trigger={seq[2]}")

# Check what sequences have EXSTART triggers
cursor.execute("""
    SELECT 
        s.name,
        s.trigger,
        ss.trigger as step_trigger,
        ss.is_entry_point,
        ss.day_number
    FROM sequences s
    JOIN sequence_steps ss ON s.id = ss.sequence_id
    WHERE ss.is_entry_point = true
    AND (ss.trigger LIKE '%EXSTART%' OR s.trigger LIKE '%EXSTART%')
    ORDER BY s.name
""")
exstart_sequences = cursor.fetchall()

print("\nEXSTART Sequences:")
for seq in exstart_sequences:
    print(f"- {seq[0]}: Sequence Trigger={seq[1]}, Step Trigger={seq[2]}")

# 4. Check for duplicate or overlapping triggers
print("\n=== CHECKING FOR TRIGGER CONFLICTS ===")
cursor.execute("""
    SELECT 
        ss.trigger,
        COUNT(*) as count,
        STRING_AGG(s.name, ', ') as sequences
    FROM sequence_steps ss
    JOIN sequences s ON s.id = ss.sequence_id
    WHERE ss.is_entry_point = true
    GROUP BY ss.trigger
    HAVING COUNT(*) > 1
    ORDER BY count DESC
""")
conflicts = cursor.fetchall()

if conflicts:
    print("Found trigger conflicts:")
    for conflict in conflicts:
        print(f"- Trigger '{conflict[0]}' used in {conflict[1]} sequences: {conflict[2]}")
else:
    print("No trigger conflicts found")

conn.close()
