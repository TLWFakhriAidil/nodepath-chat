import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # Check sequence_contacts table structure
    print("\n1. Checking sequence_contacts table structure...")
    cur.execute("""
        SELECT column_name, data_type, is_nullable, column_default
        FROM information_schema.columns
        WHERE table_name = 'sequence_contacts'
        ORDER BY ordinal_position
    """)
    
    columns = cur.fetchall()
    print("Current columns in sequence_contacts:")
    for col_name, data_type, nullable, default in columns:
        print(f"  - {col_name}: {data_type} (nullable: {nullable}, default: {default})")
    
    # Check sequence_steps table structure
    print("\n2. Checking sequence_steps table structure...")
    cur.execute("""
        SELECT column_name, data_type, is_nullable, column_default
        FROM information_schema.columns
        WHERE table_name = 'sequence_steps'
        ORDER BY ordinal_position
    """)
    
    columns = cur.fetchall()
    print("Current columns in sequence_steps:")
    for col_name, data_type, nullable, default in columns:
        print(f"  - {col_name}: {data_type} (nullable: {nullable}, default: {default})")
    
    # Check if sequence_stepid column exists
    print("\n3. Checking if sequence_stepid column exists in sequence_contacts...")
    cur.execute("""
        SELECT COUNT(*) 
        FROM information_schema.columns 
        WHERE table_name = 'sequence_contacts' 
        AND column_name = 'sequence_stepid'
    """)
    
    exists = cur.fetchone()[0]
    print(f"sequence_stepid column exists: {exists > 0}")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.close()
