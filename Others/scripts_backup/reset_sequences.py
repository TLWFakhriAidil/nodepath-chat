import psycopg2
from datetime import datetime

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def reset_sequences():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== RESET SEQUENCES TO CLEAN STATE ===")
    print("")
    
    try:
        # 1. Delete all broadcast messages
        print("1. Deleting all broadcast messages...")
        cur.execute("DELETE FROM broadcast_messages WHERE sequence_id IS NOT NULL")
        bm_deleted = cur.rowcount
        print(f"   [OK] Deleted {bm_deleted} broadcast messages")
        
        # 2. Reset all sequence contacts to pending
        print("\n2. Resetting all sequence contacts to pending...")
        cur.execute("""
            UPDATE sequence_contacts 
            SET status = 'pending',
                completed_at = NULL,
                processing_device_id = NULL,
                processing_started_at = NULL,
                last_error = NULL,
                retry_count = 0
            WHERE status != 'pending'
        """)
        sc_reset = cur.rowcount
        print(f"   [OK] Reset {sc_reset} sequence contacts to pending")
        
        # 3. Verify the state
        print("\n3. Verifying clean state...")
        
        # Check sequence contacts
        cur.execute("""
            SELECT 
                contact_phone,
                current_step,
                status,
                COUNT(*) as count
            FROM sequence_contacts
            GROUP BY contact_phone, current_step, status
            ORDER BY contact_phone, current_step
        """)
        
        print(f"\n   {'Phone':<15} {'Step':<6} {'Status':<10} {'Count':<6}")
        print("   " + "-" * 40)
        for row in cur.fetchall():
            print(f"   {row[0]:<15} {row[1]:<6} {row[2]:<10} {row[3]:<6}")
        
        # Check what will be processed next
        cur.execute("""
            WITH earliest_pending AS (
                SELECT DISTINCT ON (sc.sequence_id, sc.contact_phone)
                    sc.contact_phone,
                    sc.current_step,
                    sc.next_trigger_time
                FROM sequence_contacts sc
                WHERE sc.status = 'pending'
                ORDER BY sc.sequence_id, sc.contact_phone, sc.current_step ASC, sc.next_trigger_time ASC
            )
            SELECT * FROM earliest_pending
            ORDER BY next_trigger_time ASC
        """)
        
        print(f"\n4. Next steps to be processed:")
        print(f"   {'Phone':<15} {'Step':<6} {'Trigger Time':<20}")
        print("   " + "-" * 45)
        for row in cur.fetchall():
            trigger = row[2].strftime("%Y-%m-%d %H:%M:%S") if row[2] else "None"
            print(f"   {row[0]:<15} {row[1]:<6} {trigger:<20}")
        
        # Commit changes
        conn.commit()
        print("\n[OK] RESET COMPLETE!")
        
    except Exception as e:
        conn.rollback()
        print(f"\n[X] Error: {e}")
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    confirm = input("This will reset all sequences to pending state. Type 'RESET' to confirm: ")
    if confirm == "RESET":
        reset_sequences()
    else:
        print("Reset cancelled.")
