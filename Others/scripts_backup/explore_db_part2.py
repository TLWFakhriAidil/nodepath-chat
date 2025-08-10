# Check sequence entry points and triggers
cursor.execute("""
    SELECT 
        s.name as sequence_name,
        ss.day_number,
        ss.trigger,
        ss.is_entry_point,
        ss.next_trigger,
        ss.trigger_delay_hours
    FROM sequence_steps ss
    JOIN sequences s ON s.id = ss.sequence_id
    WHERE ss.is_entry_point = true OR ss.next_trigger IS NOT NULL
    ORDER BY s.name, ss.day_number
""")
triggers = cursor.fetchall()

print("\nSequence Triggers and Links:")
for trigger in triggers:
    seq_name, step, trig, is_entry, next_trig, delay = trigger
    if is_entry:
        print(f"\n{seq_name} - Entry Point:")
        print(f"  - Trigger: {trig}")
    if next_trig:
        print(f"  - Day {step} links to: {next_trig} (after {delay}h)")

# Check broadcast messages queue
print("\n" + "="*60)
print("\n=== BROADCAST MESSAGES QUEUE STATUS ===")

cursor.execute("""
    SELECT 
        status,
        COUNT(*) as count,
        MIN(scheduled_at) as earliest,
        MAX(scheduled_at) as latest
    FROM broadcast_messages
    GROUP BY status
    ORDER BY status
""")
queue_status = cursor.fetchall()

for status in queue_status:
    print(f"\n{status[0]}:")
    print(f"  - Count: {status[1]}")
    print(f"  - Earliest: {status[2]}")
    print(f"  - Latest: {status[3]}")

# Check devices
cursor.execute("""
    SELECT 
        COUNT(*) as total_devices,
        COUNT(CASE WHEN jid IS NOT NULL AND jid != '' THEN 1 END) as connected_devices
    FROM user_devices
""")
device_stats = cursor.fetchone()
print(f"\nDevice Status:")
print(f"  - Total devices: {device_stats[0]}")
print(f"  - Connected devices: {device_stats[1]}")

# Check recent leads with triggers
print("\n" + "="*60)
print("\n=== RECENT LEADS WITH TRIGGERS ===")

cursor.execute("""
    SELECT 
        trigger,
        COUNT(*) as count
    FROM leads
    WHERE trigger IS NOT NULL AND trigger != ''
    GROUP BY trigger
    ORDER BY count DESC
    LIMIT 10
""")
lead_triggers = cursor.fetchall()

print("\nTop Lead Triggers:")
for trigger in lead_triggers:
    print(f"  - {trigger[0]}: {trigger[1]} leads")

conn.close()
print("\n" + "="*60)
print("Database exploration complete!")