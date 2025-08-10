import psycopg2

# Connect to database
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(DB_URI)
conn.autocommit = True  # Required for CREATE INDEX CONCURRENTLY
cur = conn.cursor()

print("Executing 3000 device optimization fix...")

try:
    # Step 1: Clean up
    print("1. Cleaning up completed records...")
    cur.execute("DELETE FROM sequence_contacts WHERE status = 'completed'")
    print(f"   Deleted {cur.rowcount} completed records")
    
    # Step 2: Remove duplicates
    print("2. Removing duplicate active records...")
    cur.execute("""
        WITH duplicates AS (
            SELECT id,
                   ROW_NUMBER() OVER (PARTITION BY sequence_id, contact_phone 
                                     ORDER BY current_step ASC) as rn
            FROM sequence_contacts
            WHERE status = 'active'
        )
        DELETE FROM sequence_contacts
        WHERE id IN (SELECT id FROM duplicates WHERE rn > 1)
    """)
    print(f"   Removed {cur.rowcount} duplicates")
    
    # Step 3: Drop existing constraints
    print("3. Dropping old constraints...")
    cur.execute("DROP INDEX IF EXISTS idx_one_active_per_contact")
    
    # Step 4: Create unique index for active steps
    print("4. Creating unique index for active steps...")
    cur.execute("""
        CREATE UNIQUE INDEX idx_one_active_per_contact
        ON sequence_contacts(sequence_id, contact_phone)
        WHERE status = 'active'
    """)
    
    # Step 5: Create index for pending steps
    print("5. Creating index for pending steps lookup...")
    cur.execute("""
        CREATE INDEX IF NOT EXISTS idx_pending_steps_by_time
        ON sequence_contacts(sequence_id, contact_phone, next_trigger_time)
        WHERE status = 'pending'
    """)
    
    # Step 6: Create the concurrent-safe function
    print("6. Creating concurrent-safe progression function...")
    cur.execute("""
        CREATE OR REPLACE FUNCTION progress_sequence_contact_concurrent(
            p_contact_id UUID
        ) RETURNS BOOLEAN AS $$
        DECLARE
            v_sequence_id UUID;
            v_contact_phone VARCHAR;
            v_current_step INT;
            v_next_id UUID;
            v_activated BOOLEAN := FALSE;
        BEGIN
            -- Step 1: Try to lock and complete current step
            UPDATE sequence_contacts 
            SET status = 'completed',
                completed_at = NOW()
            WHERE id = p_contact_id 
              AND status = 'active'
            RETURNING sequence_id, contact_phone, current_step 
            INTO v_sequence_id, v_contact_phone, v_current_step;
            
            IF NOT FOUND THEN
                RETURN FALSE;
            END IF;
            
            -- Step 2: Find next pending step by EARLIEST trigger time
            SELECT id INTO v_next_id
            FROM sequence_contacts
            WHERE sequence_id = v_sequence_id
              AND contact_phone = v_contact_phone
              AND status = 'pending'
              AND next_trigger_time <= NOW()
            ORDER BY next_trigger_time ASC
            LIMIT 1
            FOR UPDATE SKIP LOCKED;
            
            IF v_next_id IS NOT NULL THEN
                UPDATE sequence_contacts
                SET status = 'active'
                WHERE id = v_next_id
                  AND status = 'pending';
                
                GET DIAGNOSTICS v_activated = ROW_COUNT;
            END IF;
            
            RETURN v_activated;
        END;
        $$ LANGUAGE plpgsql;
    """)
    
    print("\n✅ Database optimizations applied successfully!")
    
    # Show current state
    print("\nCurrent sequence contacts state:")
    cur.execute("""
        SELECT contact_phone, current_step, status, next_trigger_time
        FROM sequence_contacts
        ORDER BY contact_phone, current_step
    """)
    for row in cur.fetchall():
        print(f"  Phone: {row[0]}, Step: {row[1]}, Status: {row[2]}, Next: {row[3]}")
        
except Exception as e:
    print(f"Error: {e}")
    
cur.close()
conn.close()