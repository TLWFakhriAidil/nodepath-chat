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
    
    # Check the actual duplicate situation
    query = """
    SELECT 
        recipient_phone,
        sequence_stepid,
        COUNT(*) as count,
        GROUP_CONCAT(DISTINCT status ORDER BY status) as statuses,
        GROUP_CONCAT(id ORDER BY created_at) as message_ids
    FROM broadcast_messages
    WHERE device_id = %s
    AND sequence_stepid IS NOT NULL
    GROUP BY recipient_phone, sequence_stepid
    HAVING COUNT(*) > 1
    ORDER BY count DESC
    LIMIT 10
    """
    
    df = pd.read_sql(query, connection, params=[device_id])
    
    print("=== DUPLICATE ANALYSIS FOR DEVICE SCHQ-S105 ===")
    print(f"Device ID: {device_id}")
    print("="*80)
    
    if len(df) > 0:
        print(f"\nFound {len(df)} duplicate combinations (phone + sequence_stepid)")
        print("\nTop 10 duplicates:")
        for idx, row in df.iterrows():
            print(f"\nPhone: {row['recipient_phone']}")
            print(f"Sequence Step ID: {row['sequence_stepid']}")
            print(f"Count: {row['count']}")
            print(f"Statuses: {row['statuses']}")
            print(f"Message IDs: {row['message_ids'][:100]}...")
            
        # Count by status combination
        status_query = """
        SELECT 
            CASE 
                WHEN COUNT(DISTINCT status) = 1 THEN MAX(status)
                ELSE 'mixed'
            END as status_type,
            COUNT(*) as duplicate_groups
        FROM (
            SELECT 
                recipient_phone,
                sequence_stepid,
                GROUP_CONCAT(DISTINCT status ORDER BY status) as statuses
            FROM broadcast_messages
            WHERE device_id = %s
            AND sequence_stepid IS NOT NULL
            GROUP BY recipient_phone, sequence_stepid
            HAVING COUNT(*) > 1
        ) t
        GROUP BY status_type
        """
        
        df_status = pd.read_sql(status_query, connection, params=[device_id])
        print("\n=== DUPLICATE GROUPS BY STATUS ===")
        print(df_status.to_string(index=False))
        
        # Find duplicates that can be safely removed (all have same status)
        removable_query = """
        SELECT 
            recipient_phone,
            sequence_stepid,
            status,
            COUNT(*) as count,
            MIN(created_at) as oldest_created,
            MAX(created_at) as newest_created
        FROM broadcast_messages
        WHERE device_id = %s
        AND sequence_stepid IS NOT NULL
        GROUP BY recipient_phone, sequence_stepid, status
        HAVING COUNT(*) > 1
        ORDER BY count DESC
        """
        
        df_removable = pd.read_sql(removable_query, connection, params=[device_id])
        print(f"\n=== SAFELY REMOVABLE DUPLICATES ===")
        print(f"Found {len(df_removable)} groups with same status duplicates")
        
        if len(df_removable) > 0:
            print("\nBreakdown by status:")
            status_counts = df_removable.groupby('status')['count'].sum() - len(df_removable)
            for status, count in status_counts.items():
                print(f"  {status}: {count} duplicates can be removed")
    else:
        print("\nNo duplicates found based on recipient_phone + sequence_stepid")
        
except Exception as e:
    print(f"Error: {e}")
    
finally:
    connection.close()
