import pymysql
from datetime import datetime, timedelta

print("COMPREHENSIVE A-Z CHECK: Sequences and Campaigns")
print("="*100)

conn = pymysql.connect(
    host='159.89.198.71',
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    port=3306
)
cursor = conn.cursor()

# 1. CHECK DATABASE SCHEMA
print("\n1. DATABASE SCHEMA CHECK")
print("-"*50)

# Check if critical columns exist
columns_to_check = [
    ('broadcast_messages', 'processing_worker_id'),
    ('broadcast_messages', 'processing_started_at'),
    ('broadcast_messages', 'sequence_stepid'),
    ('broadcast_messages', 'campaign_id'),
    ('broadcast_messages', 'device_id'),
    ('broadcast_messages', 'recipient_phone')
]

for table, column in columns_to_check:
    cursor.execute(f"SHOW COLUMNS FROM {table} LIKE '{column}'")
    result = cursor.fetchone()
    print(f"  {table}.{column}: {'✓ EXISTS' if result else '✗ MISSING'}")

# Check indexes
print("\n  Indexes:")
cursor.execute("SHOW INDEX FROM broadcast_messages WHERE Key_name IN ('idx_processing_worker', 'unique_sequence_message', 'unique_campaign_message')")
indexes = cursor.fetchall()
for idx in indexes:
    print(f"    {idx[2]}: {idx[4]} (Unique: {idx[1] == 0})")

# 2. CHECK DUPLICATE PREVENTION LOGIC
print("\n\n2. DUPLICATE PREVENTION CHECK")
print("-"*50)

# Check for sequence duplicates
cursor.execute("""
SELECT 
    sequence_stepid,
    recipient_phone,
    device_id,
    COUNT(*) as count
FROM broadcast_messages 
WHERE sequence_stepid IS NOT NULL
AND created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY sequence_stepid, recipient_phone, device_id
HAVING COUNT(*) > 1
LIMIT 5
""")

seq_duplicates = cursor.fetchall()
print(f"\n  Sequence Duplicates (last 7 days): {len(seq_duplicates)}")
for dup in seq_duplicates:
    print(f"    Step: {dup[0][:20]}..., Phone: {dup[1]}, Device: {dup[2][:8]}..., Count: {dup[3]}")

# Check for campaign duplicates
cursor.execute("""
SELECT 
    campaign_id,
    recipient_phone,
    device_id,
    COUNT(*) as count
FROM broadcast_messages 
WHERE campaign_id IS NOT NULL
AND created_at >= DATE_SUB(NOW(), INTERVAL 7 DAY)
GROUP BY campaign_id, recipient_phone, device_id
HAVING COUNT(*) > 1
LIMIT 5
""")

camp_duplicates = cursor.fetchall()
print(f"\n  Campaign Duplicates (last 7 days): {len(camp_duplicates)}")
for dup in camp_duplicates:
    print(f"    Campaign: {dup[0]}, Phone: {dup[1]}, Device: {dup[2][:8]}..., Count: {dup[3]}")

# 3. CHECK WORKER ID USAGE
print("\n\n3. WORKER ID USAGE CHECK")
print("-"*50)

cursor.execute("""
SELECT 
    DATE(created_at) as date,
    COUNT(*) as total,
    COUNT(processing_worker_id) as with_worker_id,
    COUNT(*) - COUNT(processing_worker_id) as without_worker_id
FROM broadcast_messages 
WHERE created_at >= DATE_SUB(NOW(), INTERVAL 3 DAY)
GROUP BY DATE(created_at)
ORDER BY date DESC
""")

worker_stats = cursor.fetchall()
print("\n  Worker ID Statistics:")
for stat in worker_stats:
    percentage = (stat[2] / stat[1] * 100) if stat[1] > 0 else 0
    print(f"    {stat[0]}: Total: {stat[1]}, With Worker ID: {stat[2]} ({percentage:.1f}%), Without: {stat[3]}")

# 4. CHECK MESSAGE FLOW
print("\n\n4. MESSAGE FLOW CHECK")
print("-"*50)

# Check recent sequence message creation
cursor.execute("""
SELECT 
    id,
    sequence_stepid,
    recipient_phone,
    device_id,
    status,
    processing_worker_id,
    created_at,
    sent_at
FROM broadcast_messages 
WHERE sequence_stepid IS NOT NULL
AND created_at >= DATE_SUB(NOW(), INTERVAL 1 HOUR)
ORDER BY created_at DESC
LIMIT 5
""")

recent_seq = cursor.fetchall()
print("\n  Recent Sequence Messages:")
for msg in recent_seq:
    print(f"    ID: {msg[0][:8]}...")
    print(f"      Step: {msg[1][:20]}..., Phone: {msg[2]}, Device: {msg[3][:8]}...")
    print(f"      Status: {msg[4]}, Worker: {msg[5][:20] if msg[5] else 'NULL'}...")
    print(f"      Created: {msg[6]}, Sent: {msg[7]}")

# 5. CHECK STUCK MESSAGES
print("\n\n5. STUCK MESSAGES CHECK")
print("-"*50)

cursor.execute("""
SELECT 
    COUNT(*) as stuck_count,
    MIN(created_at) as oldest,
    MAX(created_at) as newest
FROM broadcast_messages 
WHERE status = 'processing'
AND processing_started_at < DATE_SUB(NOW(), INTERVAL 10 MINUTE)
""")

stuck = cursor.fetchone()
print(f"\n  Messages stuck in 'processing': {stuck[0]}")
if stuck[0] > 0:
    print(f"    Oldest: {stuck[1]}, Newest: {stuck[2]}")

# 6. VERIFY ATOMIC OPERATIONS
print("\n\n6. ATOMIC OPERATION VERIFICATION")
print("-"*50)

# Check if messages are being locked properly
cursor.execute("""
SELECT 
    processing_worker_id,
    COUNT(*) as message_count
FROM broadcast_messages 
WHERE status = 'processing'
AND processing_worker_id IS NOT NULL
GROUP BY processing_worker_id
HAVING COUNT(*) > 50
LIMIT 5
""")

worker_loads = cursor.fetchall()
print("\n  Workers with high message counts:")
for load in worker_loads:
    print(f"    Worker {load[0][:30]}...: {load[1]} messages")

conn.close()

print("\n\n" + "="*100)
print("SUMMARY OF ISSUES TO FIX:")
print("="*100)
