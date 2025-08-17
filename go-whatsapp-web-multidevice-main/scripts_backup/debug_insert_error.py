import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    print("Connected to PostgreSQL\n")
    
    # Check for leads that are trying to enroll
    print("=== Checking leads that match enrollment criteria ===")
    cur.execute("""
        SELECT l.id, l.phone, l.device_id, l.user_id, l.trigger,
               s.id as sequence_id, ss.trigger as entry_trigger,
               ss.id as step_id
        FROM leads l
        CROSS JOIN sequences s
        INNER JOIN sequence_steps ss ON ss.sequence_id = s.id
        WHERE s.is_active = true 
            AND ss.is_entry_point = true
            AND l.trigger IS NOT NULL 
            AND l.trigger != ''
            AND l.device_id IS NOT NULL 
            AND l.user_id IS NOT NULL
            AND position(ss.trigger in l.trigger) > 0
            AND l.phone IN ('601111158915', '60199204337')
        LIMIT 5
    """)
    
    for row in cur.fetchall():
        print(f"\nLead phone: {row[1]}")
        print(f"  device_id: {row[2]}")
        print(f"  user_id: {row[3]}")
        print(f"  trigger: {row[4]}")
        print(f"  sequence_id: {row[5]}")
        print(f"  step_id: {row[7]}")
    
    # Check sequence_steps for potential empty IDs
    print("\n=== Checking sequence_steps for empty values ===")
    cur.execute("""
        SELECT id, sequence_id, day_number, trigger, content
        FROM sequence_steps
        WHERE sequence_id IN (
            SELECT id FROM sequences WHERE is_active = true
        )
        AND (id IS NULL OR sequence_id IS NULL OR id::text = '' OR sequence_id::text = '')
        LIMIT 5
    """)
    
    if cur.rowcount > 0:
        print("Found sequence_steps with NULL or empty IDs:")
        for row in cur.fetchall():
            print(f"  step_id: {row[0]}, sequence_id: {row[1]}, day: {row[2]}")
    else:
        print("No sequence_steps with NULL/empty IDs found")
    
    # Check for sequence_steps details
    print("\n=== Sample sequence_steps data ===")
    cur.execute("""
        SELECT ss.id, ss.sequence_id, ss.day_number, ss.trigger, ss.is_entry_point,
               s.name as sequence_name
        FROM sequence_steps ss
        JOIN sequences s ON s.id = ss.sequence_id
        WHERE s.is_active = true AND ss.is_entry_point = true
        LIMIT 3
    """)
    
    for row in cur.fetchall():
        print(f"\nSequence: {row[5]}")
        print(f"  step_id: {row[0]}")
        print(f"  sequence_id: {row[1]}")
        print(f"  trigger: {row[3]}")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
