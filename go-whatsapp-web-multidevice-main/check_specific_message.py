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
    
    print("=== CHECKING MESSAGE fgeef10d-4ae4-4e8a-8bbd-73d2fdb6094a ===\n")
    
    query = """
    SELECT 
        id,
        status,
        recipient_phone,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at,
        DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i:%s') as now_time,
        TIMESTAMPDIFF(MINUTE, scheduled_at, NOW()) as age_minutes,
        processing_worker_id,
        device_id,
        content,
        campaign_id,
        sequence_id,
        user_id
    FROM broadcast_messages
    WHERE id = 'fgeef10d-4ae4-4e8a-8bbd-73d2fdb6094a'
    """
    
    cursor.execute(query)
    msg = cursor.fetchone()
    
    if msg:
        print(f"ID: {msg['id']}")
        print(f"Phone: {msg['recipient_phone']}")
        print(f"Status: {msg['status']}")
        print(f"Device ID: {msg['device_id']}")
        print(f"User ID: {msg['user_id']}")
        print(f"Campaign ID: {msg['campaign_id']}")
        print(f"Sequence ID: {msg['sequence_id']}")
        print(f"\nScheduled: {msg['scheduled_at']}")
        print(f"Current: {msg['now_time']}")
        print(f"Age: {msg['age_minutes']} minutes")
        print(f"Worker ID: {msg['processing_worker_id'] or 'NULL'}")
        
        print(f"\nContent: {msg['content']}")
        
        # Check all conditions for GetPendingMessagesAndLock
        print("\n=== CHECKING CONDITIONS ===")
        
        conditions = [
            ("status = 'pending'", msg['status'] == 'pending'),
            ("processing_worker_id IS NULL", msg['processing_worker_id'] is None),
            ("scheduled_at IS NOT NULL", msg['scheduled_at'] is not None),
            ("age <= 10 minutes", msg['age_minutes'] <= 10)
        ]
        
        all_pass = True
        for condition, result in conditions:
            status = "✅" if result else "❌"
            print(f"{status} {condition}")
            if not result:
                all_pass = False
        
        if all_pass:
            print("\n✅ Message SHOULD be picked up!")
            print("But worker_id is NULL, so old code is running.")
        else:
            print("\n❌ Message does NOT meet all conditions")
    
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
