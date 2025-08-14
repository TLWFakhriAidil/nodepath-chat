import psycopg2
import pandas as pd
from datetime import datetime

# Connect with SSL
conn_str = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
conn = psycopg2.connect(conn_str, connect_timeout=20)
cursor = conn.cursor()
print("[OK] Connected to Railway PostgreSQL with SSL!")

# Get all tables
cursor.execute("""
    SELECT table_name 
    FROM information_schema.tables 
    WHERE table_schema = 'public' 
    ORDER BY table_name;
""")
tables = [t[0] for t in cursor.fetchall()]

print(f"\n{'='*60}")
print(f"DATABASE OVERVIEW - {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
print(f"{'='*60}")
print(f"\nTotal tables: {len(tables)}")
print("\nAll tables:")
for i, table in enumerate(tables, 1):
    print(f"{i:2d}. {table}")

# Analyze key tables
print(f"\n{'='*60}")
print("KEY TABLES ANALYSIS")
print(f"{'='*60}")

# 1. Users
try:
    cursor.execute("SELECT COUNT(*) FROM users")
    user_count = cursor.fetchone()[0]
    cursor.execute("SELECT email, full_name FROM users LIMIT 5")
    sample_users = cursor.fetchall()
    print(f"\n1. USERS: {user_count} total")
    if sample_users:
        print("   Sample users:")
        for email, name in sample_users:
            print(f"   - {name} ({email})")
except:
    print("\n1. USERS: Table exists but empty or error")

# 2. Devices
try:
    cursor.execute("""
        SELECT COUNT(*), 
               SUM(CASE WHEN status = 'online' THEN 1 ELSE 0 END) as online
        FROM user_devices
    """)
    device_count, online_count = cursor.fetchone()
    print(f"\n2. DEVICES: {device_count or 0} total, {online_count or 0} online")
    
    cursor.execute("""
        SELECT device_name, phone, status, min_delay_seconds, max_delay_seconds
        FROM user_devices 
        ORDER BY created_at DESC
        LIMIT 5
    """)
    devices = cursor.fetchall()
    if devices:
        print("   Recent devices:")
        for dev in devices:
            print(f"   - {dev[0]}: {dev[1] or 'No phone'} [{dev[2]}] (delay: {dev[3]}-{dev[4]}s)")
except Exception as e:
    print(f"\n2. DEVICES: Error - {str(e)[:50]}")

# 3. Leads
try:
    cursor.execute("SELECT COUNT(*) FROM leads")
    lead_count = cursor.fetchone()[0]
    
    cursor.execute("""
        SELECT trigger, COUNT(*) as count 
        FROM leads 
        WHERE trigger IS NOT NULL AND trigger != ''
        GROUP BY trigger 
        ORDER BY count DESC 
        LIMIT 10
    """)
    triggers = cursor.fetchall()
    
    print(f"\n3. LEADS: {lead_count} total")
    if triggers:
        print("   Top triggers:")
        for trigger, count in triggers:
            print(f"   - {trigger}: {count} leads")
except Exception as e:
    print(f"\n3. LEADS: Error - {str(e)[:50]}")

# 4. Campaigns
try:
    cursor.execute("""
        SELECT COUNT(*), 
               SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END)
        FROM campaigns
    """)
    campaign_count, active_campaigns = cursor.fetchone()
    
    cursor.execute("""
        SELECT title, status, campaign_date, message
        FROM campaigns 
        ORDER BY created_at DESC 
        LIMIT 5
    """)
    campaigns = cursor.fetchall()
    
    print(f"\n4. CAMPAIGNS: {campaign_count or 0} total, {active_campaigns or 0} active")
    if campaigns:
        print("   Recent campaigns:")
        for camp in campaigns:
            msg_preview = camp[3][:50] + "..." if camp[3] and len(camp[3]) > 50 else camp[3]
            print(f"   - {camp[0]} [{camp[1]}] on {camp[2]}: {msg_preview}")
except Exception as e:
    print(f"\n4. CAMPAIGNS: Error - {str(e)[:50]}")

# 5. Sequences
try:
    cursor.execute("""
        SELECT COUNT(*), 
               SUM(CASE WHEN is_active = true THEN 1 ELSE 0 END)
        FROM sequences
    """)
    sequence_count, active_sequences = cursor.fetchone()
    
    cursor.execute("""
        SELECT s.name, s.is_active, 
               (SELECT COUNT(*) FROM sequence_steps ss WHERE ss.sequence_id = s.id) as steps
        FROM sequences s
        ORDER BY s.created_at DESC
        LIMIT 5
    """)
    sequences = cursor.fetchall()
    
    print(f"\n5. SEQUENCES: {sequence_count or 0} total, {active_sequences or 0} active")
    if sequences:
        print("   Recent sequences:")
        for seq in sequences:
            status = "Active" if seq[1] else "Inactive"
            print(f"   - {seq[0]} [{status}] with {seq[2]} steps")
except Exception as e:
    print(f"\n5. SEQUENCES: Error - {str(e)[:50]}")

# 6. Broadcast Messages
try:
    cursor.execute("""
        SELECT status, COUNT(*) 
        FROM broadcast_messages 
        GROUP BY status
    """)
    message_stats = dict(cursor.fetchall())
    
    total_messages = sum(message_stats.values())
    print(f"\n6. BROADCAST MESSAGES: {total_messages} total")
    for status, count in message_stats.items():
        print(f"   - {status}: {count}")
except Exception as e:
    print(f"\n6. BROADCAST MESSAGES: Error - {str(e)[:50]}")

# 7. WhatsApp specific tables
print(f"\n{'='*60}")
print("WHATSAPP TABLES")
print(f"{'='*60}")

whatsapp_tables = [t for t in tables if 'whatsapp' in t.lower() or 'whatsmeow' in t.lower()]
print(f"\nFound {len(whatsapp_tables)} WhatsApp-related tables:")
for table in whatsapp_tables:
    try:
        cursor.execute(f"SELECT COUNT(*) FROM {table}")
        count = cursor.fetchone()[0]
        print(f"   - {table}: {count} records")
    except:
        print(f"   - {table}: (unable to count)")

cursor.close()
conn.close()
print(f"\n{'='*60}")
print("[OK] Analysis complete!")
