#!/usr/bin/env python3
"""
Simplified cleanup focusing on what we can actually clean
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

def simple_cleanup():
    env = load_env()
    pg_url = env.get('DB_URI', '')
    
    if not pg_url:
        print("[ERROR] No PostgreSQL connection")
        return
        
    try:
        conn = psycopg2.connect(pg_url)
        conn.autocommit = True
        
        with conn.cursor() as cur:
            print("PostgreSQL Simple Cleanup")
            print("=" * 70)
            
            # Get initial size
            cur.execute("SELECT pg_size_pretty(pg_database_size(current_database())) as db_size")
            size_before = cur.fetchone()[0]
            print(f"\nDatabase size before: {size_before}")
            
            # 1. First, let's check whatsmeow_contacts structure
            print("\n1. Checking whatsmeow_contacts structure...")
            cur.execute("""
                SELECT column_name 
                FROM information_schema.columns 
                WHERE table_name = 'whatsmeow_contacts'
                LIMIT 10
            """)
            columns = [row[0] for row in cur.fetchall()]
            print(f"Columns: {', '.join(columns)}")
            
            # 2. Clean orphaned data based on user_devices
            print("\n2. Cleaning orphaned WhatsApp session data...")
            
            # Get active device JIDs
            cur.execute("SELECT DISTINCT jid FROM user_devices WHERE jid IS NOT NULL")
            active_jids = [row[0] for row in cur.fetchall()]
            print(f"Active device JIDs: {len(active_jids)}")
            
            if active_jids:
                # Clean message secrets for non-active devices
                placeholders = ','.join(['%s'] * len(active_jids))
                
                # Check how many secrets we have
                cur.execute("SELECT COUNT(*) FROM whatsmeow_message_secrets")
                total_before = cur.fetchone()[0]
                
                # Delete secrets not belonging to active devices
                cur.execute(f"""
                    DELETE FROM whatsmeow_message_secrets 
                    WHERE our_jid NOT IN ({placeholders})
                """, active_jids)
                deleted_secrets = cur.rowcount
                
                print(f"Message secrets: {total_before:,} -> {total_before - deleted_secrets:,} (deleted {deleted_secrets:,})")
                
                # Clean sessions for non-active devices
                cur.execute(f"""
                    DELETE FROM whatsmeow_sessions
                    WHERE jid NOT IN ({placeholders})
                """, active_jids)
                print(f"Deleted {cur.rowcount} orphaned sessions")
            
            # 3. Clean really old messages (if any)
            print("\n3. Checking for very old messages...")
            cur.execute("""
                SELECT 
                    MIN(created_at) as oldest,
                    MAX(created_at) as newest,
                    COUNT(*) as total
                FROM whatsapp_messages
            """)
            oldest, newest, total = cur.fetchone()
            print(f"Messages: {total:,} (from {oldest} to {newest})")
            
            # Delete messages older than 7 days to save more space
            cur.execute("DELETE FROM whatsapp_messages WHERE created_at < NOW() - INTERVAL '7 days'")
            if cur.rowcount > 0:
                print(f"Deleted {cur.rowcount:,} messages older than 7 days")
            
            # 4. Limit message secrets per device
            print("\n4. Limiting message secrets per device...")
            
            # Keep only last 5000 secrets per device
            cur.execute("""
                WITH numbered_secrets AS (
                    SELECT message_id,
                           ROW_NUMBER() OVER (PARTITION BY our_jid ORDER BY message_id DESC) as rn
                    FROM whatsmeow_message_secrets
                )
                DELETE FROM whatsmeow_message_secrets
                WHERE message_id IN (
                    SELECT message_id FROM numbered_secrets WHERE rn > 5000
                )
            """)
            if cur.rowcount > 0:
                print(f"Deleted {cur.rowcount:,} old message secrets (kept last 5000 per device)")
            
            # 5. Clean app state data
            print("\n5. Cleaning app state data...")
            
            # Clean old app state mutation macs
            cur.execute("""
                WITH numbered_macs AS (
                    SELECT index_mac,
                           ROW_NUMBER() OVER (PARTITION BY jid ORDER BY index_mac DESC) as rn
                    FROM whatsmeow_app_state_mutation_macs
                )
                DELETE FROM whatsmeow_app_state_mutation_macs
                WHERE index_mac IN (
                    SELECT index_mac FROM numbered_macs WHERE rn > 1000
                )
            """)
            if cur.rowcount > 0:
                print(f"Deleted {cur.rowcount:,} old app state records")
            
            # 6. VACUUM
            print("\n6. Running VACUUM FULL...")
            print("This will take a minute...")
            cur.execute("VACUUM FULL")
            print("[OK] VACUUM completed")
            
            # Final results
            print("\n" + "=" * 70)
            print("FINAL RESULTS:")
            print("=" * 70)
            
            # Get final size
            cur.execute("SELECT pg_size_pretty(pg_database_size(current_database())) as db_size")
            size_after = cur.fetchone()[0]
            
            # Get numeric size for calculation
            cur.execute("SELECT pg_database_size(current_database()) as size_bytes")
            size_bytes = cur.fetchone()[0]
            size_mb_after = size_bytes / (1024*1024)
            
            # Show largest tables
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
            
            print("\nLargest tables:")
            for row in cur.fetchall():
                print(f"  {row[0]:<35} {row[1]:>10}")
            
            print(f"\nDatabase size: {size_before} -> {size_after}")
            print(f"Space freed: ~{121 - size_mb_after:.1f} MB")
            print(f"Total reduction: {((121 - size_mb_after)/121)*100:.1f}%")
            
        conn.close()
        
    except Exception as e:
        print(f"[ERROR] {e}")
        import traceback
        traceback.print_exc()

if __name__ == "__main__":
    simple_cleanup()
