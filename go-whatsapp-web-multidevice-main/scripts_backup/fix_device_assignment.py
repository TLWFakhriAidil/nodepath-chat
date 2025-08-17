import psycopg2

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def fix_device_assignment():
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("=== FIXING DEVICE ASSIGNMENT ISSUE ===\n")
    
    try:
        # Option 1: Update sequence_contacts to use the lead's device (if that's what you want)
        print("Option 1: Update sequence_contacts to match lead device")
        print("Do you want sequence to use the SAME device as the lead? (y/n): ", end='')
        choice = input().strip().lower()
        
        if choice == 'y':
            cur.execute("""
                UPDATE sequence_contacts sc
                SET assigned_device_id = l.device_id
                FROM leads l
                WHERE sc.contact_phone = l.phone
                  AND l.device_id IS NOT NULL
            """)
            updated = cur.rowcount
            print(f"Updated {updated} sequence_contacts to use lead device")
            conn.commit()
        else:
            print("Keeping sequence_contacts with current assigned_device_id")
            
        # Show current state
        print("\n\nCurrent device assignments:")
        cur.execute("""
            SELECT DISTINCT
                sc.assigned_device_id as sequence_device,
                l.device_id as lead_device,
                ud1.device_name as sequence_device_name,
                ud2.device_name as lead_device_name,
                COUNT(*) as count
            FROM sequence_contacts sc
            LEFT JOIN leads l ON l.phone = sc.contact_phone
            LEFT JOIN user_devices ud1 ON ud1.id = sc.assigned_device_id
            LEFT JOIN user_devices ud2 ON ud2.id = l.device_id
            GROUP BY sc.assigned_device_id, l.device_id, ud1.device_name, ud2.device_name
        """)
        
        results = cur.fetchall()
        for row in results:
            print(f"\nSequence Device: {row[0]} ({row[2]})")
            print(f"Lead Device: {row[1]} ({row[3]})")
            print(f"Count: {row[4]}")
            
    except Exception as e:
        conn.rollback()
        print(f"Error: {e}")
    finally:
        cur.close()
        conn.close()

if __name__ == "__main__":
    fix_device_assignment()
