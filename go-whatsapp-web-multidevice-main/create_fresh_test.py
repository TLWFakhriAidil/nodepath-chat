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
    cursor = conn.cursor()
    
    print("=== CREATING FRESH TEST MESSAGE ===\n")
    
    # Delete old test message
    cursor.execute("DELETE FROM broadcast_messages WHERE id = '1ef22b72-df41-45c1-9c69-be5925a38080'")
    
    # Create new message with current time
    new_id = str(uuid.uuid4())
    
    insert_query = """
    INSERT INTO broadcast_messages (
        id, user_id, device_id, recipient_phone, recipient_name,
        message_type, content, status, scheduled_at, created_at, updated_at
    ) VALUES (
        %s, %s, %s, %s, %s, %s, %s, %s, NOW(), NOW(), NOW()
    )
    """
    
    values = (
        new_id,
        '8badb299-f1d1-493a-bddf-84cbaba1273b',  # Using device_id as user_id since we don't have it
        '8badb299-f1d1-493a-bddf-84cbaba1273b',  # Your device
        '60108924904',  # Your phone
        'Test User',
        'text',
        'Test message to verify worker ID fix is deployed. Time: ' + datetime.now().strftime('%H:%M:%S'),
        'pending'
    )
    
    cursor.execute(insert_query, values)
    conn.commit()
    
    print(f"✅ Created new test message")
    print(f"ID: {new_id}")
    print(f"Phone: 60108924904")
    print(f"Device: SCHQ-S54 (Wablas)")
    print(f"Scheduled: NOW")
    print(f"\n🔄 This message is DEFINITELY in the time window!")
    print("Watch the logs - it should be picked up within 5 seconds.")
    
    # Store the ID for checking
    with open('test_message_id.txt', 'w') as f:
        f.write(new_id)
    
except Exception as e:
    print(f"Error: {e}")
    import traceback
    traceback.print_exc()
finally:
    if 'cursor' in locals():
        cursor.close()
    if 'conn' in locals():
        conn.close()
