import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def check_device_issue():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== CHECKING DEVICE ID ISSUES ===\n")
    
    try:
        # 1. Check the COALESCE query result
        print("1. CHECKING COALESCE QUERY RESULT:")
        cur.execute("""
            SELECT 
                sc.contact_phone,
                sc.contact_name,
                sc.assigned_device_id,
                l.device_id as lead_device_id,
                COALESCE(sc.assigned_device_id, l.device_id) as preferred_device_id,
                CASE 
                    WHEN sc.assigned_device_id IS NOT NULL THEN 'Using sequence device'
                    WHEN l.device_id IS NOT NULL THEN 'Using lead device'
                    ELSE 'No device!'
                END as device_source
            FROM sequence_contacts sc
            LEFT JOIN leads l ON l.phone = sc.contact_phone
            WHERE sc.status = 'active'
        """)
        
        results = cur.fetchall()
        print(f"{'Phone':<15} {'Name':<10} {'Assigned Device':<40} {'Lead Device':<40} {'Preferred Device':<40} {'Source':<25}")
        print("-" * 170)
        for row in results:
            phone, name, assigned, lead, preferred, source = row
            print(f"{phone:<15} {name or '':<10} {assigned or 'NULL':<40} {lead or 'NULL':<40} {preferred or 'NULL':<40} {source:<25}")
        
        # 2. Check broadcast_messages created
        print("\n\n2. RECENT BROADCAST_MESSAGES:")
        cur.execute("""
            SELECT 
                bm.recipient_phone,
                bm.device_id,
                bm.sequence_id,
                bm.status,
                bm.created_at
            FROM broadcast_messages bm
            WHERE bm.sequence_id IS NOT NULL
            ORDER BY bm.created_at DESC
            LIMIT 10
        """)
        
        results = cur.fetchall()
        print(f"{'Phone':<15} {'Device ID':<40} {'Sequence ID':<40} {'Status':<10} {'Created':<25}")
        print("-" * 130)
        for row in results:
            phone, device, seq, status, created = row
            print(f"{phone:<15} {device:<40} {seq:<40} {status:<10} {created}")
        
        # 3. Check if leads have different devices than sequence_contacts
        print("\n\n3. DEVICE MISMATCHES:")
        cur.execute("""
            SELECT 
                sc.contact_phone,
                sc.assigned_device_id,
                l.device_id,
                CASE 
                    WHEN sc.assigned_device_id = l.device_id THEN 'MATCH'
                    WHEN sc.assigned_device_id IS NULL THEN 'NULL assigned'
                    ELSE 'MISMATCH!'
                END as status
            FROM sequence_contacts sc
            LEFT JOIN leads l ON l.phone = sc.contact_phone
            WHERE sc.status IN ('active', 'pending')
        """)
        
        results = cur.fetchall()
        mismatches = 0
        for row in results:
            if row[3] == 'MISMATCH!':
                mismatches += 1
                print(f"Phone: {row[0]}")
                print(f"  Sequence device: {row[1]}")
                print(f"  Lead device:     {row[2]}")
                print()
        
        print(f"Total mismatches found: {mismatches}")
        
        # 4. Check current active sequence_contacts
        print("\n\n4. CURRENT ACTIVE SEQUENCE CONTACTS:")
        cur.execute("""
            SELECT 
                id,
                contact_phone,
                contact_name,
                current_step,
                current_trigger,
                assigned_device_id,
                processing_device_id,
                next_trigger_time
            FROM sequence_contacts
            WHERE status = 'active'
        """)
        
        results = cur.fetchall()
        for row in results:
            print(f"ID: {row[0]}")
            print(f"  Phone: {row[1]} ({row[2]})")
            print(f"  Step: {row[3]}, Trigger: {row[4]}")
            print(f"  Assigned device: {row[5]}")
            print(f"  Processing device: {row[6]}")
            print(f"  Next trigger: {row[7]}")
            print()
            
    except Exception as e:
        print(f"Error: {e}")
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    check_device_issue()
