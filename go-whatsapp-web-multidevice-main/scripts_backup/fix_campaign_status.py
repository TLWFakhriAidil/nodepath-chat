import psycopg2
import sys

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

# Fix the campaign to target 'new' status or 'all'
print("Updating campaign to target 'new' status...")
cur.execute("""
    UPDATE campaigns 
    SET target_status = 'new',
        updated_at = NOW()
    WHERE id = 68
    AND status = 'pending'
""")

if cur.rowcount > 0:
    print(f"✅ Updated campaign 68 to target 'new' status")
    conn.commit()
else:
    print("❌ No campaign updated")

# Verify the update
cur.execute("""
    SELECT id, title, target_status, niche 
    FROM campaigns 
    WHERE id = 68
""")

campaign = cur.fetchone()
if campaign:
    print(f"\nCampaign {campaign[0]}: {campaign[1]}")
    print(f"Now targeting: {campaign[2]} status in {campaign[3]} niche")

cur.close()
conn.close()
