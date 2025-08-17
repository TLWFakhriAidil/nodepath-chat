import psycopg2
import pandas as pd
from datetime import datetime

# Connect with SSL
conn_str = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(conn_str, connect_timeout=20)
cursor = conn.cursor()
print("[OK] Connected to Railway PostgreSQL!")

print(f"\n{'='*80}")
print("FAILED SEQUENCE MESSAGES ANALYSIS")
print(f"{'='*80}")

# First, let's check the column structure
cursor.execute("""
    SELECT column_name, data_type 
    FROM information_schema.columns 
    WHERE table_name = 'broadcast_messages' 
    AND column_name IN ('sequence_stepid', 'status', 'error_message', 'message', 'phone', 'created_at', 'updated_at', 'device_id', 'scheduled_at')
    ORDER BY ordinal_position;
""")
columns = cursor.fetchall()
print("\nAvailable columns:")
for col, dtype in columns:
    print(f"  - {col}: {dtype}")

# Count failed sequence messages
cursor.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages 
    WHERE sequence_stepid IS NOT NULL 
    AND status = 'failed' 
    AND error_message IS NOT NULL
""")
failed_count = cursor.fetchone()[0]
print(f"\nTotal failed sequence messages with errors: {failed_count}")

# Get error message distribution
print("\n" + "="*80)
print("ERROR MESSAGE DISTRIBUTION")
print("="*80)

cursor.execute("""
    SELECT error_message, COUNT(*) as count
    FROM broadcast_messages 
    WHERE sequence_stepid IS NOT NULL 
    AND status = 'failed' 
    AND error_message IS NOT NULL
    GROUP BY error_message
    ORDER BY count DESC
    LIMIT 20
""")

error_dist = cursor.fetchall()
if error_dist:
    print(f"\n{'Error Message':<60} {'Count':>10}")
    print("-" * 72)
    for error, count in error_dist:
        error_msg = error[:57] + "..." if len(error) > 60 else error
        print(f"{error_msg:<60} {count:>10}")

# Get sample failed messages with details
print("\n" + "="*80)
print("SAMPLE FAILED MESSAGES (Latest 10)")
print("="*80)

cursor.execute("""
    SELECT 
        bm.id,
        bm.phone,
        bm.message,
        bm.error_message,
        bm.created_at,
        bm.scheduled_at,
        bm.updated_at,
        bm.device_id,
        ud.device_name,
        ud.status as device_status
    FROM broadcast_messages bm
    LEFT JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.sequence_stepid IS NOT NULL 
    AND bm.status = 'failed' 
    AND bm.error_message IS NOT NULL
    ORDER BY bm.updated_at DESC
    LIMIT 10
""")

failed_messages = cursor.fetchall()

for i, msg in enumerate(failed_messages, 1):
    print(f"\n--- Message {i} ---")
    print(f"ID: {msg[0]}")
    print(f"Phone: {msg[1]}")
    print(f"Message: {msg[2][:100]}..." if len(msg[2]) > 100 else f"Message: {msg[2]}")
    print(f"Error: {msg[3]}")
    print(f"Created: {msg[4]}")
    print(f"Scheduled: {msg[5]}")
    print(f"Failed at: {msg[6]}")
    print(f"Device: {msg[8]} (Status: {msg[9]}) [ID: {msg[7]}]")

# Get device-wise failure distribution
print("\n" + "="*80)
print("DEVICE-WISE FAILURE DISTRIBUTION")
print("="*80)

cursor.execute("""
    SELECT 
        ud.device_name,
        ud.phone as device_phone,
        ud.status as device_status,
        COUNT(bm.id) as failed_messages
    FROM broadcast_messages bm
    LEFT JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.sequence_stepid IS NOT NULL 
    AND bm.status = 'failed' 
    AND bm.error_message IS NOT NULL
    GROUP BY ud.device_name, ud.phone, ud.status
    ORDER BY failed_messages DESC
    LIMIT 15
""")

device_failures = cursor.fetchall()
if device_failures:
    print(f"\n{'Device Name':<30} {'Device Phone':<15} {'Status':<10} {'Failed':>8}")
    print("-" * 65)
    for device in device_failures:
        device_name = device[0][:27] + "..." if device[0] and len(device[0]) > 30 else device[0] or "Unknown"
        print(f"{device_name:<30} {device[1] or 'N/A':<15} {device[2] or 'N/A':<10} {device[3]:>8}")

# Get time-based failure pattern
print("\n" + "="*80)
print("FAILURE TIME PATTERN (Last 7 days)")
print("="*80)

cursor.execute("""
    SELECT 
        DATE(updated_at) as fail_date,
        COUNT(*) as failure_count
    FROM broadcast_messages
    WHERE sequence_stepid IS NOT NULL 
    AND status = 'failed' 
    AND error_message IS NOT NULL
    AND updated_at >= CURRENT_DATE - INTERVAL '7 days'
    GROUP BY DATE(updated_at)
    ORDER BY fail_date DESC
""")

time_pattern = cursor.fetchall()
if time_pattern:
    print(f"\n{'Date':<12} {'Failures':>10}")
    print("-" * 25)
    for date, count in time_pattern:
        print(f"{str(date):<12} {count:>10}")

# Get sequence-wise failure stats
print("\n" + "="*80)
print("SEQUENCE-WISE FAILURE STATS")
print("="*80)

cursor.execute("""
    SELECT 
        s.name as sequence_name,
        COUNT(DISTINCT bm.id) as failed_messages,
        COUNT(DISTINCT bm.phone) as unique_phones_failed
    FROM broadcast_messages bm
    JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
    JOIN sequences s ON ss.sequence_id = s.id
    WHERE bm.status = 'failed' 
    AND bm.error_message IS NOT NULL
    GROUP BY s.name
    ORDER BY failed_messages DESC
""")

sequence_failures = cursor.fetchall()
if sequence_failures:
    print(f"\n{'Sequence':<25} {'Failed Messages':>15} {'Unique Phones':>15}")
    print("-" * 58)
    for seq in sequence_failures:
        print(f"{seq[0]:<25} {seq[1]:>15} {seq[2]:>15}")

cursor.close()
conn.close()
print(f"\n{'='*80}")
print("[OK] Analysis complete!")
