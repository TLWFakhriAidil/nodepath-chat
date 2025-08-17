import psycopg2
import sys
from datetime import datetime

# Set UTF-8 encoding for output
sys.stdout.reconfigure(encoding='utf-8')

# Connect to PostgreSQL
print("Analyzing platform timeout timeline...")
conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()

# Check when our fix was applied vs when these timeouts occurred
print("\n=== Platform Timeout Timeline Analysis ===")

# Get the time range of platform timeout messages
cursor.execute("""
    SELECT 
        MIN(bm.created_at) as first_timeout,
        MAX(bm.created_at) as last_timeout,
        COUNT(*) as total_messages,
        COUNT(DISTINCT DATE_TRUNC('hour', bm.created_at)) as hours_affected
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform = 'Wablas'
""")
timeline = cursor.fetchone()

print(f"\nTimeout messages time range:")
print(f"  First timeout: {timeline[0]}")
print(f"  Last timeout:  {timeline[1]}")
print(f"  Total messages: {timeline[2]}")
print(f"  Hours affected: {timeline[3]}")

# Check if new timeouts occurred after our fix (around 13:00 today)
fix_time = datetime(2025, 7, 27, 13, 0, 0)  # Approximate time of our fix
cursor.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform = 'Wablas'
    AND bm.created_at > %s
""", (fix_time,))
new_timeouts = cursor.fetchone()[0]

print(f"\nTimeouts after our fix (after {fix_time}): {new_timeouts}")

if new_timeouts == 0:
    print("✅ No new timeouts since the fix was applied!")
    print("These are OLD timeouts from before the fix (around 4:00-5:00 AM)")
else:
    print("⚠️  New timeouts detected after the fix!")

# Since these are old messages, let's reset them to pending
print("\n=== Resetting OLD Platform Timeout Messages ===")
cursor.execute("""
    UPDATE broadcast_messages bm
    SET status = 'pending', 
        error_message = NULL,
        sent_at = NULL
    FROM user_devices ud
    WHERE bm.device_id = ud.id
    AND bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform = 'Wablas'
    RETURNING bm.id
""")
reset_ids = cursor.fetchall()
reset_count = len(reset_ids)
conn.commit()

print(f"Reset {reset_count} old platform timeout messages to 'pending' status")
print("These messages will now be retried with the fixed code")

# Verify the reset
cursor.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform IS NOT NULL AND ud.platform != ''
""")
remaining = cursor.fetchone()[0]

print(f"\nRemaining platform timeouts: {remaining}")

cursor.close()
conn.close()
print("\n✅ Analysis and cleanup completed!")
