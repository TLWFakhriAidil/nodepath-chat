import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # First, clean up any NULL sequence_stepid records
    print("\n1. Cleaning up records with NULL sequence_stepid...")
    cur.execute("""
        DELETE FROM sequence_contacts WHERE sequence_stepid IS NULL
    """)
    print(f"Deleted {cur.rowcount} records with NULL sequence_stepid")
    
    # Make sequence_stepid NOT NULL
    print("\n2. Making sequence_stepid NOT NULL...")
    cur.execute("""
        ALTER TABLE sequence_contacts 
        ALTER COLUMN sequence_stepid SET NOT NULL
    """)
    print("sequence_stepid is now NOT NULL")
    
    # Drop existing constraints
    print("\n3. Dropping existing constraints...")
    cur.execute("""
        ALTER TABLE sequence_contacts
        DROP CONSTRAINT IF EXISTS uq_sequence_contacts_full;
        
        ALTER TABLE sequence_contacts
        DROP CONSTRAINT IF EXISTS uq_sequence_contacts_stepid;
    """)
    
    # Drop the partial unique indexes
    cur.execute("""
        DROP INDEX IF EXISTS idx_sequence_contacts_unique;
        DROP INDEX IF EXISTS idx_sequence_contacts_unique_no_stepid;
    """)
    print("Dropped existing constraints and indexes")
    
    # Create the proper unique constraint
    print("\n4. Creating proper unique constraint...")
    cur.execute("""
        ALTER TABLE sequence_contacts
        ADD CONSTRAINT uq_sequence_contact_step
        UNIQUE (sequence_id, contact_phone, sequence_stepid)
    """)
    print("Created unique constraint on (sequence_id, contact_phone, sequence_stepid)")
    
    # Commit changes
    conn.commit()
    
    # Verify the schema
    print("\n5. Verifying sequence_contacts schema...")
    cur.execute("""
        SELECT column_name, data_type, is_nullable
        FROM information_schema.columns
        WHERE table_name = 'sequence_contacts'
        AND column_name = 'sequence_stepid'
    """)
    
    result = cur.fetchone()
    if result:
        col_name, data_type, nullable = result
        print(f"  - {col_name}: {data_type} (nullable: {nullable})")
    
    print("\n6. Final constraints:")
    cur.execute("""
        SELECT conname, pg_get_constraintdef(oid) as definition
        FROM pg_constraint 
        WHERE conrelid = 'sequence_contacts'::regclass
        AND contype = 'u'
        ORDER BY conname
    """)
    
    constraints = cur.fetchall()
    for name, definition in constraints:
        print(f"  - {name}: {definition}")
    
    # Test the constraint
    print("\n7. Testing ON CONFLICT with sample data...")
    cur.execute("""
        SELECT id FROM sequences WHERE name = 'WARM Sequence' LIMIT 1
    """)
    seq_result = cur.fetchone()
    
    if seq_result:
        seq_id = seq_result[0]
        test_phone = '60123456789_test'
        test_stepid = 'e51bc68b-0f83-441d-8991-5865979809b1'
        
        try:
            cur.execute("""
                INSERT INTO sequence_contacts (
                    sequence_id, contact_phone, contact_name, 
                    current_step, status, completed_at, current_trigger,
                    next_trigger_time, sequence_stepid, assigned_device_id
                ) VALUES (%s, %s, 'Test', 1, 'active', NOW(), 'WARMEXAMA', NOW(), %s, NULL)
                ON CONFLICT (sequence_id, contact_phone, sequence_stepid) DO NOTHING
                RETURNING id
            """, (seq_id, test_phone, test_stepid))
            
            result = cur.fetchone()
            if result:
                print("✅ Test insert successful!")
                # Clean up
                cur.execute("DELETE FROM sequence_contacts WHERE id = %s", (result[0],))
            else:
                print("✅ ON CONFLICT worked - duplicate prevented")
                
        except Exception as e:
            print(f"❌ Test failed: {e}")
    
    # Commit cleanup
    conn.commit()
    
    print("\n" + "=" * 60)
    print("SEQUENCE CONSTRAINT FIX COMPLETED!")
    print("- sequence_stepid is now NOT NULL")
    print("- Unique constraint on (sequence_id, contact_phone, sequence_stepid)")
    print("- ON CONFLICT should now work properly")
    print("- Each enrollment step is properly tracked")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
