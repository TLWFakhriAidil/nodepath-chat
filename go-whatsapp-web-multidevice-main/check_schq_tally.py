import pymysql
from datetime import date, datetime

conn = pymysql.connect(
    host='159.89.198.71',
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    port=3306
)
cursor = conn.cursor()

print("Checking SCHQ-509 sequence message tallies...")
print("="*100)

# First, find the SCHQ-509 sequence and its steps
cursor.execute("""
SELECT s.id, s.name, ss.id as step_id, ss.day_number
FROM sequences s
JOIN sequence_steps ss ON s.id = ss.sequence_id
WHERE s.name = 'SCHQ-509'
ORDER BY ss.day_number
""")
sequence_info = cursor.fetchall()

if not sequence_info:
    print("No SCHQ-509 sequence found!")
else:
    seq_id = sequence_info[0][0]
    print(f"Found sequence: {sequence_info[0][1]} (ID: {seq_id})")
    print("\nStep Details:")
    
    for _, _, step_id, day_num in sequence_info:
        # Get counts for each step
        cursor.execute("""
        SELECT 
            COUNT(DISTINCT CASE WHEN status = 'sent' AND DATE(sent_at) = '2025-08-07' THEN id END) as sent_aug7,
            COUNT(DISTINCT CASE WHEN status = 'sent' AND DATE(sent_at) = '2025-08-06' THEN id END) as sent_aug6,
            COUNT(DISTINCT CASE WHEN status = 'sent' AND DATE(sent_at) = '2025-08-02' THEN id END) as sent_aug2,
            COUNT(DISTINCT CASE WHEN status = 'sent' THEN id END) as sent_total,
            COUNT(DISTINCT CASE WHEN status = 'failed' THEN id END) as failed_total,
            COUNT(DISTINCT CASE WHEN status = 'pending' THEN id END) as pending_total,
            COUNT(DISTINCT id) as total_messages
        FROM broadcast_messages
        WHERE sequence_stepid = %s
        """, (step_id,))
        
        counts = cursor.fetchone()
        
        print(f"\nDay {day_num} (Step ID: {step_id[:8]}...):")
        print(f"  Sent on Aug 7: {counts[0]}")
        print(f"  Sent on Aug 6: {counts[1]}")
        print(f"  Sent on Aug 2: {counts[2]}")
        print(f"  Total Sent (all time): {counts[3]}")
        print(f"  Failed: {counts[4]}")
        print(f"  Pending: {counts[5]}")
        print(f"  Total Messages: {counts[6]}")
        
        # Show sample of sent messages for Aug 2
        if counts[2] > 0:
            cursor.execute("""
            SELECT recipient_phone, recipient_name, sent_at
            FROM broadcast_messages
            WHERE sequence_stepid = %s 
            AND status = 'sent' 
            AND DATE(sent_at) = '2025-08-02'
            LIMIT 5
            """, (step_id,))
            
            samples = cursor.fetchall()
            print(f"\n  Sample messages sent on Aug 2:")
            for phone, name, sent_time in samples:
                print(f"    - {phone} ({name}) at {sent_time}")

print("\n" + "="*100)
print("ISSUE: The modal is showing ALL sent messages (all dates), not filtered by the selected date!")
print("When you filter for Aug 7, it should only show messages sent on Aug 7, not Aug 2 messages.")

conn.close()
