import pymysql
from datetime import datetime, timedelta
from collections import defaultdict

# Database connection
connection = pymysql.connect(
    host='159.89.198.71',
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    charset='utf8mb4',
    cursorclass=pymysql.cursors.DictCursor
)

try:
    with connection.cursor() as cursor:
        print("=== ROOT CAUSE ANALYSIS: Duplicate Messages on August 6, 2025 ===\n")
        
        # The KEY FINDING from the previous output:
        # People are enrolled in MULTIPLE sequences!
        # Example: 601114161956 is in both:
        # - 0be82745-8f68-4352-abd0-0b405b43a905 (sent at 01:08)
        # - deccef4f-8ae1-4ed6-891c-bcb7d12baa8a (scheduled at 13:08)
        
        print("=== THE PROBLEM: Same Recipients in Multiple Sequences ===\n")
        
        # Check the two main sequences
        query = """
        SELECT id, name, status, `trigger`, target_status 
        FROM sequences 
        WHERE id IN ('deccef4f-8ae1-4ed6-891c-bcb7d12baa8a', '0be82745-8f68-4352-abd0-0b405b43a905')
        """
        
        cursor.execute(query)
        sequences = cursor.fetchall()
        
        for seq in sequences:
            print(f"Sequence: {seq['name']} ({seq['id'][:8]}...)")
            print(f"  Status: {seq['status']}")
            print(f"  Trigger: {seq['trigger']}")
            print(f"  Target: {seq['target_status']}")
            print()
        
        # Check overlap
        print("=== Checking Recipient Overlap ===\n")
        
        query = """
        SELECT 
            COUNT(DISTINCT recipient_phone) as total_recipients,
            SUM(CASE WHEN num_sequences = 1 THEN 1 ELSE 0 END) as single_sequence,
            SUM(CASE WHEN num_sequences = 2 THEN 1 ELSE 0 END) as in_two_sequences,
            SUM(CASE WHEN num_sequences > 2 THEN 1 ELSE 0 END) as in_many_sequences
        FROM (
            SELECT 
                recipient_phone,
                COUNT(DISTINCT sequence_id) as num_sequences
            FROM broadcast_messages
            WHERE sequence_id IS NOT NULL
                AND DATE(scheduled_at) = '2025-08-06'
            GROUP BY recipient_phone
        ) as recipient_counts
        """
        
        cursor.execute(query)
        overlap = cursor.fetchone()
        
        print(f"Total recipients with messages on Aug 6: {overlap['total_recipients']}")
        print(f"  In only 1 sequence: {overlap['single_sequence']}")
        print(f"  In 2 sequences: {overlap['in_two_sequences']} ⚠️")
        print(f"  In 3+ sequences: {overlap['in_many_sequences']} ⚠️")
        
        # Show specific examples
        print("\n=== Example Recipients in Multiple Sequences ===\n")
        
        query = """
        SELECT 
            bm.recipient_phone,
            bm.recipient_name,
            s.name as sequence_name,
            bm.scheduled_at,
            TIME(bm.scheduled_at) as scheduled_time,
            bm.status,
            bm.sent_at
        FROM broadcast_messages bm
        JOIN sequences s ON bm.sequence_id = s.id
        WHERE bm.recipient_phone IN (
            '601114161956', '601119708692', '60122712014'
        )
        AND DATE(bm.scheduled_at) = '2025-08-06'
        ORDER BY bm.recipient_phone, bm.scheduled_at
        """
        
        cursor.execute(query)
        examples = cursor.fetchall()
        
        current_phone = None
        for msg in examples:
            if current_phone != msg['recipient_phone']:
                if current_phone:
                    print()
                current_phone = msg['recipient_phone']
                print(f"📱 {msg['recipient_phone']} ({msg['recipient_name']})")
            
            print(f"  → {msg['sequence_name']}")
            print(f"     Scheduled: {msg['scheduled_time']} | Status: {msg['status']}")
            if msg['sent_at']:
                print(f"     Sent at: {msg['sent_at']}")
        
        # Check enrollment timing
        print("\n\n=== When Were They Enrolled? ===\n")
        
        query = """
        SELECT 
            recipient_phone,
            sequence_id,
            MIN(created_at) as enrolled_at,
            COUNT(*) as message_count
        FROM broadcast_messages
        WHERE recipient_phone IN ('601114161956', '60122712014')
            AND sequence_id IS NOT NULL
        GROUP BY recipient_phone, sequence_id
        ORDER BY recipient_phone, enrolled_at
        """
        
        cursor.execute(query)
        enrollments = cursor.fetchall()
        
        current_phone = None
        for enroll in enrollments:
            if current_phone != enroll['recipient_phone']:
                if current_phone:
                    print()
                current_phone = enroll['recipient_phone']
                print(f"📱 {enroll['recipient_phone']}")
            
            print(f"  Enrolled in {enroll['sequence_id'][:8]}... at {enroll['enrolled_at']}")
            print(f"  ({enroll['message_count']} messages scheduled)")
        
        # The real issue
        print("\n\n=== THE ROOT CAUSE ===\n")
        print("❌ PROBLEM: Recipients are being enrolled in MULTIPLE sequences")
        print("   - Each sequence schedules its own messages")
        print("   - Worker Pool processes ALL messages correctly")
        print("   - Result: Same person gets messages from different sequences")
        print("\n📍 This is NOT a Worker Pool issue!")
        print("   The duplicate prevention only works WITHIN a sequence,")
        print("   not ACROSS different sequences.")
        
        # Check trigger overlap
        print("\n\n=== Checking Trigger Overlap ===\n")
        
        query = """
        SELECT 
            l.phone,
            l.trigger as lead_trigger,
            GROUP_CONCAT(DISTINCT s.name) as enrolled_sequences
        FROM leads l
        JOIN broadcast_messages bm ON l.phone = bm.recipient_phone
        JOIN sequences s ON bm.sequence_id = s.id
        WHERE l.phone IN ('601114161956', '60122712014')
        GROUP BY l.phone, l.trigger
        """
        
        cursor.execute(query)
        triggers = cursor.fetchall()
        
        for row in triggers:
            print(f"Lead {row['phone']} has trigger: '{row['lead_trigger']}'")
            print(f"  Enrolled in: {row['enrolled_sequences']}")

finally:
    connection.close()

print("\n=== Analysis Complete ===")
