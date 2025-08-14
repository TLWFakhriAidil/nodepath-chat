import pymysql

# Connect to MySQL
connection = pymysql.connect(
    host='159.89.198.71',
    port=3306,
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    charset='utf8mb4'
)

try:
    cursor = connection.cursor()
    device_id = 'b2fdd012-8e14-4568-9ccb-c0d89263a8e5'
    
    print("=== REMOVING DUPLICATE PENDING MESSAGES FOR DEVICE SCHQ-S105 ===")
    print(f"Device ID: {device_id}")
    print("="*80)
    
    # Delete duplicates where same recipient has messages in same sequence
    # Keep the oldest one, delete newer pending duplicates
    delete_query = """
    DELETE bm1 FROM broadcast_messages bm1
    INNER JOIN (
        SELECT 
            recipient_phone,
            sequence_id,
            MIN(created_at) as min_created
        FROM broadcast_messages
        WHERE device_id = %s
        AND status = 'pending'
        AND sequence_id IS NOT NULL
        GROUP BY recipient_phone, sequence_id
        HAVING COUNT(*) > 1
    ) keep_records
    ON bm1.recipient_phone = keep_records.recipient_phone
    AND bm1.sequence_id = keep_records.sequence_id
    WHERE bm1.device_id = %s
    AND bm1.status = 'pending'
    AND bm1.created_at > keep_records.min_created
    """
    
    # Execute the deletion
    cursor.execute(delete_query, (device_id, device_id))
    deleted_count = cursor.rowcount
    
    # Commit the changes
    connection.commit()
    
    print(f"\nSuccessfully deleted {deleted_count} duplicate pending records!")
    
    # Verify results
    verify_query = """
    SELECT 
        COUNT(*) as total,
        SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
        SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as sent,
        SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
    FROM broadcast_messages
    WHERE device_id = %s
    """
    cursor.execute(verify_query, (device_id,))
    result = cursor.fetchone()
    
    print(f"\n=== AFTER CLEANUP ===")
    print(f"Total messages: {result[0]}")
    print(f"Pending: {result[1]}")
    print(f"Sent: {result[2]}")
    print(f"Failed: {result[3]}")
    
    # Check if any duplicates remain
    check_query = """
    SELECT 
        recipient_phone,
        COUNT(*) as count
    FROM broadcast_messages
    WHERE device_id = %s
    GROUP BY recipient_phone
    HAVING COUNT(*) > 1
    ORDER BY count DESC
    LIMIT 5
    """
    cursor.execute(check_query, (device_id,))
    remaining = cursor.fetchall()
    
    if remaining:
        print(f"\n=== REMAINING DUPLICATES (BY PHONE) ===")
        for phone, count in remaining:
            print(f"Phone: {phone}, Total messages: {count}")
    else:
        print("\nNo duplicate phone numbers remain!")
        
except Exception as e:
    print(f"Error: {e}")
    connection.rollback()
    
finally:
    cursor.close()
    connection.close()
    print("\nDatabase connection closed.")
