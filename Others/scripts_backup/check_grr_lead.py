import psycopg2
import sys

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

# Check the GRR lead details
print("=== CHECKING GRR LEADS ===")
cur.execute("""
    SELECT id, name, phone, status, niche, device_id, user_id, trigger
    FROM leads 
    WHERE niche = 'GRR'
    AND user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
""")

leads = cur.fetchall()
print(f"Found {len(leads)} GRR leads:\n")

for lead in leads:
    print(f"ID: {lead[0]}")
    print(f"Name: {lead[1]}")
    print(f"Phone: {lead[2]}")
    print(f"Status: {lead[3]}")  # This is the key!
    print(f"Niche: {lead[4]}")
    print(f"Device ID: {lead[5]}")
    print(f"User ID: {lead[6]}")
    print(f"Trigger: {lead[7]}")
    print("-" * 50)

# Check all possible statuses
print("\n=== ALL LEAD STATUSES FOR THIS USER ===")
cur.execute("""
    SELECT status, COUNT(*) 
    FROM leads 
    WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
    GROUP BY status
    ORDER BY COUNT(*) DESC
""")

for row in cur.fetchall():
    print(f"Status '{row[0]}': {row[1]} leads")

# Check devices
print("\n=== USER DEVICES ===")
cur.execute("""
    SELECT id, device_name, status, platform, phone
    FROM user_devices 
    WHERE user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
    ORDER BY status
""")

for device in cur.fetchall():
    print(f"ID: {device[0]}")
    print(f"Name: {device[1]}")
    print(f"Status: {device[2]}")
    print(f"Platform: {device[3]}")
    print(f"Phone: {device[4]}")
    print("-" * 30)

cur.close()
conn.close()
