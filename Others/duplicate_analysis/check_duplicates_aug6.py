import pymysql
from datetime import datetime, timedelta
import pandas as pd

# MySQL connection
connection = pymysql.connect(
    host='159.89.198.71',
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    charset='utf8mb4'
)

try:
    # Check for messages sent between 1:20 PM and 1:30 PM today (Aug 6)
    query = """
    SELECT 
        id,
        recipient_phone,
        recipient_name,
        device_id,
        campaign_id,
        sequence_id,
        sequence_stepid,
        status,
        created_at,
        updated_at,
        sent_at,
        scheduled_at,
        content
    FROM broadcast_messages
    WHERE recipient_phone = '+60179075761'
    AND DATE(created_at) = '2025-08-06'
    AND TIME(created_at) BETWEEN '13:20:00' AND '13:30:00'
    ORDER BY created_at ASC
    """
    
    df = pd.read_sql(query, connection)
    
    print(f"Found {len(df)} messages for +60179075761 between 1:20 PM and 1:30 PM")
    print("\nDetailed information:")
    
    for idx, row in df.iterrows():
        print(f"\n--- Message {idx + 1} ---")
        print(f"ID: {row['id']}")
        print(f"Status: {row['status']}")
        print(f"Campaign ID: {row['campaign_id']}")
        print(f"Sequence ID: {row['sequence_id']}")
        print(f"Created: {row['created_at']}")
        print(f"Updated: {row['updated_at']}")
        print(f"Sent: {row['sent_at']}")
        print(f"Scheduled: {row['scheduled_at']}")
        print(f"Content preview: {row['content'][:100] if row['content'] else 'None'}...")
    
    # Check duplicate prevention status
    print("\n\nChecking duplicate patterns:")
    
    # Group by campaign/sequence
    if len(df) > 0:
        if df['campaign_id'].notna().any():
            campaign_groups = df.groupby('campaign_id').size()
            print("\nMessages per campaign:")
            print(campaign_groups)
        
        if df['sequence_id'].notna().any():
            sequence_groups = df.groupby(['sequence_id', 'sequence_stepid']).size()
            print("\nMessages per sequence step:")
            print(sequence_groups)
    
    # Check time differences
    if len(df) > 1:
        print("\nTime differences between messages:")
        for i in range(1, len(df)):
            time_diff = df.iloc[i]['created_at'] - df.iloc[i-1]['created_at']
            print(f"Message {i} to {i+1}: {time_diff}")

finally:
    connection.close()
