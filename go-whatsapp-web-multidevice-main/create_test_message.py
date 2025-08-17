import mysql.connector
from datetime import datetime
import uuid
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
    
    # Find a message sent today
    print("=== FINDING A MESSAGE SENT TODAY TO DUPLICATE ===\n")
    
    query = """
    SELECT 
        id, user_id, device_id, campaign_id, sequence_id, sequence_stepid,
        content, media_url, message_type, group_id, group_order,
        DATE_FORMAT(scheduled_at, '%Y-%m-%d %H:%i:%s') as scheduled_at
    FROM broadcast_messages
    WHERE status = 'sent'
        AND DATE(sent_at) = CURDATE()
        AND content IS NOT NULL
        AND device_id IS NOT NULL
    ORDER BY sent_at DESC
    LIMIT 1
    """
    
    cursor.execute(query)
    original = cursor.fetchone()
    
    if not original:
        print("No sent messages found today!")
    else:
        print(f"Found message to duplicate:")
        print(f"Original ID: {original['id']}")
        print(f"Content: {original['content'][:100]}...")
        print(f"Device: {original['device_id']}")
        print(f"Scheduled at: {original['scheduled_at']}")
        
        # Create duplicate with your phone number
        new_id = str(uuid.uuid4())
        
        insert_query = """
        INSERT INTO broadcast_messages (
            id, user_id, device_id, campaign_id, sequence_id, sequence_stepid,
            recipient_phone, recipient_name, message_type, content, media_url,
            status, scheduled_at, created_at, updated_at, group_id, group_order
        ) VALUES (
            %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, NOW(), NOW(), %s, %s
        )
        """
        
        values = (
            new_id,
            original['user_id'],
            original['device_id'],
            original['campaign_id'],
            original['sequence_id'],
            original['sequence_stepid'],
            '60108924904',  # Your phone number
            'Test User',    # Name
            original['message_type'],
            original['content'],
            original['media_url'],
            'pending',      # Status
            original['scheduled_at'],
            original['group_id'],
            original['group_order']
        )
        
        cursor.execute(insert_query, values)
        conn.commit()
        
        print(f"\n✅ SUCCESSFULLY CREATED TEST MESSAGE!")
        print(f"New ID: {new_id}")
        print(f"Phone: 60108924904")
        print(f"Status: pending")
        print(f"Device: {original['device_id']}")
        print(f"\nThis message should be picked up by the broadcast processor within 5 seconds.")
        
        # Verify it was created
        cursor.execute("SELECT id, recipient_phone, status FROM broadcast_messages WHERE id = %s", (new_id,))
        verify = cursor.fetchone()
        if verify:
            print(f"\nVerified: Message exists in database")
            print(f"Status: {verify['status']}")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
