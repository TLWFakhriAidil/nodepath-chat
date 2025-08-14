import psycopg2
import json
from datetime import datetime
import sys

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

# Database connection
conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    print("Connecting to PostgreSQL database...")
    conn = psycopg2.connect(conn_string)
    cursor = conn.cursor()
    print("[SUCCESS] Connected successfully!\n")
    
    # 1. Check messages with status 'failed' that have error messages
    print("[1. FAILED MESSAGES WITH ERRORS]:")
    print("-" * 100)
    cursor.execute("""
        SELECT 
            id,
            recipient_phone,
            status,
            error_message,
            created_at,
            scheduled_at,
            device_id,
            campaign_id,
            sequence_id
        FROM broadcast_messages
        WHERE status = 'failed' AND error_message IS NOT NULL
        ORDER BY created_at DESC
        LIMIT 20
    """)
    failed_with_errors = cursor.fetchall()
    
    if failed_with_errors:
        for msg in failed_with_errors:
            print(f"ID: {msg[0]}")
            print(f"  Phone: {msg[1]} | Status: {msg[2]}")
            print(f"  Error: {msg[3]}")
            print(f"  Created: {msg[4]} | Scheduled: {msg[5]}")
            print(f"  Campaign: {msg[7]} | Sequence: {msg[8]}")
            print()
    else:
        print("No failed messages with error messages found")
    
    # 2. Check messages with status 'sent' but have error messages (inconsistent)
    print("\n[2. SENT MESSAGES WITH ERROR MESSAGES (INCONSISTENT)]:")
    print("-" * 100)
    cursor.execute("""
        SELECT 
            id,
            recipient_phone,
            status,
            error_message,
            sent_at,
            device_id
        FROM broadcast_messages
        WHERE status = 'sent' AND error_message IS NOT NULL
        ORDER BY sent_at DESC
        LIMIT 20
    """)
    sent_with_errors = cursor.fetchall()
    
    if sent_with_errors:
        for msg in sent_with_errors:
            print(f"ID: {msg[0]}")
            print(f"  Phone: {msg[1]} | Status: {msg[2]}")
            print(f"  Error (shouldn't exist): {msg[3]}")
            print(f"  Sent at: {msg[4]}")
            print()
    else:
        print("No inconsistent sent messages with errors found (Good!)")
    
    # 3. Count of messages by status
    print("\n[3. MESSAGE STATUS SUMMARY]:")
    print("-" * 100)
    cursor.execute("""
        SELECT 
            status,
            COUNT(*) as count,
            COUNT(CASE WHEN error_message IS NOT NULL THEN 1 END) as with_errors
        FROM broadcast_messages
        GROUP BY status
        ORDER BY count DESC
    """)
    status_summary = cursor.fetchall()
    
    for status, count, with_errors in status_summary:
        print(f"Status: {status:15} | Total: {count:6} | With Errors: {with_errors:6}")
    
    # 4. Most common error messages
    print("\n[4. MOST COMMON ERROR MESSAGES]:")
    print("-" * 100)
    cursor.execute("""
        SELECT 
            error_message,
            COUNT(*) as count,
            MAX(created_at) as last_occurrence
        FROM broadcast_messages
        WHERE error_message IS NOT NULL
        GROUP BY error_message
        ORDER BY count DESC
        LIMIT 10
    """)
    common_errors = cursor.fetchall()
    
    if common_errors:
        for error, count, last_time in common_errors:
            print(f"Error: {error[:80]}...")
            print(f"  Count: {count} | Last occurred: {last_time}")
            print()
    else:
        print("No error messages found")
    
    # 5. Failed messages by device
    print("\n[5. FAILED MESSAGES BY DEVICE]:")
    print("-" * 100)
    cursor.execute("""
        SELECT 
            d.device_name,
            d.jid,
            d.status as device_status,
            COUNT(bm.id) as failed_count,
            MAX(bm.created_at) as last_failure
        FROM broadcast_messages bm
        JOIN user_devices d ON bm.device_id = d.id
        WHERE bm.status = 'failed'
        GROUP BY d.id, d.device_name, d.jid, d.status
        ORDER BY failed_count DESC
        LIMIT 10
    """)
    device_failures = cursor.fetchall()
    
    if device_failures:
        for device_name, jid, status, count, last_fail in device_failures:
            print(f"Device: {device_name} | Status: {status}")
            print(f"  JID: {jid[:50]}...")
            print(f"  Failed messages: {count} | Last failure: {last_fail}")
            print()
    else:
        print("No device-specific failures found")
    
    # 6. Messages stuck in pending for too long
    print("\n[6. STUCK PENDING MESSAGES (>24 hours)]:")
    print("-" * 100)
    cursor.execute("""
        SELECT 
            id,
            recipient_phone,
            status,
            created_at,
            scheduled_at,
            device_id
        FROM broadcast_messages
        WHERE status = 'pending' 
        AND created_at < NOW() - INTERVAL '24 hours'
        ORDER BY created_at
        LIMIT 10
    """)
    stuck_messages = cursor.fetchall()
    
    if stuck_messages:
        for msg in stuck_messages:
            print(f"ID: {msg[0]}")
            print(f"  Phone: {msg[1]} | Status: {msg[2]}")
            print(f"  Created: {msg[3]} | Scheduled: {msg[4]}")
            print(f"  Device: {msg[5]}")
            print()
    else:
        print("No stuck pending messages found")
    
    # 7. Recent errors (last 24 hours)
    print("\n[7. RECENT ERRORS (Last 24 hours)]:")
    print("-" * 100)
    cursor.execute("""
        SELECT 
            COUNT(*) as total_errors,
            COUNT(DISTINCT device_id) as devices_affected,
            COUNT(DISTINCT recipient_phone) as recipients_affected
        FROM broadcast_messages
        WHERE error_message IS NOT NULL
        AND created_at > NOW() - INTERVAL '24 hours'
    """)
    recent_stats = cursor.fetchone()
    print(f"Total errors: {recent_stats[0]}")
    print(f"Devices affected: {recent_stats[1]}")
    print(f"Recipients affected: {recent_stats[2]}")
    
    cursor.close()
    conn.close()
    print("\n[SUCCESS] Analysis complete!")
    
except Exception as e:
    print(f"[ERROR] {str(e)}")
    import traceback
    traceback.print_exc()
