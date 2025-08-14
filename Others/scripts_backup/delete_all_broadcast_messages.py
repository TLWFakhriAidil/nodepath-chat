import psycopg2
import time

# Wait a bit for connections to clear
time.sleep(2)

try:
    # Connect
    conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
    cursor = conn.cursor()
    
    # Delete all records
    cursor.execute("DELETE FROM broadcast_messages")
    deleted_count = cursor.rowcount
    
    # Commit
    conn.commit()
    
    # Verify
    cursor.execute("SELECT COUNT(*) FROM broadcast_messages")
    remaining = cursor.fetchone()[0]
    
    print(f"Successfully deleted {deleted_count} records")
    print(f"Records remaining: {remaining}")
    
    # Close
    cursor.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
