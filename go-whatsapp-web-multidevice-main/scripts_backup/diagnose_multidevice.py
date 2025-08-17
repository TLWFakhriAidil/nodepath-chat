import psycopg2
import requests
import json

conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(conn_string)
    cur = conn.cursor()
    
    print("=== MULTI-DEVICE CONNECTION ISSUE DIAGNOSIS ===")
    
    # Get all online WhatsApp devices (not platform devices)
    cur.execute("""
        SELECT id, phone, jid, status, last_seen
        FROM user_devices 
        WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
        AND platform IS NULL
        AND status = 'online'
        ORDER BY last_seen DESC
    """)
    
    devices = cur.fetchall()
    print(f"\nFound {len(devices)} online WhatsApp devices:")
    
    for device in devices:
        print(f"\nDevice: {device[0]}")
        print(f"  Phone: {device[1]}")
        print(f"  JID: {device[2]}")
        print(f"  Status: {device[3]}")
        print(f"  Last Seen: {device[4]}")
    
    print("\n=== THE PROBLEM ===")
    print("When multiple devices connect:")
    print("1. Each device gets registered with ClientManager")
    print("2. But something is causing previous clients to be removed")
    print("3. Only the last connected device remains in ClientManager")
    
    print("\n=== SOLUTION NEEDED ===")
    print("We need to ensure ClientManager maintains ALL device clients")
    print("Not just the most recent one")
    
    # Check if there are any patterns in the JIDs
    print("\n=== CHECKING JID PATTERNS ===")
    cur.execute("""
        SELECT phone, COUNT(*) as device_count, 
               STRING_AGG(id, ', ') as device_ids
        FROM user_devices 
        WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
        AND platform IS NULL
        AND phone IS NOT NULL
        GROUP BY phone
        HAVING COUNT(*) > 1
    """)
    
    duplicates = cur.fetchall()
    if duplicates:
        print("Found devices with same phone number:")
        for dup in duplicates:
            print(f"Phone {dup[0]}: {dup[1]} devices")
            print(f"  Device IDs: {dup[2]}")
    else:
        print("No duplicate phone numbers found")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
