import psycopg2

conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(conn_string)
    cur = conn.cursor()
    
    print("=== CHECKING CAMPAIGN ISSUE ===")
    
    # Check if d409cadc device has the lead
    device_id = "d409cadc-75e2-4004-a789-c2bad0b31393"
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE device_id = %s 
        AND niche = 'GRR' 
        AND target_status = 'prospect'
    """, (device_id,))
    
    count = cur.fetchone()[0]
    print(f"Device {device_id} has {count} matching leads")
    
    # Check why campaign might not be processing this device
    print("\n=== CHECKING CAMPAIGN PROCESSING ===")
    
    # Simulate the campaign query
    cur.execute("""
        SELECT d.id, d.platform, 
               (SELECT COUNT(*) FROM leads l 
                WHERE l.device_id = d.id 
                AND l.niche = 'GRR' 
                AND l.target_status = 'prospect') as matching_leads
        FROM user_devices d
        WHERE d.user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
        AND (d.status = 'online' OR d.platform IS NOT NULL)
        ORDER BY matching_leads DESC
        LIMIT 10
    """)
    
    print("\nTop 10 devices by matching leads:")
    for row in cur.fetchall():
        print(f"Device: {row[0][:8]}... Platform: {row[1] or 'WhatsApp'} Leads: {row[2]}")
    
    # Check if messages were created but failed
    print("\n=== CHECKING BROADCAST MESSAGES ===")
    cur.execute("""
        SELECT device_id, status, COUNT(*) 
        FROM broadcast_messages 
        WHERE campaign_id = 59
        GROUP BY device_id, status
    """)
    
    results = cur.fetchall()
    if results:
        print("Messages created for campaign 59:")
        for row in results:
            print(f"Device: {row[0][:8]}... Status: {row[1]} Count: {row[2]}")
    else:
        print("No messages found for campaign 59")
    
    # Check the exact campaign trigger time
    print("\n=== CAMPAIGN TRIGGER CHECK ===")
    cur.execute("""
        SELECT campaign_date, time_schedule, 
               (campaign_date || ' ' || time_schedule)::timestamp as scheduled_time,
               NOW() as current_time,
               CASE 
                   WHEN (campaign_date || ' ' || time_schedule)::timestamp <= NOW() 
                   THEN 'Should trigger'
                   ELSE 'Not yet'
               END as trigger_status
        FROM campaigns 
        WHERE id = 59
    """)
    
    result = cur.fetchone()
    print(f"Campaign Date: {result[0]}, Time: {result[1]}")
    print(f"Scheduled: {result[2]}")
    print(f"Current: {result[3]}")
    print(f"Status: {result[4]}")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
