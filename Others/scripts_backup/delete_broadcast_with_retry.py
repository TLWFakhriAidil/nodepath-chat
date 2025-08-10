import psycopg2
import time
import sys

print("Attempting to delete all records from broadcast_messages table...")
print("Will retry if database is busy...")

max_retries = 10
retry_count = 0

while retry_count < max_retries:
    try:
        # Try to connect
        print(f"\nAttempt {retry_count + 1}/{max_retries}...")
        conn = psycopg2.connect(
            "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway",
            connect_timeout=5
        )
        cursor = conn.cursor()
        print("Connected successfully!")
        
        # Check current count
        cursor.execute("SELECT COUNT(*) FROM broadcast_messages")
        count_before = cursor.fetchone()[0]
        print(f"Records before deletion: {count_before:,}")
        
        if count_before == 0:
            print("Table is already empty!")
        else:
            # Delete all records
            print("Deleting all records...")
            cursor.execute("DELETE FROM broadcast_messages")
            deleted = cursor.rowcount
            
            # Commit the transaction
            conn.commit()
            print(f"Deleted {deleted:,} records successfully!")
            
            # Verify deletion
            cursor.execute("SELECT COUNT(*) FROM broadcast_messages")
            count_after = cursor.fetchone()[0]
            print(f"Records remaining: {count_after}")
        
        # Close connection
        cursor.close()
        conn.close()
        print("\n✅ Operation completed successfully!")
        break
        
    except psycopg2.OperationalError as e:
        retry_count += 1
        if "too many clients" in str(e):
            print(f"Database busy (too many connections). Waiting 5 seconds before retry...")
            time.sleep(5)
        else:
            print(f"Connection error: {e}")
            break
    except Exception as e:
        print(f"Unexpected error: {e}")
        break

if retry_count >= max_retries:
    print("\n❌ Failed to connect after maximum retries.")
    print("The database has too many active connections.")
    print("Please try again later or close some connections.")
