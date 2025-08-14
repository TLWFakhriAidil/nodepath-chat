#!/usr/bin/env python3
"""
Run PostgreSQL Additional Cleanup
"""

import os
import psycopg2
from datetime import datetime
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

def run_additional_cleanup():
    """Run additional PostgreSQL cleanup commands"""
    env = load_env()
    pg_url = env.get('DB_URI') or env.get('DATABASE_URL', '')
    
    if not pg_url:
        print("[ERROR] No PostgreSQL connection string found")
        return
        
    try:
        conn = psycopg2.connect(pg_url)
        conn.autocommit = True
        
        with conn.cursor() as cur:
            print("PostgreSQL Additional Cleanup")
            print("=" * 70)
            
            # 1. Check current sizes
            print("\n1. CURRENT TABLE SIZES:")
            print("-" * 50)
            cur.execute("""
                SELECT 
                    tablename,
                    pg_size_pretty(pg_total_relation_size('public.'||tablename)) as size,
                    pg_total_relation_size('public.'||tablename)/1024/1024 as size_mb
                FROM pg_tables
                WHERE schemaname = 'public' 
                AND tablename IN ('whatsmeow_message_secrets', 'whatsapp_messages', 'whatsmeow_contacts', 'whatsapp_chats')
                ORDER BY size_mb DESC
            """)
            
            total_before = 0
            for row in cur.fetchall():
                print(f"{row[0]:<30} {row[1]:>10} ({row[2]:.1f} MB)")
                total_before += row[2]
            print(f"\nTotal of these tables: {total_before:.1f} MB")
            
            # Get overall DB size before
            cur.execute("SELECT pg_size_pretty(pg_database_size(current_database())) as db_size")
            db_size_before = cur.fetchone()[0]
            print(f"Total database size: {db_size_before}")
            
            # 2. Clean old message encryption keys
            print("\n2. CLEANING OLD MESSAGE ENCRYPTION KEYS...")
            print("-" * 50)
            
            # First check how many will be deleted
            cur.execute("""
                SELECT COUNT(*) FROM whatsmeow_message_secrets 
                WHERE jid IN (
                    SELECT jid FROM user_devices 
                    WHERE status != 'online' 
                    OR last_seen < NOW() - INTERVAL '30 days'
                )
            """)
            keys_to_delete = cur.fetchone()[0]
            print(f"Keys to delete: {keys_to_delete:,}")
            
            # Delete them
            cur.execute("""
                DELETE FROM whatsmeow_message_secrets 
                WHERE jid IN (
                    SELECT jid FROM user_devices 
                    WHERE status != 'online' 
                    OR last_seen < NOW() - INTERVAL '30 days'
                )
            """)
            print(f"[OK] Deleted {cur.rowcount:,} encryption keys")
            
            # 3. Archive old messages
            print("\n3. ARCHIVING OLD MESSAGES...")
            print("-" * 50)
            
            # Check how many messages are older than 30 days
            cur.execute("SELECT COUNT(*) FROM whatsapp_messages WHERE created_at < NOW() - INTERVAL '30 days'")
            old_messages = cur.fetchone()[0]
            print(f"Messages older than 30 days: {old_messages:,}")
            
            # Delete old messages
            cur.execute("DELETE FROM whatsapp_messages WHERE created_at < NOW() - INTERVAL '30 days'")
            print(f"[OK] Deleted {cur.rowcount:,} old messages")
            
            # 4. Clean duplicate contacts
            print("\n4. CLEANING DUPLICATE CONTACTS...")
            print("-" * 50)
            
            # Find duplicates
            cur.execute("""
                WITH duplicates AS (
                    SELECT COUNT(*) as dup_count 
                    FROM (
                        SELECT jid, COUNT(*) as cnt 
                        FROM whatsmeow_contacts 
                        GROUP BY jid 
                        HAVING COUNT(*) > 1
                    ) t
                )
                SELECT COALESCE(SUM(dup_count), 0) FROM duplicates
            """)
            duplicates = cur.fetchone()[0]
            print(f"Duplicate contacts found: {duplicates}")
            
            # Remove duplicates
            cur.execute("""
                WITH duplicates AS (
                    SELECT id, 
                           ROW_NUMBER() OVER (PARTITION BY jid ORDER BY id DESC) as rn
                    FROM whatsmeow_contacts
                )
                DELETE FROM whatsmeow_contacts
                WHERE id IN (SELECT id FROM duplicates WHERE rn > 1)
            """)
            print(f"[OK] Removed {cur.rowcount:,} duplicate contacts")
            
            # 5. Clean old chats
            print("\n5. CLEANING OLD CHATS...")
            print("-" * 50)
            
            cur.execute("""
                DELETE FROM whatsapp_chats
                WHERE last_message_time < NOW() - INTERVAL '60 days'
                AND chat_jid NOT IN (
                    SELECT DISTINCT chat_jid 
                    FROM whatsapp_messages 
                    WHERE created_at > NOW() - INTERVAL '60 days'
                )
            """)
            print(f"[OK] Removed {cur.rowcount:,} old chats")
            
            # 6. Clean orphaned session data
            print("\n6. CLEANING ORPHANED SESSION DATA...")
            print("-" * 50)
            
            # Clean sessions
            cur.execute("""
                DELETE FROM whatsmeow_sessions
                WHERE jid NOT IN (SELECT jid FROM user_devices)
            """)
            sessions_deleted = cur.rowcount
            
            # Clean identity keys
            cur.execute("""
                DELETE FROM whatsmeow_identity_keys
                WHERE jid NOT IN (SELECT jid FROM user_devices)
            """)
            identity_deleted = cur.rowcount
            
            # Clean sender keys
            cur.execute("""
                DELETE FROM whatsmeow_sender_keys
                WHERE chat_jid NOT IN (SELECT DISTINCT chat_jid FROM whatsapp_chats)
            """)
            sender_deleted = cur.rowcount
            
            print(f"[OK] Cleaned: {sessions_deleted} sessions, {identity_deleted} identity keys, {sender_deleted} sender keys")
            
            # 7. Clean app state data
            print("\n7. CLEANING APP STATE DATA...")
            print("-" * 50)
            
            cur.execute("""
                DELETE FROM whatsmeow_app_state_mutation_macs
                WHERE jid NOT IN (SELECT jid FROM user_devices WHERE status = 'online')
            """)
            print(f"[OK] Removed {cur.rowcount:,} app state records")
            
            # 8. Run VACUUM
            print("\n8. RUNNING VACUUM FULL...")
            print("-" * 50)
            print("This may take a minute...")
            cur.execute("VACUUM FULL")
            print("[OK] VACUUM completed - disk space reclaimed")
            
            # 9. Check sizes after cleanup
            print("\n9. FINAL RESULTS:")
            print("-" * 50)
            
            cur.execute("""
                SELECT 
                    tablename,
                    pg_size_pretty(pg_total_relation_size('public.'||tablename)) as size,
                    pg_total_relation_size('public.'||tablename)/1024/1024 as size_mb
                FROM pg_tables
                WHERE schemaname = 'public' 
                AND tablename IN ('whatsmeow_message_secrets', 'whatsapp_messages', 'whatsmeow_contacts', 'whatsapp_chats')
                ORDER BY size_mb DESC
            """)
            
            total_after = 0
            for row in cur.fetchall():
                print(f"{row[0]:<30} {row[1]:>10} ({row[2]:.1f} MB)")
                total_after += row[2]
            
            # Get overall DB size after
            cur.execute("SELECT pg_size_pretty(pg_database_size(current_database())) as db_size")
            db_size_after = cur.fetchone()[0]
            
            print(f"\nTotal of these tables: {total_after:.1f} MB")
            print(f"Space saved from these tables: {total_before - total_after:.1f} MB")
            print(f"\nTotal database size: {db_size_after}")
            print(f"Previous size: {db_size_before}")
            
            # Show all large tables
            print("\n\nALL TABLES OVER 1 MB:")
            print("-" * 60)
            cur.execute("""
                SELECT 
                    tablename,
                    pg_size_pretty(pg_total_relation_size('public.'||tablename)) as size
                FROM pg_tables
                WHERE schemaname = 'public' 
                AND pg_total_relation_size('public.'||tablename) > 1024*1024
                ORDER BY pg_total_relation_size('public.'||tablename) DESC
                LIMIT 15
            """)
            
            for row in cur.fetchall():
                print(f"{row[0]:<35} {row[1]:>10}")
                
        conn.close()
        print("\n[OK] Additional cleanup completed successfully!")
        
    except Exception as e:
        print(f"[ERROR] Error during cleanup: {e}")

if __name__ == "__main__":
    run_additional_cleanup()
