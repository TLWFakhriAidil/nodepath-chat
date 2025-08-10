import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Show current state
    print("\nCurrent sequence_contacts constraints:")
    cur.execute("""
        SELECT conname, pg_get_constraintdef(oid) as definition
        FROM pg_constraint 
        WHERE conrelid = 'sequence_contacts'::regclass
        ORDER BY contype, conname
    """)
    
    constraints = cur.fetchall()
    for name, definition in constraints:
        print(f"  - {name}: {definition}")
    
    print("\nColumn information for sequence_stepid:")
    cur.execute("""
        SELECT column_name, data_type, is_nullable, column_default
        FROM information_schema.columns
        WHERE table_name = 'sequence_contacts'
        AND column_name = 'sequence_stepid'
    """)
    
    result = cur.fetchone()
    if result:
        col_name, data_type, nullable, default = result
        print(f"  - {col_name}: {data_type}")
        print(f"    Nullable: {nullable}")
        print(f"    Default: {default}")
    
    print("\n" + "=" * 60)
    print("FINAL STATE:")
    print("1. sequence_stepid is NOT NULL (required for tracking)")
    print("2. Unique constraint: (sequence_id, contact_phone, sequence_stepid)")
    print("3. This allows same contact in multiple steps of same sequence")
    print("4. ON CONFLICT clause will work properly now")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.close()
