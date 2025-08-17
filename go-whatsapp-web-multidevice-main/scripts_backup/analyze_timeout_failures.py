import psycopg2
import pandas as pd
from datetime import datetime

# Connect with SSL
conn_str = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(conn_str, connect_timeout=20)
cursor = conn.cursor()
print("[OK] Connected to Railway PostgreSQL!")

print("\n" + "="*80)
print("ANALYZING FAILED SEQUENCE MESSAGES WITH TIMEOUT ERRORS")
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

# Get sample failed messages with correct column names
print("\n\n--- DETAILED ANALYSIS OF FAILED MESSAGES ---")
cursor.execute("""
    SELECT 
        bm.id,
        bm.recipient_phone,
        bm.recipient_name,
        bm.content,
        bm.error_message,
        bm.created_at,
        bm.scheduled_at,
        bm.sent_at,
        bm.device_id,
        ud.device_name,
        ud.phone as device_phone,
        ud.status as device_status,
        ss.step_number,
        ss.message as step_template,
        s.name as sequence_name
    FROM broadcast_messages bm
    LEFT JOIN user_devices ud ON bm.device_id = ud.id
    LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
    LEFT JOIN sequences s ON bm.sequence_id = s.id
    WHERE bm.sequence_stepid IS NOT NULL 
    AND bm.status = 'failed' 
    AND bm.error_message = 'Message timeout - could not be delivered within 12 hours'
    ORDER BY bm.created_at DESC
    LIMIT 10
""")

failed_messages = cursor.fetchall()

print(f"\nShowing {len(failed_messages)} sample timeout failures:")

for i, msg in enumerate(failed_messages, 1):
    print(f"\n{'='*70}")
    print(f"FAILED MESSAGE #{i}")
    print(f"{'='*70}")
    print(f"Message ID: {msg[0]}")
    print(f"Recipient: {msg[2]} ({msg[1]})")
    print(f"Content: {msg[3][:100]}..." if msg[3] and len(msg[3]) > 100 else f"Content: {msg[3]}")
    print(f"\nTiming:")
    print(f"  Created: {msg[5]}")
    print(f"  Scheduled: {msg[6]}")
    print(f"  Failed at: {msg[7]}")
    time_diff = (msg[7] - msg[6]).total_seconds() / 3600 if msg[7] and msg[6] else None
    if time_diff:
        print(f"  Time until timeout: {time_diff:.1f} hours")
    
    print(f"\nDevice Info:")
    print(f"  Device: {msg[9]} ({msg[10]})")
    print(f"  Device Status: {msg[11]}")
    
    print(f"\nSequence Info:")
    print(f"  Sequence: {msg[14]}")
    print(f"  Step: {msg[12]}")
    print(f"  Template: {msg[13][:80]}..." if msg[13] and len(msg[13]) > 80 else f"  Template: {msg[13]}")

# Analyze timeout patterns
print("\n\n--- TIMEOUT PATTERN ANALYSIS ---")

# Check when messages were created vs when they timed out
cursor.execute("""
    SELECT 
        DATE(created_at) as created_date,
        DATE(sent_at) as failed_date,
        COUNT(*) as count,
        AVG(EXTRACT(EPOCH FROM (sent_at - scheduled_at))/3600) as avg_hours_to_timeout
    FROM broadcast_messages 
    WHERE sequence_stepid IS NOT NULL 
    AND status = 'failed' 
    AND error_message = 'Message timeout - could not be delivered within 12 hours'
    AND sent_at IS NOT NULL
    GROUP BY DATE(created_at), DATE(sent_at)
    ORDER BY created_date DESC
""")

timeout_patterns = cursor.fetchall()
print("\nTimeout patterns by date:")
for created, failed, count, avg_hours in timeout_patterns:
    print(f"  Created: {created}, Failed: {failed}, Count: {count}, Avg timeout: {avg_hours:.1f} hrs")

# Check which devices are having timeout issues
print("\n\n--- DEVICES WITH TIMEOUT ISSUES ---")
cursor.execute("""
    SELECT 
        ud.device_name,
        ud.phone,
        ud.status,
        COUNT(bm.id) as timeout_count,
        MIN(bm.created_at) as first_timeout,
        MAX(bm.created_at) as last_timeout
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.sequence_stepid IS NOT NULL 
    AND bm.status = 'failed' 
    AND bm.error_message = 'Message timeout - could not be delivered within 12 hours'
    GROUP BY ud.id, ud.device_name, ud.phone, ud.status
    ORDER BY timeout_count DESC
""")

device_timeouts = cursor.fetchall()
for device in device_timeouts:
    print(f"\n{device[0]}")
    print(f"  Phone: {device[1]}")
    print(f"  Current Status: {device[2]}")
    print(f"  Timeout Count: {device[3]}")
    print(f"  First Timeout: {device[4]}")
    print(f"  Last Timeout: {device[5]}")

# Check if these phones have any successful messages
print("\n\n--- CHECKING SUCCESS RATE FOR AFFECTED PHONES ---")
cursor.execute("""
    WITH failed_phones AS (
        SELECT DISTINCT recipient_phone
        FROM broadcast_messages
        WHERE sequence_stepid IS NOT NULL 
        AND status = 'failed' 
        AND error_message = 'Message timeout - could not be delivered within 12 hours'
    )
    SELECT 
        fp.recipient_phone,
        SUM(CASE WHEN bm.status = 'sent' THEN 1 ELSE 0 END) as sent_count,
        SUM(CASE WHEN bm.status = 'failed' THEN 1 ELSE 0 END) as failed_count,
        SUM(CASE WHEN bm.status = 'pending' THEN 1 ELSE 0 END) as pending_count
    FROM failed_phones fp
    LEFT JOIN broadcast_messages bm ON fp.recipient_phone = bm.recipient_phone
    GROUP BY fp.recipient_phone
    ORDER BY failed_count DESC
    LIMIT 10
""")

phone_stats = cursor.fetchall()
print("\nMessage statistics for phones with timeouts:")
for phone, sent, failed, pending in phone_stats:
    total = sent + failed + pending
    success_rate = (sent / total * 100) if total > 0 else 0
    print(f"\n{phone}:")
    print(f"  Sent: {sent}, Failed: {failed}, Pending: {pending}")
    print(f"  Success Rate: {success_rate:.1f}%")

cursor.close()
conn.close()
print("\n[OK] Analysis complete!")
