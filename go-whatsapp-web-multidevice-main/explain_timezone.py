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
    
    print("=== UNDERSTANDING THE TIMEZONE LOGIC ===\n")
    
    # Check the actual SQL condition used in GetPendingMessagesAndLock
    print("The processor uses this condition:")
    print("scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)")
    print("\nThis means:")
    print("- Database stores times in UTC")
    print("- Malaysia time is UTC+8")
    print("- So it adds 8 hours to current UTC time to get Malaysia time")
    print("- Then checks if scheduled_at is before or equal to Malaysia time")
    
    # Let's see what this means for your message
    query = """
    SELECT 
        DATE_FORMAT(NOW(), '%Y-%m-%d %H:%i:%s') as current_utc,
        DATE_FORMAT(DATE_ADD(NOW(), INTERVAL 8 HOUR), '%Y-%m-%d %H:%i:%s') as current_myt,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_utc,
        DATE_FORMAT(DATE_ADD(scheduled_at, INTERVAL 8 HOUR), '%Y-%m-%d %H:%i:%s') as scheduled_myt,
        CASE 
            WHEN scheduled_at <= NOW() THEN 'Ready (UTC comparison)'
            ELSE 'Not ready (UTC comparison)'
        END as utc_check,
        CASE 
            WHEN scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR) THEN 'Ready (MYT comparison)'
            ELSE 'Not ready (MYT comparison)'
        END as myt_check
    FROM broadcast_messages
    WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'
    """
    
    cursor.execute(query)
    result = cursor.fetchone()
    
    if result:
        print("\n=== TIME COMPARISON ===")
        print(f"Current Time UTC: {result['current_utc']}")
        print(f"Current Time MYT: {result['current_myt']}")
        print(f"\nScheduled Time UTC: {result['scheduled_utc']}")
        print(f"Scheduled Time MYT: {result['scheduled_myt']}")
        print(f"\nUTC Check: {result['utc_check']}")
        print(f"MYT Check: {result['myt_check']}")
    
    # Show example
    print("\n=== EXAMPLE ===")
    print("If you want to send at 3:00 PM Malaysia time:")
    print("- Malaysia time: 15:00 (3:00 PM)")
    print("- Store in DB as: 07:00 UTC (15:00 - 8 hours)")
    print("- Processor checks: Is 07:00 <= DATE_ADD(NOW(), INTERVAL 8 HOUR)?")
    print("- When Malaysia time reaches 3:00 PM, this condition becomes true")
    
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
