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
    
    # First, let's see what we actually have
    query1 = """
    SELECT 
        recipient_phone,
        COUNT(*) as total_messages,
        COUNT(DISTINCT sequence_id) as unique_sequences,
        COUNT(DISTINCT sequence_stepid) as unique_steps,
        COUNT(DISTINCT campaign_id) as unique_campaigns,
        GROUP_CONCAT(DISTINCT status) as statuses
    FROM broadcast_messages
    WHERE device_id = %s
    GROUP BY recipient_phone
    HAVING COUNT(*) > 1
    ORDER BY total_messages DESC
    LIMIT 5
    """
    
    df1 = pd.read_sql(query1, connection, params=[device_id])
    
    print("=== DUPLICATE MESSAGES BY RECIPIENT FOR DEVICE SCHQ-S105 ===")
    print(df1.to_string(index=False))
    
    # Let's look at one specific phone number in detail
    if len(df1) > 0:
        sample_phone = df1.iloc[0]['recipient_phone']
        
        query2 = """
        SELECT 
            id,
            recipient_phone,
            sequence_id,
            sequence_stepid,
            status,
            created_at
        FROM broadcast_messages
        WHERE device_id = %s AND recipient_phone = %s
        ORDER BY created_at, id
        """
        
        df2 = pd.read_sql(query2, connection, params=[device_id, sample_phone])
        
        print(f"\n=== DETAILED VIEW FOR PHONE {sample_phone} ===")
        print(df2.to_string(index=False))
        
        # Check if any have duplicate sequence_stepid
        print("\n=== CHECKING FOR DUPLICATE SEQUENCE STEPS ===")
        step_counts = df2['sequence_stepid'].value_counts()
        duplicated_steps = step_counts[step_counts > 1]
        
        if len(duplicated_steps) > 0:
            print("Found duplicate sequence steps:")
            print(duplicated_steps)
        else:
            print("No duplicate sequence steps found - each message has a unique step ID")
            
        # Now let's create a proper delete query for actual duplicates
        # Delete records where same phone has multiple pending messages for same sequence
        print("\n=== PREPARING TO REMOVE DUPLICATES ===")
        
        # Count how many we would delete
        count_query = """
        WITH duplicates AS (
            SELECT 
                id,
                recipient_phone,
                sequence_id,
                status,
                ROW_NUMBER() OVER (
                    PARTITION BY recipient_phone, sequence_id 
                    ORDER BY 
                        CASE WHEN status = 'sent' THEN 1 
                             WHEN status = 'failed' THEN 2 
                             ELSE 3 END,
                        created_at
                ) as rn
            FROM broadcast_messages
            WHERE device_id = %s
            AND status = 'pending'
        )
        SELECT COUNT(*) as duplicates_to_remove
        FROM duplicates
        WHERE rn > 1
        """
        
        cursor = connection.cursor()
        cursor.execute(count_query, (device_id,))
        result = cursor.fetchone()
        
        if result and result[0] > 0:
            print(f"Found {result[0]} duplicate pending messages to remove")
            
            confirm = input("\nDo you want to remove these duplicates? (yes/no): ")
            
            if confirm.lower() == 'yes':
                # MySQL doesn't support CTEs in DELETE, so we need a different approach
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
                    GROUP BY recipient_phone, sequence_id
                    HAVING COUNT(*) > 1
                ) keep_records
                ON bm1.recipient_phone = keep_records.recipient_phone
                AND bm1.sequence_id = keep_records.sequence_id
                WHERE bm1.device_id = %s
                AND bm1.status = 'pending'
                AND bm1.created_at > keep_records.min_created
                """
                
                cursor.execute(delete_query, (device_id, device_id))
                deleted = cursor.rowcount
                connection.commit()
                
                print(f"\n✅ Successfully deleted {deleted} duplicate pending records!")
            else:
                print("\nDeletion cancelled.")
        else:
            print("No duplicate pending messages found to remove")
            
except Exception as e:
    print(f"Error: {e}")
    connection.rollback()
    
finally:
    connection.close()
    print("\nDatabase connection closed.")
