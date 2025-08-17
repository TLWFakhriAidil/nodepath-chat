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
    
    print("=== DIAGNOSTIC CHECK ===\n")
    
    # 1. Check if the device is online
    print("1. CHECKING DEVICE STATUS:")
    query1 = """
    SELECT 
        id,
        device_name,
        status,
        platform,
        DATE_FORMAT(last_seen, '%Y-%m-%d %H:%i:%s') as last_seen
    FROM user_devices
    WHERE id = '8badb299-f1d1-493a-bddf-84cbaba1273b'
    """
    
    cursor.execute(query1)
    device = cursor.fetchone()
    
    if device:
        print(f"Device: {device['device_name']}")
        print(f"Status: {device['status']}")
        print(f"Platform: {device['platform'] or 'WhatsApp Web'}")
        print(f"Last seen: {device['last_seen']}")
        
        if device['status'] not in ['connected', 'online']:
            print("❌ ISSUE: Device is not online!")
    else:
        print("❌ Device not found!")
    
    # 2. Check if there are other pending messages
    print("\n\n2. CHECKING PENDING MESSAGES FOR THIS DEVICE:")
    query2 = """
    SELECT 
        COUNT(*) as total_pending,
        MIN(DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')) as oldest_pending,
        MAX(DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')) as newest_pending
    FROM broadcast_messages
    WHERE device_id = '8badb299-f1d1-493a-bddf-84cbaba1273b'
        AND status = 'pending'
    """
    
    cursor.execute(query2)
    pending = cursor.fetchone()
    
    print(f"Total pending: {pending['total_pending']}")
    print(f"Oldest pending: {pending['oldest_pending']}")
    print(f"Newest pending: {pending['newest_pending']}")
    
    # 3. Check recent processing activity
    print("\n\n3. RECENT PROCESSING ACTIVITY:")
    query3 = """
    SELECT 
        status,
        COUNT(*) as count,
        MAX(DATE_FORMAT(processing_started_at, '%Y-%m-%d %H:%i:%s')) as last_processed
    FROM broadcast_messages
    WHERE device_id = '8badb299-f1d1-493a-bddf-84cbaba1273b'
        AND processing_started_at IS NOT NULL
        AND processing_started_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
    GROUP BY status
    """
    
    cursor.execute(query3)
    activity = cursor.fetchall()
    
    if activity:
        for act in activity:
            print(f"Status: {act['status']}, Count: {act['count']}, Last: {act['last_processed']}")
    else:
        print("No recent processing activity!")
    
    # 4. Check if any messages are stuck in processing
    print("\n\n4. CHECKING FOR STUCK MESSAGES:")
    query4 = """
    SELECT 
        id,
        status,
        processing_worker_id,
        DATE_FORMAT(processing_started_at, '%Y-%m-%d %H:%i:%s') as started_at,
        TIMESTAMPDIFF(MINUTE, processing_started_at, NOW()) as minutes_stuck
    FROM broadcast_messages
    WHERE status = 'processing'
        AND processing_started_at < DATE_SUB(NOW(), INTERVAL 5 MINUTE)
    LIMIT 5
    """
    
    cursor.execute(query4)
    stuck = cursor.fetchall()
    
    if stuck:
        print(f"Found {len(stuck)} stuck messages:")
        for s in stuck:
            print(f"- {s['id'][:8]}... stuck for {s['minutes_stuck']} minutes")
    else:
        print("No stuck messages found")
    
    # 5. Check scheduled_at time
    print("\n\n5. CHECKING SCHEDULED TIME:")
    query5 = """
    SELECT 
        id,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at_utc,
        DATE_FORMAT(DATE_ADD(scheduled_at, INTERVAL 8 HOUR), '%Y-%m-%d %H:%i:%s') as scheduled_at_myt,
        DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i:%s') as current_time_utc,
        DATE_FORMAT(DATE_ADD(NOW(), INTERVAL 8 HOUR), '%Y-%m-%d %H:%i:%s') as current_time_myt,
        CASE 
            WHEN scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR) THEN 'Ready to send'
            ELSE 'Not ready yet'
        END as send_status
    FROM broadcast_messages
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(query5)
    schedule = cursor.fetchone()
    
    if schedule:
        print(f"Scheduled at (UTC): {schedule['scheduled_at_utc']}")
        print(f"Scheduled at (MYT): {schedule['scheduled_at_myt']}")
        print(f"Current time (UTC): {schedule['current_time_utc']}")
        print(f"Current time (MYT): {schedule['current_time_myt']}")
        print(f"Send status: {schedule['send_status']}")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
