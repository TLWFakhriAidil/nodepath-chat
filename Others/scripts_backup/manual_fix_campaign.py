import psycopg2

conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(conn_string)
    cur = conn.cursor()
    
    print("=== FINAL DIAGNOSIS FOR CAMPAIGN 59 ===")
    
    device_id = "d409cadc-75e2-4004-a789-c2bad0b31393"
    
    # Check if this specific device has leads matching campaign criteria
    cur.execute("""
        SELECT COUNT(*) 
        FROM leads 
        WHERE device_id = %s 
        AND niche = 'GRR' 
        AND target_status = 'prospect'
    """, (device_id,))
    
    lead_count = cur.fetchone()[0]
    print(f"Device {device_id} has {lead_count} leads matching campaign criteria")
    
    # Check if messages were created in broadcast_messages table
    cur.execute("""
        SELECT id, status, created_at, error_message
        FROM broadcast_messages 
        WHERE device_id = %s 
        AND campaign_id = 59
        ORDER BY created_at DESC
        LIMIT 5
    """, (device_id,))
    
    messages = cur.fetchall()
    if messages:
        print(f"\nFound {len(messages)} broadcast messages for this device:")
        for msg in messages:
            print(f"- ID: {msg[0][:8]}, Status: {msg[1]}, Created: {msg[2]}, Error: {msg[3]}")
    else:
        print("\nNo broadcast messages found for this device and campaign 59")
    
    # The real fix - manually create the broadcast message
    if lead_count > 0 and not messages:
        print("\n=== MANUAL FIX ===")
        print("The campaign should have created messages but didn't.")
        print("Let's manually create a broadcast message:")
        
        # Get lead details
        cur.execute("""
            SELECT phone, name 
            FROM leads 
            WHERE device_id = %s 
            AND niche = 'GRR' 
            AND target_status = 'prospect'
            LIMIT 1
        """, (device_id,))
        
        lead = cur.fetchone()
        if lead:
            print(f"\nCreating message for lead: {lead[0]} ({lead[1]})")
            
            # Get campaign message
            cur.execute("SELECT message FROM campaigns WHERE id = 59")
            message = cur.fetchone()[0]
            
            # Insert broadcast message
            cur.execute("""
                INSERT INTO broadcast_messages 
                (user_id, device_id, campaign_id, recipient_phone, recipient_name, 
                 message_type, content, status, scheduled_at, created_at)
                VALUES 
                ('de078f16-3266-4ab3-8153-a248b015228f', %s, 59, %s, %s, 
                 'text', %s, 'pending', NOW(), NOW())
                RETURNING id
            """, (device_id, lead[0], lead[1], message))
            
            msg_id = cur.fetchone()[0]
            conn.commit()
            
            print(f"✓ Created broadcast message with ID: {msg_id}")
            print("The message should now be processed by the broadcast worker!")
    
    # Check current worker status
    print("\n=== CHECKING IF WORKERS ARE PROCESSING ===")
    cur.execute("""
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE status = 'pending' 
        AND scheduled_at <= NOW()
    """)
    pending = cur.fetchone()[0]
    print(f"Total pending messages ready to send: {pending}")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
