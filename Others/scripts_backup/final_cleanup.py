#!/usr/bin/env python3
"""
Final cleanup with correct column names
"""

import os
import psycopg2

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

def final_cleanup():
    env = load_env()
    pg_url = env.get('DB_URI', '')
    
    if not pg_url:
        return
        
    try:
        conn = psycopg2.connect(pg_url)
        conn.autocommit = True
        
        with conn.cursor() as cur:
            print("PostgreSQL Final Cleanup")
            print("=" * 70)
            
            # Continue from where we left off
            print("\n1. Continuing cleanup after deleting 70,404 message secrets...")
            
            # Get active JIDs again
            cur.execute("SELECT DISTINCT jid FROM user_devices WHERE jid IS NOT NULL")
            active_jids = [row[0] for row in cur.fetchall()]
            
            # Clean contacts - using correct column name
            print("\n2. Cleaning duplicate contacts...")
            cur.execute("""
                WITH duplicates AS (
                    SELECT their_jid, COUNT(*) as cnt,
                           MIN(LENGTH(COALESCE(full_name, '') || COALESCE(push_name, ''))) as min_name_len
                    FROM whatsmeow_contacts
                    GROUP BY their_jid
                    HAVING COUNT(*) > 1
                )
                SELECT COUNT(*) FROM duplicates
            """)
            dup_count = cur.fetchone()[0]
            print(f"Found {dup_count} JIDs with duplicates")
            
            # Remove duplicates keeping the one with most info
            cur.execute("""
                DELETE FROM whatsmeow_contacts a
                USING whatsmeow_contacts b
                WHERE a.their_jid = b.their_jid
                AND a.our_jid = b.our_jid
                AND (
                    LENGTH(COALESCE(a.full_name, '') || COALESCE(a.push_name, '')) < 
                    LENGTH(COALESCE(b.full_name, '') || COALESCE(b.push_name, ''))
                    OR (LENGTH(COALESCE(a.full_name, '') || COALESCE(a.push_name, '')) = 
                        LENGTH(COALESCE(b.full_name, '') || COALESCE(b.push_name, ''))
                        AND a.ctid < b.ctid)
                )
            """)
            print(f"Removed {cur.rowcount:,} duplicate contacts")
            
            # Clean old messages (keep last 3 days only for aggressive cleanup)
            print("\n3. Aggressive message cleanup...")
            cur.execute("SELECT COUNT(*) FROM whatsapp_messages WHERE created_at < NOW() - INTERVAL '3 days'")
            old_msgs = cur.fetchone()[0]
            
            if old_msgs > 0:
                cur.execute("DELETE FROM whatsapp_messages WHERE created_at < NOW() - INTERVAL '3 days'")
                print(f"Deleted {cur.rowcount:,} messages older than 3 days")
            
            # Clean chats without recent messages
            print("\n4. Cleaning old chats...")
            cur.execute("""
                DELETE FROM whatsapp_chats
                WHERE chat_jid NOT IN (
                    SELECT DISTINCT chat_jid FROM whatsapp_messages
                )
            """)
            print(f"Removed {cur.rowcount:,} chats without messages")
            
            # More aggressive message secrets cleanup
            print("\n5. More aggressive message secrets cleanup...")
            cur.execute("""
                WITH numbered AS (
                    SELECT message_id,
                           ROW_NUMBER() OVER (PARTITION BY our_jid ORDER BY message_id DESC) as rn
                    FROM whatsmeow_message_secrets
                )
                DELETE FROM whatsmeow_message_secrets
                WHERE message_id IN (
                    SELECT message_id FROM numbered WHERE rn > 2000
                )
            """)
            print(f"Deleted {cur.rowcount:,} more old message secrets (kept last 2000 per device)")
            
            # Clean app state
            print("\n6. Cleaning app state...")
            cur.execute("""
                DELETE FROM whatsmeow_app_state_mutation_macs
                WHERE jid NOT IN (SELECT jid FROM user_devices WHERE status = 'online')
            """)
            print(f"Deleted {cur.rowcount:,} app state records for offline devices")
            
            # VACUUM
            print("\n7. Running VACUUM FULL...")
            cur.execute("VACUUM FULL")
            print("[OK] VACUUM completed")
            
            # Final results
            print("\n" + "=" * 70)
            
            cur.execute("SELECT pg_size_pretty(pg_database_size(current_database()))")
            final_size = cur.fetchone()[0]
            
            cur.execute("SELECT pg_database_size(current_database())")
            size_bytes = cur.fetchone()[0]
            final_mb = size_bytes / (1024*1024)
            
            # Show top tables
            cur.execute("""
                SELECT tablename, pg_size_pretty(pg_total_relation_size('public.'||tablename))
                FROM pg_tables
                WHERE schemaname = 'public'
                AND pg_total_relation_size('public.'||tablename) > 1024*1024
                ORDER BY pg_total_relation_size('public.'||tablename) DESC
                LIMIT 10
            """)
            
            print("\nFinal table sizes:")
            for row in cur.fetchall():
                print(f"  {row[0]:<35} {row[1]:>10}")
            
            print(f"\nFinal database size: {final_size}")
            print(f"Reduced from 121 MB to {final_mb:.1f} MB")
            print(f"Total space freed: {121 - final_mb:.1f} MB ({((121 - final_mb)/121)*100:.1f}% reduction)")
            
        conn.close()
        
    except Exception as e:
        print(f"[ERROR] {e}")

if __name__ == "__main__":
    final_cleanup()
