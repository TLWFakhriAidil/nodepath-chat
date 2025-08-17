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
    
    print("=== CHECKING SCHEDULED TIME FOR YOUR MESSAGE ===\n")
    
    # Check the scheduled time and current time
    query = """
    SELECT 
        id,
        recipient_phone,
        status,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at_utc,
        DATE_FORMAT(DATE_ADD(scheduled_at, INTERVAL 8 HOUR), '%Y-%m-%d %H:%i:%s') as scheduled_at_myt,
        DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i:%s') as current_time_utc,
        DATE_FORMAT(DATE_ADD(NOW(), INTERVAL 8 HOUR), '%Y-%m-%d %H:%i:%s') as current_time_myt,
        TIMESTAMPDIFF(MINUTE, NOW(), scheduled_at) as minutes_until_send,
        TIMESTAMPDIFF(MINUTE, NOW(), DATE_SUB(scheduled_at, INTERVAL 8 HOUR)) as minutes_until_send_adjusted,
        device_id,
        processing_worker_id
    FROM broadcast_messages
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(query)
    msg = cursor.fetchone()
    
    if msg:
        print(f"Message ID: {msg['id']}")
        print(f"Phone: {msg['recipient_phone']}")
        print(f"Status: {msg['status']}")
        print(f"\nScheduled Time (UTC): {msg['scheduled_at_utc']}")
        print(f"Scheduled Time (MYT +8): {msg['scheduled_at_myt']}")
        print(f"\nCurrent Time (UTC): {msg['current_time_utc']}")
        print(f"Current Time (MYT +8): {msg['current_time_myt']}")
        
        # Check the SQL condition that the processor uses
        print("\n=== PROCESSOR CONDITION CHECK ===")
        
        check_query = """
        SELECT 
            CASE 
                WHEN scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR) THEN 'READY TO SEND'
                ELSE 'NOT READY YET'
            END as send_status,
            CASE
                WHEN scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR) THEN 'WITHIN TIME WINDOW'
                ELSE 'OUTSIDE TIME WINDOW (too old)'
            END as time_window_status
        FROM broadcast_messages
        WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
        """
        
        cursor.execute(check_query)
        check = cursor.fetchone()
        
        print(f"Send Status: {check['send_status']}")
        print(f"Time Window: {check['time_window_status']}")
        
        # Check if this message matches the processor's WHERE conditions
        print("\n=== CHECKING IF MESSAGE MATCHES PROCESSOR CONDITIONS ===")
        
        processor_check = """
        SELECT COUNT(*) as matches
        FROM broadcast_messages
        WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
            AND status = 'pending'
            AND processing_worker_id IS NULL
            AND scheduled_at IS NOT NULL
            AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
            AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
        """
        
        cursor.execute(processor_check)
        matches = cursor.fetchone()
        
        if matches['matches'] > 0:
            print("✅ Message SHOULD be picked up by processor")
        else:
            print("❌ Message does NOT match processor conditions")
            
            # Check which condition is failing
            print("\nChecking individual conditions:")
            
            conditions = [
                ("status = 'pending'", "SELECT COUNT(*) FROM broadcast_messages WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080' AND status = 'pending'"),
                ("processing_worker_id IS NULL", "SELECT COUNT(*) FROM broadcast_messages WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080' AND processing_worker_id IS NULL"),
                ("scheduled_at IS NOT NULL", "SELECT COUNT(*) FROM broadcast_messages WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080' AND scheduled_at IS NOT NULL"),
                ("scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)", "SELECT COUNT(*) FROM broadcast_messages WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080' AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)"),
                ("scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)", "SELECT COUNT(*) FROM broadcast_messages WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080' AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)")
            ]
            
            for condition_name, condition_query in conditions:
                cursor.execute(condition_query)
                result = cursor.fetchone()
                status = "✅" if result['COUNT(*)'] > 0 else "❌"
                print(f"{status} {condition_name}")
    
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
