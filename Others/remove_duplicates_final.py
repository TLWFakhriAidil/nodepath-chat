import pymysql
import pandas as pd

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
    device_id = 'b2fdd012-8e14-4568-9ccb-c0d89263a8e5'
    cursor = connection.cursor()
    
    print("=== ANALYZING DUPLICATES FOR DEVICE SCHQ-S105 ===")
    print(f"Device ID: {device_id}")
    print("="*80)
    
    # Since each message has a unique sequence_stepid, the "duplicates" are actually
    # multiple messages from different sequences sent to the same recipient
    # Let's identify which ones are truly unwanted duplicates
    
    # Find pending messages that are duplicates within the same sequence
    query = """
    SELECT 
        bm1.id,
        bm1.recipient_phone,
        bm1.sequence_id,
        bm1.sequence_stepid,
        bm1.status,
        bm1.created_at
    FROM broadcast_messages bm1
    WHERE bm1.device_id = %s
    AND bm1.status = 'pending'
    AND EXISTS (
        SELECT 1 
        FROM broadcast_messages bm2
        WHERE bm2.recipient_phone = bm1.recipient_phone
        AND bm2.sequence_id = bm1.sequence_id
        AND bm2.device_id = bm1.device_id
        AND bm2.id != bm1.id
        AND (bm2.status = 'sent' OR (bm2.status = 'pending' AND bm2.created_at < bm1.created_at))
    )
    ORDER BY bm1.recipient_phone, bm1.sequence_id, bm1.created_at
    """
    
    cursor.execute(query, (device_id,))
    duplicates = cursor.fetchall()
    
    if duplicates:
        print(f"\nFound {len(duplicates)} pending messages that are duplicates")
        print("\nSample duplicates to be removed:")
        
        df = pd.DataFrame(duplicates, columns=['id', 'phone', 'sequence_id', 'step_id', 'status', 'created_at'])
        print(df.head(10).to_string(index=False))
        
        # Group by phone and sequence to show the pattern
        print("\n=== DUPLICATE PATTERN ===")
        summary_query = """
        SELECT 
            recipient_phone,
            sequence_id,
            COUNT(*) as total_messages,
            SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as sent,
            SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
            SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed
        FROM broadcast_messages
        WHERE device_id = %s
        AND recipient_phone IN (
            SELECT DISTINCT recipient_phone 
            FROM broadcast_messages 
            WHERE device_id = %s
            GROUP BY recipient_phone 
            HAVING COUNT(*) > 1
        )
        GROUP BY recipient_phone, sequence_id
        HAVING COUNT(*) > 1
        ORDER BY recipient_phone, sequence_id
        LIMIT 20
        """
        
        cursor.execute(summary_query, (device_id, device_id))
        summary = cursor.fetchall()
        
        print("\nMessages per recipient per sequence:")
        for row in summary:
            print(f"Phone: {row[0]}, Sequence: {row[1]}, Total: {row[2]} (Sent: {row[3]}, Pending: {row[4]}, Failed: {row[5]})")
        
        confirm = input(f"\nDo you want to delete {len(duplicates)} duplicate pending messages? (yes/no): ")
        
        if confirm.lower() == 'yes':
            # Delete the identified duplicates
            ids_to_delete = [dup[0] for dup in duplicates]
            
            # Delete in batches of 100
            deleted_total = 0
            batch_size = 100
            
            for i in range(0, len(ids_to_delete), batch_size):
                batch = ids_to_delete[i:i+batch_size]
                placeholders = ','.join(['%s'] * len(batch))
                delete_query = f"DELETE FROM broadcast_messages WHERE id IN ({placeholders})"
                
                cursor.execute(delete_query, batch)
                deleted_total += cursor.rowcount
                
            connection.commit()
            print(f"\n✅ Successfully deleted {deleted_total} duplicate pending messages!")
            
            # Verify results
            verify_query = """
            SELECT 
                COUNT(*) as total,
                SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
                SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as sent
            FROM broadcast_messages
            WHERE device_id = %s
            """
            cursor.execute(verify_query, (device_id,))
            result = cursor.fetchone()
            
            print(f"\n=== AFTER CLEANUP ===")
            print(f"Total messages: {result[0]}")
            print(f"Pending: {result[1]}")
            print(f"Sent: {result[2]}")
            
        else:
            print("\nDeletion cancelled.")
    else:
        print("\nNo duplicate pending messages found.")
        
        # Show current status
        status_query = """
        SELECT 
            status,
            COUNT(*) as count
        FROM broadcast_messages
        WHERE device_id = %s
        GROUP BY status
        """
        cursor.execute(status_query, (device_id,))
        status_results = cursor.fetchall()
        
        print("\n=== CURRENT MESSAGE STATUS ===")
        for status, count in status_results:
            print(f"{status}: {count}")
            
except Exception as e:
    print(f"Error: {e}")
    connection.rollback()
    
finally:
    cursor.close()
    connection.close()
    print("\nDatabase connection closed.")
