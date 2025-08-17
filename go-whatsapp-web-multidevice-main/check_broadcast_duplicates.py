import pymysql
import pandas as pd
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
    # First get the device ID for SCHQ-S105
    device_id = 'b2fdd012-8e14-4568-9ccb-c0d89263a8e5'
    print(f"Checking broadcast_messages for device SCHQ-S105 (ID: {device_id})")
    print("="*80)
    
    # Check for duplicate recipients in broadcast_messages
    query1 = """
    SELECT 
        recipient_phone,
        COUNT(*) as message_count,
        COUNT(DISTINCT campaign_id) as campaigns,
        COUNT(DISTINCT sequence_id) as sequences,
        COUNT(DISTINCT status) as different_statuses,
        GROUP_CONCAT(DISTINCT status ORDER BY status SEPARATOR ', ') as all_statuses,
        MIN(created_at) as first_message,
        MAX(created_at) as last_message
    FROM broadcast_messages
    WHERE device_id = %s
    GROUP BY recipient_phone
    HAVING COUNT(*) > 1
    ORDER BY message_count DESC
    """
    
    df_duplicates = pd.read_sql(query1, connection, params=[device_id])
    
    if len(df_duplicates) > 0:
        print(f"\n!!! FOUND {len(df_duplicates)} PHONE NUMBERS WITH DUPLICATE MESSAGES !!!")
        print(f"\nShowing duplicates (total found: {len(df_duplicates)}):")
        
        # Show summary of duplicates
        for idx, row in df_duplicates.head(10).iterrows():
            print(f"\nPhone: {row['recipient_phone']}")
            print(f"  - Total messages: {row['message_count']}")
            print(f"  - Campaigns: {row['campaigns']}, Sequences: {row['sequences']}")
            print(f"  - Statuses: {row['all_statuses']}")
            print(f"  - Date range: {row['first_message']} to {row['last_message']}")
        
        # Get more details for one duplicate
        top_phone = df_duplicates.iloc[0]['recipient_phone']
        query2 = """
        SELECT 
            id,
            recipient_phone,
            campaign_id,
            sequence_id,
            sequence_stepid,
            status,
            created_at,
            scheduled_at,
            sent_at
        FROM broadcast_messages
        WHERE device_id = %s AND recipient_phone = %s
        ORDER BY created_at DESC
        """
        
        df_details = pd.read_sql(query2, connection, params=[device_id, top_phone])
        print(f"\n=== SAMPLE DUPLICATE DETAILS FOR {top_phone} ===")
        for idx, row in df_details.iterrows():
            print(f"\nMessage {idx+1}:")
            print(f"  ID: {row['id']}")
            print(f"  Campaign ID: {row['campaign_id']}")
            print(f"  Sequence ID: {row['sequence_id']}")
            print(f"  Sequence Step ID: {row['sequence_stepid']}")
            print(f"  Status: {row['status']}")
            print(f"  Created: {row['created_at']}")
            
    # Check overall statistics
    query3 = """
    SELECT 
        COUNT(*) as total_messages,
        COUNT(DISTINCT recipient_phone) as unique_recipients,
        SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as sent_count,
        SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending_count,
        SUM(CASE WHEN status = 'failed' THEN 1 ELSE 0 END) as failed_count
    FROM broadcast_messages
    WHERE device_id = %s
    """
    
    df_stats = pd.read_sql(query3, connection, params=[device_id])
    print(f"\n=== OVERALL STATISTICS FOR SCHQ-S105 ===")
    stats = df_stats.iloc[0]
    print(f"Total messages: {stats['total_messages']}")
    print(f"Unique recipients: {stats['unique_recipients']}")
    print(f"Sent: {stats['sent_count']}")
    print(f"Pending: {stats['pending_count']}")
    print(f"Failed: {stats['failed_count']}")
    
    # Analyze the duplicate pattern
    query4 = """
    SELECT 
        DATE(created_at) as created_date,
        TIME(created_at) as created_time,
        COUNT(*) as messages_created,
        COUNT(DISTINCT recipient_phone) as unique_recipients
    FROM broadcast_messages
    WHERE device_id = %s
    GROUP BY DATE(created_at), TIME(created_at)
    HAVING messages_created > 50
    ORDER BY created_date DESC, created_time DESC
    LIMIT 10
    """
    
    df_pattern = pd.read_sql(query4, connection, params=[device_id])
    if len(df_pattern) > 0:
        print(f"\n=== HIGH VOLUME MESSAGE CREATION TIMES ===")
        print("(Times when many messages were created at once)")
        print(df_pattern.to_string(index=False))

finally:
    connection.close()
    print("\nDatabase connection closed.")
