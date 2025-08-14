import psycopg2

DB_URI = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    print("Connected to PostgreSQL\n")
    
    # Check current count
    cur.execute("SELECT COUNT(*) FROM broadcast_messages")
    count_before = cur.fetchone()[0]
    print(f"Current records in broadcast_messages: {count_before}")
    
    if count_before > 0:
        # Delete all records
        print("\nDeleting all records...")
        cur.execute("DELETE FROM broadcast_messages")
        
        # Commit the changes
        conn.commit()
        
        # Verify deletion
        cur.execute("SELECT COUNT(*) FROM broadcast_messages")
        count_after = cur.fetchone()[0]
        print(f"Records after deletion: {count_after}")
        print(f"Successfully deleted {count_before} records!")
    else:
        print("No records to delete.")
    
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
