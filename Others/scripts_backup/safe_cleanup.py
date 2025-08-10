#!/usr/bin/env python3
"""
Safe PostgreSQL Additional Cleanup
Based on actual table structures
"""

import os
import psycopg2
from datetime import datetime

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

def run_safe_cleanup():
    """Run safe PostgreSQL cleanup based on actual structure"""
    env = load_env()
    pg_url = env.get('DB_URI', '')
    
    if not pg_url:
        print("[ERROR] No PostgreSQL connection")
        return
        
    try:
        conn = psycopg2.connect(pg_url)
        conn.autocommit = True
        
        with conn.cursor() as cur:
            print("Safe PostgreSQL Additional Cleanup")
            print("=" * 70)
            
            # Get initial database size
            cur.execute("SELECT pg_size_pretty(pg_database_size(current_database())) as db_size")
            db_size_before = cur.fetchone()[0]
            print(f"\nInitial database size: {db_size_before}")
            
            # 1. Clean old messages (has created_at column)
            print("\n1. CLEANING OLD MESSAGES...")
            print("-" * 50)
            
            # Check message age distribution
            cur.execute("""
                SELECT 
                    COUNT(CASE WHEN created_at > NOW() - INTERVAL '7 days' THEN 1 END) as last_week,
                    COUNT(CASE WHEN created_at > NOW() - INTERVAL '30 days' THEN 1 END) as last_month,
                    COUNT(*) as total
                FROM whatsapp_messages
            """)
            last_week, last_month, total = cur.fetchone()
            print(f"Total messages: {total:,}")
            print(f"Last 7 days: {last_week:,}")
            print(f"Last 30 days: {last_month:,}")
            print(f"Older than 30 days: {total - last_month:,}")
            
            if total - last_month > 0:
                # Delete messages older than 30 days
                cur.execute("DELETE FROM whatsapp_messages WHERE created_at < NOW() - INTERVAL '30 days'")
                print(f"[OK] Deleted {cur.rowcount:,} old messages")
            else:
                print("[INFO] No old messages to delete")
            
            # 2. Clean duplicate contacts
            print("\n2. CLEANING DUPLICATE CONTACTS...")
            print("-" * 50)
            
            # Check for duplicates
            cur.execute("""
                SELECT jid, COUNT(*) as cnt 
                FROM whatsmeow_contacts 
                GROUP BY jid 
                HAVING COUNT(*) > 1
                LIMIT 10
            """)
            duplicates = cur.fetchall()
            
            if duplicates:
                print(f"Found duplicates for {len(duplicates)} JIDs")
                
                # Remove duplicates keeping newest
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
            else:
                print("[INFO] No duplicate contacts found")
            
            # 3. Clean orphaned message secrets
            print("\n3. CLEANING ORPHANED MESSAGE SECRETS...")
            print("-" * 50)
            
            # Get active JIDs from user_devices
            cur.execute("SELECT DISTINCT jid FROM user_devices WHERE status = 'online'")
            active_jids = [row[0] for row in cur.fetchall()]
            print(f"Active devices: {len(active_jids)}")
            
            # Clean secrets for inactive devices
            if active_jids:
                placeholders = ','.join(['%s'] * len(active_jids))
                cur.execute(f"""
                    DELETE FROM whatsmeow_message_secrets 
                    WHERE our_jid NOT IN ({placeholders})
                """, active_jids)
                print(f"[OK] Removed {cur.rowcount:,} message secrets for inactive devices")
            
            # 4. Clean old chats without recent messages
            print("\n4. CLEANING OLD CHATS...")
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
            
            # 5. Clean whatsmeow sessions for inactive devices
            print("\n5. CLEANING INACTIVE SESSIONS...")
            print("-" * 50)
            
            cur.execute("""
                DELETE FROM whatsmeow_sessions
                WHERE jid NOT IN (SELECT jid FROM user_devices WHERE status = 'online')
            """)
            print(f"[OK] Removed {cur.rowcount:,} inactive sessions")
            
            # 6. Aggressive cleanup for message secrets (optional)
            print("\n6. AGGRESSIVE MESSAGE SECRETS CLEANUP...")
            print("-" * 50)
            
            # Keep only recent message secrets
            cur.execute("""
                SELECT COUNT(*) FROM whatsmeow_message_secrets
            """)
            total_secrets = cur.fetchone()[0]
            
            # Delete very old ones (keep last 10k per device)
            cur.execute("""
                WITH ranked_secrets AS (
                    SELECT *, ROW_NUMBER() OVER (PARTITION BY our_jid ORDER BY message_id DESC) as rn
                    FROM whatsmeow_message_secrets
                )
                DELETE FROM whatsmeow_message_secrets
                WHERE message_id IN (
                    SELECT message_id FROM ranked_secrets WHERE rn > 10000
                )
            """)
            
            if cur.rowcount > 0:
                print(f"[OK] Removed {cur.rowcount:,} old message secrets (kept 10k per device)")
            
            # 7. Run VACUUM FULL
            print("\n7. RUNNING VACUUM FULL...")
            print("-" * 50)
            print("This may take a few minutes...")
            cur.execute("VACUUM FULL")
            print("[OK] VACUUM completed")
            
            # 8. Final results
            print("\n8. FINAL RESULTS:")
            print("=" * 70)
            
            # Get final database size
            cur.execute("SELECT pg_size_pretty(pg_database_size(current_database())) as db_size")
            db_size_after = cur.fetchone()[0]
            
            # Get table sizes
            cur.execute("""
                SELECT 
                    tablename,
                    pg_size_pretty(pg_total_relation_size('public.'||tablename)) as size
                FROM pg_tables
                WHERE schemaname = 'public' 
                AND pg_total_relation_size('public.'||tablename) > 1024*1024
                ORDER BY pg_total_relation_size('public.'||tablename) DESC
                LIMIT 10
            """)
            
            print("\nLargest tables after cleanup:")
            print("-" * 50)
            for row in cur.fetchall():
                print(f"{row[0]:<35} {row[1]:>10}")
            
            print(f"\nDatabase size before: {db_size_before}")
            print(f"Database size after: {db_size_after}")
            
            # Calculate reduction
            cur.execute("""
                SELECT 
                    pg_database_size(current_database()) as size_bytes
            """)
            size_bytes = cur.fetchone()[0]
            size_mb = size_bytes / (1024*1024)
            reduction = 121 - size_mb
            
            print(f"\nSpace freed: ~{reduction:.1f} MB")
            print(f"Reduction: {(reduction/121)*100:.1f}%")
            
        conn.close()
        print("\n[OK] Safe cleanup completed successfully!")
        
    except Exception as e:
        print(f"[ERROR] {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    run_safe_cleanup()
