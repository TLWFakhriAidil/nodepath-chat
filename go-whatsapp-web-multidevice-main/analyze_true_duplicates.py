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
    
    # Let's analyze one specific phone with duplicates
    query = """
    SELECT 
        id,
        recipient_phone,
        sequence_id,
        sequence_stepid,
        status,
        created_at,
        scheduled_at
    FROM broadcast_messages
    WHERE device_id = %s 
    AND recipient_phone = '601125777864'
    ORDER BY created_at
    """
    
    df = pd.read_sql(query, connection, params=[device_id])
    
    print("=== ANALYSIS OF DUPLICATE MESSAGES FOR PHONE 601125777864 ===")
    print(f"\nTotal messages: {len(df)}")
    print("\nDetailed breakdown:")
    
    for idx, row in df.iterrows():
        print(f"\n{idx+1}. ID: {row['id']}")
        print(f"   Sequence: {row['sequence_id']}")
        print(f"   Step ID: {row['sequence_stepid']}")
        print(f"   Status: {row['status']}")
        print(f"   Created: {row['created_at']}")
        
    # Now let's see the pattern - are these truly duplicates?
    print("\n=== CHECKING FOR TRUE DUPLICATES ===")
    
    # Check if any messages have exact same sequence_stepid
    duplicate_check = df.groupby('sequence_stepid').size()
    true_duplicates = duplicate_check[duplicate_check > 1]
    
    if len(true_duplicates) > 0:
        print("\nFound TRUE duplicates (same sequence_stepid):")
        for stepid, count in true_duplicates.items():
            print(f"  Step ID {stepid}: {count} messages")
    else:
        print("\nNO TRUE DUPLICATES found!")
        print("Each message has a unique sequence_stepid.")
        print("\nThese are different sequence steps for the same recipient.")
        print("This is NORMAL behavior - sequences have multiple steps.")
        
    # Show sequence breakdown
    print("\n=== MESSAGES BY SEQUENCE ===")
    sequence_counts = df.groupby(['sequence_id', 'status']).size().reset_index(name='count')
    print(sequence_counts.to_string(index=False))
    
except Exception as e:
    print(f"Error: {e}")
    
finally:
    connection.close()
    print("\nDatabase connection closed.")
