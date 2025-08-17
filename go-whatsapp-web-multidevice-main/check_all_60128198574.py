import mysql.connector
import sys
from datetime import datetime, timedelta

sys.stdout.reconfigure(encoding='utf-8')

config = {
    'host': '159.89.198.71',
    'port': 3306,
    'database': 'admin_railway',
    'user': 'admin_aqil',
    'password': 'admin_aqil'
}

try:
    conn = mysql.connector.connect(**config)
    cursor = conn.cursor(dictionary=True)
    
    print("=== ALL MESSAGES FOR 60128198574 IN LAST 2 DAYS ===\n")
    
    # Get all messages for this phone
    query = """
    SELECT 
        id,
        sequence_stepid,
        status,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at_utc,
        DATE_FORMAT(DATE_ADD(sent_at, INTERVAL 8 HOUR), '%Y-%m-%d %H:%i:%s') as sent_at_myt,
        processing_worker_id,
        device_id,
        LEFT(content, 100) as content_preview
    FROM broadcast_messages
    WHERE recipient_phone LIKE '%128198574%'
        AND created_at >= DATE_SUB(NOW(), INTERVAL 2 DAY)
    ORDER BY created_at DESC
    """
    
    cursor.execute(query)
    all_messages = cursor.fetchall()
    
    print(f"Total messages found: {len(all_messages)}\n")
    
    # Group by sequence_stepid to find duplicates
    step_groups = {}
    for msg in all_messages:
        step_id = msg['sequence_stepid']
        if step_id:
            if step_id not in step_groups:
                step_groups[step_id] = []
            step_groups[step_id].append(msg)
    
    # Show duplicates
    duplicates_found = False
    for step_id, messages in step_groups.items():
        if len(messages) > 1:
            duplicates_found = True
            print(f"❌ DUPLICATE FOUND for sequence_stepid: {step_id}")
            print(f"Number of duplicates: {len(messages)}")
            for i, msg in enumerate(messages, 1):
                print(f"\n  Message {i}:")
                print(f"  ID: {msg['id']}")
                print(f"  Status: {msg['status']}")
                print(f"  Created: {msg['created_at']}")
                print(f"  Sent UTC: {msg['sent_at_utc']}")
                print(f"  Sent MYT: {msg['sent_at_myt']}")
                print(f"  Worker ID: {msg['processing_worker_id']}")
                print(f"  Device: {msg['device_id']}")
                print(f"  Content: {msg['content_preview']}...")
            print("-" * 80)
    
    if not duplicates_found:
        print("✅ No duplicates found by sequence_stepid\n")
        print("All messages:")
        for msg in all_messages:
            print(f"\nID: {msg['id']}")
            print(f"Step ID: {msg['sequence_stepid']}")
            print(f"Status: {msg['status']}")
            print(f"Created: {msg['created_at']}")
            if msg['sent_at_myt']:
                print(f"Sent MYT: {msg['sent_at_myt']}")
            print(f"Content: {msg['content_preview'][:50]}...")
            
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
