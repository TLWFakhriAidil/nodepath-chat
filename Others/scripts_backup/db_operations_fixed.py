#!/usr/bin/env python3
"""
Database Operations Script for WhatsApp Multi-Device System
Handles both PostgreSQL and MySQL connections
"""

import os
import psycopg2
import pymysql
from datetime import datetime
import json
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

# Parse database URLs
def parse_db_url(url):
    """Parse database URL into connection parameters"""
    parsed = urlparse(url)
    return {
        'host': parsed.hostname,
        'port': parsed.port,
        'user': parsed.username,
        'password': parsed.password,
        'database': parsed.path.lstrip('/')
    }

class DatabaseManager:
    def __init__(self):
        self.env = load_env()
        self.mysql_conn = None
        self.pg_conn = None
        
    def connect_mysql(self):
        """Connect to MySQL database"""
        try:
            mysql_url = self.env.get('MYSQL_URI', '')
            if mysql_url:
                params = parse_db_url(mysql_url)
                self.mysql_conn = pymysql.connect(
                    host=params['host'],
                    port=params['port'] or 3306,
                    user=params['user'],
                    password=params['password'],
                    database=params['database'],
                    charset='utf8mb4'
                )
                print(f"[OK] Connected to MySQL: {params['host']}:{params['port']}/{params['database']}")
                return True
        except Exception as e:
            print(f"[ERROR] MySQL connection failed: {e}")
            return False
    
    def connect_postgresql(self):
        """Connect to PostgreSQL database"""
        try:
            # Check for DB_URI first (local), then DATABASE_URL (Railway)
            pg_url = self.env.get('DB_URI') or self.env.get('DATABASE_URL', '')
            
            if not pg_url:
                print("[ERROR] No PostgreSQL connection string found. Please set DB_URI or DATABASE_URL in .env")
                print("\nTo get PostgreSQL URL from Railway:")
                print("1. Go to your Railway project")
                print("2. Click on the PostgreSQL service")
                print("3. Go to 'Connect' tab")
                print("4. Copy the DATABASE_URL")
                print("5. Add to .env as: DB_URI=<your-postgresql-url>")
                return False
                
            # Handle Railway internal URL
            if 'railway.internal' in pg_url:
                print("[WARNING] Internal Railway URL detected. Please use the public URL instead.")
                print("Go to Railway > PostgreSQL > Connect > DATABASE_PUBLIC_URL")
                return False
                
            self.pg_conn = psycopg2.connect(pg_url)
            self.pg_conn.autocommit = True
            
            # Get connection info
            with self.pg_conn.cursor() as cur:
                cur.execute("SELECT current_database(), current_user, inet_server_addr(), inet_server_port()")
                db_info = cur.fetchone()
                print(f"[OK] Connected to PostgreSQL: {db_info[2]}:{db_info[3]}/{db_info[0]} as {db_info[1]}")
            return True
        except Exception as e:
            print(f"[ERROR] PostgreSQL connection failed: {e}")
            return False
    
    def export_mysql_schema(self):
        """Export MySQL schema to a documentation file"""
        if not self.mysql_conn:
            print("[ERROR] Not connected to MySQL")
            return
            
        schema_doc = "# MySQL Database Schema Documentation\n\n"
        schema_doc += f"Generated on: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}\n\n"
        
        try:
            with self.mysql_conn.cursor() as cursor:
                # Get all tables
                cursor.execute("SHOW TABLES")
                tables = cursor.fetchall()
                
                schema_doc += f"## Tables ({len(tables)} total)\n\n"
                
                for (table_name,) in tables:
                    schema_doc += f"### Table: `{table_name}`\n\n"
                    
                    # Get table structure
                    cursor.execute(f"SHOW CREATE TABLE `{table_name}`")
                    create_stmt = cursor.fetchone()[1]
                    
                    # Get column details
                    cursor.execute(f"DESCRIBE `{table_name}`")
                    columns = cursor.fetchall()
                    
                    schema_doc += "| Column | Type | Null | Key | Default | Extra |\n"
                    schema_doc += "|--------|------|------|-----|---------|-------|\n"
                    
                    for col in columns:
                        schema_doc += f"| {col[0]} | {col[1]} | {col[2]} | {col[3] or '-'} | {col[4] or '-'} | {col[5] or '-'} |\n"
                    
                    schema_doc += "\n"
                    
                    # Get indexes
                    cursor.execute(f"SHOW INDEX FROM `{table_name}`")
                    indexes = cursor.fetchall()
                    if indexes:
                        schema_doc += "**Indexes:**\n"
                        for idx in indexes:
                            schema_doc += f"- {idx[2]} on `{idx[4]}` {'(UNIQUE)' if not idx[1] else ''}\n"
                        schema_doc += "\n"
                    
                    # Add CREATE statement
                    schema_doc += "<details>\n<summary>CREATE Statement</summary>\n\n```sql\n"
                    schema_doc += create_stmt
                    schema_doc += "\n```\n</details>\n\n---\n\n"
            
            # Save to file
            with open('MYSQL_SCHEMA_DOCUMENTATION.md', 'w', encoding='utf-8') as f:
                f.write(schema_doc)
            
            print("[OK] MySQL schema exported to MYSQL_SCHEMA_DOCUMENTATION.md")
            
        except Exception as e:
            print(f"[ERROR] Error exporting MySQL schema: {e}")
    
    def clear_postgresql_tables(self):
        """Clear specified tables in PostgreSQL to free up disk space"""
        if not self.pg_conn:
            print("[ERROR] Not connected to PostgreSQL")
            return
            
        tables_to_clear = [
            'leads',
            'leads_ai',
            'sequences',
            'sequence_contacts',
            'broadcast_messages',
            'campaigns'
        ]
        
        try:
            with self.pg_conn.cursor() as cur:
                print("\n[CLEANUP] Clearing PostgreSQL tables to free disk space...\n")
                
                for table in tables_to_clear:
                    try:
                        # Check if table exists
                        cur.execute("""
                            SELECT EXISTS (
                                SELECT FROM information_schema.tables 
                                WHERE table_name = %s
                            )
                        """, (table,))
                        
                        if cur.fetchone()[0]:
                            # Get count before deletion
                            cur.execute(f"SELECT COUNT(*) FROM {table}")
                            count = cur.fetchone()[0]
                            
                            # Clear the table
                            cur.execute(f"TRUNCATE TABLE {table} CASCADE")
                            
                            print(f"[OK] Cleared {table}: {count:,} records removed")
                        else:
                            print(f"[WARNING] Table {table} does not exist")
                            
                    except Exception as e:
                        print(f"[ERROR] Error clearing {table}: {e}")
                
                # Run VACUUM to reclaim disk space
                print("\n[TOOLS] Running VACUUM to reclaim disk space...")
                cur.execute("VACUUM FULL")
                print("[OK] VACUUM completed - disk space reclaimed")
                
                # Show database size
                cur.execute("""
                    SELECT pg_database.datname,
                           pg_size_pretty(pg_database_size(pg_database.datname)) AS size
                    FROM pg_database
                    WHERE datname = current_database()
                """)
                db_size = cur.fetchone()
                print(f"\n[STATS] Current database size: {db_size[1]}")
                
        except Exception as e:
            print(f"[ERROR] Error during PostgreSQL cleanup: {e}")
    
    def show_postgresql_disk_usage(self):
        """Show disk usage statistics for PostgreSQL"""
        if not self.pg_conn:
            print("[ERROR] Not connected to PostgreSQL")
            return
            
        try:
            with self.pg_conn.cursor() as cur:
                print("\n[STATS] PostgreSQL Disk Usage Report\n")
                
                # Database size
                cur.execute("""
                    SELECT pg_size_pretty(pg_database_size(current_database())) as db_size
                """)
                print(f"Total database size: {cur.fetchone()[0]}")
                
                # Table sizes
                cur.execute("""
                    SELECT 
                        schemaname,
                        tablename,
                        pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS size,
                        pg_total_relation_size(schemaname||'.'||tablename) AS size_bytes
                    FROM pg_tables
                    WHERE schemaname NOT IN ('pg_catalog', 'information_schema')
                    ORDER BY size_bytes DESC
                    LIMIT 20
                """)
                
                print("\nTop 20 tables by size:")
                print("-" * 60)
                for row in cur.fetchall():
                    print(f"{row[0]}.{row[1]:<30} {row[2]:>10}")
                
        except Exception as e:
            print(f"[ERROR] Error getting disk usage: {e}")
    
    def close_connections(self):
        """Close all database connections"""
        if self.mysql_conn:
            self.mysql_conn.close()
            print("[OK] MySQL connection closed")
        if self.pg_conn:
            self.pg_conn.close()
            print("[OK] PostgreSQL connection closed")

def main():
    print("WhatsApp Multi-Device Database Operations\n")
    
    db = DatabaseManager()
    
    # Connect to databases
    print("1. Connecting to databases...\n")
    mysql_connected = db.connect_mysql()
    pg_connected = db.connect_postgresql()
    
    if mysql_connected:
        print("\n2. Exporting MySQL schema documentation...\n")
        db.export_mysql_schema()
    
    if pg_connected:
        print("\n3. PostgreSQL disk usage before cleanup:\n")
        db.show_postgresql_disk_usage()
        
        print("\n4. Clearing PostgreSQL tables...\n")
        db.clear_postgresql_tables()
        
        print("\n5. PostgreSQL disk usage after cleanup:\n")
        db.show_postgresql_disk_usage()
    
    # Close connections
    print("\n6. Closing connections...\n")
    db.close_connections()
    
    print("\n[OK] All operations completed!")

if __name__ == "__main__":
    main()
