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
    
    print("=== TRACING COMPLETE SEQUENCE FLOW ===\n")
    
    # 1. Check if sequence exists and is active
    print("1. SEQUENCE CHECK:")
    cursor.execute("""
    SELECT id, name, status, is_active, device_id
    FROM sequences
    WHERE id = '06bc88b9-155e-4fd7-96a3-ced3532a84f8'
    """)
    sequence = cursor.fetchone()
    if sequence:
        print(f"Sequence: {sequence['name']}")
        print(f"Status: {sequence['status']}")
        print(f"Active: {sequence['is_active']}")
        print(f"Device: {sequence['device_id']}")
    
    # 2. Check sequence contact enrollment
    print("\n2. SEQUENCE CONTACT ENROLLMENT:")
    cursor.execute("""
    SELECT sc.*, s.name as sequence_name
    FROM sequence_contacts sc
    JOIN sequences s ON sc.sequence_id = s.id
    WHERE sc.contact_phone = '60108924904'
    """)
    enrollments = cursor.fetchall()
    if enrollments:
        for e in enrollments:
            print(f"Sequence: {e['sequence_name']}")
            print(f"Status: {e['status']}")
            print(f"Current Step: {e['current_step']}")
    
    # 3. Check the message details
    print("\n3. MESSAGE IN BROADCAST_MESSAGES:")
    cursor.execute("""
    SELECT 
        bm.*,
        ud.device_name,
        ud.status as device_status,
        ud.platform
    FROM broadcast_messages bm
    LEFT JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.id = 'fgeef10d-4ae4-4e8a-8bbd-73d2fdb6094a'
    """)
    msg = cursor.fetchone()
    if msg:
        print(f"Message ID: {msg['id']}")
        print(f"Status: {msg['status']}")
        print(f"Device: {msg['device_name']} ({msg['device_status']})")
        print(f"Platform: {msg['platform']}")
        print(f"Processing Worker ID: {msg['processing_worker_id'] or 'NULL'}")
        
    # 4. Check why GetPendingMessagesAndLock might not be working
    print("\n4. CHECKING GetPendingMessagesAndLock CONDITIONS:")
    
    # Simulate the exact query
    cursor.execute("""
    SELECT COUNT(*) as eligible
    FROM broadcast_messages
    WHERE device_id = '8badb299-f1d1-493a-bddf-84cbaba1273b'
    AND status = 'pending'
    AND processing_worker_id IS NULL
    AND scheduled_at IS NOT NULL
    AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
    AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
    """)
    eligible = cursor.fetchone()
    print(f"Messages eligible for atomic locking: {eligible['eligible']}")
    
    # 5. Check if processor is running but not processing
    print("\n5. PROCESSOR ACTIVITY CHECK:")
    cursor.execute("""
    SELECT 
        COUNT(*) as total,
        SUM(CASE WHEN processing_worker_id IS NOT NULL THEN 1 ELSE 0 END) as with_worker,
        MAX(processing_started_at) as last_processing,
        MAX(sent_at) as last_sent
    FROM broadcast_messages
    WHERE created_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
    """)
    activity = cursor.fetchone()
    print(f"Messages in last hour: {activity['total']}")
    print(f"With worker ID: {activity['with_worker']}")
    print(f"Last processing: {activity['last_processing'] or 'Never'}")
    print(f"Last sent: {activity['last_sent'] or 'Never'}")
    
    # 6. Check if this is a platform issue
    print("\n6. PLATFORM DEVICE CHECK:")
    if msg and msg['platform']:
        print(f"This is a {msg['platform']} platform device")
        print("Platform devices might be processed differently")
        
        # Check platform message sending
        cursor.execute("""
        SELECT COUNT(*) as platform_sent
        FROM broadcast_messages bm
        JOIN user_devices ud ON bm.device_id = ud.id
        WHERE ud.platform IS NOT NULL
        AND bm.status = 'sent'
        AND DATE(bm.sent_at) = CURDATE()
        """)
        platform_stats = cursor.fetchone()
        print(f"Platform messages sent today: {platform_stats['platform_sent']}")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
