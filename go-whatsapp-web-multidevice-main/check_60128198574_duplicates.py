import mysql.connector
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
    
    print("=== CHECKING DUPLICATES FOR 60128198574 BY SEQUENCE_STEPID ===\n")
    
    # Check for duplicates for this specific phone number
    query = """
    SELECT 
        sequence_stepid,
        COUNT(*) as count,
        GROUP_CONCAT(id) as message_ids,
        GROUP_CONCAT(status) as statuses,
        GROUP_CONCAT(DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')) as created_times,
        GROUP_CONCAT(DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s')) as sent_times,
        GROUP_CONCAT(DATE_FORMAT(DATE_ADD(sent_at, INTERVAL 8 HOUR), '%H:%i:%s')) as sent_times_myt,
        GROUP_CONCAT(processing_worker_id) as worker_ids,
        LEFT(MAX(content), 100) as content_preview
    FROM broadcast_messages
    WHERE recipient_phone LIKE '%128198574%'
        AND sequence_stepid IS NOT NULL
        AND DATE(created_at) = CURDATE()
    GROUP BY sequence_stepid
    ORDER BY MAX(created_at) DESC
    """
    
    cursor.execute(query)
    results = cursor.fetchall()
    
    if results:
        duplicate_found = False
        for result in results:
            if result['count'] > 1:
                duplicate_found = True
                print(f"❌ DUPLICATE FOUND!")
                print(f"Sequence Step ID: {result['sequence_stepid']}")
                print(f"Count: {result['count']} messages")
                print(f"Message IDs: {result['message_ids']}")
                print(f"Statuses: {result['statuses']}")
                print(f"Created at: {result['created_times']}")
                print(f"Sent at UTC: {result['sent_times']}")
                print(f"Sent at MYT (+8): {result['sent_times_myt']}")
                print(f"Worker IDs: {result['worker_ids']}")
                print(f"Content: {result['content_preview']}...")
                print("-" * 80)
        
        if not duplicate_found:
            print("✅ No duplicates found for 60128198574")
            print(f"\nShowing all {len(results)} unique messages:")
            for result in results:
                print(f"\nStep ID: {result['sequence_stepid']}")
                print(f"Sent at MYT: {result['sent_times_myt']}")
                print(f"Status: {result['statuses']}")
                print(f"Content: {result['content_preview'][:50]}...")
    else:
        print("No messages found for 60128198574 today")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
