import mysql.connector
import time
import sys

sys.stdout.reconfigure(encoding='utf-8')

config = {
    'host': '159.89.198.71',
    'port': 3306,
    'database': 'admin_railway',
    'user': 'admin_aqil',
    'password': 'admin_aqil'
}

# Wait 10 seconds
print("Waiting 10 seconds for processor to pick up the message...")
time.sleep(10)

try:
    conn = mysql.connector.connect(**config)
    cursor = conn.cursor()
    
    print("\n=== CHECKING TEST MESSAGE STATUS ===\n")
    
    query = """
    SELECT 
        id,
        status,
        processing_worker_id,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
        error_message
    FROM broadcast_messages 
    WHERE id = '9d36c1a5-3bd3-468d-a5f6-43db174f58e9'
    """
    
    cursor.execute(query)
    result = cursor.fetchone()
    
    if result:
        print(f"ID: {result[0]}")
        print(f"Status: {result[1]}")
        print(f"Worker ID: {result[2] or 'NULL ❌'}")
        print(f"Sent at: {result[3] or 'Not sent'}")
        print(f"Error: {result[4] or 'None'}")
        
        if result[2] is None:
            print("\n❌ WORKER ID IS STILL NULL!")
            print("This confirms the OLD CODE is running on the server.")
            print("The fix has NOT been deployed.")
        else:
            print(f"\n✅ WORKER ID SET: {result[2]}")
            print("The fix is working!")
    else:
        print("Message not found!")
        
except Exception as e:
    print(f"Error: {e}")
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
