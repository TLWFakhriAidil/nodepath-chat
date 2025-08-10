import pymysql
from datetime import datetime

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
    
    # First, let's identify duplicates for device SCHQ-S105
    device_id = 'b2fdd012-8e14-4568-9ccb-c0d89263a8e5'
    
    print("=== DUPLICATE REMOVAL PROCESS FOR DEVICE SCHQ-S105 ===")
    print(f"Device ID: {device_id}")
    print("="*80)
    
    # Count duplicates before removal (only pending status)
    count_query = """
    SELECT COUNT(*) as total_duplicates
    FROM broadcast_messages bm1
    WHERE bm1.device_id = %s
    AND bm1.status = 'pending'
    AND bm1.sequence_stepid IS NOT NULL
    AND EXISTS (
        SELECT 1 
        FROM broadcast_messages bm2
        WHERE bm2.recipient_phone = bm1.recipient_phone
        AND bm2.device_id = bm1.device_id
        AND bm2.sequence_stepid = bm1.sequence_stepid
        AND bm2.status = 'pending'
        AND bm2.id != bm1.id
        AND bm2.created_at < bm1.created_at
    )
    """
    
    cursor.execute(count_query, (device_id,))
    result = cursor.fetchone()
    duplicates_count = result[0] if result else 0
    
    print(f"\nFound {duplicates_count} duplicate 'pending' records to remove")
    
    if duplicates_count > 0:
        # Show sample of duplicates before deletion
        sample_query = """
        SELECT bm1.id, bm1.recipient_phone, bm1.sequence_stepid, bm1.created_at
        FROM broadcast_messages bm1
        WHERE bm1.device_id = %s
        AND bm1.status = 'pending'
        AND bm1.sequence_stepid IS NOT NULL
        AND EXISTS (
            SELECT 1 
            FROM broadcast_messages bm2
            WHERE bm2.recipient_phone = bm1.recipient_phone
            AND bm2.device_id = bm1.device_id
            AND bm2.sequence_stepid = bm1.sequence_stepid
            AND bm2.status = 'pending'
            AND bm2.id != bm1.id
            AND bm2.created_at < bm1.created_at
        )
        LIMIT 5
        """
        
        cursor.execute(sample_query, (device_id,))
        samples = cursor.fetchall()
        
        print("\nSample of duplicates to be removed:")
        for sample in samples:
            print(f"  ID: {sample[0]}, Phone: {sample[1]}, Step: {sample[2]}, Created: {sample[3]}")
        
        # Ask for confirmation
        print(f"\nThis will delete {duplicates_count} duplicate 'pending' records.")
        confirm = input("Do you want to proceed? (yes/no): ")
        
        if confirm.lower() == 'yes':
            # Delete duplicates - keep the oldest record for each combination
            delete_query = """
            DELETE bm1 
            FROM broadcast_messages bm1
            INNER JOIN broadcast_messages bm2 
            ON bm1.recipient_phone = bm2.recipient_phone
            AND bm1.device_id = bm2.device_id
            AND bm1.sequence_stepid = bm2.sequence_stepid
            WHERE bm1.device_id = %s
            AND bm1.status = 'pending'
            AND bm2.status = 'pending'
            AND bm1.sequence_stepid IS NOT NULL
            AND bm1.created_at > bm2.created_at
            """
            
            cursor.execute(delete_query, (device_id,))
            deleted_count = cursor.rowcount
            
            # Commit the changes
            connection.commit()
            
            print(f"\n✅ Successfully deleted {deleted_count} duplicate records!")
            
            # Verify the results
            verify_query = """
            SELECT 
                COUNT(*) as total_messages,
                SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_count,
                COUNT(DISTINCT CONCAT(recipient_phone, '|', device_id, '|', COALESCE(sequence_stepid, ''))) as unique_combinations
            FROM broadcast_messages
            WHERE device_id = %s
            """
            
            cursor.execute(verify_query, (device_id,))
            result = cursor.fetchone()
            
            print("\n=== AFTER CLEANUP ===")
            print(f"Total messages: {result[0]}")
            print(f"Pending messages: {result[1]}")
            print(f"Unique combinations: {result[2]}")
            
        else:
            print("\nDeletion cancelled.")
    else:
        print("\nNo duplicate 'pending' records found for this device.")
        
except Exception as e:
    print(f"\nError: {e}")
    connection.rollback()
    
finally:
    cursor.close()
    connection.close()
    print("\nDatabase connection closed.")
