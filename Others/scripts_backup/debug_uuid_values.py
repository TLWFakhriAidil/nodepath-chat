import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    # Check for sequences with empty IDs
    print("=== Checking for empty sequence IDs ===")
    cur.execute("""
        SELECT id, name FROM sequences 
        WHERE id IS NULL OR id::text = ''
    """)
    
    if cur.rowcount > 0:
        print("Found sequences with empty IDs:")
        for row in cur.fetchall():
            print(f"  ID: '{row[0]}', Name: {row[1]}")
    else:
        print("No sequences with empty IDs")
    
    # Check sequence_steps
    print("\n=== Checking sequence_steps for empty IDs ===")
    cur.execute("""
        SELECT id, sequence_id, day_number 
        FROM sequence_steps 
        WHERE id IS NULL OR id::text = '' 
        OR sequence_id IS NULL OR sequence_id::text = ''
        LIMIT 10
    """)
    
    if cur.rowcount > 0:
        print("Found sequence_steps with empty IDs:")
        for row in cur.fetchall():
            print(f"  step_id: '{row[0]}', sequence_id: '{row[1]}', day: {row[2]}")
    else:
        print("No sequence_steps with empty IDs")
    
    # Check specific leads that are failing
    print("\n=== Checking specific failing leads ===")
    phones = ['60102203990', '601111158915']
    for phone in phones:
        cur.execute("""
            SELECT phone, device_id, user_id, trigger
            FROM leads 
            WHERE phone = %s
        """, (phone,))
        
        result = cur.fetchone()
        if result:
            print(f"\nLead {phone}:")
            print(f"  device_id: '{result[1]}'")
            print(f"  user_id: '{result[2]}'")
            print(f"  trigger: {result[3]}")
            
            # Check if UUIDs are valid
            if result[1]:
                try:
                    cur.execute("SELECT %s::uuid", (result[1],))
                    print(f"  device_id is valid UUID")
                except:
                    print(f"  device_id is INVALID UUID!")
                    
            if result[2]:
                try:
                    cur.execute("SELECT %s::uuid", (result[2],))
                    print(f"  user_id is valid UUID")
                except:
                    print(f"  user_id is INVALID UUID!")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
