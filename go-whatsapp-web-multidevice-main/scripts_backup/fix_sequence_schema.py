import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Fix the sequence_stepid column type
    print("\n1. Fixing sequence_stepid column type from integer to uuid...")
    
    try:
        # First, remove any existing data in sequence_stepid (since it's the wrong type)
        cur.execute("""
            UPDATE sequence_contacts 
            SET sequence_stepid = NULL
            WHERE sequence_stepid IS NOT NULL
        """)
        print("Cleared existing sequence_stepid values")
        
        # Now alter the column type
        cur.execute("""
            ALTER TABLE sequence_contacts 
            ALTER COLUMN sequence_stepid TYPE uuid USING NULL
        """)
        print("Successfully changed sequence_stepid from integer to uuid")
        
    except Exception as e:
        print(f"Error changing column type: {e}")
        # Try alternative approach
        print("\nTrying alternative approach...")
        cur.execute("""
            ALTER TABLE sequence_contacts 
            DROP COLUMN IF EXISTS sequence_stepid
        """)
        cur.execute("""
            ALTER TABLE sequence_contacts 
            ADD COLUMN sequence_stepid uuid
        """)
        print("Successfully recreated sequence_stepid as uuid")
    
    # Add any missing columns that might be needed
    print("\n2. Checking for other missing columns...")
    
    # Add user_id column to sequence_contacts if missing
    cur.execute("""
        SELECT COUNT(*) 
        FROM information_schema.columns 
        WHERE table_name = 'sequence_contacts' 
        AND column_name = 'user_id'
    """)
    
    if cur.fetchone()[0] == 0:
        print("Adding user_id column to sequence_contacts...")
        cur.execute("""
            ALTER TABLE sequence_contacts 
            ADD COLUMN user_id uuid
        """)
        print("Added user_id column")
    
    # Create index for better performance
    print("\n3. Creating indexes for better performance...")
    cur.execute("""
        CREATE INDEX IF NOT EXISTS idx_sequence_contacts_stepid 
        ON sequence_contacts(sequence_stepid)
    """)
    cur.execute("""
        CREATE INDEX IF NOT EXISTS idx_sequence_contacts_phone 
        ON sequence_contacts(contact_phone)
    """)
    cur.execute("""
        CREATE INDEX IF NOT EXISTS idx_sequence_contacts_status 
        ON sequence_contacts(status)
    """)
    print("Indexes created/verified")
    
    # Commit all changes
    conn.commit()
    
    # Verify the changes
    print("\n4. Verifying changes...")
    cur.execute("""
        SELECT column_name, data_type 
        FROM information_schema.columns 
        WHERE table_name = 'sequence_contacts' 
        AND column_name = 'sequence_stepid'
    """)
    
    result = cur.fetchone()
    if result:
        print(f"sequence_stepid is now: {result[1]}")
    
    print("\n" + "=" * 60)
    print("SCHEMA UPDATE COMPLETED!")
    print("The sequence_stepid column is now uuid type to match sequence_steps.id")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
