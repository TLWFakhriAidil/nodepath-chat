import mysql.connector
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
    
    print("=== CHECKING SEQUENCE MESSAGE STRUCTURE ===\n")
    
    # Check the full message structure
    query = """
    SELECT 
        bm.*,
        ss.content as step_content,
        ss.message_text as step_message
    FROM broadcast_messages bm
    LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
    WHERE bm.id = 'fgeef10d-4ae4-4e8a-8bbd-73d2fdb6094a'
    """
    
    cursor.execute(query)
    msg = cursor.fetchone()
    
    if msg:
        print("BROADCAST MESSAGE:")
        print(f"ID: {msg['id']}")
        print(f"Content: {msg['content']}")
        print(f"Message Type: {msg['message_type']}")
        print(f"Media URL: {msg['media_url']}")
        print(f"Sequence Step ID: {msg['sequence_stepid']}")
        
        print("\nSEQUENCE STEP:")
        print(f"Step Content: {msg['step_content']}")
        print(f"Step Message: {msg['step_message']}")
        
        # Check if content is NULL or empty
        if not msg['content'] or msg['content'] == '':
            print("\n❌ CONTENT IS EMPTY!")
        elif msg['content'] == msg['sequence_stepid']:
            print("\n❌ CONTENT IS SEQUENCE_STEPID!")
            print("This might be causing the issue!")
    
    # Check if there are other messages with proper content
    print("\n\n=== CHECKING OTHER PENDING MESSAGES ===")
    
    query2 = """
    SELECT 
        id,
        recipient_phone,
        LEFT(content, 50) as content_preview,
        message_type,
        LENGTH(content) as content_length,
        campaign_id,
        sequence_id
    FROM broadcast_messages
    WHERE status = 'pending'
        AND device_id = '8badb299-f1d1-493a-bddf-84cbaba1273b'
        AND scheduled_at >= DATE_ADD(DATE_SUB(NOW(), INTERVAL 10 MINUTE), INTERVAL 8 HOUR)
    LIMIT 5
    """
    
    cursor.execute(query2)
    others = cursor.fetchall()
    
    if others:
        print(f"\nFound {len(others)} pending messages for this device:")
        for msg in others:
            print(f"\nID: {msg['id']}")
            print(f"Phone: {msg['recipient_phone']}")
            print(f"Content Length: {msg['content_length']} chars")
            print(f"Content: {msg['content_preview']}...")
            print(f"Type: {msg['message_type']}")
            
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
