import mysql.connector

# MySQL connection
conn = mysql.connector.connect(
    host='159.89.198.71',
    port=3306,
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway'
)

cursor = conn.cursor()

# Check users
print("Users in MySQL database:")
cursor.execute("SELECT id, email, full_name, is_active FROM users")
users = cursor.fetchall()
for user in users:
    print(f"  - {user[1]} ({user[2]}) - Active: {user[3]}")

# Check sessions
print("\nActive sessions:")
cursor.execute("""
    SELECT u.email, s.token, s.expires_at 
    FROM user_sessions s 
    JOIN users u ON s.user_id = u.id 
    WHERE s.expires_at > NOW()
""")
sessions = cursor.fetchall()
for session in sessions:
    print(f"  - {session[0]} - Expires: {session[2]}")

# Check devices
print("\nUser devices:")
cursor.execute("""
    SELECT u.email, d.device_name, d.status, d.jid 
    FROM user_devices d 
    JOIN users u ON d.user_id = u.id 
    ORDER BY u.email, d.device_name
    LIMIT 10
""")
devices = cursor.fetchall()
for device in devices:
    print(f"  - {device[0]} -> {device[1]} ({device[2]})")

cursor.close()
conn.close()
