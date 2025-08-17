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
    
    print("=== DEBUGGING WHY MESSAGE NOT PICKED UP ===\n")
    
    # Check your test message
    query = """
    SELECT 
        id,
        status,
        processing_worker_id,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at,
        DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i:%s') as now_time,
        DATE_FORMAT(DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR), '%Y-%m-%d %H:%i:%s') as min_allowed_time,
        CASE 
            WHEN scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR) 
            THEN 'YES - In window' 
            ELSE 'NO - Too old' 
        END as in_time_window
    FROM broadcast_messages
    WHERE id = '9d36c1a5-3bd3-468d-a5f6-43db174f58e9'
    """
    
    cursor.execute(query)
    msg = cursor.fetchone()
    
    if msg:
        print(f"Message ID: {msg['id']}")
        print(f"Status: {msg['status']}")
        print(f"Worker ID: {msg['processing_worker_id'] or 'NULL'}")
        print(f"\nScheduled at: {msg['scheduled_at']}")
        print(f"Current time: {msg['now_time']}")
        print(f"Min allowed: {msg['min_allowed_time']}")
        print(f"In time window: {msg['in_time_window']}")
        
        # Calculate how old the message is
        cursor.execute("""
        SELECT TIMESTAMPDIFF(MINUTE, scheduled_at, NOW()) as age_minutes
        FROM broadcast_messages
        WHERE id = '9d36c1a5-3bd3-468d-a5f6-43db174f58e9'
        """)
        age = cursor.fetchone()
        print(f"\nMessage age: {age['age_minutes']} minutes")
        
        if age['age_minutes'] > 10:
            print("❌ Message is older than 10 minutes - WILL NOT BE PICKED UP!")
            print("\nThe code has a 10-minute time window restriction.")
            print("Messages older than 10 minutes are ignored.")
    
    # Check if there are ANY messages in the time window
    print("\n\n=== MESSAGES IN TIME WINDOW ===")
    
    window_query = """
    SELECT 
        COUNT(*) as total,
        SUM(CASE WHEN device_id = '8badb299-f1d1-493a-bddf-84cbaba1273b' THEN 1 ELSE 0 END) as your_device
    FROM broadcast_messages
    WHERE status = 'pending'
        AND processing_worker_id IS NULL
        AND scheduled_at IS NOT NULL
        AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
        AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
    """
    
    cursor.execute(window_query)
    window = cursor.fetchone()
    
    print(f"Total messages in window: {window['total']}")
    print(f"For your device: {window['your_device']}")
    
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
