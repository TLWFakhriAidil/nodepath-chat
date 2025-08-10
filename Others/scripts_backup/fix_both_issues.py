import psycopg2

DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(DB_URI)
cur = conn.cursor()

print("=== CHECKING DEVICE PLATFORM ISSUE ===\n")

# 1. Check device with the error ID
device_id = "5f841a19-4c4d-47d8-832c-371027721cea"
print(f"1. Checking device {device_id}:")
cur.execute("""
    SELECT id, device_name, platform, status, jid
    FROM user_devices
    WHERE id = %s
""", (device_id,))

device = cur.fetchone()
if device:
    print(f"   Name: {device[1]}")
    print(f"   Platform: {device[2]}")
    print(f"   Status: {device[3]}")
    print(f"   JID: {device[4]}")
else:
    print("   Device not found!")

# 2. Check all platform devices
print("\n2. All devices with platform set:")
cur.execute("""
    SELECT id, device_name, platform, status
    FROM user_devices
    WHERE platform IS NOT NULL AND platform != ''
    ORDER BY platform
""")

platform_devices = cur.fetchall()
if platform_devices:
    for d in platform_devices:
        print(f"   {d[1]}: platform={d[2]}, status={d[3]}")
else:
    print("   No platform devices found")

# 3. Apply the constraint fix
print("\n3. Fixing sequence_contacts constraint...")
try:
    # Drop old constraint
    cur.execute("ALTER TABLE sequence_contacts DROP CONSTRAINT IF EXISTS uq_sequence_contact_step")
    
    # Add new constraint
    cur.execute("""
        ALTER TABLE sequence_contacts
        ADD CONSTRAINT uq_sequence_contact_step 
        UNIQUE (sequence_id, contact_phone, sequence_stepid)
    """)
    
    conn.commit()
    print("   ✓ Constraint fixed successfully!")
except Exception as e:
    conn.rollback()
    print(f"   ✗ Error: {e}")

# 4. Clean up the duplicate/bad data
print("\n4. Cleaning up sequence data...")
try:
    # Delete completed records
    cur.execute("DELETE FROM sequence_contacts WHERE status = 'completed'")
    deleted = cur.rowcount
    print(f"   Deleted {deleted} completed records")
    
    # Delete records where step 1 is missing but other steps exist
    cur.execute("""
        DELETE FROM sequence_contacts sc
        WHERE NOT EXISTS (
            SELECT 1 FROM sequence_contacts sc2
            WHERE sc2.sequence_id = sc.sequence_id
            AND sc2.contact_phone = sc.contact_phone
            AND sc2.current_step = 1
        )
    """)
    deleted = cur.rowcount
    print(f"   Deleted {deleted} orphaned records (no step 1)")
    
    conn.commit()
    print("   ✓ Cleanup complete!")
except Exception as e:
    conn.rollback()
    print(f"   ✗ Error: {e}")

cur.close()
conn.close()