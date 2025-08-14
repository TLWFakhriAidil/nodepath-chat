import psycopg2
import pandas as pd

# Connect to the database
conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"
conn = psycopg2.connect(conn_string)
cursor = conn.cursor()

print("Connected to PostgreSQL database successfully!")
print("\n" + "="*60)

# Get all tables
cursor.execute("""
    SELECT table_name 
    FROM information_schema.tables 
    WHERE table_schema = 'public'
    ORDER BY table_name
""")
tables = cursor.fetchall()

print("\n=== ALL TABLES IN DATABASE ===")
for table in tables:
    print(f"- {table[0]}")

# Key tables to examine
key_tables = [
    'leads', 
    'sequences', 
    'sequence_steps', 
    'broadcast_messages', 
    'user_devices',
    'campaigns',
    'campaign_leads'
]

print("\n" + "="*60)
print("\n=== DETAILED SCHEMA FOR KEY TABLES ===")

for table_name in key_tables:
    if table_name in [t[0] for t in tables]:
        print(f"\n\n### TABLE: {table_name} ###")
        
        # Get column information
        cursor.execute("""
            SELECT 
                column_name,
                data_type,
                is_nullable,
                column_default
            FROM information_schema.columns
            WHERE table_name = %s
            ORDER BY ordinal_position
        """, (table_name,))
        
        columns = cursor.fetchall()
        
        # Print column details
        print(f"{'Column':<30} {'Type':<20} {'Nullable':<10} {'Default':<30}")
        print("-" * 90)
        for col in columns:
            col_name, data_type, is_nullable, default = col
            default_str = str(default)[:28] if default else "NULL"
            print(f"{col_name:<30} {data_type:<20} {is_nullable:<10} {default_str:<30}")
        
        # Get row count
        cursor.execute(f"SELECT COUNT(*) FROM {table_name}")
        count = cursor.fetchone()[0]
        print(f"\nRow count: {count}")

# Examine sequence flow
print("\n" + "="*60)
print("\n=== SEQUENCE FLOW ANALYSIS ===")

# Check active sequences
cursor.execute("""
    SELECT 
        s.id,
        s.name,
        s.is_active,
        COUNT(ss.id) as step_count
    FROM sequences s
    LEFT JOIN sequence_steps ss ON s.id = ss.sequence_id
    GROUP BY s.id, s.name, s.is_active
    ORDER BY s.name
""")
sequences = cursor.fetchall()

print("\nSequences in system:")
for seq in sequences:
    status = "ACTIVE" if seq[2] else "INACTIVE"
    print(f"- {seq[1]} ({status}) - {seq[3]} steps")

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
    seq_name, day, trig, is_entry, next_trig, delay = trigger
    if is_entry:
        print(f"\n{seq_name} - Entry Point:")
        print(f"  - Trigger: {trig}")
    if next_trig:
        print(f"  - Day {day} links to: {next_trig} (after {delay}h)")

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

# Sample sequence step details
print("\n" + "="*60)
print("\n=== SAMPLE SEQUENCE STEPS ===")

cursor.execute("""
    SELECT 
        s.name as sequence_name,
        ss.day_number,
        ss.message_type,
        SUBSTRING(ss.content, 1, 50) as content_preview,
        ss.trigger,
        ss.is_entry_point,
        ss.next_trigger
    FROM sequence_steps ss
    JOIN sequences s ON s.id = ss.sequence_id
    ORDER BY s.name, ss.day_number
    LIMIT 15
""")
steps = cursor.fetchall()

print("\nSequence Message Flow:")
current_seq = None
for step in steps:
    seq_name, day, msg_type, content, trigger, is_entry, next_trig = step
    if seq_name != current_seq:
        print(f"\n{seq_name}:")
        current_seq = seq_name
    entry_flag = " [ENTRY]" if is_entry else ""
    next_flag = f" → {next_trig}" if next_trig else ""
    print(f"  Day {day}: {msg_type}{entry_flag}{next_flag}")
    if content:
        print(f"    Message: {content}...")

conn.close()
print("\n" + "="*60)
print("Database exploration complete!")
