import psycopg2
import mysql.connector
from datetime import datetime

# PostgreSQL connection (Railway)
pg_conn = psycopg2.connect(
    "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"
)

# MySQL connection
mysql_conn = mysql.connector.connect(
    host='159.89.198.71',
    port=3306,
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway'
)

pg_cursor = pg_conn.cursor()
mysql_cursor = mysql_conn.cursor()

print("Connected to both databases")

# 1. First, let's check what tables exist in PostgreSQL
pg_cursor.execute("""
    SELECT table_name 
    FROM information_schema.tables 
    WHERE table_schema = 'public' 
    AND table_type = 'BASE TABLE'
""")
pg_tables = pg_cursor.fetchall()
print("\nPostgreSQL tables:")
for table in pg_tables:
    print(f"  - {table[0]}")

# 2. Check if users table exists in PostgreSQL
pg_cursor.execute("""
    SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'users'
    )
""")
users_exists = pg_cursor.fetchone()[0]

if users_exists:
    # Get users data
    pg_cursor.execute("""
        SELECT id, email, full_name, password_hash, is_active, 
               created_at, updated_at, last_login 
        FROM users
    """)
    users = pg_cursor.fetchall()
    print(f"\nFound {len(users)} users in PostgreSQL")
    
    # Create users table in MySQL if not exists
    mysql_cursor.execute("""
        CREATE TABLE IF NOT EXISTS users (
            id VARCHAR(36) PRIMARY KEY,
            email VARCHAR(255) UNIQUE NOT NULL,
            full_name VARCHAR(255) NOT NULL,
            password_hash VARCHAR(255) NOT NULL,
            is_active BOOLEAN DEFAULT true,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
            last_login TIMESTAMP NULL
        )
    """)
    
    # Insert users into MySQL
    for user in users:
        try:
            mysql_cursor.execute("""
                INSERT INTO users (id, email, full_name, password_hash, is_active, 
                                 created_at, updated_at, last_login)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                ON DUPLICATE KEY UPDATE
                    full_name = VALUES(full_name),
                    password_hash = VALUES(password_hash),
                    is_active = VALUES(is_active),
                    updated_at = VALUES(updated_at),
                    last_login = VALUES(last_login)
            """, user)
            print(f"  + Migrated user: {user[1]}")
        except Exception as e:
            print(f"  - Error migrating user {user[1]}: {e}")

# 3. Check if user_sessions table exists
pg_cursor.execute("""
    SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'user_sessions'
    )
""")
sessions_exists = pg_cursor.fetchone()[0]

if sessions_exists:
    # Get user_sessions data
    pg_cursor.execute("""
        SELECT id, user_id, token, expires_at, created_at 
        FROM user_sessions
        WHERE expires_at > NOW()
    """)
    sessions = pg_cursor.fetchall()
    print(f"\nFound {len(sessions)} active sessions in PostgreSQL")
    
    # Create user_sessions table in MySQL if not exists
    mysql_cursor.execute("""
        CREATE TABLE IF NOT EXISTS user_sessions (
            id VARCHAR(36) PRIMARY KEY,
            user_id VARCHAR(36) NOT NULL,
            token VARCHAR(255) UNIQUE NOT NULL,
            expires_at TIMESTAMP NOT NULL,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        )
    """)
    
    # Insert sessions into MySQL
    for session in sessions:
        try:
            mysql_cursor.execute("""
                INSERT INTO user_sessions (id, user_id, token, expires_at, created_at)
                VALUES (%s, %s, %s, %s, %s)
                ON DUPLICATE KEY UPDATE
                    expires_at = VALUES(expires_at)
            """, session)
            print(f"  + Migrated session for user_id: {session[1]}")
        except Exception as e:
            print(f"  - Error migrating session: {e}")

# 4. Check if user_devices table exists
pg_cursor.execute("""
    SELECT EXISTS (
        SELECT FROM information_schema.tables 
        WHERE table_schema = 'public' 
        AND table_name = 'user_devices'
    )
""")
devices_exists = pg_cursor.fetchone()[0]

if devices_exists:
    # Get user_devices data
    pg_cursor.execute("""
        SELECT id, user_id, device_name, phone, jid, status, 
               last_seen, created_at, updated_at 
        FROM user_devices
    """)
    devices = pg_cursor.fetchall()
    print(f"\nFound {len(devices)} devices in PostgreSQL")
    
    # Create user_devices table in MySQL if not exists
    mysql_cursor.execute("""
        CREATE TABLE IF NOT EXISTS user_devices (
            id VARCHAR(36) PRIMARY KEY,
            user_id VARCHAR(36) NOT NULL,
            device_name VARCHAR(255) NOT NULL,
            phone VARCHAR(50),
            jid VARCHAR(255),
            status VARCHAR(50) DEFAULT 'offline',
            last_seen TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
            updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
            UNIQUE KEY unique_user_jid (user_id, jid),
            FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
        )
    """)
    
    # Insert devices into MySQL
    for device in devices:
        try:
            mysql_cursor.execute("""
                INSERT INTO user_devices (id, user_id, device_name, phone, jid, 
                                        status, last_seen, created_at, updated_at)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON DUPLICATE KEY UPDATE
                    device_name = VALUES(device_name),
                    phone = VALUES(phone),
                    status = VALUES(status),
                    last_seen = VALUES(last_seen),
                    updated_at = VALUES(updated_at)
            """, device)
            print(f"  + Migrated device: {device[2]}")
        except Exception as e:
            print(f"  - Error migrating device: {e}")

# Commit changes
mysql_conn.commit()
print("\n[SUCCESS] Migration completed!")

# Show summary
mysql_cursor.execute("SELECT COUNT(*) FROM users")
user_count = mysql_cursor.fetchone()[0]
print(f"\nMySQL now has {user_count} users")

# Close connections
pg_cursor.close()
pg_conn.close()
mysql_cursor.close()
mysql_conn.close()
