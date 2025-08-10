#!/usr/bin/env python3
"""
PostgreSQL Detailed Disk Usage Analysis
Shows what's taking up space in the database
"""

import os
import psycopg2
from datetime import datetime
from urllib.parse import urlparse

# Load environment variables
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

def analyze_postgresql_storage():
    """Analyze PostgreSQL storage in detail"""
    env = load_env()
    pg_url = env.get('DB_URI') or env.get('DATABASE_URL', '')
    
    if not pg_url:
        print("[ERROR] No PostgreSQL connection string found")
        return
        
    try:
        conn = psycopg2.connect(pg_url)
        conn.autocommit = True
        
        with conn.cursor() as cur:
            print("PostgreSQL Detailed Storage Analysis")
            print("=" * 70)
            
            # Overall database size
            cur.execute("SELECT pg_size_pretty(pg_database_size(current_database())) as db_size")
            print(f"\nTotal Database Size: {cur.fetchone()[0]}")
            print("-" * 70)
            
            # Detailed table sizes with row counts
            cur.execute("""
                SELECT 
                    schemaname,
                    tablename,
                    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
                    pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
                    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS indexes_size,
                    pg_total_relation_size(schemaname||'.'||tablename) AS size_bytes,
                    (SELECT COUNT(*) FROM information_schema.tables t WHERE t.table_schema = schemaname AND t.table_name = tablename) as exists
                FROM pg_tables
                WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
                ORDER BY size_bytes DESC
            """)
            
            tables = cur.fetchall()
            
            print("\nTable Storage Breakdown:")
            print(f"{'Table':<40} {'Total Size':>12} {'Table':>12} {'Indexes':>12} {'Rows':>12}")
            print("-" * 90)
            
            total_size_bytes = 0
            
            for table in tables[:25]:  # Top 25 tables
                schema, tablename, total_size, table_size, indexes_size, size_bytes, exists = table
                
                # Get row count
                try:
                    cur.execute(f"SELECT COUNT(*) FROM {schema}.{tablename}")
                    row_count = cur.fetchone()[0]
                except:
                    row_count = "N/A"
                
                print(f"{schema}.{tablename:<30} {total_size:>12} {table_size:>12} {indexes_size:>12} {str(row_count):>12}")
                total_size_bytes += size_bytes
            
            print("-" * 90)
            
            # WhatsApp session tables analysis
            print("\n\nWhatsApp Session Tables (whatsmeow_*):")
            print("-" * 70)
            
            cur.execute("""
                SELECT 
                    tablename,
                    pg_size_pretty(pg_total_relation_size('public.'||tablename)) AS size,
                    pg_total_relation_size('public.'||tablename) AS size_bytes
                FROM pg_tables
                WHERE schemaname = 'public' AND tablename LIKE 'whatsmeow_%'
                ORDER BY size_bytes DESC
            """)
            
            whatsmeow_tables = cur.fetchall()
            whatsmeow_total = 0
            
            for table in whatsmeow_tables:
                tablename, size, size_bytes = table
                
                # Get row count and sample data
                try:
                    cur.execute(f"SELECT COUNT(*) FROM public.{tablename}")
                    row_count = cur.fetchone()[0]
                    
                    # Get table description
                    cur.execute(f"""
                        SELECT column_name, data_type 
                        FROM information_schema.columns 
                        WHERE table_name = '{tablename}'
                        LIMIT 3
                    """)
                    columns = cur.fetchall()
                    col_info = ", ".join([f"{col[0]} ({col[1]})" for col in columns])
                    
                except:
                    row_count = "N/A"
                    col_info = "N/A"
                
                print(f"{tablename:<35} {size:>10} ({row_count:>8} rows)")
                whatsmeow_total += size_bytes
            
            print(f"\nTotal WhatsApp Session Storage: {whatsmeow_total / (1024*1024):.1f} MB")
            
            # Application data tables
            print("\n\nApplication Data Tables:")
            print("-" * 70)
            
            app_tables = ['users', 'user_devices', 'leads', 'campaigns', 'sequences', 
                         'broadcast_messages', 'whatsapp_messages', 'whatsapp_chats']
            
            for tablename in app_tables:
                try:
                    cur.execute(f"""
                        SELECT 
                            pg_size_pretty(pg_total_relation_size('public.{tablename}')),
                            COUNT(*)
                        FROM public.{tablename}
                    """)
                    size, count = cur.fetchone()
                    print(f"{tablename:<30} {size:>10} ({count:>8} rows)")
                except:
                    print(f"{tablename:<30} {'N/A':>10} ({'N/A':>8} rows)")
            
            # Storage by category
            print("\n\nStorage by Category:")
            print("-" * 50)
            
            # WhatsApp session storage
            cur.execute("""
                SELECT SUM(pg_total_relation_size('public.'||tablename))
                FROM pg_tables
                WHERE schemaname = 'public' AND tablename LIKE 'whatsmeow_%'
            """)
            whatsmeow_size = cur.fetchone()[0] or 0
            
            # WhatsApp chat/message storage
            cur.execute("""
                SELECT SUM(pg_total_relation_size('public.'||tablename))
                FROM pg_tables
                WHERE schemaname = 'public' 
                AND tablename IN ('whatsapp_messages', 'whatsapp_chats')
            """)
            whatsapp_data_size = cur.fetchone()[0] or 0
            
            # Application data
            cur.execute("""
                SELECT SUM(pg_total_relation_size('public.'||tablename))
                FROM pg_tables
                WHERE schemaname = 'public' 
                AND tablename NOT LIKE 'whatsmeow_%'
                AND tablename NOT IN ('whatsapp_messages', 'whatsapp_chats')
            """)
            app_data_size = cur.fetchone()[0] or 0
            
            print(f"WhatsApp Session Data (whatsmeow_*): {whatsmeow_size / (1024*1024):.1f} MB ({whatsmeow_size / (1024*1024) / 121 * 100:.1f}%)")
            print(f"WhatsApp Chat/Message Data: {whatsapp_data_size / (1024*1024):.1f} MB ({whatsapp_data_size / (1024*1024) / 121 * 100:.1f}%)")
            print(f"Application Data: {app_data_size / (1024*1024):.1f} MB ({app_data_size / (1024*1024) / 121 * 100:.1f}%)")
            
            # Recommendations
            print("\n\nRecommendations for Further Space Reduction:")
            print("-" * 70)
            
            # Check message_secrets
            cur.execute("SELECT COUNT(*) FROM public.whatsmeow_message_secrets")
            secrets_count = cur.fetchone()[0]
            print(f"1. whatsmeow_message_secrets has {secrets_count:,} records (37 MB)")
            print("   - This stores encryption keys for messages")
            print("   - Can be cleaned for old/inactive devices")
            
            # Check old messages
            cur.execute("""
                SELECT COUNT(*) 
                FROM public.whatsapp_messages 
                WHERE timestamp < NOW() - INTERVAL '30 days'
            """)
            old_msgs = cur.fetchone()[0]
            print(f"\n2. whatsapp_messages has {old_msgs:,} messages older than 30 days")
            print("   - Consider archiving or deleting old messages")
            
            # Check contacts
            cur.execute("SELECT COUNT(*) FROM public.whatsmeow_contacts")
            contacts = cur.fetchone()[0]
            print(f"\n3. whatsmeow_contacts has {contacts:,} contacts (18 MB)")
            print("   - May contain duplicate or old contacts")
            
            print("\n\nSQL Commands to Free More Space:")
            print("-" * 70)
            print("-- Delete old WhatsApp messages (older than 30 days)")
            print("DELETE FROM whatsapp_messages WHERE timestamp < NOW() - INTERVAL '30 days';")
            print("\n-- Clean up message secrets for inactive devices")
            print("DELETE FROM whatsmeow_message_secrets WHERE jid NOT IN (SELECT jid FROM user_devices WHERE status = 'online');")
            print("\n-- Remove old session data")
            print("DELETE FROM whatsmeow_sessions WHERE last_seen < NOW() - INTERVAL '7 days';")
            print("\n-- Run VACUUM FULL after cleanup")
            print("VACUUM FULL;")
            
        conn.close()
        
    except Exception as e:
        print(f"[ERROR] {e}")

if __name__ == "__main__":
    analyze_postgresql_storage()
