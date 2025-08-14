import psycopg2
import pandas as pd
from datetime import datetime

# Connect with SSL
conn_str = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(conn_str, connect_timeout=20)
cursor = conn.cursor()
print("[OK] Connected to Railway PostgreSQL!")

print("\n" + "="*80)
print("ANALYZING FAILED SEQUENCE MESSAGES")
print("="*80)

# First, get the count
cursor.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages 
    WHERE sequence_stepid IS NOT NULL 
    AND status = 'failed' 
    AND error_message IS NOT NULL
""")
failed_count = cursor.fetchone()[0]
print(f"\nTotal failed sequence messages with errors: {failed_count}")

# Get detailed breakdown of error types
print("\n--- ERROR MESSAGE BREAKDOWN ---")
cursor.execute("""
    SELECT error_message, COUNT(*) as count
    FROM broadcast_messages 
    WHERE sequence_stepid IS NOT NULL 
    AND status = 'failed' 
    AND error_message IS NOT NULL
    GROUP BY error_message
    ORDER BY count DESC
""")
error_breakdown = cursor.fetchall()

for error, count in error_breakdown:
    print(f"\n[{count} occurrences]")
    print(f"Error: {error}")

# Get sample failed messages with full details
print("\n\n--- SAMPLE FAILED MESSAGES (First 10) ---")
cursor.execute("""
    SELECT 
        bm.id,
        bm.recipient_phone,
        bm.message,
        bm.error_message,
        bm.created_at,
        bm.scheduled_at,
        bm.sent_at,
        ss.step_number,
        ss.message as step_template,
        s.name as sequence_name
    FROM broadcast_messages bm
    LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
    LEFT JOIN sequences s ON ss.sequence_id = s.id
    WHERE bm.sequence_stepid IS NOT NULL 
    AND bm.status = 'failed' 
    AND bm.error_message IS NOT NULL
    ORDER BY bm.created_at DESC
    LIMIT 10
""")

failed_messages = cursor.fetchall()

for i, msg in enumerate(failed_messages, 1):
    print(f"\n{'='*60}")
    print(f"Failed Message #{i}")
    print(f"{'='*60}")
    print(f"ID: {msg[0]}")
    print(f"Phone: {msg[1]}")
    print(f"Message: {msg[2][:100]}..." if msg[2] and len(msg[2]) > 100 else f"Message: {msg[2]}")
    print(f"Error: {msg[3]}")
    print(f"Created: {msg[4]}")
    print(f"Scheduled: {msg[5]}")
    print(f"Failed at: {msg[6]}")
    print(f"Sequence: {msg[9]} - Step {msg[7]}")
    print(f"Template: {msg[8][:100]}..." if msg[8] and len(msg[8]) > 100 else f"Template: {msg[8]}")

# Check which devices are failing
print("\n\n--- DEVICES WITH FAILED MESSAGES ---")
cursor.execute("""
    SELECT 
        ud.device_name,
        ud.phone as device_phone,
        ud.status as device_status,
        COUNT(bm.id) as failed_messages
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.sequence_stepid IS NOT NULL 
    AND bm.status = 'failed' 
    AND bm.error_message IS NOT NULL
    GROUP BY ud.id, ud.device_name, ud.phone, ud.status
    ORDER BY failed_messages DESC
""")

device_failures = cursor.fetchall()
for device in device_failures:
    print(f"\n{device[0]}")
    print(f"  Phone: {device[1]}")
    print(f"  Status: {device[2]}")
    print(f"  Failed messages: {device[3]}")

# Check timing patterns
print("\n\n--- FAILURE TIMING ANALYSIS ---")
cursor.execute("""
    SELECT 
        DATE(sent_at) as failure_date,
        COUNT(*) as failures
    FROM broadcast_messages 
    WHERE sequence_stepid IS NOT NULL 
    AND status = 'failed' 
    AND error_message IS NOT NULL
    AND sent_at IS NOT NULL
    GROUP BY DATE(sent_at)
    ORDER BY failure_date DESC
    LIMIT 10
""")

timing_data = cursor.fetchall()
for date, count in timing_data:
    print(f"{date}: {count} failures")

# Check if there are any successful sequence messages
cursor.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages 
    WHERE sequence_stepid IS NOT NULL 
    AND status = 'sent'
""")
success_count = cursor.fetchone()[0]
print(f"\n\nFor comparison - Successful sequence messages: {success_count}")

cursor.close()
conn.close()
print("\n[OK] Analysis complete!")
