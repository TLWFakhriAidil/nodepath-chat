import psycopg2
import time
import sys

# Try to connect with retries
conn_str = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

for attempt in range(3):
    try:
        print(f"Connection attempt {attempt + 1}...")
        conn = psycopg2.connect(conn_str, connect_timeout=30)
        cursor = conn.cursor()
        print("[OK] Connected!")
        break
    except Exception as e:
        print(f"[FAILED] {str(e)}")
        if attempt < 2:
            print("Retrying in 5 seconds...")
            time.sleep(5)
        else:
            print("All connection attempts failed.")
            sys.exit(1)

print("\n" + "="*80)
print("FAILED SEQUENCE MESSAGES WITH TIMEOUT ERROR")
print("="*80)

# Quick query to get the main info
cursor.execute("""
    SELECT 
        bm.recipient_phone,
        bm.recipient_name,
        bm.content,
        bm.error_message,
        bm.created_at,
        bm.scheduled_at,
        bm.sent_at,
        ud.device_name,
        ud.status as device_status
    FROM broadcast_messages bm
    LEFT JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.sequence_stepid IS NOT NULL 
    AND bm.status = 'failed' 
    AND bm.error_message = 'Message timeout - could not be delivered within 12 hours'
    ORDER BY bm.created_at DESC
    LIMIT 5
""")

results = cursor.fetchall()
print(f"\nFound {len(results)} timeout failures. Here are the details:")

for i, row in enumerate(results, 1):
    print(f"\n--- Failure #{i} ---")
    print(f"Phone: {row[0]}")
    print(f"Name: {row[1]}")
    print(f"Message: {row[2][:80]}..." if row[2] and len(row[2]) > 80 else f"Message: {row[2]}")
    print(f"Error: {row[3]}")
    print(f"Created: {row[4]}")
    print(f"Scheduled: {row[5]}")
    print(f"Failed: {row[6]}")
    print(f"Device: {row[7]} (Status: {row[8]})")
    
    # Calculate timeout duration
    if row[6] and row[5]:
        timeout_hours = (row[6] - row[5]).total_seconds() / 3600
        print(f"Timeout after: {timeout_hours:.1f} hours")

# Summary
cursor.execute("""
    SELECT COUNT(DISTINCT device_id) as devices,
           COUNT(DISTINCT recipient_phone) as phones,
           COUNT(*) as total_timeouts
    FROM broadcast_messages 
    WHERE sequence_stepid IS NOT NULL 
    AND status = 'failed' 
    AND error_message = 'Message timeout - could not be delivered within 12 hours'
""")

summary = cursor.fetchone()
print(f"\n\nSUMMARY:")
print(f"- Total timeout failures: {summary[2]}")
print(f"- Unique phones affected: {summary[1]}")
print(f"- Devices involved: {summary[0]}")

cursor.close()
conn.close()
print("\n[OK] Done!")
