import psycopg2

DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

print("=== FIXING SEQUENCE_CONTACTS TABLE ===")
print("This will:")
print("1. Remove any updated_at references from triggers")
print("2. Ensure the table structure is correct")
print("")

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    # Read and execute the SQL file
    with open('fix_sequence_contacts.sql', 'r') as f:
        sql = f.read()
    
    # Execute each statement separately
    statements = sql.split(';')
    for stmt in statements:
        stmt = stmt.strip()
        if stmt:
            try:
                print(f"Executing: {stmt[:50]}...")
                cur.execute(stmt + ';')
                print("Success")
            except Exception as e:
                print(f"Error: {e}")
    
    conn.commit()
    print("\n=== FIX COMPLETE ===")
    
    # Verify the fix
    print("\n=== VERIFYING FIX ===")
    cur.execute("""
        SELECT column_name 
        FROM information_schema.columns 
        WHERE table_name = 'sequence_contacts' 
        AND column_name = 'updated_at'
    """)
    if cur.fetchone():
        print("WARNING: updated_at column still exists!")
    else:
        print("GOOD: updated_at column does not exist!")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"\nERROR: {e}")
    print("Fix failed!")
