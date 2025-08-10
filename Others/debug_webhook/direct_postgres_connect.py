import psycopg2
import json
from datetime import datetime
import sys

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

# Database connection
conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    print("Connecting to PostgreSQL database...")
    conn = psycopg2.connect(conn_string)
    cursor = conn.cursor()
    print("[SUCCESS] Connected successfully!\n")
    
    # Get database version
    cursor.execute("SELECT version();")
    version = cursor.fetchone()[0]
    print(f"PostgreSQL Version: {version.split(',')[0]}\n")
    
    # List all tables
    print("[DATABASE TABLES]:")
    print("-" * 80)
    cursor.execute("""
        SELECT table_name, 
               pg_size_pretty(pg_total_relation_size(quote_ident(table_name)::regclass)) as size
        FROM information_schema.tables 
        WHERE table_schema = 'public' 
        ORDER BY pg_total_relation_size(quote_ident(table_name)::regclass) DESC
    """)
    tables = cursor.fetchall()
    for table, size in tables:
        print(f"{table:30} | Size: {size}")
    
    # Get row counts for main tables
    print("\n[TABLE ROW COUNTS]:")
    print("-" * 80)
    main_tables = ['users', 'user_devices', 'leads', 'campaigns', 'sequences', 
                   'broadcast_messages', 'sequence_steps', 'message_analytics']
    
    for table in main_tables:
        try:
            cursor.execute(f"SELECT COUNT(*) FROM {table}")
            count = cursor.fetchone()[0]
            print(f"{table:30} | {count:,} rows")
        except:
            pass
    
    # Check active devices
    print("\n[DEVICE STATUS]:")
    print("-" * 80)
    cursor.execute("""
        SELECT status, COUNT(*) as count 
        FROM user_devices 
        GROUP BY status 
        ORDER BY count DESC
    """)
    statuses = cursor.fetchall()
    for status, count in statuses:
        print(f"Status: {status or 'NULL':15} | {count:5} devices")
    
    # Check recent activity
    print("\n[RECENT ACTIVITY (Last 24 hours)]:")
    print("-" * 80)
    cursor.execute("""
        SELECT 
            COUNT(*) as messages_sent
        FROM broadcast_messages 
        WHERE sent_at > NOW() - INTERVAL '24 hours'
    """)
    messages_sent = cursor.fetchone()[0]
    print(f"Messages sent: {messages_sent:,}")
    
    cursor.execute("""
        SELECT 
            COUNT(*) as leads_created
        FROM leads 
        WHERE created_at > NOW() - INTERVAL '24 hours'
    """)
    leads_created = cursor.fetchone()[0]
    print(f"Leads created: {leads_created:,}")
    
    # Platform distribution
    print("\n[PLATFORM DISTRIBUTION]:")
    print("-" * 80)
    cursor.execute("""
        SELECT platform, COUNT(*) as count 
        FROM leads 
        WHERE platform IS NOT NULL 
        GROUP BY platform 
        ORDER BY count DESC
    """)
    platforms = cursor.fetchall()
    for platform, count in platforms:
        print(f"{platform:20} | {count:,} leads")
    
    # Active sequences
    print("\n[ACTIVE SEQUENCES]:")
    print("-" * 80)
    cursor.execute("""
        SELECT name, is_active, total_contacts, active_contacts 
        FROM sequences 
        WHERE is_active = true 
        ORDER BY active_contacts DESC 
        LIMIT 10
    """)
    sequences = cursor.fetchall()
    if sequences:
        for name, active, total, active_contacts in sequences:
            print(f"{name:30} | Total: {total or 0:5} | Active: {active_contacts or 0:5}")
    else:
        print("No active sequences found")
    
    # Check users
    print("\n[USERS]:")
    print("-" * 80)
    cursor.execute("""
        SELECT id, email, full_name, is_active, last_login
        FROM users
        ORDER BY last_login DESC NULLS LAST
        LIMIT 5
    """)
    users = cursor.fetchall()
    for user_id, email, name, active, last_login in users:
        print(f"User: {email:30} | Name: {name:20} | Active: {active}")
    
    # Recent campaigns
    print("\n[RECENT CAMPAIGNS]:")
    print("-" * 80)
    cursor.execute("""
        SELECT title, status, campaign_date, created_at
        FROM campaigns
        ORDER BY created_at DESC
        LIMIT 5
    """)
    campaigns = cursor.fetchall()
    if campaigns:
        for title, status, date, created in campaigns:
            print(f"{title:30} | Status: {status:10} | Date: {date}")
    else:
        print("No campaigns found")
    
    # Database size
    print("\n[DATABASE SIZE]:")
    print("-" * 80)
    cursor.execute("""
        SELECT pg_database_size(current_database()) as size,
               pg_size_pretty(pg_database_size(current_database())) as pretty_size
    """)
    db_size = cursor.fetchone()
    print(f"Total database size: {db_size[1]}")
    
    # Connection info
    print("\n[CONNECTION INFO]:")
    print("-" * 80)
    print(f"Host: yamanote.proxy.rlwy.net")
    print(f"Port: 49914")
    print(f"Database: railway")
    print(f"User: postgres")
    
    cursor.close()
    conn.close()
    print("\n[SUCCESS] Connection closed successfully!")
    
except Exception as e:
    print(f"[ERROR] {str(e)}")
    import traceback
    traceback.print_exc()
