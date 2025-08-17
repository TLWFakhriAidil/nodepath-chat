import mysql.connector
from datetime import datetime
import sys

sys.stdout.reconfigure(encoding='utf-8')

config = {
    'host': '159.89.198.71',
    'port': 3306,
    'database': 'admin_railway',
    'user': 'admin_aqil',
    'password': 'admin_aqil'
}

try:
    conn = mysql.connector.connect(**config)
    cursor = conn.cursor(dictionary=True)
    
    print("=== ANALYZING THE ISSUE ===\n")
    
    # Check when the processor last worked
    query1 = """
    SELECT 
        MAX(sent_at) as last_sent,
        MAX(processing_started_at) as last_processing_started,
        TIMESTAMPDIFF(HOUR, MAX(sent_at), NOW()) as hours_since_last_sent
    FROM broadcast_messages
    WHERE sent_at IS NOT NULL
    """
    
    cursor.execute(query1)
    last_activity = cursor.fetchone()
    
    print(f"Last message sent: {last_activity['last_sent']}")
    print(f"Hours since last sent: {last_activity['hours_since_last_sent']} hours")
    print(f"Last processing started: {last_activity['last_processing_started']}")
    
    # Check the 2 messages that should be processed NOW
    print("\n=== MESSAGES THAT SHOULD BE PROCESSED NOW ===")
    
    query2 = """
    SELECT 
        id,
        device_id,
        recipient_phone,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at,
        processing_worker_id,
        campaign_id,
        sequence_id,
        LEFT(content, 50) as content_preview
    FROM broadcast_messages
    WHERE status = 'pending'
        AND scheduled_at <= DATE_ADD(NOW(), INTERVAL 8 HOUR)
        AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
        AND processing_worker_id IS NULL
    """
    
    cursor.execute(query2)
    eligible = cursor.fetchall()
    
    if eligible:
        for msg in eligible:
            print(f"\nID: {msg['id']}")
            print(f"Phone: {msg['recipient_phone']}")
            print(f"Scheduled: {msg['scheduled_at']}")
            print(f"Device: {msg['device_id']}")
            print(f"Content: {msg['content_preview']}...")
            
            # Check device status
            cursor.execute("SELECT device_name, status FROM user_devices WHERE id = %s", (msg['device_id'],))
            device = cursor.fetchone()
            if device:
                print(f"Device Name: {device['device_name']}")
                print(f"Device Status: {device['status']}")
    
    # Check campaign/sequence status
    print("\n\n=== CAMPAIGN/SEQUENCE ANALYSIS ===")
    
    query3 = """
    SELECT 
        CASE 
            WHEN campaign_id IS NOT NULL THEN 'Campaign'
            WHEN sequence_id IS NOT NULL THEN 'Sequence'
            ELSE 'Direct'
        END as message_type,
        COUNT(*) as count,
        SUM(CASE WHEN processing_worker_id IS NOT NULL THEN 1 ELSE 0 END) as with_worker
    FROM broadcast_messages
    WHERE status = 'pending'
        AND DATE(scheduled_at) = '2025-08-11'
    GROUP BY message_type
    """
    
    cursor.execute(query3)
    types = cursor.fetchall()
    
    for msg_type in types:
        print(f"\n{msg_type['message_type']} messages:")
        print(f"Total pending: {msg_type['count']}")
        print(f"With worker ID: {msg_type['with_worker']}")
    
    # Check if there's a pattern
    print("\n\n=== LOOKING FOR PATTERNS ===")
    
    # Are all pending messages from specific sequences?
    query4 = """
    SELECT 
        s.name as sequence_name,
        COUNT(bm.id) as pending_count
    FROM broadcast_messages bm
    JOIN sequences s ON bm.sequence_id = s.id
    WHERE bm.status = 'pending'
        AND DATE(bm.scheduled_at) = '2025-08-11'
    GROUP BY s.id, s.name
    """
    
    cursor.execute(query4)
    sequences = cursor.fetchall()
    
    if sequences:
        print("\nPending messages by sequence:")
        for seq in sequences:
            print(f"- {seq['sequence_name']}: {seq['pending_count']} messages")
    
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
