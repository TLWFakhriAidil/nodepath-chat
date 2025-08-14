import psycopg2

conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(conn_string)
    cur = conn.cursor()
    
    print("=== CAMPAIGN 59 DETAILS ===")
    cur.execute("""
        SELECT id, title, status, user_id, niche, target_status, message, 
               campaign_date, time_schedule, created_at
        FROM campaigns 
        WHERE id = 59
    """)
    result = cur.fetchone()
    if result:
        print(f"ID: {result[0]}")
        print(f"Title: {result[1]}")
        print(f"Status: {result[2]}")
        print(f"User ID: {result[3]}")
        print(f"Niche: {result[4]}")
        print(f"Target Status: {result[5]}")
        print(f"Message: {result[6][:50]}...")
        print(f"Campaign Date: {result[7]}")
        print(f"Time Schedule: {result[8]}")
        print(f"Created At: {result[9]}")
    
    print("\n=== LEADS FOR THIS CAMPAIGN ===")
    # Check leads matching the campaign criteria
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE device_id = 'd409cadc-75e2-4004-a789-c2bad0b31393'
        AND niche = 'GRR'
        AND target_status = 'prospect'
    """)
    lead_count = cur.fetchone()[0]
    print(f"Matching leads: {lead_count}")
    
    # Show a sample lead
    cur.execute("""
        SELECT phone, name, device_id, niche, target_status
        FROM leads 
        WHERE device_id = 'd409cadc-75e2-4004-a789-c2bad0b31393'
        AND niche = 'GRR'
        AND target_status = 'prospect'
        LIMIT 5
    """)
    
    print("\nSample leads:")
    for row in cur.fetchall():
        print(f"Phone: {row[0]}, Name: {row[1]}, Device: {row[2]}, Niche: {row[3]}, Status: {row[4]}")
    
    # Check if the campaign trigger is running
    print("\n=== CHECKING CAMPAIGN TRIGGER CONDITIONS ===")
    cur.execute("""
        SELECT 
            CASE 
                WHEN campaign_date <= CURRENT_DATE AND time_schedule <= CURRENT_TIME THEN 'Should trigger NOW'
                ELSE 'Not time yet'
            END as trigger_status,
            campaign_date,
            time_schedule,
            CURRENT_DATE as today,
            CURRENT_TIME as now
        FROM campaigns 
        WHERE id = 59
    """)
    result = cur.fetchone()
    print(f"Trigger Status: {result[0]}")
    print(f"Campaign Date: {result[1]}, Time: {result[2]}")
    print(f"Current Date: {result[3]}, Time: {result[4]}")
    
    # Check if there are ANY broadcast messages
    cur.execute("SELECT COUNT(*) FROM broadcast_messages WHERE campaign_id = 59")
    msg_count = cur.fetchone()[0]
    print(f"\nTotal broadcast messages for campaign 59: {msg_count}")
    
    # Check the campaign processor query
    print("\n=== CHECKING CAMPAIGN PROCESSOR QUERY ===")
    cur.execute("""
        SELECT id, title, status, campaign_date, time_schedule
        FROM campaigns 
        WHERE status = 'pending' 
        AND campaign_date <= CURRENT_DATE
        AND user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
        ORDER BY campaign_date, time_schedule
    """)
    
    print("Campaigns ready to process:")
    for row in cur.fetchall():
        print(f"- ID: {row[0]}, Title: {row[1]}, Status: {row[2]}, Date: {row[3]}, Time: {row[4]}")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
