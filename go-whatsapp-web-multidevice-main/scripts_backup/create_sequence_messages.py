import psycopg2
from datetime import datetime

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def create_pending_messages():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== CREATING BROADCAST MESSAGES FOR READY STEPS ===")
    
    try:
        # Get ready steps
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
        
        created_count = 0
        for step in ready_steps:
            (sc_id, seq_id, phone, name, step_num, trigger, device_id, 
             user_id, step_id, content, msg_type, media_url, min_delay, max_delay) = step
            
            print(f"\nProcessing {phone} - Step {step_num}:")
            print(f"  Content: {content[:50]}...")
            print(f"  Device: {device_id}")
            
            # Check if message already exists
            cur.execute("""
                SELECT id FROM broadcast_messages 
                WHERE sequence_stepid = %s 
                AND recipient_phone = %s
            """, (step_id, phone))
            
            existing = cur.fetchone()
            if existing:
                print(f"  Message already exists (id: {existing[0]})")
                continue
            
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
            print(f"  Created broadcast_message id: {msg_id}")
            
            # Update sequence_contact to completed
            cur.execute("""
                UPDATE sequence_contacts
                SET status = 'completed',
                    completed_at = NOW(),
                    processing_device_id = %s
                WHERE id = %s
            """, (device_id, sc_id))
            
            print(f"  Marked step as completed")
            created_count += 1
        
        conn.commit()
        print(f"\n[OK] Successfully created {created_count} broadcast messages!")
        
        # Verify
        cur.execute("""
            SELECT bm.id, bm.recipient_phone, bm.status, bm.device_id,
                   ud.device_name, ud.status as device_status
            FROM broadcast_messages bm
            JOIN user_devices ud ON ud.id = bm.device_id
            WHERE bm.sequence_id IS NOT NULL 
            AND bm.status = 'pending'
            ORDER BY bm.created_at DESC
            LIMIT 10
        """)
        
        messages = cur.fetchall()
        print("\n=== PENDING MESSAGES ===")
        for msg in messages:
            print(f"ID: {msg[0]}, To: {msg[1]}, Device: {msg[4]} ({msg[5]})")
        
    except Exception as e:
        print(f"\nERROR: {e}")
        import traceback
        traceback.print_exc()
        conn.rollback()
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    create_pending_messages()
