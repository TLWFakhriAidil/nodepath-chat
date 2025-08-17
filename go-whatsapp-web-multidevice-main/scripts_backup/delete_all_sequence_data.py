import psycopg2
from datetime import datetime

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def delete_all_sequence_data():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== DELETE ALL SEQUENCE CONTACTS AND BROADCAST MESSAGES ===")
    print("!!! WARNING: This will DELETE ALL DATA from:")
    print("   - sequence_contacts table")
    print("   - broadcast_messages table")
    print("")
    
    try:
        # First, show current counts
        print("Current data counts:")
        
        cur.execute("SELECT COUNT(*) FROM sequence_contacts")
        sc_count = cur.fetchone()[0]
        print(f"  - sequence_contacts: {sc_count} records")
        
        cur.execute("SELECT COUNT(*) FROM broadcast_messages")
        bm_count = cur.fetchone()[0]
        print(f"  - broadcast_messages: {bm_count} records")
        
        if sc_count == 0 and bm_count == 0:
            print("\n[OK] Both tables are already empty!")
            return
        
        # Confirm deletion
        print("\n" + "="*50)
        confirm = input("Are you SURE you want to delete ALL records? Type 'DELETE ALL' to confirm: ")
        
        if confirm != "DELETE ALL":
            print("\n[X] Deletion cancelled. No data was deleted.")
            return
        
        print("\nDeleting data...")
        
        # Delete broadcast_messages first (it may have foreign keys)
        print("\n1. Deleting all broadcast_messages...")
        cur.execute("DELETE FROM broadcast_messages")
        bm_deleted = cur.rowcount
        print(f"   [OK] Deleted {bm_deleted} broadcast messages")
        
        # Delete sequence_contacts
        print("\n2. Deleting all sequence_contacts...")
        cur.execute("DELETE FROM sequence_contacts")
        sc_deleted = cur.rowcount
        print(f"   [OK] Deleted {sc_deleted} sequence contacts")
        
        # Reset any sequences if needed
        print("\n3. Resetting sequences to 'scheduled' status...")
        cur.execute("""
            UPDATE sequences 
            SET status = 'scheduled' 
            WHERE status IN ('processing', 'completed')
        """)
        seq_reset = cur.rowcount
        print(f"   [OK] Reset {seq_reset} sequences")
        
        # Also reset campaigns if needed
        print("\n4. Resetting campaigns to 'scheduled' status...")
        cur.execute("""
            UPDATE campaigns 
            SET status = 'scheduled' 
            WHERE status IN ('triggered', 'processing', 'completed')
              AND campaign_date >= CURRENT_DATE
        """)
        camp_reset = cur.rowcount
        print(f"   [OK] Reset {camp_reset} campaigns")
        
        # Commit the changes
        conn.commit()
        
        print("\n" + "="*50)
        print("[OK] DELETION COMPLETE!")
        print(f"   - Deleted {bm_deleted} broadcast messages")
        print(f"   - Deleted {sc_deleted} sequence contacts")
        print(f"   - Reset {seq_reset} sequences")
        print(f"   - Reset {camp_reset} campaigns")
        
        # Verify deletion
        print("\nVerifying deletion...")
        cur.execute("SELECT COUNT(*) FROM sequence_contacts")
        new_sc_count = cur.fetchone()[0]
        cur.execute("SELECT COUNT(*) FROM broadcast_messages")
        new_bm_count = cur.fetchone()[0]
        
        print(f"  - sequence_contacts: {new_sc_count} records (should be 0)")
        print(f"  - broadcast_messages: {new_bm_count} records (should be 0)")
        
    except Exception as e:
        conn.rollback()
        print(f"\n[X] Error: {e}")
        print("No data was deleted due to error.")
        raise
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    delete_all_sequence_data()
