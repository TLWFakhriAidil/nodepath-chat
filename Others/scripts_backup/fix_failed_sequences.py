import psycopg2
import sys
from datetime import datetime

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

print("=== FIXING FAILED SEQUENCE MESSAGES ===\n")

# First, let's check the current status
cur.execute("""
    SELECT COUNT(*) 
    FROM broadcast_messages 
    WHERE sequence_stepid IS NOT NULL
    AND status = 'failed'
    AND error_message = 'Message timeout - could not be delivered within 12 hours'
""")

failed_count = cur.fetchone()[0]
print(f"Found {failed_count} messages with timeout error")

# These messages should not have failed - they were scheduled for the future
# Let's reset them to pending
cur.execute("""
    UPDATE broadcast_messages
    SET status = 'pending',
        error_message = NULL,
        updated_at = NOW()
    WHERE sequence_stepid IS NOT NULL
    AND status = 'failed'
    AND error_message = 'Message timeout - could not be delivered within 12 hours'
    RETURNING id, recipient_phone
""")

updated = cur.fetchall()
print(f"\nReset {len(updated)} messages to pending status:")
for msg in updated:
    print(f"  - Message {msg[0]} for phone {msg[1]}")

conn.commit()

# Check if there are any other types of failures
print("\n=== CHECKING OTHER SEQUENCE FAILURES ===")
cur.execute("""
    SELECT error_message, COUNT(*)
    FROM broadcast_messages
    WHERE sequence_stepid IS NOT NULL
    AND status = 'failed'
    AND error_message IS NOT NULL
    GROUP BY error_message
""")

other_errors = cur.fetchall()
if other_errors:
    print("Other error types found:")
    for error in other_errors:
        print(f"  - {error[0]}: {error[1]} messages")
else:
    print("No other sequence failures found!")

# Summary of sequence messages
print("\n=== SEQUENCE MESSAGE SUMMARY ===")
cur.execute("""
    SELECT 
        status,
        COUNT(*) as count
    FROM broadcast_messages
    WHERE sequence_stepid IS NOT NULL
    GROUP BY status
    ORDER BY count DESC
""")

summary = cur.fetchall()
for status in summary:
    print(f"{status[0]}: {status[1]} messages")

cur.close()
conn.close()

print("\n✅ Failed sequence messages have been reset to pending!")
print("They should now be processed correctly by the broadcast workers.")
