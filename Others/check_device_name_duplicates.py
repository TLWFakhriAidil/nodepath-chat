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
    # Find device with name SCHQ-S105
    query1 = """
    SELECT id, device_name, phone, status, platform
    FROM user_devices 
    WHERE device_name = 'SCHQ-S105'
    """
    
    df_device = pd.read_sql(query1, connection)
    print("=== DEVICE SCHQ-S105 DETAILS ===")
    if len(df_device) > 0:
        print(df_device.to_string(index=False))
        device_id = df_device.iloc[0]['id']
        print(f"\nDevice ID: {device_id}")
        
        # Now check for duplicate leads for this device
        query2 = """
        SELECT phone, name, COUNT(*) as count
        FROM leads
        WHERE device_id = %s
        GROUP BY phone, name
        HAVING COUNT(*) > 1
        ORDER BY count DESC
        """
        
        df_duplicates = pd.read_sql(query2, connection, params=[device_id])
        print(f"\n=== DUPLICATE LEADS FOR DEVICE SCHQ-S105 ===")
        if len(df_duplicates) > 0:
            print(f"Found {len(df_duplicates)} duplicate phone/name combinations:")
            print(df_duplicates.to_string(index=False))
        else:
            print("No duplicate leads found for this device")
            
        # Get total lead statistics
        query3 = """
        SELECT 
            COUNT(*) as total_leads,
            COUNT(DISTINCT phone) as unique_phones,
            COUNT(DISTINCT name) as unique_names,
            COUNT(DISTINCT niche) as unique_niches,
            COUNT(DISTINCT status) as unique_statuses
        FROM leads
        WHERE device_id = %s
        """
        
        df_stats = pd.read_sql(query3, connection, params=[device_id])
        print(f"\n=== LEAD STATISTICS FOR DEVICE SCHQ-S105 ===")
        print(df_stats.to_string(index=False))
        
        # Check if same phone numbers exist with different data
        query4 = """
        SELECT phone, 
               COUNT(*) as occurrences,
               COUNT(DISTINCT name) as different_names,
               GROUP_CONCAT(DISTINCT name ORDER BY name SEPARATOR ' | ') as all_names,
               COUNT(DISTINCT niche) as different_niches,
               GROUP_CONCAT(DISTINCT niche ORDER BY niche SEPARATOR ' | ') as all_niches
        FROM leads
        WHERE device_id = %s
        GROUP BY phone
        HAVING COUNT(*) > 1
        ORDER BY occurrences DESC
        LIMIT 20
        """
        
        df_phone_dups = pd.read_sql(query4, connection, params=[device_id])
        print(f"\n=== DUPLICATE PHONE NUMBERS WITH DIFFERENT DATA ===")
        if len(df_phone_dups) > 0:
            print(f"Found {len(df_phone_dups)} phone numbers with multiple entries:")
            print("\n" + df_phone_dups.to_string(index=False))
        else:
            print("No phone numbers with multiple entries found")
            
        # Get sample of duplicate entries
        if len(df_phone_dups) > 0:
            sample_phone = df_phone_dups.iloc[0]['phone']
            query5 = """
            SELECT id, phone, name, niche, status, created_at
            FROM leads
            WHERE device_id = %s AND phone = %s
            ORDER BY created_at
            """
            df_sample = pd.read_sql(query5, connection, params=[device_id, sample_phone])
            print(f"\n=== SAMPLE DUPLICATE ENTRIES FOR PHONE {sample_phone} ===")
            print(df_sample.to_string(index=False))
            
    else:
        print("No device found with name 'SCHQ-S105'")
        
        # Check similar device names
        print("\n=== CHECKING SIMILAR DEVICE NAMES ===")
        query6 = """
        SELECT device_name, id, status, platform
        FROM user_devices
        WHERE device_name LIKE '%S105%' OR device_name LIKE '%SCHQ%'
        ORDER BY device_name
        """
        df_similar = pd.read_sql(query6, connection)
        if len(df_similar) > 0:
            print(f"Found {len(df_similar)} devices with similar names:")
            print(df_similar.to_string(index=False))
        else:
            print("No devices found with similar names")

finally:
    connection.close()
    print("\nDatabase connection closed.")
