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
    
    print("=== CHECKING ALL MESSAGES FOR YOUR NUMBER ===\n")
    
    # Check ALL messages for your number in last 2 days
    query = """
    SELECT 
        id,
        recipient_phone,
        status,
        processing_worker_id,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
        device_id,
        LEFT(content, 100) as content_preview
    FROM broadcast_messages
    WHERE recipient_phone IN ('60108924904', '+60108924904', '0108924904')
        AND created_at >= DATE_SUB(NOW(), INTERVAL 2 DAY)
    ORDER BY created_at DESC
    """
    
    cursor.execute(query)
    messages = cursor.fetchall()
    
    print(f"Found {len(messages)} messages for your number:\n")
    
    for msg in messages:
        print(f"ID: {msg['id']}")
        print(f"Phone: {msg['recipient_phone']}")
        print(f"Status: {msg['status']}")
        print(f"Worker ID: {msg['processing_worker_id'] or 'NULL'}")
        print(f"Created: {msg['created_at']}")
        print(f"Sent: {msg['sent_at'] or 'Not sent'}")
        print(f"Device: {msg['device_id']}")
        print(f"Content: {msg['content_preview']}...")
        print("-" * 80)
    
    # Check if the old processor is running
    print("\n\n=== CHECKING RECENT PROCESSING ACTIVITY ===\n")
    
    query2 = """
    SELECT 
        COUNT(*) as total_processed,
        SUM(CASE WHEN processing_worker_id IS NOT NULL THEN 1 ELSE 0 END) as with_worker_id,
        SUM(CASE WHEN processing_worker_id IS NULL THEN 1 ELSE 0 END) as without_worker_id,
        MAX(sent_at) as last_sent
    FROM broadcast_messages
    WHERE sent_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
    """
    
    cursor.execute(query2)
    activity = cursor.fetchone()
    
    print(f"Messages sent in last hour: {activity['total_processed']}")
    print(f"With worker ID: {activity['with_worker_id']}")
    print(f"Without worker ID (OLD CODE): {activity['without_worker_id']}")
    print(f"Last sent: {activity['last_sent']}")
    
    if activity['without_worker_id'] > 0:
        print("\n⚠️ WARNING: The OLD code (without worker ID) is still running!")
        print("The application needs to be restarted with the new code.")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
