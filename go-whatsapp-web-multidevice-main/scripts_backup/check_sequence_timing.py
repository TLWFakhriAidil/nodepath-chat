import psycopg2
from datetime import datetime
import pytz

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def check_sequence_timing():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== SEQUENCE TIMING CHECK ===")
    
    try:
        # Get current server time
        cur.execute("SELECT NOW(), CURRENT_TIMESTAMP")
        server_time = cur.fetchone()[0]
        print(f"Server time: {server_time}")
        print(f"Local time:  {datetime.now()}")
        
        # Check all pending steps
        cur.execute("""
            SELECT 
                sc.contact_phone, sc.contact_name, sc.current_step,
                sc.next_trigger_time,
                sc.next_trigger_time - NOW() as time_remaining,
                sc.status, sc.current_trigger
            FROM sequence_contacts sc
            WHERE sc.status IN ('pending', 'active')
            ORDER BY sc.contact_phone, sc.current_step
        """)
        
        steps = cur.fetchall()
        print(f"\n=== PENDING/ACTIVE STEPS ({len(steps)} total) ===")
        
        for step in steps:
            phone, name, step_num, trigger_time, remaining, status, trigger = step
            print(f"\n{phone} ({name}) - Step {step_num}:")
            print(f"  Status: {status}")
            print(f"  Trigger: {trigger}")
            print(f"  Scheduled: {trigger_time}")
            
            if remaining.total_seconds() > 0:
                hours = int(remaining.total_seconds() // 3600)
                minutes = int((remaining.total_seconds() % 3600) // 60)
                print(f"  Remaining: {hours}h {minutes}m")
            else:
                overdue_minutes = int(abs(remaining.total_seconds()) // 60)
                print(f"  OVERDUE BY: {overdue_minutes} minutes!")
        
        # Check the first step specifically
        print("\n=== CHECKING STEP 1 SPECIFICALLY ===")
        cur.execute("""
            SELECT 
                sc.*, 
                ss.content,
                sc.next_trigger_time <= NOW() as is_ready
            FROM sequence_contacts sc
            JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
            WHERE sc.current_step = 1
            AND sc.status = 'pending'
        """)
        
        step1_records = cur.fetchall()
        print(f"Found {len(step1_records)} Step 1 records in pending status")
        
        # Force process overdue steps
        print("\n=== FORCING OVERDUE STEPS ===")
        cur.execute("""
            UPDATE sequence_contacts
            SET next_trigger_time = NOW() - INTERVAL '1 minute'
            WHERE current_step = 1
            AND status = 'pending'
            AND next_trigger_time > NOW()
            RETURNING id, contact_phone, next_trigger_time
        """)
        
        updated = cur.fetchall()
        if updated:
            conn.commit()
            print(f"Updated {len(updated)} steps to be immediately ready")
            for u in updated:
                print(f"  - {u[1]}: new time {u[2]}")
        else:
            print("No updates needed - steps might already be ready")
            
    except Exception as e:
        print(f"\nERROR: {e}")
        import traceback
        traceback.print_exc()
        conn.rollback()
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    check_sequence_timing()
