import psycopg2
from datetime import datetime

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def check_sequence_issues():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== CHECKING SEQUENCE ISSUES ===")
    print("")
    
    try:
        # 1. Check sequence contacts order
        print("1. Sequence Contacts Order:")
        cur.execute("""
            SELECT 
                contact_phone,
                current_step,
                status,
                next_trigger_time,
                assigned_device_id,
                sequence_stepid
            FROM sequence_contacts
            WHERE sequence_id = 'deccef4f-8ae1-4ed6-891c-bcb7d12baa8a'
            ORDER BY contact_phone, current_step
        """)
        
        print(f"{'Phone':<15} {'Step':<6} {'Status':<12} {'Next Trigger':<20} {'Device ID':<40}")
        print("-" * 100)
        for row in cur.fetchall():
            phone = row[0]
            step = row[1]
            status = row[2]
            trigger = row[3].strftime("%Y-%m-%d %H:%M:%S") if row[3] else "None"
            device = row[4][:8] + "..." if row[4] else "None"
            print(f"{phone:<15} {step:<6} {status:<12} {trigger:<20} {device:<40}")
        
        # 2. Check broadcast messages
        print("\n2. Broadcast Messages Created:")
        cur.execute("""
            SELECT 
                bm.recipient_phone,
                bm.device_id,
                bm.status,
                bm.created_at,
                bm.sequence_stepid,
                sc.current_step,
                sc.assigned_device_id
            FROM broadcast_messages bm
            LEFT JOIN sequence_contacts sc ON sc.sequence_stepid = bm.sequence_stepid::uuid
            WHERE bm.sequence_id = 'deccef4f-8ae1-4ed6-891c-bcb7d12baa8a'
            ORDER BY bm.created_at DESC
        """)
        
        print(f"\n{'Phone':<15} {'BM Device':<15} {'SC Device':<15} {'Step':<6} {'Status':<10} {'Created':<20}")
        print("-" * 100)
        for row in cur.fetchall():
            phone = row[0]
            bm_device = row[1][:8] + "..." if row[1] else "None"
            status = row[2]
            created = row[3].strftime("%Y-%m-%d %H:%M:%S") if row[3] else "None"
            step = row[5] if row[5] else "?"
            sc_device = row[6][:8] + "..." if row[6] else "None"
            print(f"{phone:<15} {bm_device:<15} {sc_device:<15} {step:<6} {status:<10} {created:<20}")
        
        # 3. Check which steps should be processed first
        print("\n3. What SHOULD be processed (earliest pending per contact):")
        cur.execute("""
            WITH earliest_pending AS (
                SELECT DISTINCT ON (sc.sequence_id, sc.contact_phone)
                    sc.contact_phone,
                    sc.current_step,
                    sc.next_trigger_time,
                    sc.assigned_device_id
                FROM sequence_contacts sc
                WHERE sc.status = 'pending'
                    AND sc.sequence_id = 'deccef4f-8ae1-4ed6-891c-bcb7d12baa8a'
                ORDER BY sc.sequence_id, sc.contact_phone, sc.current_step ASC, sc.next_trigger_time ASC
            )
            SELECT * FROM earliest_pending
            ORDER BY next_trigger_time ASC
        """)
        
        print(f"\n{'Phone':<15} {'Step':<6} {'Next Trigger':<20} {'Will Process?':<15}")
        print("-" * 60)
        for row in cur.fetchall():
            phone = row[0]
            step = row[1]
            trigger = row[2]
            will_process = "YES" if trigger <= datetime.now() else f"NO (in {trigger - datetime.now()})"
            print(f"{phone:<15} {step:<6} {trigger.strftime('%Y-%m-%d %H:%M:%S'):<20} {will_process:<15}")
            
    except Exception as e:
        print(f"\nError: {e}")
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    check_sequence_issues()
