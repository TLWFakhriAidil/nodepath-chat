import mysql.connector
from datetime import datetime, timedelta
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
    cursor = conn.cursor()
    
    print("=== FIXING TIME WINDOW ISSUE ===\n")
    
    # Set scheduled_at to 5 minutes ago to ensure it's within the window
    update_query = """
    UPDATE broadcast_messages 
    SET scheduled_at = DATE_SUB(NOW(), INTERVAL 5 MINUTE),
        updated_at = NOW()
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(update_query)
    conn.commit()
    
    print("✅ Set scheduled_at to 5 minutes ago")
    
    # Now check if it meets conditions
    check_query = """
    SELECT 
        CASE 
            WHEN scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR) 
                AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
            THEN 'YES - Within window'
            ELSE 'NO - Outside window'
        END as meets_conditions,
        DATE_FORMAT(scheduled_at, '%H:%i:%s') as scheduled_time,
        DATE_FORMAT(NOW(), '%H:%i:%s') as current_time
    FROM broadcast_messages
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(check_query)
    result = cursor.fetchone()
    
    print(f"\nMeets conditions: {result[0]}")
    print(f"Scheduled: {result[1]} UTC")
    print(f"Current: {result[2]} UTC")
    
    print("\n⚠️ However, the processor appears to be STOPPED!")
    print("Last message was sent over 2 hours ago.")
    print("The application needs to be restarted.")
    
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
