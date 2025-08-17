import mysql.connector
from datetime import datetime
import sys

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
    
    print("=== ALL PENDING MESSAGES FOR YOUR PHONE ===\n")
    
    query = """
    SELECT 
        id,
        status,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        TIMESTAMPDIFF(MINUTE, scheduled_at, NOW()) as age_minutes,
        processing_worker_id,
        device_id,
        LEFT(content, 50) as content_preview
    FROM broadcast_messages
    WHERE recipient_phone = '60108924904'
        AND status = 'pending'
    ORDER BY created_at DESC
    """
    
    cursor.execute(query)
    messages = cursor.fetchall()
    
    if messages:
        print(f"Found {len(messages)} pending messages:\n")
        for msg in messages:
            print(f"ID: {msg['id']}")
            print(f"Created: {msg['created_at']}")
            print(f"Scheduled: {msg['scheduled_at']}")
            print(f"Age: {msg['age_minutes']} minutes")
            print(f"Worker ID: {msg['processing_worker_id'] or 'NULL'}")
            print(f"Content: {msg['content_preview']}...")
            
            if msg['age_minutes'] > 10:
                print("❌ TOO OLD - Outside 10-minute window")
            else:
                print("✅ In time window")
            print("-" * 60)
    else:
        print("No pending messages found for your phone")
    
    # Check the specific test message
    print("\n\n=== CHECKING TEST MESSAGE ===")
    
    test_query = """
    SELECT id, status, sent_at, error_message
    FROM broadcast_messages
    WHERE id = '9d36c1a5-3bd3-468d-a5f6-43db174f58e9'
    """
    
    cursor.execute(test_query)
    test = cursor.fetchone()
    
    if test:
        print(f"Test message status: {test['status']}")
        if test['sent_at']:
            print(f"Sent at: {test['sent_at']}")
        if test['error_message']:
            print(f"Error: {test['error_message']}")
    else:
        print("Test message not found - might have been sent")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
