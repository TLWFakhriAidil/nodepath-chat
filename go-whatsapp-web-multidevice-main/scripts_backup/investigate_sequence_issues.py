import psycopg2
from datetime import datetime

# Database connection
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

try:
    # Connect to database
    conn = psycopg2.connect(DB_URI)
    cur = conn.cursor()
    
    print("Connected to PostgreSQL database...")
    print("=" * 60)
    
    # 1. Check why current_step is all 4
    print("\n1. Investigating current_step values...")
    cur.execute("""
        SELECT 
            sc.id,
            sc.contact_phone,
            sc.sequence_stepid,
            ss.day_number as step_day_number,
            sc.current_step,
            sc.status,
            sc.completed_at,
            sc.next_trigger_time,
            ss.trigger_delay_hours
        FROM sequence_contacts sc
        JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
        WHERE sc.contact_phone = '60199204337'
        AND sc.sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
        ORDER BY ss.day_number
    """)
    
    results = cur.fetchall()
    print(f"Details for phone 60199204337:")
    for row in results:
        print(f"\n  Record ID: {row[0][:8]}...")
        print(f"  Step Day Number: {row[3]} (from sequence_steps)")
        print(f"  Current Step: {row[4]} (from sequence_contacts)")
        print(f"  Status: {row[5]}")
        print(f"  Completed At: {row[6]}")
        print(f"  Next Trigger Time: {row[7]}")
        print(f"  Delay Hours: {row[8]}")
    
    # 2. Check timing - were delays respected?
    print("\n\n2. Checking if delays were respected...")
    cur.execute("""
        SELECT 
            sc.contact_phone,
            ss.day_number,
            sc.completed_at,
            sc.next_trigger_time,
            ss.trigger_delay_hours,
            LAG(sc.completed_at) OVER (PARTITION BY sc.contact_phone ORDER BY ss.day_number) as prev_completed
        FROM sequence_contacts sc
        JOIN sequence_steps ss ON ss.id = sc.sequence_stepid
        WHERE sc.contact_phone IN ('60199204337', '60196153796')
        AND sc.sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
        ORDER BY sc.contact_phone, ss.day_number
    """)
    
    timing_results = cur.fetchall()
    current_phone = None
    for row in timing_results:
        phone = row[0]
        if phone != current_phone:
            print(f"\nPhone: {phone}")
            current_phone = phone
        
        day = row[1]
        completed = row[2]
        next_trigger = row[3]
        delay_hours = row[4]
        prev_completed = row[5]
        
        print(f"  Day {day}:")
        print(f"    Completed: {completed}")
        print(f"    Next Trigger Was: {next_trigger}")
        print(f"    Configured Delay: {delay_hours} hours")
        
        if prev_completed and completed:
            time_diff = (completed - prev_completed).total_seconds() / 3600
            print(f"    Actual time from previous: {time_diff:.2f} hours")
    
    # 3. Check broadcast_messages to see if messages were actually sent
    print("\n\n3. Checking if messages were actually sent...")
    cur.execute("""
        SELECT 
            bm.recipient_phone,
            bm.status as message_status,
            bm.sent_at,
            bm.created_at,
            bm.content
        FROM broadcast_messages bm
        WHERE bm.sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
        AND bm.recipient_phone IN ('60199204337', '60196153796')
        ORDER BY bm.recipient_phone, bm.created_at
        LIMIT 10
    """)
    
    messages = cur.fetchall()
    if messages:
        print(f"Found {len(messages)} broadcast messages:")
        for msg in messages:
            print(f"\n  Phone: {msg[0]}")
            print(f"  Status: {msg[1]}")
            print(f"  Created: {msg[2]}")
            print(f"  Sent: {msg[3]}")
            print(f"  Content preview: {msg[4][:50]}...")
    else:
        print("No broadcast messages found for these sequence contacts!")
    
    # 4. Check the sequence processor logic
    print("\n\n4. Understanding the enrollment logic...")
    cur.execute("""
        SELECT 
            COUNT(*) as total_records,
            COUNT(DISTINCT contact_phone) as unique_contacts,
            COUNT(DISTINCT sequence_stepid) as unique_steps
        FROM sequence_contacts
        WHERE sequence_id = (SELECT id FROM sequences WHERE name = 'WARM Sequence')
    """)
    
    summary = cur.fetchone()
    print(f"Sequence enrollment summary:")
    print(f"  Total records: {summary[0]}")
    print(f"  Unique contacts: {summary[1]}")
    print(f"  Unique steps: {summary[2]}")
    print(f"  Records per contact: {summary[0] / summary[1] if summary[1] > 0 else 0}")
    
    print("\n" + "=" * 60)
    print("FINDINGS:")
    print("1. current_step is showing 4 for all because it's being set wrong")
    print("2. All marked 'completed' at the same time - delays NOT respected")
    print("3. Need to check if messages were actually sent")
    
    cur.close()
    conn.close()
    
except Exception as e:
    print(f"Error: {e}")
    if conn:
        conn.close()
