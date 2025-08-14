import psycopg2
import pandas as pd
from datetime import datetime

# Connect to database
conn = psycopg2.connect(
    "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway",
    connect_timeout=30
)
cursor = conn.cursor()
print("[OK] Connected to Railway PostgreSQL!")

# Get all tables
cursor.execute("""
    SELECT table_name 
    FROM information_schema.tables 
    WHERE table_schema = 'public' 
    ORDER BY table_name;
""")
tables = cursor.fetchall()

print(f"\n=== DATABASE OVERVIEW ===")
print(f"Total tables: {len(tables)}")
print("\nTable list:")
for i, table in enumerate(tables):
    print(f"{i+1:2d}. {table[0]}")

# Check important tables
print("\n=== KEY TABLES ANALYSIS ===")

# 1. Check users
cursor.execute("SELECT COUNT(*) FROM users")
user_count = cursor.fetchone()[0]
print(f"\n1. USERS: {user_count} users registered")

# 2. Check devices
cursor.execute("SELECT COUNT(*) FROM user_devices")
device_count = cursor.fetchone()[0]
cursor.execute("SELECT COUNT(*) FROM user_devices WHERE status = 'online'")
online_count = cursor.fetchone()[0]
print(f"\n2. DEVICES: {device_count} total devices, {online_count} online")

# 3. Check leads
cursor.execute("SELECT COUNT(*) FROM leads")
lead_count = cursor.fetchone()[0]
cursor.execute("SELECT COUNT(DISTINCT trigger) FROM leads WHERE trigger IS NOT NULL")
trigger_types = cursor.fetchone()[0]
print(f"\n3. LEADS: {lead_count} total leads, {trigger_types} unique triggers")

# 4. Check campaigns
cursor.execute("SELECT COUNT(*) FROM campaigns")
campaign_count = cursor.fetchone()[0]
cursor.execute("SELECT COUNT(*) FROM campaigns WHERE status = 'active'")
active_campaigns = cursor.fetchone()[0]
print(f"\n4. CAMPAIGNS: {campaign_count} total campaigns, {active_campaigns} active")

# 5. Check sequences
cursor.execute("SELECT COUNT(*) FROM sequences")
sequence_count = cursor.fetchone()[0]
cursor.execute("SELECT COUNT(*) FROM sequences WHERE is_active = true")
active_sequences = cursor.fetchone()[0]
print(f"\n5. SEQUENCES: {sequence_count} total sequences, {active_sequences} active")

# 6. Check broadcast messages
cursor.execute("SELECT COUNT(*) FROM broadcast_messages")
total_messages = cursor.fetchone()[0]
cursor.execute("SELECT COUNT(*) FROM broadcast_messages WHERE status = 'pending'")
pending_messages = cursor.fetchone()[0]
cursor.execute("SELECT COUNT(*) FROM broadcast_messages WHERE status = 'sent'")
sent_messages = cursor.fetchone()[0]
print(f"\n6. BROADCAST MESSAGES:")
print(f"   - Total: {total_messages}")
print(f"   - Pending: {pending_messages}")
print(f"   - Sent: {sent_messages}")

# Show sample triggers
print("\n=== SAMPLE TRIGGERS ===")
cursor.execute("""
    SELECT trigger, COUNT(*) as count 
    FROM leads 
    WHERE trigger IS NOT NULL 
    GROUP BY trigger 
    ORDER BY count DESC 
    LIMIT 10
""")
triggers = cursor.fetchall()
for trigger, count in triggers:
    print(f"   - {trigger}: {count} leads")

# Show active sequences
print("\n=== ACTIVE SEQUENCES ===")
cursor.execute("""
    SELECT name, description, niche 
    FROM sequences 
    WHERE is_active = true 
    ORDER BY created_at DESC 
    LIMIT 10
""")
sequences = cursor.fetchall()
for seq in sequences:
    print(f"   - {seq[0]} ({seq[2] or 'No niche'}): {seq[1] or 'No description'}")

cursor.close()
conn.close()
print("\n[OK] Analysis complete!")
