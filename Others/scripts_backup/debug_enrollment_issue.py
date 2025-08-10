import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Check why leads aren't enrolling
    print("\n1. Checking WARMEXAMA leads and their enrollment status...")
    cur.execute("""
        SELECT 
            l.phone, 
            l.trigger as lead_trigger,
            COUNT(sc.id) as enrollment_count,
            STRING_AGG(sc.current_step::text || ':' || sc.status, ', ') as enrollments
        FROM leads l
        LEFT JOIN sequence_contacts sc ON sc.contact_phone = l.phone
        WHERE l.trigger = 'WARMEXAMA'
        GROUP BY l.phone, l.trigger
        ORDER BY l.phone
    """)
    
    results = cur.fetchall()
    print(f"Found {len(results)} WARMEXAMA leads:")
    for phone, trigger, count, enrollments in results:
        print(f"  Phone: {phone}")
        print(f"    Trigger: {trigger}")
        print(f"    Enrollments: {count}")
        if enrollments:
            print(f"    Details: {enrollments}")
    
    # Check the exact NOT EXISTS condition
    print("\n2. Checking the NOT EXISTS condition...")
    cur.execute("""
        SELECT l.phone, sc.current_step, sc.status, sc.sequence_stepid
        FROM leads l
        LEFT JOIN sequence_contacts sc ON sc.contact_phone = l.phone
            AND sc.sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
            AND sc.current_step = 1
        WHERE l.trigger = 'WARMEXAMA'
        AND sc.id IS NOT NULL
    """)
    
    blocking_enrollments = cur.fetchall()
    if blocking_enrollments:
        print(f"Found {len(blocking_enrollments)} leads with current_step = 1:")
        for phone, step, status, stepid in blocking_enrollments:
            print(f"  Phone: {phone}, Step: {step}, Status: {status}, StepID: {stepid}")
    else:
        print("No leads have current_step = 1 - they should be eligible for enrollment!")
    
    # Check if the issue is with sequence_stepid
    print("\n3. Let's test the actual enrollment logic...")
    
    # Get a lead that should be enrolled
    cur.execute("""
        SELECT l.id, l.phone, l.name, l.device_id, l.user_id
        FROM leads l
        WHERE l.trigger = 'WARMEXAMA'
        AND NOT EXISTS (
            SELECT 1 FROM sequence_contacts sc
            WHERE sc.contact_phone = l.phone
            AND sc.current_step = 1
        )
        LIMIT 1
    """)
    
    lead = cur.fetchone()
    if lead:
        lead_id, phone, name, device_id, user_id = lead
        print(f"Testing enrollment for: {phone}")
        
        # Get sequence and steps
        cur.execute("""
            SELECT s.id, ss.id, ss.day_number, ss.trigger
            FROM sequences s
            JOIN sequence_steps ss ON ss.sequence_id = s.id
            WHERE s.name = 'WARM Sequence'
            AND ss.is_entry_point = true
        """)
        
        seq_info = cur.fetchone()
        if seq_info:
            seq_id, step_id, day_num, trigger = seq_info
            
            # Try to insert exactly as the Go code would
            try:
                cur.execute("""
                    INSERT INTO sequence_contacts (
                        sequence_id, contact_phone, contact_name, 
                        current_step, status, completed_at, current_trigger,
                        next_trigger_time, sequence_stepid, assigned_device_id
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                    ON CONFLICT (sequence_id, contact_phone, sequence_stepid) DO NOTHING
                    RETURNING id
                """, (
                    seq_id,           # sequence_id
                    phone,            # contact_phone
                    name,             # contact_name
                    day_num,          # current_step (day_number)
                    'active',         # status
                    'NOW()',          # completed_at
                    trigger,          # current_trigger
                    'NOW()',          # next_trigger_time
                    step_id,          # sequence_stepid (UUID)
                    device_id         # assigned_device_id
                ))
                
                result = cur.fetchone()
                if result:
                    print(f"SUCCESS: Created enrollment with ID: {result[0]}")
                    # Don't delete - let's keep it for real
                    conn.commit()
                else:
                    print("ON CONFLICT prevented insertion - checking why...")
                    
                    # Check if this exact combination exists
                    cur.execute("""
                        SELECT id, status, current_step
                        FROM sequence_contacts
                        WHERE sequence_id = %s 
                        AND contact_phone = %s 
                        AND sequence_stepid = %s
                    """, (seq_id, phone, step_id))
                    
                    existing = cur.fetchone()
                    if existing:
                        print(f"Found existing enrollment: ID={existing[0]}, Status={existing[1]}, Step={existing[2]}")
                    
            except Exception as e:
                print(f"ERROR during insertion: {e}")
                print(f"  seq_id type: {type(seq_id)}")
                print(f"  step_id type: {type(step_id)}")
                print(f"  device_id type: {type(device_id)}")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.close()
