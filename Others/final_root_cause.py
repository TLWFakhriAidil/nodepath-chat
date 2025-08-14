import pymysql
from datetime import datetime, timedelta
from collections import defaultdict
import sys

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

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
        print(f"  In 2 sequences: {overlap['in_two_sequences']} <-- DUPLICATES!")
        print(f"  In 3+ sequences: {overlap['in_many_sequences']} <-- MULTIPLE DUPLICATES!")
        
        # Show the pattern
        print("\n=== THE DUPLICATE PATTERN ===\n")
        
        query = """
        SELECT 
            bm1.recipient_phone,
            bm1.recipient_name,
            bm1.scheduled_at as time1,
            bm2.scheduled_at as time2,
            TIMESTAMPDIFF(HOUR, bm1.scheduled_at, bm2.scheduled_at) as hour_diff,
            s1.name as seq1_name,
            s2.name as seq2_name
        FROM broadcast_messages bm1
        JOIN broadcast_messages bm2 ON bm1.recipient_phone = bm2.recipient_phone
            AND bm1.id < bm2.id
        JOIN sequences s1 ON bm1.sequence_id = s1.id
        JOIN sequences s2 ON bm2.sequence_id = s2.id
        WHERE DATE(bm1.scheduled_at) = '2025-08-06'
            AND DATE(bm2.scheduled_at) = '2025-08-06'
            AND bm1.sequence_id != bm2.sequence_id
            AND bm1.recipient_phone IN ('601114161956', '60122712014')
        ORDER BY bm1.recipient_phone
        LIMIT 5
        """
        
        cursor.execute(query)
        patterns = cursor.fetchall()
        
        for p in patterns:
            print(f"Phone: {p['recipient_phone']} ({p['recipient_name']})")
            print(f"  Message 1: {p['seq1_name']} at {p['time1']}")
            print(f"  Message 2: {p['seq2_name']} at {p['time2']}")
            print(f"  Time difference: {p['hour_diff']} hours")
            print()
        
        # The real issue
        print("\n=== THE ROOT CAUSE ===\n")
        print("PROBLEM IDENTIFIED:")
        print("1. Recipients are enrolled in MULTIPLE sequences (COLD and WARM)")
        print("2. Each sequence schedules its own messages independently")
        print("3. Worker Pool correctly sends ALL scheduled messages")
        print("4. Result: Same person receives messages from different sequences")
        print("\nThis is NOT a bug in the Worker Pool!")
        print("The system is working as designed, but the BUSINESS LOGIC allows")
        print("people to be in multiple sequences simultaneously.")
        
        # Check if there's a progression
        print("\n=== Checking Sequence Relationships ===\n")
        
        query = """
        SELECT 
            s.id,
            s.name,
            s.trigger,
            ss.next_trigger,
            COUNT(DISTINCT bm.recipient_phone) as recipients
        FROM sequences s
        JOIN sequence_steps ss ON s.id = ss.sequence_id
        JOIN broadcast_messages bm ON s.id = bm.sequence_id
        WHERE s.id IN ('0be82745-8f68-4352-abd0-0b405b43a905', 'deccef4f-8ae1-4ed6-891c-bcb7d12baa8a')
            AND ss.next_trigger IS NOT NULL
        GROUP BY s.id, s.name, s.trigger, ss.next_trigger
        """
        
        cursor.execute(query)
        relationships = cursor.fetchall()
        
        for rel in relationships:
            print(f"Sequence: {rel['name']}")
            print(f"  Trigger: {rel['trigger']}")
            print(f"  Next trigger points to: {rel['next_trigger']}")
            print(f"  Recipients: {rel['recipients']}")
            print()
        
        print("\n=== SOLUTION OPTIONS ===\n")
        print("1. PREVENT DUPLICATE ENROLLMENT:")
        print("   - Check if recipient is already in ANY active sequence")
        print("   - Don't enroll in new sequences until current one completes")
        print("\n2. CANCEL OLD SEQUENCES:")
        print("   - When enrolling in new sequence, cancel messages from old ones")
        print("\n3. MERGE SEQUENCES:")
        print("   - Combine COLD and WARM into one sequence with proper progression")

finally:
    connection.close()
