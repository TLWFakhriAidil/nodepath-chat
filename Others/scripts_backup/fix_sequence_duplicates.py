import psycopg2
from datetime import datetime

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def fix_sequence_duplicates():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== Fixing Sequence Contact Duplicates ===")
    
    try:
        # 1. Check current duplicates
        print("\n1. Checking for duplicates...")
        cur.execute("""
            SELECT sequence_id, contact_phone, current_step, COUNT(*) as count
            FROM sequence_contacts
            GROUP BY sequence_id, contact_phone, current_step
            HAVING COUNT(*) > 1
        """)
        
        duplicates = cur.fetchall()
        if duplicates:
            print(f"Found {len(duplicates)} duplicate groups:")
            for dup in duplicates:
                print(f"  - Sequence: {dup[0][:8]}..., Phone: {dup[1]}, Step: {dup[2]}, Count: {dup[3]}")
        else:
            print("No duplicates found!")
        
        # 2. Delete duplicates (keep active/completed over pending)
        print("\n2. Removing duplicates...")
        cur.execute("""
            WITH duplicates AS (
                SELECT id,
                       ROW_NUMBER() OVER (
                           PARTITION BY sequence_id, contact_phone, current_step 
                           ORDER BY 
                               CASE status 
                                   WHEN 'completed' THEN 1
                                   WHEN 'active' THEN 2
                                   WHEN 'pending' THEN 3
                                   ELSE 4
                               END,
                               id ASC
                       ) as rn
                FROM sequence_contacts
            )
            DELETE FROM sequence_contacts
            WHERE id IN (
                SELECT id FROM duplicates WHERE rn > 1
            )
        """)
        
        deleted_count = cur.rowcount
        print(f"Deleted {deleted_count} duplicate records")
        
        # 3. Add unique constraints
        print("\n3. Adding unique constraints...")
        
        # Drop existing constraints
        cur.execute("ALTER TABLE sequence_contacts DROP CONSTRAINT IF EXISTS uk_sequence_contact_step")
        cur.execute("ALTER TABLE sequence_contacts DROP CONSTRAINT IF EXISTS uk_sequence_contact_stepid")
        
        # Add constraint for step number
        cur.execute("""
            ALTER TABLE sequence_contacts
            ADD CONSTRAINT uk_sequence_contact_step 
            UNIQUE (sequence_id, contact_phone, current_step)
        """)
        print("Added unique constraint on (sequence_id, contact_phone, current_step)")
        
        # Add constraint for stepid (needed for ON CONFLICT)
        cur.execute("""
            ALTER TABLE sequence_contacts
            ADD CONSTRAINT uk_sequence_contact_stepid
            UNIQUE (sequence_id, contact_phone, sequence_stepid)
        """)
        print("Added unique constraint on (sequence_id, contact_phone, sequence_stepid)")
        
        # 4. Fix stuck active records
        print("\n4. Resetting stuck active records...")
        cur.execute("""
            UPDATE sequence_contacts
            SET status = 'pending',
                processing_device_id = NULL,
                processing_started_at = NULL
            WHERE status = 'active'
              AND processing_started_at < NOW() - INTERVAL '30 minutes'
        """)
        
        reset_count = cur.rowcount
        print(f"Reset {reset_count} stuck active records")
        
        # 5. Ensure only one active per contact
        print("\n5. Checking for multiple active steps per contact...")
        cur.execute("""
            WITH ranked_active AS (
                SELECT id,
                       ROW_NUMBER() OVER (
                           PARTITION BY sequence_id, contact_phone 
                           ORDER BY current_step ASC
                       ) as rn
                FROM sequence_contacts
                WHERE status = 'active'
            )
            UPDATE sequence_contacts
            SET status = 'pending'
            WHERE id IN (
                SELECT id FROM ranked_active WHERE rn > 1
            )
        """)
        
        fixed_active = cur.rowcount
        print(f"Fixed {fixed_active} contacts with multiple active steps")
        
        # 6. Show final state
        print("\n6. Final state check:")
        cur.execute("""
            SELECT 
                contact_name,
                contact_phone,
                current_step,
                status,
                COUNT(*) as count
            FROM sequence_contacts
            WHERE sequence_id = '1-4ed6-891c-bcb7d12baa8a'
            GROUP BY contact_name, contact_phone, current_step, status
            ORDER BY contact_name, current_step
        """)
        
        results = cur.fetchall()
        print("\nSequence contacts for your sequence:")
        print(f"{'Name':<15} {'Phone':<15} {'Step':<6} {'Status':<12} {'Count':<6}")
        print("-" * 60)
        for row in results:
            print(f"{row[0]:<15} {row[1]:<15} {row[2]:<6} {row[3]:<12} {row[4]:<6}")
        
        # Commit changes
        conn.commit()
        print("\n[OK] All fixes applied successfully!")
        
    except Exception as e:
        conn.rollback()
        print(f"\n[ERROR] Error: {e}")
        raise
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    fix_sequence_duplicates()
