import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # First, check existing constraints
    print("\n1. Checking existing constraints on sequence_contacts...")
    cur.execute("""
        SELECT conname, contype, 
               pg_get_constraintdef(oid) as definition
        FROM pg_constraint 
        WHERE conrelid = 'sequence_contacts'::regclass
        ORDER BY conname
    """)
    
    constraints = cur.fetchall()
    print("Current constraints:")
    for name, type, definition in constraints:
        print(f"  - {name} ({type}): {definition}")
    
    # Drop the old unique constraint if it exists
    print("\n2. Dropping old unique constraint...")
    cur.execute("""
        ALTER TABLE sequence_contacts 
        DROP CONSTRAINT IF EXISTS sequence_contacts_sequence_id_contact_phone_key
    """)
    print("Old constraint dropped (if existed)")
    
    # Create the new unique constraint that the code expects
    print("\n3. Creating new unique constraint...")
    cur.execute("""
        CREATE UNIQUE INDEX IF NOT EXISTS idx_sequence_contacts_unique 
        ON sequence_contacts(sequence_id, contact_phone, sequence_stepid) 
        WHERE sequence_stepid IS NOT NULL
    """)
    print("Created unique index on (sequence_id, contact_phone, sequence_stepid)")
    
    # Also create a unique constraint for contacts without stepid (for backward compatibility)
    cur.execute("""
        CREATE UNIQUE INDEX IF NOT EXISTS idx_sequence_contacts_unique_no_stepid 
        ON sequence_contacts(sequence_id, contact_phone) 
        WHERE sequence_stepid IS NULL
    """)
    print("Created unique index for NULL sequence_stepid")
    
    # Commit changes
    conn.commit()
    
    # Verify the new constraints
    print("\n4. Verifying new constraints...")
    cur.execute("""
        SELECT indexname, indexdef 
        FROM pg_indexes 
        WHERE tablename = 'sequence_contacts' 
        AND indexname LIKE '%unique%'
    """)
    
    indexes = cur.fetchall()
    print("Unique indexes:")
    for name, definition in indexes:
        print(f"  - {name}")
        print(f"    {definition}")
    
    print("\n" + "=" * 60)
    print("CONSTRAINT UPDATE COMPLETED!")
    print("The sequence enrollment should now work without ON CONFLICT errors")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
