import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Check for NULL sequence_stepid values
    print("\n1. Checking for NULL sequence_stepid values...")
    cur.execute("""
        SELECT COUNT(*) FROM sequence_contacts WHERE sequence_stepid IS NULL
    """)
    null_count = cur.fetchone()[0]
    print(f"Records with NULL sequence_stepid: {null_count}")
    
    if null_count > 0:
        # Delete old records with NULL sequence_stepid (these are from old enrollments)
        print("\n2. Cleaning up old records with NULL sequence_stepid...")
        cur.execute("""
            DELETE FROM sequence_contacts WHERE sequence_stepid IS NULL
        """)
        print(f"Deleted {cur.rowcount} old records")
    
    # Now let's make sure the constraint handles NULLs properly
    print("\n3. Recreating constraint to handle NULLs properly...")
    
    # Drop the constraint we just created
    cur.execute("""
        ALTER TABLE sequence_contacts
        DROP CONSTRAINT IF EXISTS uq_sequence_contacts_full;
    """)
    
    # PostgreSQL treats NULL values as distinct in unique constraints
    # So we don't need to worry about NULL handling
    cur.execute("""
        ALTER TABLE sequence_contacts
        ADD CONSTRAINT uq_sequence_contacts_stepid
        UNIQUE (sequence_id, contact_phone, sequence_stepid);
    """)
    
    print("Created unique constraint that properly handles NULLs")
    
    # Test if the enrollment will work now
    print("\n4. Testing enrollment with sample data...")
    
    # Get a sample sequence and lead
    cur.execute("""
        SELECT s.id, l.phone 
        FROM sequences s, leads l 
        WHERE s.name = 'WARM Sequence' 
        AND l.trigger = 'WARMEXAMA'
        LIMIT 1
    """)
    
    result = cur.fetchone()
    if result:
        seq_id, phone = result
        test_stepid = 'e51bc68b-0f83-441d-8991-5865979809b1'  # First step ID
        
        # Try the insert with ON CONFLICT
        cur.execute("""
            INSERT INTO sequence_contacts (
                sequence_id, contact_phone, contact_name, 
                current_step, status, completed_at, current_trigger,
                next_trigger_time, sequence_stepid, assigned_device_id
            ) VALUES (%s, %s, 'Test', 1, 'active', NOW(), 'WARMEXAMA', NOW(), %s, NULL)
            ON CONFLICT (sequence_id, contact_phone, sequence_stepid) DO NOTHING
            RETURNING id
        """, (seq_id, phone + '_test', test_stepid))
        
        if cur.fetchone():
            print("✅ Test insert with ON CONFLICT successful!")
        else:
            print("⚠️ Insert was skipped due to conflict (expected if duplicate)")
        
        # Clean up test
        cur.execute("DELETE FROM sequence_contacts WHERE contact_phone = %s", (phone + '_test',))
    
    # Commit all changes
    conn.commit()
    
    print("\n5. Final constraint check:")
    cur.execute("""
        SELECT conname, pg_get_constraintdef(oid) as definition
        FROM pg_constraint 
        WHERE conrelid = 'sequence_contacts'::regclass
        AND contype = 'u'
    """)
    
    constraints = cur.fetchall()
    for name, definition in constraints:
        print(f"  - {name}: {definition}")
    
    print("\n" + "=" * 60)
    print("FINAL FIX COMPLETED!")
    print("The ON CONFLICT clause should now work properly")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
