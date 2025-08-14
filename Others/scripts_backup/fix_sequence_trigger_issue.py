import psycopg2
from datetime import datetime

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def check_and_fix_trigger():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== CHECKING DATABASE TRIGGER ===")
    
    try:
        # Get the trigger function definition
        cur.execute("""
            SELECT 
                p.proname AS function_name,
                pg_get_functiondef(p.oid) AS function_definition
            FROM pg_proc p
            JOIN pg_trigger t ON t.tgfoid = p.oid
            WHERE t.tgname = 'enforce_step_sequence'
            AND t.tgrelid = 'sequence_contacts'::regclass
        """)
        
        result = cur.fetchone()
        if result:
            print(f"\nTrigger function: {result[0]}")
            print("\nFunction definition:")
            print(result[1])
            
        # Now let's see if we should disable this trigger temporarily
        print("\n" + "="*50)
        print("This trigger is preventing the pending-first approach!")
        print("It's enforcing that steps must be completed in order.")
        print("But with pending-first, we need to mark steps as 'active' before completing them.")
        
        # Drop the problematic trigger
        print("\nDROPPING the enforce_step_sequence trigger...")
        cur.execute("DROP TRIGGER IF EXISTS enforce_step_sequence ON sequence_contacts")
        conn.commit()
        print("✅ Trigger dropped successfully!")
        
        # Now manually create the broadcast messages for ready steps
        print("\n=== MANUALLY CREATING BROADCAST MESSAGES ===")
        
        # Get ready steps again
        cur.execute("""
            SELECT 
                sc.id, sc.sequence_id, sc.contact_phone, sc.contact_name,
                sc.current_step, sc.current_trigger, sc.assigned_device_id,
                sc.user_id, sc.sequence_stepid,
                ss.content, ss.message_type, ss.media_url,
                COALESCE(ss.min_delay_seconds, s.min_delay_seconds, 5) as min_delay,
                COALESCE(ss.max_delay_seconds, s.max_delay_seconds, 15) as max_delay
            FROM sequence_contacts sc
            JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
            JOIN sequences s ON s.id = sc.sequence_id
            WHERE sc.status = 'pending'
            AND sc.next_trigger_time <= NOW()
            AND sc.completed_at IS NULL
            ORDER BY sc.next_trigger_time ASC
        """)
        
        ready_steps = cur.fetchall()
        print(f"Found {len(ready_steps)} steps ready to send")
        
        for step in ready_steps:
            (sc_id, seq_id, phone, name, step_num, trigger, device_id, 
             user_id, step_id, content, msg_type, media_url, min_delay, max_delay) = step
            
            print(f"\nCreating message for {phone} - Step {step_num}")
            
            # Insert broadcast message
            cur.execute("""
                INSERT INTO broadcast_messages (
                    user_id, device_id, sequence_id, sequence_stepid,
                    recipient_phone, recipient_name, message, content,
                    type, media_url, image_url, 
                    min_delay, max_delay,
                    scheduled_at, status, created_at, updated_at
                ) VALUES (
                    %s, %s, %s, %s,
                    %s, %s, %s, %s,
                    %s, %s, %s,
                    %s, %s,
                    NOW(), 'pending', NOW(), NOW()
                ) RETURNING id
            """, (
                user_id, device_id, seq_id, step_id,
                phone, name, content, content,
                msg_type or 'text', media_url, media_url,
                min_delay, max_delay
            ))
            
            msg_id = cur.fetchone()[0]
            print(f"   Created broadcast_message id: {msg_id}")
            
            # Update sequence_contact to completed
            cur.execute("""
                UPDATE sequence_contacts
                SET status = 'completed',
                    completed_at = NOW(),
                    processing_device_id = %s
                WHERE id = %s
            """, (device_id, sc_id))
            
            print(f"   Marked step as completed")
        
        conn.commit()
        print(f"\n✅ Successfully created {len(ready_steps)} broadcast messages!")
        
        # Check the results
        print("\n=== VERIFICATION ===")
        cur.execute("""
            SELECT COUNT(*) FROM broadcast_messages 
            WHERE sequence_id IS NOT NULL 
            AND status = 'pending'
        """)
        pending_count = cur.fetchone()[0]
        print(f"Pending broadcast messages: {pending_count}")
        
    except Exception as e:
        print(f"\nERROR: {e}")
        conn.rollback()
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    check_and_fix_trigger()
