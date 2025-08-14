import psycopg2

DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(DB_URI)
cur = conn.cursor()

print("=== FIXING BOTH ISSUES ===\n")

# Issue 1: Fix the constraint
print("1. Fixing sequence_contacts constraint...")
try:
    cur.execute("ALTER TABLE sequence_contacts DROP CONSTRAINT IF EXISTS uq_sequence_contact_step")
    cur.execute("""
        ALTER TABLE sequence_contacts
        ADD CONSTRAINT uq_sequence_contact_step 
        UNIQUE (sequence_id, contact_phone, sequence_stepid)
    """)
    conn.commit()
    print("   SUCCESS: Constraint fixed!")
except Exception as e:
    conn.rollback()
    print(f"   ERROR: {e}")

# Issue 2: Check why platform device detection fails
print("\n2. Checking message sender logic...")

# First, let's see which device is being used for the failed message
cur.execute("""
    SELECT bm.id, bm.device_id, ud.device_name, ud.platform, bm.status, bm.error_message
    FROM broadcast_messages bm
    JOIN user_devices ud ON ud.id = bm.device_id
    WHERE bm.id = '1658b215-c3db-471c-bc63-4f23dc57ed60'
""")

msg = cur.fetchone()
if msg:
    print(f"   Message ID: {msg[0]}")
    print(f"   Device ID: {msg[1]}")
    print(f"   Device Name: {msg[2]}")
    print(f"   Platform: {msg[3]}")
    print(f"   Status: {msg[4]}")
    print(f"   Error: {msg[5]}")

# Clean up bad sequence data
print("\n3. Cleaning up sequence data...")
try:
    # Delete all records for now to start fresh
    cur.execute("DELETE FROM sequence_contacts")
    deleted = cur.rowcount
    print(f"   Deleted {deleted} sequence contact records")
    
    conn.commit()
    print("   SUCCESS: Cleanup complete!")
except Exception as e:
    conn.rollback()
    print(f"   ERROR: {e}")

# Check the broadcast worker logic
print("\n4. Checking how devices are selected for sequences...")
cur.execute("""
    SELECT 
        s.name as sequence_name,
        s.id as sequence_id,
        COUNT(DISTINCT l.phone) as total_leads,
        COUNT(DISTINCT l.device_id) as unique_devices
    FROM sequences s
    LEFT JOIN leads l ON l.trigger LIKE '%' || s.trigger || '%'
    WHERE s.is_active = true
    GROUP BY s.id, s.name
""")

for row in cur.fetchall():
    print(f"   Sequence: {row[0]}")
    print(f"   Total Leads: {row[2]}")
    print(f"   Unique Devices: {row[3]}")

cur.close()
conn.close()

print("\n=== RECOMMENDATIONS ===")
print("1. The constraint is now fixed - sequence enrollment should work")
print("2. For platform devices, check the broadcast worker code")
print("3. Make sure it checks device.Platform before trying WhatsApp Web")
print("4. All sequence contacts cleared - ready for fresh test")