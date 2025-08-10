import psycopg2

conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(conn_string)
    cur = conn.cursor()
    
    print("=== DEBUGGING CAMPAIGN 59 EXECUTION ===")
    
    # Get campaign details
    cur.execute("""
        SELECT user_id, niche, target_status 
        FROM campaigns 
        WHERE id = 59
    """)
    campaign = cur.fetchone()
    user_id = campaign[0]
    niche = campaign[1]
    target_status = campaign[2]
    
    print(f"Campaign 59:")
    print(f"- User ID: {user_id}")
    print(f"- Niche: {niche}")
    print(f"- Target Status: {target_status}")
    
    # Check devices for this user
    print(f"\n=== DEVICES FOR USER {user_id} ===")
    cur.execute("""
        SELECT id, phone, status, platform
        FROM user_devices 
        WHERE user_id = %s
    """, (user_id,))
    
    devices = cur.fetchall()
    for device in devices:
        print(f"Device: {device[0]}")
        print(f"  Phone: {device[1]}")
        print(f"  Status: {device[2]}")
        print(f"  Platform: {device[3]}")
        
        # Check if device is considered "connected" by the campaign processor
        if device[2] in ['connected', 'Connected', 'online', 'Online'] or device[3]:
            print(f"  ✓ Device is considered CONNECTED")
        else:
            print(f"  ✗ Device is NOT connected")
    
    # Check leads for each device
    print(f"\n=== LEADS MATCHING CAMPAIGN CRITERIA ===")
    for device in devices:
        device_id = device[0]
        cur.execute("""
            SELECT COUNT(*) 
            FROM leads 
            WHERE device_id = %s 
            AND niche = %s 
            AND target_status = %s
        """, (device_id, niche, target_status))
        
        lead_count = cur.fetchone()[0]
        print(f"Device {device_id}: {lead_count} matching leads")
        
        if lead_count > 0:
            # Show sample leads
            cur.execute("""
                SELECT phone, name 
                FROM leads 
                WHERE device_id = %s 
                AND niche = %s 
                AND target_status = %s
                LIMIT 3
            """, (device_id, niche, target_status))
            
            for lead in cur.fetchall():
                print(f"  - {lead[0]} ({lead[1]})")
    
    # Check if broadcast messages are being created
    print(f"\n=== CHECKING BROADCAST MESSAGE CREATION ===")
    cur.execute("""
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE user_id = %s 
        AND created_at > NOW() - INTERVAL '10 minutes'
    """, (user_id,))
    
    recent_count = cur.fetchone()[0]
    print(f"Broadcast messages created in last 10 minutes: {recent_count}")
    
    # Force trigger the campaign
    print("\n=== FORCING CAMPAIGN TRIGGER ===")
    print("To manually trigger the campaign, update its timestamp:")
    print(f"UPDATE campaigns SET updated_at = NOW() WHERE id = 59;")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
