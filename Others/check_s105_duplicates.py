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
    # First check if any device contains S105
    query1 = """
    SELECT DISTINCT device_id 
    FROM leads 
    WHERE device_id LIKE '%S105%'
    ORDER BY device_id
    """
    
    df_devices = pd.read_sql(query1, connection)
    print(f"Found {len(df_devices)} device(s) containing 'S105':")
    if len(df_devices) > 0:
        for device_id in df_devices['device_id']:
            print(f"  - {device_id}")
    else:
        print("  No devices found with S105 in the ID")
        
    # Check all device IDs to see the pattern
    print("\nChecking all unique device IDs (first 20):")
    query2 = "SELECT DISTINCT device_id FROM leads ORDER BY device_id LIMIT 20"
    df_all_devices = pd.read_sql(query2, connection)
    for i, device_id in enumerate(df_all_devices['device_id']):
        print(f"  {i+1}. {device_id}")
    
    # Look for duplicates by phone number across all devices
    print("\nChecking for duplicate phone numbers across all devices:")
    query3 = """
    SELECT phone, COUNT(*) as count, GROUP_CONCAT(DISTINCT device_id) as devices
    FROM leads
    GROUP BY phone
    HAVING COUNT(*) > 1
    ORDER BY count DESC
    LIMIT 20
    """
    
    df_duplicates = pd.read_sql(query3, connection)
    if len(df_duplicates) > 0:
        print(f"\nFound {len(df_duplicates)} phone numbers with duplicates:")
        print(df_duplicates.to_string(index=False))
    else:
        print("No duplicate phone numbers found")
        
    # Check for exact device pattern
    print("\nLooking for devices with pattern containing 'SCHQ':")
    query4 = """
    SELECT device_id, COUNT(*) as total_leads, COUNT(DISTINCT phone) as unique_phones
    FROM leads
    WHERE device_id LIKE '%SCHQ%'
    GROUP BY device_id
    ORDER BY device_id
    """
    df_schq = pd.read_sql(query4, connection)
    if len(df_schq) > 0:
        print(f"Found {len(df_schq)} devices with SCHQ pattern:")
        print(df_schq.to_string(index=False))
    else:
        print("No devices found with SCHQ pattern")

finally:
    connection.close()
    print("\nDatabase connection closed.")
