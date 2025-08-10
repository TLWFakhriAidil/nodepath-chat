import psycopg2

# Connect to PostgreSQL
print("Connecting to PostgreSQL...")
conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()
print("Connected successfully!")

# First, let's check the actual columns in broadcast_messages
cursor.execute("""
    SELECT column_name 
    FROM information_schema.columns 
    WHERE table_name = 'broadcast_messages'
    ORDER BY ordinal_position
""")
columns = cursor.fetchall()
print("\nColumns in broadcast_messages table:")
for col in columns:
    print(f"  - {col[0]}")

# Now let's check the platform timeout issue with correct columns
print("\n=== Platform Device Timeout Analysis ===")
cursor.execute("""
    SELECT ud.id, ud.phone, ud.platform, ud.status, 
           COUNT(bm.id) as total_messages,
           MIN(bm.created_at) as first_message,
           MAX(bm.created_at) as last_message
    FROM user_devices ud
    JOIN broadcast_messages bm ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform IS NOT NULL AND ud.platform != ''
    GROUP BY ud.id, ud.phone, ud.platform, ud.status
""")
platform_devices = cursor.fetchall()

for device in platform_devices:
    print(f"\nDevice ID: {device[0]}")
    print(f"  Phone: {device[1] if device[1] else 'NO PHONE'}")
    print(f"  Platform: {device[2]}")
    print(f"  Device Status: {device[3]}")
    print(f"  Total timeout messages: {device[4]}")
    print(f"  Time range: {device[5]} to {device[6]}")

# Get sample messages with correct column names
cursor.execute("""
    SELECT bm.id, bm.recipient_phone, bm.type, bm.created_at, 
           ud.platform, bm.campaign_id, bm.sequence_id, bm.sequence_stepid
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform = 'Wablas'
    LIMIT 5
""")
sample_messages = cursor.fetchall()

print("\n\nSample Wablas messages with timeout:")
for msg in sample_messages:
    print(f"\nMessage ID: {msg[0]}")
    print(f"  To: {msg[1]}")
    print(f"  Type: {msg[2]}")
    print(f"  Created: {msg[3]}")
    print(f"  Campaign: {msg[5]}, Sequence: {msg[6]}, Step: {msg[7]}")

# Summary and recommendation
print("\n=== ISSUE FOUND ===")
print("Platform devices (Wablas/Whacenter) are being checked for WhatsApp Web availability")
print("This is incorrect because:")
print("1. Platform devices use API, not WhatsApp Web")
print("2. They should always be considered 'available'")
print("3. The broadcast worker needs to skip availability check for platform devices")

# Close connection
cursor.close()
conn.close()
print("\nAnalysis completed!")
