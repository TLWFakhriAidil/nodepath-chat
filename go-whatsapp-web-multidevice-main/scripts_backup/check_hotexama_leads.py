import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Checking HOTEXAMA Leads ===\n")
    
    # Check HOTEXAMA leads
    cur.execute("""
        SELECT 
            COUNT(*) as total,
            COUNT(CASE WHEN device_id IS NOT NULL THEN 1 END) as with_device,
            COUNT(CASE WHEN user_id IS NOT NULL THEN 1 END) as with_user,
            COUNT(CASE WHEN device_id IS NOT NULL AND user_id IS NOT NULL THEN 1 END) as with_both
        FROM leads 
        WHERE trigger = 'HOTEXAMA'
    """)
    
    result = cur.fetchone()
    print(f"Total HOTEXAMA leads: {result[0]}")
    print(f"With device_id: {result[1]}")
    print(f"With user_id: {result[2]}")
    print(f"With both: {result[3]}")
    
    # Check if HOT sequence exists and is active
    print("\n=== Checking HOT Sequence ===")
    cur.execute("""
        SELECT s.id, s.name, s.is_active, ss.trigger, ss.is_entry_point
        FROM sequences s
        JOIN sequence_steps ss ON s.id = ss.sequence_id
        WHERE ss.trigger = 'HOTEXAMA'
    """)
    
    sequences = cur.fetchall()
    for seq in sequences:
        print(f"Sequence: {seq[1]} (ID: {seq[0]})")
        print(f"  Active: {seq[2]}")
        print(f"  Entry trigger: {seq[3]}")
        print(f"  Is entry point: {seq[4]}")
    
    # Check if any HOTEXAMA leads already have messages
    print("\n=== Checking Existing Messages ===")
    cur.execute("""
        SELECT COUNT(DISTINCT l.phone)
        FROM leads l
        JOIN broadcast_messages bm ON l.phone = bm.recipient_phone
        WHERE l.trigger = 'HOTEXAMA'
        AND bm.status IN ('pending', 'sent')
    """)
    
    count = cur.fetchone()[0]
    print(f"HOTEXAMA leads with existing messages: {count}")
    
    # Sample some HOTEXAMA leads
    print("\n=== Sample HOTEXAMA Leads ===")
    cur.execute("""
        SELECT phone, device_id, user_id,
               (SELECT COUNT(*) FROM broadcast_messages bm 
                WHERE bm.recipient_phone = l.phone) as msg_count
        FROM leads l
        WHERE trigger = 'HOTEXAMA'
        LIMIT 5
    """)
    
    leads = cur.fetchall()
    for lead in leads:
        print(f"Phone: {lead[0]}")
        print(f"  device_id: {lead[1]}")
        print(f"  user_id: {lead[2]}")
        print(f"  existing messages: {lead[3]}")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
