import psycopg2
from datetime import datetime, date

# Connect
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

# Get today's date
today = date.today()
print(f"\n=== SEQUENCE STEP REPORT FOR TODAY ({today}) ===\n")

# Get sequence step statistics for today
cursor.execute("""
SELECT 
    ss.id as step_id,
    s.name as sequence_name,
    ss.day_number,
    ss.trigger,
    COUNT(bm.id) as total_messages,
    COUNT(DISTINCT bm.device_id) as devices_involved,
    SUM(CASE WHEN bm.status = 'pending' THEN 1 ELSE 0 END) as pending,
    SUM(CASE WHEN bm.status = 'sent' THEN 1 ELSE 0 END) as sent,
    SUM(CASE WHEN bm.status = 'failed' THEN 1 ELSE 0 END) as failed,
    SUM(CASE WHEN bm.status = 'processing' THEN 1 ELSE 0 END) as processing,
    SUM(CASE WHEN bm.status = 'queued' THEN 1 ELSE 0 END) as queued
FROM broadcast_messages bm
JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
JOIN sequences s ON ss.sequence_id = s.id
WHERE DATE(bm.created_at) = CURRENT_DATE
GROUP BY ss.id, s.name, ss.day_number, ss.trigger
ORDER BY s.name, ss.day_number;
""")

results = cursor.fetchall()

if not results:
    print("No sequence messages found for today.")
else:
    current_sequence = None
    sequence_totals = {}
    
    for row in results:
        step_id, seq_name, day_num, trigger, total, devices, pending, sent, failed, processing, queued = row
        
        # Print sequence header if new sequence
        if seq_name != current_sequence:
            if current_sequence:
                print()  # Add spacing between sequences
            current_sequence = seq_name
            print(f"\n[SEQUENCE: {seq_name}]")
            print("=" * 80)
            
            # Initialize totals for this sequence
            sequence_totals[seq_name] = {
                'total': 0, 'devices': set(), 'pending': 0, 
                'sent': 0, 'failed': 0, 'processing': 0, 'queued': 0
            }
        
        # Print step details
        print(f"\n  Step Day {day_num} (Trigger: {trigger or 'None'})")
        print(f"  Step ID: {step_id}")
        print(f"  Total Messages: {total}")
        print(f"  Devices Involved: {devices}")
        print(f"  Status Breakdown:")
        print(f"    - Sent: {sent}")
        print(f"    - Failed: {failed}")
        print(f"    - Pending: {pending}")
        if processing > 0:
            print(f"    - Processing: {processing}")
        if queued > 0:
            print(f"    - Queued: {queued}")
        
        if total > 0:
            success_rate = (sent / total) * 100
            print(f"  Success Rate: {success_rate:.1f}%")
        
        # Update sequence totals
        sequence_totals[seq_name]['total'] += total
        sequence_totals[seq_name]['pending'] += pending
        sequence_totals[seq_name]['sent'] += sent
        sequence_totals[seq_name]['failed'] += failed
        sequence_totals[seq_name]['processing'] += processing
        sequence_totals[seq_name]['queued'] += queued

# Get device breakdown for today's sequence messages
cursor.execute("""
SELECT 
    ud.device_name,
    ud.platform,
    COUNT(DISTINCT ss.sequence_id) as sequences_handled,
    COUNT(DISTINCT bm.sequence_stepid) as unique_steps,
    COUNT(bm.id) as total_messages,
    SUM(CASE WHEN bm.status = 'sent' THEN 1 ELSE 0 END) as sent,
    SUM(CASE WHEN bm.status = 'failed' THEN 1 ELSE 0 END) as failed,
    SUM(CASE WHEN bm.status = 'pending' THEN 1 ELSE 0 END) as pending
FROM broadcast_messages bm
JOIN user_devices ud ON bm.device_id = ud.id
JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
WHERE DATE(bm.created_at) = CURRENT_DATE
GROUP BY ud.device_name, ud.platform
ORDER BY total_messages DESC;
""")

device_results = cursor.fetchall()

print("\n\n=== DEVICES INVOLVED IN TODAY'S SEQUENCES ===")
print(f"\n{'Device Name':<25} {'Platform':<12} {'Sequences':<10} {'Steps':<8} {'Total':<8} {'Sent':<8} {'Failed':<8} {'Pending':<8}")
print("=" * 95)

total_devices = 0
grand_total = 0
grand_sent = 0
grand_failed = 0
grand_pending = 0

for device in device_results:
    name, platform, seqs, steps, total, sent, failed, pending = device
    platform_str = platform or "WhatsApp"
    print(f"{name[:24]:<25} {platform_str[:11]:<12} {seqs:<10} {steps:<8} {total:<8} {sent:<8} {failed:<8} {pending:<8}")
    
    total_devices += 1
    grand_total += total
    grand_sent += sent
    grand_failed += failed
    grand_pending += pending

print("=" * 95)
print(f"{'TOTAL (' + str(total_devices) + ' devices)':<25} {'':<12} {'':<10} {'':<8} {grand_total:<8} {grand_sent:<8} {grand_failed:<8} {grand_pending:<8}")

# Overall summary
print(f"\n\n=== TODAY'S SUMMARY ===")
print(f"Total Devices Involved: {total_devices}")
print(f"Total Messages: {grand_total}")
print(f"Messages Sent: {grand_sent}")
print(f"Messages Failed: {grand_failed}")
print(f"Messages Pending: {grand_pending}")
if grand_total > 0:
    print(f"Overall Success Rate: {(grand_sent/grand_total*100):.1f}%")
    print(f"Overall Failure Rate: {(grand_failed/grand_total*100):.1f}%")
    print(f"Overall Pending Rate: {(grand_pending/grand_total*100):.1f}%")

# Sequence totals summary
if sequence_totals:
    print(f"\n\n=== SEQUENCE TOTALS FOR TODAY ===")
    for seq_name, totals in sequence_totals.items():
        print(f"\n{seq_name}:")
        print(f"  Total Messages: {totals['total']}")
        print(f"  Sent: {totals['sent']} | Failed: {totals['failed']} | Pending: {totals['pending']}")

cursor.close()
conn.close()
