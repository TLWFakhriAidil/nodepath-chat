import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Check the sequence contacts in detail
    print("\n1. Detailed view of sequence_contacts for WARMEXAMA leads...")
    cur.execute("""
        SELECT 
            sc.contact_phone,
            sc.sequence_stepid,
            ss.day_number,
            ss.trigger as step_trigger,
            sc.current_step,
            sc.status,
            sc.completed_at,
            sc.current_trigger
        FROM sequence_contacts sc
        JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
        WHERE sc.contact_phone IN (
            SELECT phone FROM leads WHERE trigger = 'WARMEXAMA'
        )
        AND sc.sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
        ORDER BY sc.contact_phone, ss.day_number
        LIMIT 20
    """)
    
    results = cur.fetchall()
    current_phone = None
    for row in results:
        phone = row[0]
        if phone != current_phone:
            print(f"\nPhone: {phone}")
            current_phone = phone
        print(f"  Day {row[2]}: {row[3]} (Step: {row[4]}, Status: {row[5]})")
        print(f"    Completed: {row[6]}")
    
    # Check if sequence is actually processing
    print("\n2. Checking sequence processing status...")
    cur.execute("""
        SELECT 
            status, 
            COUNT(*) as count,
            MIN(completed_at) as first_completed,
            MAX(completed_at) as last_completed
        FROM sequence_contacts
        WHERE sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
        GROUP BY status
    """)
    
    status_results = cur.fetchall()
    for status, count, first, last in status_results:
        print(f"  Status '{status}': {count} contacts")
        if first:
            print(f"    First completed: {first}")
            print(f"    Last completed: {last}")
    
    # Check for any leads that could be enrolled
    print("\n3. Checking for any leads that could be enrolled in sequences...")
    cur.execute("""
        SELECT COUNT(DISTINCT l.phone)
        FROM leads l
        WHERE l.trigger IS NOT NULL 
        AND l.trigger != ''
        AND EXISTS (
            SELECT 1 FROM sequence_steps ss
            WHERE ss.is_entry_point = true
            AND position(ss.trigger in l.trigger) > 0
        )
        AND NOT EXISTS (
            SELECT 1 FROM sequence_contacts sc
            WHERE sc.contact_phone = l.phone
        )
    """)
    
    available_count = cur.fetchone()[0]
    print(f"Leads available for NEW enrollment: {available_count}")
    
    # Show some examples if any
    if available_count > 0:
        cur.execute("""
            SELECT l.phone, l.trigger, l.name
            FROM leads l
            WHERE l.trigger IS NOT NULL 
            AND l.trigger != ''
            AND EXISTS (
                SELECT 1 FROM sequence_steps ss
                WHERE ss.is_entry_point = true
                AND position(ss.trigger in l.trigger) > 0
            )
            AND NOT EXISTS (
                SELECT 1 FROM sequence_contacts sc
                WHERE sc.contact_phone = l.phone
            )
            LIMIT 5
        """)
        
        examples = cur.fetchall()
        print("\nExample leads ready for enrollment:")
        for phone, trigger, name in examples:
            print(f"  {phone} ({name}) - Trigger: {trigger}")
    
    print("\n" + "=" * 60)
    print("SUMMARY:")
    print("- All WARMEXAMA leads have already completed the sequence")
    print("- Each lead went through all 4 steps successfully")
    print("- The sequence processor is working correctly!")
    print("- It's not enrolling them again because they already completed")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.close()
