import pymysql

conn = pymysql.connect(
    host='159.89.198.71',
    user='admin_aqil',
    password='admin_aqil',
    database='admin_railway',
    port=3306
)
cursor = conn.cursor()

# Find sequences with SCHQ in name
cursor.execute("SELECT DISTINCT name FROM sequences WHERE name LIKE '%SCHQ%' LIMIT 10")
results = cursor.fetchall()
print("Sequences with SCHQ:")
for r in results:
    print(f"  - {r[0]}")

# Check the issue with date filtering
print("\n" + "="*80)
print("MAIN ISSUE IDENTIFIED:")
print("="*80)
print("The success/failed modal is showing ALL messages ever sent for that step,")
print("NOT respecting the date filter you selected (August 7).")
print("\nWhen you:")
print("1. Filter for August 7 in the dashboard")
print("2. See Step 1: Total 2, Success 2")
print("3. Click on the '2' success count")
print("\nThe modal SHOULD show only 2 messages sent on August 7")
print("But INSTEAD it shows 14 messages from August 2!")
print("\nThis is a BUG in the modal query - it's not using the date filter.")

conn.close()
