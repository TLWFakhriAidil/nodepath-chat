#!/usr/bin/env python3
"""
Check PostgreSQL table structures before cleanup
"""

import os
import psycopg2
from urllib.parse import urlparse

def load_env():
    env_vars = {}
    env_path = '.env'
    if os.path.exists(env_path):
        with open(env_path, 'r') as f:
            for line in f:
                if '=' in line and not line.strip().startswith('#'):
                    key, value = line.strip().split('=', 1)
                    env_vars[key] = value.strip()
    return env_vars

def check_table_structure():
    """Check table structures to write correct cleanup queries"""
    env = load_env()
    pg_url = env.get('DB_URI', '')
    
    if not pg_url:
        print("[ERROR] No PostgreSQL connection")
        return
        
    try:
        conn = psycopg2.connect(pg_url)
        cur = conn.cursor()
        
        print("Checking Table Structures for Cleanup")
        print("=" * 70)
        
        # Check whatsmeow_message_secrets columns
        print("\n1. whatsmeow_message_secrets structure:")
        cur.execute("""
            SELECT column_name, data_type 
            FROM information_schema.columns 
            WHERE table_name = 'whatsmeow_message_secrets'
            ORDER BY ordinal_position
        """)
        for col in cur.fetchall():
            print(f"   - {col[0]} ({col[1]})")
            
        # Check user_devices columns
        print("\n2. user_devices structure:")
        cur.execute("""
            SELECT column_name, data_type 
            FROM information_schema.columns 
            WHERE table_name = 'user_devices' 
            AND column_name IN ('jid', 'status', 'last_seen')
        """)
        for col in cur.fetchall():
            print(f"   - {col[0]} ({col[1]})")
            
        # Check if there's a link between tables
        print("\n3. Checking relationships...")
        
        # Count records in each table
        tables = ['whatsmeow_message_secrets', 'user_devices', 'whatsapp_messages', 'whatsmeow_contacts']
        for table in tables:
            cur.execute(f"SELECT COUNT(*) FROM {table}")
            count = cur.fetchone()[0]
            print(f"   - {table}: {count:,} records")
            
        # Safe cleanup based on actual structure
        print("\n4. SAFE CLEANUP OPTIONS:")
        print("-" * 50)
        
        # For message secrets - check if we can clean by age
        cur.execute("""
            SELECT COUNT(*) 
            FROM whatsmeow_message_secrets 
            WHERE created_at < NOW() - INTERVAL '30 days'
        """)
        old_secrets = cur.fetchone()[0] if cur.lastrowid else 0
        print(f"Message secrets older than 30 days: {old_secrets:,}")
        
        # For messages
        cur.execute("""
            SELECT COUNT(*) 
            FROM whatsapp_messages 
            WHERE created_at < NOW() - INTERVAL '30 days'
        """)
        old_messages = cur.fetchone()[0]
        print(f"Messages older than 30 days: {old_messages:,}")
        
        # Show safe cleanup SQL
        print("\n5. SAFE CLEANUP COMMANDS:")
        print("-" * 50)
        print("-- Delete old messages")
        print("DELETE FROM whatsapp_messages WHERE created_at < NOW() - INTERVAL '30 days';")
        print("\n-- Clean duplicate contacts (if jid column exists)")
        print("""WITH duplicates AS (
    SELECT id, ROW_NUMBER() OVER (PARTITION BY jid ORDER BY id DESC) as rn
    FROM whatsmeow_contacts
)
DELETE FROM whatsmeow_contacts WHERE id IN (SELECT id FROM duplicates WHERE rn > 1);""")
        print("\n-- Run VACUUM")
        print("VACUUM FULL;")
        
        conn.close()
        
    except Exception as e:
        print(f"[ERROR] {e}")

if __name__ == "__main__":
    check_table_structure()
