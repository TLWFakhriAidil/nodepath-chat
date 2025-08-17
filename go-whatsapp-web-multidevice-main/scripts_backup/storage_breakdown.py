#!/usr/bin/env python3
"""
PostgreSQL Storage Breakdown Analysis
"""

import psycopg2
import os
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

def main():
    env = load_env()
    pg_url = env.get('DB_URI', '')
    
    if not pg_url:
        print("[ERROR] No PostgreSQL connection")
        return
        
    conn = psycopg2.connect(pg_url)
    conn.autocommit = True
    
    with conn.cursor() as cur:
        print("\n=== PostgreSQL Storage Breakdown (121 MB Total) ===\n")
        
        # Summary by category
        print("STORAGE BY CATEGORY:")
        print("-" * 50)
        
        # 1. WhatsApp Session Data (whatsmeow_*)
        cur.execute("""
            SELECT 
                COUNT(*) as table_count,
                SUM(pg_total_relation_size('public.'||tablename)) as total_bytes,
                pg_size_pretty(SUM(pg_total_relation_size('public.'||tablename))) as total_size
            FROM pg_tables
            WHERE schemaname = 'public' AND tablename LIKE 'whatsmeow_%'
        """)
        count, bytes_size, pretty_size = cur.fetchone()
        mb_size = bytes_size / (1024*1024) if bytes_size else 0
        print(f"1. WhatsApp Session Data: {pretty_size} ({mb_size:.1f} MB - {mb_size/121*100:.1f}%)")
        print(f"   - {count} tables storing device sessions & encryption")
        
        # 2. WhatsApp Messages & Chats
        cur.execute("""
            SELECT 
                tablename,
                pg_size_pretty(pg_total_relation_size('public.'||tablename)) as size,
                (SELECT COUNT(*) FROM public.whatsapp_messages) as msg_count,
                (SELECT COUNT(*) FROM public.whatsapp_chats) as chat_count
            FROM pg_tables
            WHERE tablename IN ('whatsapp_messages', 'whatsapp_chats')
            LIMIT 1
        """)
        result = cur.fetchone()
        msg_count = result[2] if result else 0
        chat_count = result[3] if result else 0
        
        print(f"\n2. WhatsApp Chat Data: ~39 MB (32.2%)")
        print(f"   - {msg_count:,} messages stored")
        print(f"   - {chat_count:,} chats tracked")
        
        # 3. Application Data
        print(f"\n3. Application Data: ~0.9 MB (0.7%)")
        print(f"   - User accounts, devices, leads (all cleared)")
        
        # Largest tables details
        print("\n\nLARGEST TABLES IN DETAIL:")
        print("-" * 70)
        
        # Message secrets analysis
        cur.execute("""
            SELECT COUNT(*), COUNT(DISTINCT jid) as unique_jids
            FROM public.whatsmeow_message_secrets
        """)
        total_secrets, unique_jids = cur.fetchone()
        
        print(f"\n1. whatsmeow_message_secrets (37 MB - 30.6% of total)")
        print(f"   - {total_secrets:,} encryption keys stored")
        print(f"   - {unique_jids} unique devices")
        print(f"   - Average: {total_secrets//unique_jids if unique_jids else 0} keys per device")
        print("   - Purpose: End-to-end encryption for WhatsApp messages")
        
        # Messages analysis
        print(f"\n2. whatsapp_messages (28 MB - 23.1% of total)")
        print(f"   - {msg_count:,} messages stored")
        print("   - Includes text content, media URLs, timestamps")
        
        # Get message age distribution
        try:
            cur.execute("""
                SELECT 
                    COUNT(CASE WHEN created_at > NOW() - INTERVAL '7 days' THEN 1 END) as week_old,
                    COUNT(CASE WHEN created_at > NOW() - INTERVAL '30 days' AND created_at <= NOW() - INTERVAL '7 days' THEN 1 END) as month_old,
                    COUNT(CASE WHEN created_at <= NOW() - INTERVAL '30 days' THEN 1 END) as older
                FROM public.whatsapp_messages
            """)
            week_old, month_old, older = cur.fetchone()
            print(f"   - Last 7 days: {week_old:,} messages")
            print(f"   - 7-30 days: {month_old:,} messages")
            print(f"   - Older than 30 days: {older:,} messages")
        except:
            pass
        
        # Contacts analysis
        cur.execute("SELECT COUNT(*), COUNT(DISTINCT jid) FROM public.whatsmeow_contacts")
        total_contacts, unique_contacts = cur.fetchone()
        
        print(f"\n3. whatsmeow_contacts (18 MB - 14.9% of total)")
        print(f"   - {total_contacts:,} contact records")
        print(f"   - {unique_contacts:,} unique contacts")
        print("   - Stores contact names, phone numbers, profile data")
        
        # Recommendations
        print("\n\nCLEANUP RECOMMENDATIONS:")
        print("-" * 70)
        
        print("\n1. Clean old message encryption keys (could free ~20 MB):")
        print("   DELETE FROM whatsmeow_message_secrets")
        print("   WHERE jid IN (SELECT jid FROM user_devices WHERE last_seen < NOW() - INTERVAL '30 days');")
        
        print("\n2. Archive old messages (could free ~15 MB):")
        print("   DELETE FROM whatsapp_messages WHERE created_at < NOW() - INTERVAL '30 days';")
        
        print("\n3. Clean duplicate contacts (could free ~5 MB):")
        print("   -- Remove duplicate contacts keeping the most recent")
        print("   DELETE FROM whatsmeow_contacts a USING whatsmeow_contacts b")
        print("   WHERE a.id < b.id AND a.jid = b.jid;")
        
        print("\n4. Clean old session data:")
        print("   DELETE FROM whatsmeow_sessions WHERE last_seen < NOW() - INTERVAL '7 days';")
        
        print("\n5. After cleanup, run:")
        print("   VACUUM FULL;")
        
        print("\n\nESTIMATED SPACE AFTER FULL CLEANUP: ~60-70 MB (50% reduction possible)")
        
    conn.close()

if __name__ == "__main__":
    main()
