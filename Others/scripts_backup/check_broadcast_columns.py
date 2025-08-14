import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    # Check broadcast_messages columns
    print("=== Broadcast Messages Columns ===")
    cur.execute("""
        SELECT column_name, data_type, is_nullable
        FROM information_schema.columns
        WHERE table_name = 'broadcast_messages'
        ORDER BY ordinal_position
    """)
    
    for col in cur.fetchall():
        print(f"{col[0]}: {col[1]} (nullable: {col[2]})")
    
    # Check if sequence_stepid allows NULL
    print("\n=== Checking sequence_stepid column ===")
    cur.execute("""
        SELECT column_name, is_nullable, column_default
        FROM information_schema.columns
        WHERE table_name = 'broadcast_messages' 
        AND column_name = 'sequence_stepid'
    """)
    
    result = cur.fetchone()
    if result:
        print(f"sequence_stepid - Nullable: {result[1]}, Default: {result[2]}")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
