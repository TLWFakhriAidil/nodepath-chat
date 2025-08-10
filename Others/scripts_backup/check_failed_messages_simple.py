import psycopg2
import time
import sys

# Retry connection up to 3 times
conn_str = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

for attempt in range(3):
    try:
        print(f"Connection attempt {attempt + 1}...")
        conn = psycopg2.connect(conn_str, connect_timeout=30)
        cursor = conn.cursor()
        print("[OK] Connected!")
        break
    except Exception as e:
        print(f"Failed: {e}")
        if attempt < 2:
            print("Retrying in 5 seconds...")
            time.sleep(5)
        else:
            print("All connection attempts failed.")
            sys.exit(1)

# Query for failed sequence messages
print("\nQuerying failed sequence messages...")

query = """
SELECT 
    bm.id,
    bm.phone,
    LEFT(bm.message, 100) as message_preview,
    bm.error_message,
    bm.status,
    bm.created_at::date as created_date,
    bm.updated_at::date as failed_date,
    ud.device_name,
    s.name as sequence_name
FROM broadcast_messages bm
LEFT JOIN user_devices ud ON bm.device_id = ud.id
LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
LEFT JOIN sequences s ON ss.sequence_id = s.id
WHERE bm.sequence_stepid IS NOT NULL 
AND bm.status = 'failed' 
AND bm.error_message IS NOT NULL
ORDER BY bm.updated_at DESC
LIMIT 20;
"""

try:
    cursor.execute(query)
    results = cursor.fetchall()
    
    print(f"\nFound {len(results)} failed sequence messages (showing first 20)")
    print("="*100)
    
    for i, row in enumerate(results, 1):
        print(f"\n[{i}] Message ID: {row[0]}")
        print(f"    Phone: {row[1]}")
        print(f"    Message: {row[2]}...")
        print(f"    ERROR: {row[3]}")
        print(f"    Created: {row[5]}, Failed: {row[6]}")
        print(f"    Device: {row[7] or 'Unknown'}")
        print(f"    Sequence: {row[8] or 'Unknown'}")
        print("-"*80)
    
    # Get error summary
    cursor.execute("""
        SELECT error_message, COUNT(*) as count
        FROM broadcast_messages 
        WHERE sequence_stepid IS NOT NULL 
        AND status = 'failed' 
        AND error_message IS NOT NULL
        GROUP BY error_message
        ORDER BY count DESC
        LIMIT 10
    """)
    
    error_summary = cursor.fetchall()
    
    print("\n" + "="*100)
    print("ERROR SUMMARY (Top 10)")
    print("="*100)
    
    for error, count in error_summary:
        print(f"{count:>5} messages: {error}")
    
except Exception as e:
    print(f"Query error: {e}")
finally:
    cursor.close()
    conn.close()
    print("\n[OK] Done!")
