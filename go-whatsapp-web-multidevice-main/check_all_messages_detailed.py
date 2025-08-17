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
    
    # Get ALL messages for this phone in last 2 days
    print("=== ALL MESSAGES FOR PHONE NUMBER IN LAST 2 DAYS ===")
    query = """
    SELECT 
        id,
        recipient_phone,
        status,
        DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
        DATE_FORMAT(sent_at, '%Y-%m-%d %H:%i:%s') as sent_at,
        device_id,
        campaign_id,
        sequence_id,
        sequence_stepid,
        processing_worker_id,
        LENGTH(content) as content_length,
        LEFT(content, 50) as preview
    FROM broadcast_messages
    WHERE (recipient_phone LIKE '%128198574%')
        AND created_at >= '2025-08-09'
    ORDER BY created_at DESC
    """
    
    cursor.execute(query)
    results = cursor.fetchall()
    
    print(f"\nTotal messages found: {len(results)}")
    
    if results:
        # Group by date
        by_date = {}
        for msg in results:
            date = msg['created_at'].split(' ')[0] if msg['created_at'] else 'Unknown'
            if date not in by_date:
                by_date[date] = []
            by_date[date].append(msg)
        
        for date, msgs in by_date.items():
            print(f"\n--- Date: {date} ({len(msgs)} messages) ---")
            for msg in msgs:
                print(f"\nID: {msg['id']}")
                print(f"Phone: {msg['recipient_phone']}")
                print(f"Status: {msg['status']}")
                print(f"Created: {msg['created_at']}")
                print(f"Sent: {msg['sent_at']}")
                print(f"Content length: {msg['content_length']} chars")
                try:
                    print(f"Preview: {msg['preview']}...")
                except:
                    print("Preview: [Special characters]")
                if msg['sequence_stepid']:
                    print(f"Sequence Step: {msg['sequence_stepid']}")
    
    # Check whatsapp_messages table too (if it exists)
    print("\n\n=== CHECKING whatsapp_messages TABLE ===")
    try:
        cursor.execute("SHOW TABLES LIKE 'whatsapp_messages'")
        if cursor.fetchone():
            query2 = """
            SELECT 
                id,
                phone_number,
                message_text,
                DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at,
                status
            FROM whatsapp_messages
            WHERE phone_number LIKE '%128198574%'
                AND created_at >= '2025-08-09'
            ORDER BY created_at DESC
            LIMIT 20
            """
            cursor.execute(query2)
            whatsapp_msgs = cursor.fetchall()
            
            if whatsapp_msgs:
                print(f"\nFound {len(whatsapp_msgs)} messages in whatsapp_messages table:")
                for msg in whatsapp_msgs:
                    print(f"\nID: {msg['id']}")
                    print(f"Phone: {msg['phone_number']}")
                    print(f"Created: {msg['created_at']}")
                    print(f"Status: {msg['status']}")
                    try:
                        print(f"Message: {msg['message_text'][:100]}...")
                    except:
                        print("Message: [Special characters]")
        else:
            print("whatsapp_messages table does not exist")
    except Exception as e:
        print(f"Could not check whatsapp_messages: {e}")
        
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
