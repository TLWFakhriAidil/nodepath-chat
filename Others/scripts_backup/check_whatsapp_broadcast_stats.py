import psycopg2

# Connect
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

# Get stats for all WhatsApp Web devices
cursor.execute("""
SELECT 
    ud.device_name,
    COUNT(bm.id) as total,
    SUM(CASE WHEN bm.status = 'sent' THEN 1 ELSE 0 END) as sent,
    SUM(CASE WHEN bm.status = 'failed' THEN 1 ELSE 0 END) as failed,
    SUM(CASE WHEN bm.status = 'pending' THEN 1 ELSE 0 END) as pending,
    SUM(CASE WHEN bm.status = 'processing' THEN 1 ELSE 0 END) as processing,
    SUM(CASE WHEN bm.status = 'queued' THEN 1 ELSE 0 END) as queued,
    SUM(CASE WHEN bm.campaign_id IS NOT NULL THEN 1 ELSE 0 END) as campaigns,
    SUM(CASE WHEN bm.sequence_id IS NOT NULL THEN 1 ELSE 0 END) as sequences
FROM user_devices ud
LEFT JOIN broadcast_messages bm ON ud.id = bm.device_id
WHERE (ud.platform IS NULL OR ud.platform = '')
GROUP BY ud.device_name
ORDER BY ud.device_name;
""")

results = cursor.fetchall()

print("\n=== BROADCAST MESSAGE STATS FOR ALL WHATSAPP WEB DEVICES ===\n")
print(f"{'Device Name':<25} {'Total':<8} {'Sent':<8} {'Failed':<8} {'Pending':<8} {'Campaign':<10} {'Sequence':<10}")
print("="*90)

total_all = 0
total_sent = 0
total_failed = 0

for row in results:
    name, total, sent, failed, pending, processing, queued, campaigns, sequences = row
    
    if total > 0:
        print(f"{name[:24]:<25} {total:<8} {sent:<8} {failed:<8} {pending:<8} {campaigns:<10} {sequences:<10}")
        total_all += total
        total_sent += sent
        total_failed += failed
    else:
        print(f"{name[:24]:<25} {'0':<8} {'-':<8} {'-':<8} {'-':<8} {'-':<10} {'-':<10}")

print("="*90)
print(f"{'TOTALS:':<25} {total_all:<8} {total_sent:<8} {total_failed:<8}")

if total_sent + total_failed > 0:
    success_rate = (total_sent / (total_sent + total_failed)) * 100
    print(f"\nOverall Success Rate: {success_rate:.1f}%")
else:
    print("\nNo messages have been sent or failed yet.")

cursor.close()
conn.close()
