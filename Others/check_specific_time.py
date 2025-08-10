import pymysql
from datetime import datetime

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
        print("=== CHECKING YOUR SPECIFIC EXAMPLE ===\n")
        
        # Check messages sent around 6:49 AM on Aug 6
        query = """
        SELECT 
            id,
            recipient_phone,
            recipient_name,
            LEFT(content, 50) as content_preview,
            status,
            created_at,
            scheduled_at,
            sent_at,
            device_id,
            campaign_id,
            sequence_id,
            sequence_stepid
        FROM broadcast_messages
        WHERE sent_at BETWEEN '2025-08-06 06:45:00' AND '2025-08-06 06:55:00'
        ORDER BY recipient_phone, sent_at
        """
        
        cursor.execute(query)
        messages = cursor.fetchall()
        
        print(f"Messages sent between 6:45-6:55 AM: {len(messages)}\n")
        
        # Group by recipient
        recipients = {}
        for msg in messages:
            phone = msg['recipient_phone']
            if phone not in recipients:
                recipients[phone] = []
            recipients[phone].append(msg)
        
        # Show recipients with multiple messages
        for phone, msgs in recipients.items():
            if len(msgs) > 1:
                print(f"\nPhone: {phone} ({msgs[0]['recipient_name']})")
                print(f"Received {len(msgs)} messages:")
                
                for i, msg in enumerate(msgs, 1):
                    print(f"\n  Message {i}:")
                    print(f"    ID: {msg['id']}")
                    print(f"    Sent at: {msg['sent_at']}")
                    print(f"    Device: {msg['device_id'][:8]}...")
                    print(f"    Content: {msg['content_preview']}...")
                    print(f"    Sequence: {msg['sequence_id']}")
                    print(f"    Step ID: {msg['sequence_stepid']}")

finally:
    connection.close()
