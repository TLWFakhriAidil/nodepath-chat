import psycopg2
import sys

sys.stdout.reconfigure(encoding='utf-8')

conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cur = conn.cursor()

# Reset campaign 69 to pending to test again
print("=== RESETTING CAMPAIGN 69 FOR TESTING ===")

# Delete any broadcast messages
cur.execute("DELETE FROM broadcast_messages WHERE campaign_id = 69")
deleted = cur.rowcount
print(f"Deleted {deleted} broadcast messages")

# Reset campaign to pending
cur.execute("""
    UPDATE campaigns 
    SET status = 'pending', 
        target_status = 'new',  -- Match the actual lead status
        updated_at = NOW()
    WHERE id = 69
""")
print("Reset campaign to pending with target_status='new'")

# Verify the lead exists with correct data
cur.execute("""
    SELECT l.phone, l.name, l.status, l.niche, l.device_id, l.user_id,
           ud.status as device_status
    FROM leads l
    LEFT JOIN user_devices ud ON ud.id = l.device_id
    WHERE l.user_id = 'de078f16-3266-4ab3-8153-a248b015228f'
    AND l.niche = 'GRR'
""")

leads = cur.fetchall()
print(f"\nFound {len(leads)} leads with niche 'GRR':")
for lead in leads:
    print(f"  Phone: {lead[0]}")
    print(f"  Name: {lead[1]}")
    print(f"  Status: {lead[2]}")
    print(f"  Niche: {lead[3]}")
    print(f"  Device ID: {lead[4]}")
    print(f"  Device Status: {lead[6]}")
    print()

conn.commit()
cur.close()
conn.close()

print("Campaign 69 is now ready to be triggered again.")
print("The campaign trigger processor should pick it up within 1 minute.")
