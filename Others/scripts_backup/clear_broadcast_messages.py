import psycopg2
import sys

try:
    # Connect to database
    print("Connecting to database...")
    conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
    cursor = conn.cursor()
    
    # Check current count
    cursor.execute("SELECT COUNT(*) FROM broadcast_messages")
    before_count = cursor.fetchone()[0]
    print(f"Records before deletion: {before_count:,}")
    
    if before_count == 0:
        print("Table is already empty!")
    else:
        # Delete all records
        print("Deleting all records...")
        cursor.execute("DELETE FROM broadcast_messages")
        deleted = cursor.rowcount
        
        # Commit changes
        conn.commit()
        print(f"Deleted {deleted:,} records")
        
        # Verify
        cursor.execute("SELECT COUNT(*) FROM broadcast_messages")
        after_count = cursor.fetchone()[0]
        print(f"Records after deletion: {after_count}")
        
        if after_count == 0:
            print("✅ All broadcast messages deleted successfully!")
        else:
            print(f"⚠️ {after_count} records still remain")
    
    cursor.close()
    conn.close()
    print("Database connection closed.")
    
except Exception as e:
    print(f"Error: {e}")
    sys.exit(1)
