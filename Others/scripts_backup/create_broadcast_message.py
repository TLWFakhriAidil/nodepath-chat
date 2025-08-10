import psycopg2
import uuid

conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(conn_string)
    cur = conn.cursor()
    
    print("=== CREATING BROADCAST MESSAGE MANUALLY ===")
    
    device_id = "d409cadc-75e2-4004-a789-c2bad0b31393"
    
    # Get lead phone number only
    cur.execute("""
        SELECT phone 
        FROM leads 
        WHERE device_id = %s 
        AND niche = 'GRR' 
        AND target_status = 'prospect'
        LIMIT 1
    """, (device_id,))
    
    lead_phone = cur.fetchone()[0]
    print(f"Creating message for phone: {lead_phone}")
    
    # Get campaign message
    cur.execute("SELECT message FROM campaigns WHERE id = 59")
    message = cur.fetchone()[0]
    print(f"Message content: {message}")
    
    # Generate UUID
    msg_id = str(uuid.uuid4())
    
    # Insert broadcast message
    cur.execute("""
        INSERT INTO broadcast_messages 
        (id, user_id, device_id, campaign_id, recipient_phone, 
         message_type, content, status, scheduled_at, created_at)
        VALUES 
        (%s, 'de078f16-3266-4ab3-8153-a248b015228f', %s, 59, %s, 
         'text', %s, 'pending', NOW(), NOW())
    """, (msg_id, device_id, lead_phone, message))
    
    conn.commit()
    
    print(f"\nSUCCESS! Created broadcast message:")
    print(f"- ID: {msg_id}")
    print(f"- Device: {device_id}")
    print(f"- Recipient: {lead_phone}")
    print(f"- Status: pending")
    print("\nThe broadcast processor should pick this up within 2 seconds!")
    
    # Verify it was created
    cur.execute("""
        SELECT COUNT(*) 
        FROM broadcast_messages 
        WHERE id = %s
    """, (msg_id,))
    
    if cur.fetchone()[0] == 1:
        print("\nVerified: Message is in the database and ready to be processed!")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
