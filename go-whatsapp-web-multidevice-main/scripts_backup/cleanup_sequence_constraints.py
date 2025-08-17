import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Drop the conflicting constraint
    print("\n1. Dropping conflicting unique constraint...")
    cur.execute("""
        ALTER TABLE sequence_contacts 
        DROP CONSTRAINT IF EXISTS uq_sequence_contact
    """)
    print("Dropped uq_sequence_contact constraint")
    
    # Commit changes
    conn.commit()
    
    # Verify constraints
    print("\n2. Final constraints on sequence_contacts:")
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
    
    # Show indexes
    print("\n3. Indexes on sequence_contacts:")
    cur.execute("""
        SELECT indexname, indexdef 
        FROM pg_indexes 
        WHERE tablename = 'sequence_contacts'
        ORDER BY indexname
    """)
    
    indexes = cur.fetchall()
    for name, definition in indexes:
        print(f"  - {name}")
    
    print("\n" + "=" * 60)
    print("CLEANUP COMPLETED!")
    print("Now only the correct unique constraints exist")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.rollback()
        conn.close()
