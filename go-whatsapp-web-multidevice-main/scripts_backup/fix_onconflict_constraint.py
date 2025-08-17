import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Check ALL constraints and indexes
    print("\n1. ALL constraints on sequence_contacts:")
    cur.execute("""
        SELECT conname, contype, 
               pg_get_constraintdef(oid) as definition
        FROM pg_constraint 
        WHERE conrelid = 'sequence_contacts'::regclass
        ORDER BY conname
    """)
    
    constraints = cur.fetchall()
    for name, type, definition in constraints:
        print(f"  - {name} ({type}): {definition}")
    
    print("\n2. ALL indexes on sequence_contacts:")
    cur.execute("""
        SELECT indexname, indexdef 
        FROM pg_indexes 
        WHERE tablename = 'sequence_contacts'
        ORDER BY indexname
    """)
    
    indexes = cur.fetchall()
    for name, definition in indexes:
        print(f"  - {name}")
        print(f"    Definition: {definition}")
    
    # Check if we can manually insert with ON CONFLICT
    print("\n3. Testing ON CONFLICT manually...")
    
    # First, let's see what columns we're dealing with
    cur.execute("""
        SELECT column_name, data_type, is_nullable
        FROM information_schema.columns
        WHERE table_name = 'sequence_contacts'
        AND column_name IN ('sequence_id', 'contact_phone', 'sequence_stepid')
        ORDER BY ordinal_position
    """)
    
    columns = cur.fetchall()
    print("\nRelevant columns:")
    for col_name, data_type, nullable in columns:
        print(f"  - {col_name}: {data_type} (nullable: {nullable})")
    
    # Let's create the exact constraint the code expects
    print("\n4. Creating exact constraint for ON CONFLICT...")
    
    # First drop any partial unique indexes
    cur.execute("""
        DROP INDEX IF EXISTS idx_sequence_contacts_unique;
        DROP INDEX IF EXISTS idx_sequence_contacts_unique_no_stepid;
    """)
    
    # Create a proper unique constraint (not just an index)
    cur.execute("""
        ALTER TABLE sequence_contacts
        ADD CONSTRAINT uq_sequence_contacts_full
        UNIQUE (sequence_id, contact_phone, sequence_stepid);
    """)
    
    print("Created full unique constraint on (sequence_id, contact_phone, sequence_stepid)")
    
    # Commit changes
    conn.commit()
    
    print("\n5. Final constraints:")
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
    
    print("\n" + "=" * 60)
    print("CONSTRAINT FIX COMPLETED!")
    print("Created proper UNIQUE constraint for ON CONFLICT clause")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
