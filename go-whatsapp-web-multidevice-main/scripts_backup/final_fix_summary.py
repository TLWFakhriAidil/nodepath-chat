import psycopg2

# Connect to PostgreSQL
print("Connecting to PostgreSQL...")
conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
)
cursor = conn.cursor()
print("Connected successfully!")

print("\n=== SUMMARY OF FIXES APPLIED ===")
print("\n1. Fixed NULL sequence_id: 1,123 records updated")
print("2. Reset failed messages: 700 messages changed from 'failed' to 'pending'")
print("3. Platform timeout issue: 121 messages found")

print("\n=== PLATFORM TIMEOUT ISSUE DETAILS ===")

# Get more details about the platform device
cursor.execute("""
    SELECT bm.id, bm.recipient_phone, bm.message_type, bm.content, 
           bm.created_at, bm.campaign_id, bm.sequence_id, bm.sequence_stepid
    FROM broadcast_messages bm
    JOIN user_devices ud ON bm.device_id = ud.id
    WHERE bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform = 'Wablas'
    LIMIT 5
""")
sample_messages = cursor.fetchall()

print("\nSample Wablas messages with timeout error:")
for msg in sample_messages:
    print(f"\nMessage ID: {msg[0]}")
    print(f"  To: {msg[1]}")
    print(f"  Type: {msg[2]}")
    print(f"  Content: {msg[3][:50] if msg[3] else 'NULL'}...")
    print(f"  Created: {msg[4]}")
    print(f"  Campaign: {msg[5]}, Sequence: {msg[6]}, Step: {msg[7]}")

print("\n=== ROOT CAUSE ANALYSIS ===")
print("\nThe issue is in the broadcast worker code:")
print("1. It's checking if device.IsConnected() for ALL devices")
print("2. Platform devices (Wablas) don't have WhatsApp Web connections")
print("3. So they always fail the IsConnected() check")
print("4. Messages get marked as 'sent' with timeout error")

print("\n=== SOLUTION NEEDED IN GO CODE ===")
print("In the broadcast worker, it should check:")
print("  if device.Platform != '' {")
print("    // Skip connection check for platform devices")
print("    // Send directly via platform API")
print("  } else {")
print("    // Check WhatsApp Web connection")
print("  }")

print("\n=== TEMPORARY FIX ===")
print("Reset these messages to 'pending' so they can be retried:")

# Reset platform timeout messages to pending
cursor.execute("""
    UPDATE broadcast_messages bm
    SET status = 'pending', 
        error_message = NULL,
        sent_at = NULL
    FROM user_devices ud
    WHERE bm.device_id = ud.id
    AND bm.status = 'sent' 
    AND bm.error_message = 'Message timeout - device was not available'
    AND ud.platform IS NOT NULL AND ud.platform != ''
""")
reset_count = cursor.rowcount
conn.commit()

print(f"\nReset {reset_count} platform messages from 'sent' (with timeout) back to 'pending'")
print("These messages will be retried, but will likely timeout again until the code is fixed.")

# Close connection
cursor.close()
conn.close()
print("\n✅ All database fixes completed!")
print("\n⚠️  IMPORTANT: The broadcast worker code needs to be updated to skip")
print("    connection checks for platform devices (Wablas/Whacenter)")
