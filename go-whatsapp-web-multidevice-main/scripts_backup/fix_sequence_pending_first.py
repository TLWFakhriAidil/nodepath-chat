import psycopg2
from datetime import datetime

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def fix_sequence_logic():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== FIXING SEQUENCE PENDING-FIRST LOGIC ===")
    print(f"Time: {datetime.now()}")
    
    try:
        # 1. First, let's see the current state
        print("\n1. Current sequence_contacts state:")
        cur.execute("""
            SELECT contact_phone, contact_name, current_step, status, 
                   next_trigger_time, current_trigger
            FROM sequence_contacts
            ORDER BY contact_phone, current_step
        """)
        current_state = cur.fetchall()
        for row in current_state:
            print(f"   {row[0]} ({row[1]}) - Step {row[2]}: {row[3]} - Trigger at: {row[4]}")
        
        # 2. Update all 'active' steps back to 'pending' if not completed
        print("\n2. Fixing status - changing all 'active' to 'pending':")
        cur.execute("""
            UPDATE sequence_contacts 
            SET status = 'pending' 
            WHERE status = 'active' 
            AND completed_at IS NULL
            RETURNING id, contact_phone, current_step
        """)
        updated = cur.fetchall()
        conn.commit()
        print(f"   Updated {len(updated)} records to pending")
        
        # 3. Check if we have any broadcast messages
        print("\n3. Checking broadcast_messages:")
        cur.execute("""
            SELECT COUNT(*) FROM broadcast_messages 
            WHERE sequence_id IS NOT NULL
            AND created_at > NOW() - INTERVAL '1 hour'
        """)
        msg_count = cur.fetchone()[0]
        print(f"   Found {msg_count} sequence messages in last hour")
        
        # 4. Find steps that should be processed NOW
        print("\n4. Steps ready to process NOW:")
        cur.execute("""
            SELECT DISTINCT ON (sequence_id, contact_phone)
                sc.id, sc.contact_phone, sc.contact_name, sc.current_step,
                sc.next_trigger_time, sc.current_trigger,
                ss.content, sc.sequence_id, sc.assigned_device_id,
                sc.user_id, sc.sequence_stepid
            FROM sequence_contacts sc
            JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
            WHERE sc.status = 'pending'
            AND sc.next_trigger_time <= NOW()
            ORDER BY sc.sequence_id, sc.contact_phone, sc.next_trigger_time ASC
        """)
        ready_steps = cur.fetchall()
        print(f"   Found {len(ready_steps)} steps ready to send:")
        for step in ready_steps:
            print(f"   - {step[1]} ({step[2]}) Step {step[3]} - Was due at {step[4]}")
        
        # 5. Check the constraint that's blocking
        print("\n5. Checking constraints:")
        cur.execute("""
            SELECT conname, pg_get_constraintdef(oid) 
            FROM pg_constraint 
            WHERE conrelid = 'sequence_contacts'::regclass
            AND contype = 'c'
        """)
        constraints = cur.fetchall()
        for name, definition in constraints:
            if 'step' in definition.lower() or 'status' in definition.lower():
                print(f"   Found constraint: {name}")
                print(f"   Definition: {definition}")
        
        # 6. Check if there's a trigger preventing activation
        print("\n6. Checking for database triggers:")
        cur.execute("""
            SELECT trigger_name, event_manipulation, action_statement
            FROM information_schema.triggers
            WHERE event_object_table = 'sequence_contacts'
        """)
        triggers = cur.fetchall()
        if triggers:
            for trigger in triggers:
                print(f"   Trigger: {trigger[0]} on {trigger[1]}")
        else:
            print("   No triggers found on sequence_contacts table")
        
    except Exception as e:
        print(f"\nERROR: {e}")
        conn.rollback()
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    fix_sequence_logic()
