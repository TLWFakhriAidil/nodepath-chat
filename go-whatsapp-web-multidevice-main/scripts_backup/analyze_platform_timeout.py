import psycopg2
from datetime import datetime

# Connect to PostgreSQL
print("Connecting to PostgreSQL...")
conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()
print("Connected successfully!")

# Task 4: Deep dive into platform timeout issue
print("\n=== Task 4: Deep Analysis of Platform Timeout Issue ===")

# Get detailed info about the platform device with timeouts
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

print("\nPlatform devices with timeout errors:")
for device in platform_devices:
    print(f"\nDevice ID: {device[0]}")
    print(f"  Phone: {device[1] if device[1] else 'NO PHONE'}")
    print(f"  Platform: {device[2]}")
    print(f"  Device Status: {device[3]}")
    print(f"  Total timeout messages: {device[4]}")
    print(f"  First message: {device[5]}")
    print(f"  Last message: {device[6]}")

# Check the actual broadcast worker code logic
print("\n=== Understanding the Timeout Issue ===")
print("The timeout error for platform devices suggests:")
print("1. The system is checking device availability even for platform devices")
print("2. Platform devices (Wablas/Whacenter) should bypass device availability checks")
print("3. These are API-based services that don't need WhatsApp Web connection")

# Get some sample messages to understand the pattern
cursor.execute("""
    SELECT bm.id, bm.recipient_phone, bm.message, bm.created_at, 
           ud.platform, bm.campaign_id, bm.sequence_id
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform = 'Wablas'
    LIMIT 5
""")
sample_messages = cursor.fetchall()

print("\nSample timeout messages for Wablas platform:")
for msg in sample_messages:
    print(f"\nMessage ID: {msg[0]}")
    print(f"  To: {msg[1]}")
    print(f"  Message: {msg[2][:50]}...")
    print(f"  Created: {msg[3]}")
    print(f"  Campaign ID: {msg[5]}, Sequence ID: {msg[6]}")

# Check if we should reset these to pending
cursor.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform IS NOT NULL AND ud.platform != ''
""")
platform_timeout_count = cursor.fetchone()[0]

print(f"\n\nTotal platform messages with timeout error: {platform_timeout_count}")
print("\nRecommendation:")
print("1. These messages should be reset to 'pending' status")
print("2. The broadcast worker should be fixed to skip device availability check for platform devices")
print("3. Platform devices should send messages directly via API without checking WhatsApp Web status")

# Ask for confirmation before resetting
print("\nWould you like to reset these platform timeout messages to pending? (Run separately if yes)")

# Close connection
cursor.close()
conn.close()
print("\nAnalysis completed!")
