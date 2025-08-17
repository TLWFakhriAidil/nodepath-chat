import psycopg2

# Connect
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

# Check for messages with sequence_stepid and error_message
cursor.execute("""
SELECT 
    COUNT(*) as total_errors,
    bm.error_message,
    COUNT(DISTINCT bm.device_id) as devices_affected,
    COUNT(DISTINCT bm.sequence_id) as sequences_affected,
    MIN(bm.created_at) as first_occurrence,
    MAX(bm.created_at) as last_occurrence
FROM broadcast_messages bm
WHERE bm.sequence_stepid IS NOT NULL 
AND bm.error_message IS NOT NULL
GROUP BY bm.error_message
ORDER BY total_errors DESC;
""")

results = cursor.fetchall()

print("\n=== SEQUENCE ERRORS (where sequence_stepid IS NOT NULL) ===\n")
print(f"Found {len(results)} different error types\n")

total_errors_all = 0

for i, row in enumerate(results, 1):
    total_errors, error_msg, devices, sequences, first, last = row
    print(f"{i}. Error: {error_msg}")
    print(f"   Count: {total_errors}")
    print(f"   Devices Affected: {devices}")
    print(f"   Sequences Affected: {sequences}")
    print(f"   First Seen: {first}")
    print(f"   Last Seen: {last}")
    print("-" * 80)
    total_errors_all += total_errors

# Get device breakdown
cursor.execute("""
SELECT 
    ud.device_name,
    ud.platform,
    COUNT(*) as error_count,
    COUNT(DISTINCT bm.error_message) as unique_errors
FROM broadcast_messages bm
JOIN user_devices ud ON bm.device_id = ud.id
WHERE bm.sequence_stepid IS NOT NULL 
AND bm.error_message IS NOT NULL
GROUP BY ud.device_name, ud.platform
ORDER BY error_count DESC
LIMIT 20;
""")

device_results = cursor.fetchall()

print(f"\n=== TOP DEVICES WITH SEQUENCE ERRORS ===\n")
for device, platform, count, unique_errors in device_results:
    platform_info = f" [{platform}]" if platform else " [WhatsApp Web]"
    print(f"{device}{platform_info}: {count} errors ({unique_errors} unique error types)")

# Get status breakdown for messages with sequence_stepid
cursor.execute("""
SELECT 
    status,
    COUNT(*) as count,
    COUNT(CASE WHEN error_message IS NOT NULL THEN 1 END) as with_errors
FROM broadcast_messages
WHERE sequence_stepid IS NOT NULL
GROUP BY status
ORDER BY count DESC;
""")

status_results = cursor.fetchall()

print(f"\n=== STATUS BREAKDOWN FOR SEQUENCE MESSAGES (sequence_stepid NOT NULL) ===\n")
for status, count, with_errors in status_results:
    print(f"{status}: {count} messages ({with_errors} with errors)")

# Get total summary
cursor.execute("""
SELECT 
    COUNT(*) as total_with_stepid,
    COUNT(CASE WHEN error_message IS NOT NULL THEN 1 END) as total_with_errors,
    COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent,
    COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed,
    COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending
FROM broadcast_messages
WHERE sequence_stepid IS NOT NULL;
""")

summary = cursor.fetchone()
total, with_errors, sent, failed, pending = summary

print(f"\n=== SUMMARY ===")
print(f"Total messages with sequence_stepid: {total}")
print(f"Messages with errors: {with_errors}")
print(f"Error rate: {(with_errors/total*100):.1f}%" if total > 0 else "0%")
print(f"\nStatus breakdown:")
print(f"- Sent: {sent}")
print(f"- Failed: {failed}")
print(f"- Pending: {pending}")

cursor.close()
conn.close()
