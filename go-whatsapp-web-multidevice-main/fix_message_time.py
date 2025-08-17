import mysql.connector

config = {
    'host': '159.89.198.71',
    'port': 3306,
    'database': 'admin_railway',
    'user': 'admin_aqil',
    'password': 'admin_aqil'
}

conn = mysql.connector.connect(**config)
cursor = conn.cursor()

cursor.execute("""
UPDATE broadcast_messages 
SET scheduled_at = NOW(),
    updated_at = NOW()
WHERE id = 'fgeef10d-4ae4-4e8a-8bbd-73d2fdb6094a'
""")

conn.commit()
print("✅ Updated message scheduled time to NOW")
print("Message is now within the 10-minute window!")

cursor.close()
conn.close()