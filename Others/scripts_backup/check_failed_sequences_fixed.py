import psycopg2
import pandas as pd
from datetime import datetime

# Connect with SSL
conn_str = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(conn_str, connect_timeout=20)
cursor = conn.cursor()
print("[OK] Connected to Railway PostgreSQL!")

# First, check the actual columns in broadcast_messages
print("\n--- BROADCAST_MESSAGES TABLE STRUCTURE ---")
cursor.execute("""
    SELECT column_name, data_type 
    FROM information_schema.columns 
    WHERE table_name = 'broadcast_messages' 
    ORDER BY ordinal_position
""")
columns = cursor.fetchall()
print("Columns in broadcast_messages:")
for col, dtype in columns:
    print(f"  - {col}: {dtype}")

print("\n" + "="*80)
print("ANALYZING FAILED SEQUENCE MESSAGES")
print("="*80)

# Get the count of failed sequence messages
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

# Get sample failed messages - using correct column names
print("\n\n--- SAMPLE FAILED MESSAGES (First 10) ---")
cursor.execute("""
    SELECT 
        bm.id,
        bm.recipient_phone,
        bm.recipient_name,
        bm.error_message,
        bm.created_at,
        bm.scheduled_at,
        bm.sent_at,
        bm.device_id,
        bm.sequence_stepid,
        bm.message_content
    FROM broadcast_messages bm
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
    print(f"Name: {msg[2]}")
    print(f"Error: {msg[3]}")
    print(f"Created: {msg[4]}")
    print(f"Scheduled: {msg[5]}")
    print(f"Failed at: {msg[6]}")
    print(f"Device ID: {msg[7]}")
    print(f"Sequence Step ID: {msg[8]}")
    if msg[9]:
        print(f"Message: {msg[9][:100]}..." if len(msg[9]) > 100 else f"Message: {msg[9]}")

# Get sequence step details for these failures
print("\n\n--- SEQUENCE STEPS THAT ARE FAILING ---")
cursor.execute("""
    SELECT 
        ss.id as step_id,
        s.name as sequence_name,
        ss.step_number,
        ss.message as step_message,
        COUNT(bm.id) as failure_count
    FROM broadcast_messages bm
    JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
    JOIN sequences s ON ss.sequence_id = s.id
    WHERE bm.status = 'failed' 
    AND bm.error_message IS NOT NULL
    GROUP BY ss.id, s.name, ss.step_number, ss.message
    ORDER BY failure_count DESC
""")

failing_steps = cursor.fetchall()
for step in failing_steps:
    print(f"\n{step[1]} - Step {step[2]}")
    print(f"  Step ID: {step[0]}")
    print(f"  Failures: {step[4]}")
    print(f"  Message: {step[3][:100]}..." if step[3] and len(step[3]) > 100 else f"  Message: {step[3]}")

# Check devices with failures
print("\n\n--- DEVICES WITH FAILED SEQUENCE MESSAGES ---")
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
    LIMIT 10
""")

device_failures = cursor.fetchall()
for device in device_failures:
    print(f"\n{device[0]}")
    print(f"  Phone: {device[1]}")
    print(f"  Status: {device[2]}")
    print(f"  Failed messages: {device[3]}")

cursor.close()
conn.close()
print("\n[OK] Analysis complete!")
