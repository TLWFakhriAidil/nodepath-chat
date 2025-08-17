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
        print("=== DUPLICATE MESSAGE ANALYSIS - August 6, 2025 ===\n")
        
        # Key Finding: Check for messages from different sequences
        query = """
        SELECT 
            recipient_phone,
            COUNT(DISTINCT sequence_id) as num_sequences,
            COUNT(*) as total_messages,
            GROUP_CONCAT(DISTINCT sequence_id) as sequence_ids,
            GROUP_CONCAT(DISTINCT sequence_stepid) as step_ids,
            MIN(created_at) as first_created,
            MAX(created_at) as last_created
        FROM broadcast_messages
        WHERE (DATE(scheduled_at) = '2025-08-06' OR DATE(sent_at) = '2025-08-06')
            AND sequence_id IS NOT NULL
        GROUP BY recipient_phone
        HAVING COUNT(DISTINCT sequence_id) > 1
        ORDER BY num_sequences DESC
        LIMIT 20
        """
        
        cursor.execute(query)
        multi_sequence = cursor.fetchall()
        
        print("=== CRITICAL FINDING: Recipients in Multiple Sequences ===\n")
        for row in multi_sequence:
            print(f"Phone: {row['recipient_phone']}")
            print(f"  In {row['num_sequences']} different sequences!")
            print(f"  Total messages: {row['total_messages']}")
            print(f"  Sequences: {row['sequence_ids']}")
            print(f"  Created between: {row['first_created']} and {row['last_created']}")
            print()
        
        # Check specific sequences
        print("\n=== Sequence Details ===\n")
        
        sequence_ids = [
            'deccef4f-8ae1-4ed6-891c-bcb7d12baa8a',
            '0be82745-8f68-4352-abd0-0b405b43a905'
        ]
        
        for seq_id in sequence_ids:
            cursor.execute("""
                SELECT name, status, trigger, target_status 
                FROM sequences 
                WHERE id = %s
            """, (seq_id,))
            seq_info = cursor.fetchone()
            
            if seq_info:
                print(f"Sequence: {seq_id}")
                print(f"  Name: {seq_info['name']}")
                print(f"  Status: {seq_info['status']}")
                print(f"  Trigger: {seq_info['trigger']}")
                print(f"  Target Status: {seq_info['target_status']}")
                
                # Count messages
                cursor.execute("""
                    SELECT 
                        COUNT(*) as total,
                        COUNT(DISTINCT recipient_phone) as unique_recipients,
                        SUM(CASE WHEN status = 'sent' THEN 1 ELSE 0 END) as sent,
                        SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending
                    FROM broadcast_messages
                    WHERE sequence_id = %s
                        AND DATE(scheduled_at) = '2025-08-06'
                """, (seq_id,))
                
                stats = cursor.fetchone()
                print(f"  Messages on Aug 6: {stats['total']} total")
                print(f"    Unique recipients: {stats['unique_recipients']}")
                print(f"    Sent: {stats['sent']}")
                print(f"    Pending: {stats['pending']}")
                print()
        
        # Check enrollment patterns
        print("\n=== Enrollment Pattern Analysis ===\n")
        
        query = """
        SELECT 
            l.phone,
            l.trigger as lead_trigger,
            COUNT(DISTINCT bm.sequence_id) as num_sequences,
            GROUP_CONCAT(DISTINCT s.name) as sequence_names,
            GROUP_CONCAT(DISTINCT s.trigger) as sequence_triggers
        FROM leads l
        JOIN broadcast_messages bm ON l.phone = bm.recipient_phone
        JOIN sequences s ON bm.sequence_id = s.id
        WHERE bm.sequence_id IS NOT NULL
            AND l.phone IN ('601114161956', '601119708692', '601125726197')
        GROUP BY l.phone, l.trigger
        """
        
        cursor.execute(query)
        enrollment = cursor.fetchall()
        
        for row in enrollment:
            print(f"Phone: {row['phone']}")
            print(f"  Lead trigger: {row['lead_trigger']}")
            print(f"  Enrolled in {row['num_sequences']} sequences")
            print(f"  Sequences: {row['sequence_names']}")
            print(f"  Sequence triggers: {row['sequence_triggers']}")
            print()
        
        # Check time differences
        print("\n=== Time Pattern Analysis ===\n")
        
        query = """
        SELECT 
            recipient_phone,
            sequence_id,
            MIN(scheduled_at) as first_scheduled,
            MAX(scheduled_at) as last_scheduled,
            TIMESTAMPDIFF(HOUR, MIN(scheduled_at), MAX(scheduled_at)) as hour_diff,
            COUNT(*) as message_count
        FROM broadcast_messages
        WHERE DATE(scheduled_at) = '2025-08-06'
            AND sequence_id IS NOT NULL
            AND recipient_phone IN ('601114161956', '601119708692')
        GROUP BY recipient_phone, sequence_id
        HAVING COUNT(*) > 1
        """
        
        cursor.execute(query)
        time_patterns = cursor.fetchall()
        
        for row in time_patterns:
            print(f"Phone: {row['recipient_phone']}")
            print(f"  Sequence: {row['sequence_id']}")
            print(f"  Messages: {row['message_count']}")
            print(f"  First scheduled: {row['first_scheduled']}")
            print(f"  Last scheduled: {row['last_scheduled']}")
            print(f"  Time difference: {row['hour_diff']} hours")
            print()

finally:
    connection.close()

print("\n=== Analysis Complete ===")
