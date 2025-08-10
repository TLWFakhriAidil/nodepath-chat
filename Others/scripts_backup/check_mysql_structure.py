import mysql.connector

# Connect to MySQL
conn = mysql.connector.connect(
    host='159.89.198.71',
    port=3306,
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway'
)

cursor = conn.cursor()

# Check tables
cursor.execute("SHOW TABLES")
tables = cursor.fetchall()
print("Tables in database:")
for table in tables:
    print(f"  - {table[0]}")

# Check specific important tables
important_tables = ['users', 'user_devices', 'leads', 'campaigns', 'sequences', 'broadcast_messages']
for table in important_tables:
    cursor.execute(f"DESCRIBE {table}")
    print(f"\n{table} structure:")
    for col in cursor.fetchall():
        print(f"  {col[0]} - {col[1]}")

cursor.close()
conn.close()
