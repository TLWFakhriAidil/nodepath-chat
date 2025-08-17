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
    
    print("=== CHECKING FOR DUPLICATES BY sequence_stepid + recipient_phone + device_id ===\n")
    
    # Check for duplicates with the exact combination
    query = """
    SELECT 
        sequence_stepid,
        recipient_phone,
        device_id,
        COUNT(*) as duplicate_count,
        GROUP_CONCAT(id) as message_ids,
        GROUP_CONCAT(status) as statuses,
        GROUP_CONCAT(DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s')) as created_times,
        GROUP_CONCAT(DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s')) as sent_times
    FROM broadcast_messages
    WHERE sequence_stepid IS NOT NULL
    GROUP BY sequence_stepid, recipient_phone, device_id
    HAVING COUNT(*) > 1
    ORDER BY duplicate_count DESC, recipient_phone
    """
    
    cursor.execute(query)
    duplicates = cursor.fetchall()
    
    if duplicates:
        print(f"❌ FOUND {len(duplicates)} DUPLICATE COMBINATIONS:\n")
        for i, dup in enumerate(duplicates, 1):
            print(f"--- Duplicate #{i} ---")
            print(f"Sequence Step ID: {dup['sequence_stepid']}")
            print(f"Phone: {dup['recipient_phone']}")
            print(f"Device ID: {dup['device_id']}")
            print(f"Count: {dup['duplicate_count']} duplicates")
            print(f"Message IDs: {dup['message_ids']}")
            print(f"Statuses: {dup['statuses']}")
            print(f"Created times: {dup['created_times']}")
            print(f"Sent times: {dup['sent_times']}")
            print()
    else:
        print("✅ NO DUPLICATES FOUND!")
        print("All messages have unique combination of sequence_stepid + recipient_phone + device_id")
    
    # Also check specifically for the phone number 60128198574
    print("\n=== CHECKING SPECIFIC PHONE NUMBER 60128198574 ===\n")
    
    query2 = """
    SELECT 
        sequence_stepid,
        device_id,
        COUNT(*) as count,
        GROUP_CONCAT(id) as message_ids,
        GROUP_CONCAT(status) as statuses
    FROM broadcast_messages
    WHERE recipient_phone LIKE '%128198574%'
      AND sequence_stepid IS NOT NULL
    GROUP BY sequence_stepid, device_id
    ORDER BY count DESC
    """
    
    cursor.execute(query2)
    phone_results = cursor.fetchall()
    
    if phone_results:
        print(f"Found {len(phone_results)} unique sequence steps for this phone:")
        has_duplicates = False
        for result in phone_results:
            if result['count'] > 1:
                has_duplicates = True
                print(f"\n❌ DUPLICATE FOUND:")
                print(f"Sequence Step: {result['sequence_stepid']}")
                print(f"Device: {result['device_id']}")
                print(f"Count: {result['count']}")
                print(f"Message IDs: {result['message_ids']}")
                print(f"Statuses: {result['statuses']}")
        
        if not has_duplicates:
            print("✅ No duplicates for phone 60128198574")
            print("Each sequence_stepid + device_id combination appears only once")
    else:
        print("No sequence messages found for this phone number")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
