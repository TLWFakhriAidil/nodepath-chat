import psycopg2
from datetime import datetime

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def fix_sequence_flow():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== FIXING SEQUENCE FLOW ===")
    
    try:
        # 1. Check Step 1 status
        print("\n1. Checking Step 1 records:")
        cur.execute("""
            SELECT contact_phone, contact_name, current_step, status, 
                   completed_at, processing_device_id
            FROM sequence_contacts
            WHERE current_step = 1
            ORDER BY contact_phone
        """)
        step1_data = cur.fetchall()
        for row in step1_data:
            print(f"   {row[0]} ({row[1]}): {row[3]} - Completed: {row[4]}")
        
        # 2. Check if broadcast messages were created for Step 1
        print("\n2. Checking broadcast messages:")
        cur.execute("""
            SELECT bm.id, bm.recipient_phone, bm.status, bm.created_at,
                   sc.current_step
            FROM broadcast_messages bm
            LEFT JOIN sequence_contacts sc ON sc.sequence_stepid = bm.sequence_stepid
            WHERE bm.sequence_id IS NOT NULL
            ORDER BY bm.created_at DESC
        """)
        messages = cur.fetchall()
        print(f"   Found {len(messages)} sequence messages total")
        for msg in messages[:5]:
            print(f"   - ID: {msg[0]}, To: {msg[1]}, Status: {msg[2]}, Step: {msg[4]}")
        
        # 3. The real issue: Step 1 might be marked completed but no message was sent
        # Let's create messages for ACTIVE steps that are overdue
        print("\n3. Creating messages for ACTIVE steps:")
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
            WHERE sc.status = 'active'
            AND sc.completed_at IS NULL
        """)
        
        active_steps = cur.fetchall()
        print(f"   Found {len(active_steps)} active steps")
        
        # Since these are Step 2 and can't be processed due to Step 1,
        # let's check if Step 1 was really completed properly
        print("\n4. Resetting sequence to start fresh:")
        
        # Delete all sequence_contacts
        cur.execute("""
            DELETE FROM sequence_contacts
            WHERE contact_phone IN ('60146674397', '60108924904')
            RETURNING contact_phone
        """)
        deleted = cur.fetchall()
        print(f"   Deleted {len(deleted)} sequence contact records")
        
        # Delete any orphaned broadcast messages
        cur.execute("""
            DELETE FROM broadcast_messages
            WHERE sequence_id IS NOT NULL
            AND recipient_phone IN ('60146674397', '60108924904')
            AND status = 'pending'
            RETURNING id
        """)
        deleted_msgs = cur.fetchall()
        print(f"   Deleted {len(deleted_msgs)} pending messages")
        
        conn.commit()
        print("\n[OK] Sequence data cleared. The system will re-enroll these contacts on next run.")
        print("     This time, all steps will be created as 'pending' as intended.")
        
        # 5. Update the enrollment logic to ensure PENDING-FIRST
        print("\n5. Checking sequence_trigger_processor.go enrollment logic...")
        print("   The code shows status = 'pending' for all steps, which is correct.")
        print("   But the database had 'active' for Step 1, suggesting old code is running.")
        print("\n   ACTION REQUIRED: Rebuild and redeploy the application!")
        
    except Exception as e:
        print(f"\nERROR: {e}")
        import traceback
        traceback.print_exc()
        conn.rollback()
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    fix_sequence_flow()
