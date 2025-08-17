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
    
    print("=== CHECKING STATUS OF TEST MESSAGE ===\n")
    
    # Check the specific test message
    query = """
    SELECT 
        id,
        recipient_phone,
        status,
        processing_worker_id,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        DATE_FORMAT(processing_started_at, '%Y-%m-%d %H:%i:%s') as processing_started_at,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
        error_message,
        device_id,
        LEFT(content, 100) as content_preview
    FROM broadcast_messages
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(query)
    message = cursor.fetchone()
    
    if message:
        print(f"Message ID: {message['id']}")
        print(f"Phone: {message['recipient_phone']}")
        print(f"Status: {message['status']}")
        print(f"Worker ID: {message['processing_worker_id'] or 'NULL (NOT PICKED UP)'}")
        print(f"Created: {message['created_at']}")
        print(f"Processing Started: {message['processing_started_at'] or 'Not started'}")
        print(f"Sent At: {message['sent_at'] or 'Not sent yet'}")
        print(f"Error: {message['error_message'] or 'None'}")
        print(f"Content: {message['content_preview']}...")
        
        # Status interpretation
        print("\n📊 STATUS ANALYSIS:")
        if message['status'] == 'pending' and message['processing_worker_id'] is None:
            print("❌ Message NOT picked up yet - Still waiting in queue")
        elif message['status'] == 'processing' and message['processing_worker_id']:
            print(f"⏳ Message is being processed by worker: {message['processing_worker_id']}")
        elif message['status'] == 'queued':
            print("📤 Message queued for sending")
        elif message['status'] == 'sent':
            print("✅ Message successfully sent!")
        elif message['status'] == 'failed':
            print(f"❌ Message failed: {message['error_message']}")
        elif message['status'] == 'skipped':
            print(f"⚠️ Message skipped: {message['error_message']}")
    else:
        print("Message not found!")
    
    # Also check all messages for your phone number today
    print("\n\n=== ALL MESSAGES FOR YOUR NUMBER TODAY ===\n")
    
    query2 = """
    SELECT 
        id,
        status,
        processing_worker_id,
        DATE_FORMAT(created_at, '%H:%i:%s') as created_time,
        DATE_FORMAT(sent_at, '%H:%i:%s') as sent_time,
        LEFT(content, 50) as content_preview
    FROM broadcast_messages
    WHERE recipient_phone = '60108924904'
        AND DATE(created_at) = CURDATE()
    ORDER BY created_at DESC
    """
    
    cursor.execute(query2)
    all_messages = cursor.fetchall()
    
    if all_messages:
        print(f"Found {len(all_messages)} messages for your number today:\n")
        for msg in all_messages:
            worker_status = "✅" if msg['processing_worker_id'] else "❌"
            print(f"ID: {msg['id'][:8]}...")
            print(f"Status: {msg['status']}")
            print(f"Worker ID: {worker_status} {msg['processing_worker_id'] or 'NULL'}")
            print(f"Created: {msg['created_time']}, Sent: {msg['sent_time'] or 'Not sent'}")
            print(f"Content: {msg['content_preview']}...")
            print("-" * 50)
    
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
