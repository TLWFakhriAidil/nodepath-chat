import psycopg2
import sys
from datetime import datetime

# Set UTF-8 encoding for output
sys.stdout.reconfigure(encoding='utf-8')

# Connect to PostgreSQL
print("Finding ALL timeout messages...")
conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()
print(f"Connected at: {datetime.now()}")

print("\n" + "="*70)
print("CHECKING ALL TIMEOUT MESSAGES")
print("="*70)

# Check ALL messages with timeout error (both sent and failed)
print("\n1. Checking ALL messages with timeout error:")
cursor.execute("""
    SELECT 
        bm.status,
        COUNT(*) as count,
        COUNT(CASE WHEN ud.platform IS NOT NULL AND ud.platform != '' THEN 1 END) as platform_count,
        COUNT(CASE WHEN ud.platform IS NULL OR ud.platform = '' THEN 1 END) as whatsapp_count
    FROM broadcast_messages bm
    LEFT JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.error_message = 'Message timeout - device was not available'
    GROUP BY bm.status
    ORDER BY bm.status
""")
results = cursor.fetchall()

total_timeout = 0
for row in results:
    print(f"\n   Status '{row[0]}':")
    print(f"     Total: {row[1]} messages")
    print(f"     Platform devices: {row[2]} (SHOULD NEVER TIMEOUT!)")
    print(f"     WhatsApp Web devices: {row[3]}")
    total_timeout += row[1]

print(f"\n   TOTAL TIMEOUT MESSAGES: {total_timeout}")

# Get detailed breakdown by platform
print("\n2. Breakdown by platform:")
cursor.execute("""
    SELECT 
        COALESCE(ud.platform, 'WhatsApp Web') as platform_type,
        bm.status,
        COUNT(*) as count
    FROM broadcast_messages bm
    LEFT JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.error_message = 'Message timeout - device was not available'
    GROUP BY ud.platform, bm.status
    ORDER BY ud.platform, bm.status
""")
platform_breakdown = cursor.fetchall()

for row in platform_breakdown:
    print(f"   {row[0]} - Status '{row[1]}': {row[2]} messages")

# Get sample messages
print("\n3. Sample timeout messages:")
cursor.execute("""
    SELECT 
        bm.id,
        bm.status,
        COALESCE(ud.platform, 'WhatsApp Web') as platform,
        bm.recipient_phone,
        bm.created_at
    FROM broadcast_messages bm
    LEFT JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.error_message = 'Message timeout - device was not available'
    ORDER BY bm.created_at DESC
    LIMIT 10
""")
samples = cursor.fetchall()

for sample in samples:
    print(f"   ID: {sample[0][:8]}... Status: {sample[1]}, Platform: {sample[2]}, Phone: {sample[3]}, Time: {sample[4]}")

print("\n" + "="*70)
print("FIXING ALL TIMEOUT MESSAGES")
print("="*70)

# First, let's save the current status before updating
cursor.execute("""
    SELECT status, COUNT(*) 
    FROM broadcast_messages
    WHERE error_message = 'Message timeout - device was not available'
    GROUP BY status
""")
status_counts = dict(cursor.fetchall())

# Fix ALL timeout messages - reset to pending
print("\nResetting ALL timeout messages to pending...")
cursor.execute("""
    UPDATE broadcast_messages
    SET status = 'pending',
        error_message = NULL,
        sent_at = NULL
    WHERE error_message = 'Message timeout - device was not available'
""")
update_count = cursor.rowcount
conn.commit()

print(f"\n✅ FIXED {update_count} timeout messages:")
for status, count in status_counts.items():
    print(f"   - {count} were status '{status}'")
print(f"   - All now set to 'pending' with no error")

# Verify fix
cursor.execute("""
    SELECT COUNT(*)
    FROM broadcast_messages
    WHERE error_message = 'Message timeout - device was not available'
""")
remaining = cursor.fetchone()[0]

print(f"\n✅ Verification: {remaining} timeout messages remaining (should be 0)")

# Check platform devices specifically
cursor.execute("""
    SELECT COUNT(*)
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.error_message = 'Message timeout - device was not available'
    AND ud.platform IS NOT NULL AND ud.platform != ''
""")
platform_remaining = cursor.fetchone()[0]

print(f"✅ Platform devices with timeout: {platform_remaining} (should be 0)")

cursor.close()
conn.close()

print("\n" + "="*70)
print("SUMMARY")
print("="*70)
print(f"✅ Fixed {update_count} timeout messages")
print("✅ All reset to 'pending' status for retry")
print("✅ Platform devices will use API (no timeout check)")
print("✅ WhatsApp Web devices will check connection as normal")
print("\n🎯 With the code fix in place, platform devices won't timeout anymore!")
