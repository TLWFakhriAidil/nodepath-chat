import psycopg2
import sys
import io
sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

# Connect to database
conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
cursor = conn.cursor()

phone = '60138847465'

print("="*100)
print(f"ANALYZING MESSAGES FOR PHONE: {phone}")
print("="*100)

# Get lead information first
cursor.execute("""
    SELECT 
        l.id,
        l.name,
        l.phone,
        l.niche,
        l.trigger,
        l.status,
        l.device_id,
        l.created_at
    FROM leads l
    WHERE l.phone = %s
""", (phone,))

lead = cursor.fetchone()
if lead:
    print("\n📱 LEAD INFORMATION:")
    print(f"  - Name: {lead[1]}")
    print(f"  - Phone: {lead[2]}")
    print(f"  - Niche: {lead[3]}")
    print(f"  - Trigger: {lead[4]}")
    print(f"  - Status: {lead[5]}")
    print(f"  - Created: {lead[7]}")
else:
    print("\n❌ No lead found with this phone number")

# Get all messages with detailed sequence step info
cursor.execute("""
    SELECT 
        bm.id,
        bm.recipient_phone,
        bm.recipient_name,
        bm.status,
        bm.scheduled_at,
        bm.sent_at,
        bm.content,
        bm.sequence_id,
        bm.campaign_id,
        bm.sequence_stepid,
        s.name as sequence_name,
        s.trigger as sequence_trigger,
        c.title as campaign_title,
        ss.day_number,
        ss.trigger as step_trigger,
        ss.is_entry_point,
        ss.next_trigger
    FROM broadcast_messages bm
    LEFT JOIN sequences s ON s.id = bm.sequence_id
    LEFT JOIN campaigns c ON c.id = bm.campaign_id
    LEFT JOIN sequence_steps ss ON ss.id = bm.sequence_stepid
    WHERE bm.recipient_phone = %s
    ORDER BY bm.scheduled_at ASC
""", (phone,))

messages = cursor.fetchall()
print(f"\n📨 TOTAL MESSAGES: {len(messages)}")
print("="*100)

# Group by sequence/campaign
sequences_used = {}
campaigns_used = {}

for msg in messages:
    (msg_id, phone_num, name, status, scheduled, sent, content, seq_id, camp_id, 
     step_id, seq_name, seq_trigger, camp_title, day_num, step_trigger, is_entry, next_trigger) = msg
    
    if seq_id:
        if seq_id not in sequences_used:
            sequences_used[seq_id] = {
                'name': seq_name,
                'trigger': seq_trigger,
                'messages': []
            }
        sequences_used[seq_id]['messages'].append(msg)
    elif camp_id:
        if camp_id not in campaigns_used:
            campaigns_used[camp_id] = {
                'title': camp_title,
                'messages': []
            }
        campaigns_used[camp_id]['messages'].append(msg)

# Display sequence details
if sequences_used:
    print("\n🔄 SEQUENCES USED:")
    print("-"*100)
    
    for seq_id, seq_data in sequences_used.items():
        print(f"\n📋 Sequence: {seq_data['name']}")
        print(f"   Trigger: {seq_data['trigger']}")
        print(f"   Messages: {len(seq_data['messages'])}")
        print("\n   Message Flow:")
        
        for i, msg in enumerate(seq_data['messages'], 1):
            (msg_id, phone_num, name, status, scheduled, sent, content, seq_id, camp_id, 
             step_id, seq_name, seq_trigger, camp_title, day_num, step_trigger, is_entry, next_trigger) = msg
            
            print(f"\n   {i}. Day {day_num if day_num else '?'} - Status: {status}")
            print(f"      Scheduled: {scheduled}")
            if sent:
                print(f"      Sent: {sent}")
            print(f"      Step Trigger: {step_trigger}")
            print(f"      Entry Point: {'YES' if is_entry else 'NO'}")
            if next_trigger:
                print(f"      Next Trigger: {next_trigger}")
            if content:
                # Show first 150 chars of content
                content_preview = content[:150].replace('\n', ' ')
                print(f"      Message: {content_preview}...")

# Display campaign details
if campaigns_used:
    print("\n\n📢 CAMPAIGNS USED:")
    print("-"*100)
    
    for camp_id, camp_data in campaigns_used.items():
        print(f"\n📣 Campaign: {camp_data['title']}")
        print(f"   Messages: {len(camp_data['messages'])}")

# Summary statistics
print("\n\n📊 SUMMARY:")
print("-"*100)
total_sent = sum(1 for msg in messages if msg[3] == 'sent')
total_pending = sum(1 for msg in messages if msg[3] == 'pending')
total_failed = sum(1 for msg in messages if msg[3] == 'failed')

print(f"  - Total Messages: {len(messages)}")
print(f"  - Sent: {total_sent}")
print(f"  - Pending: {total_pending}")
print(f"  - Failed: {total_failed}")
print(f"  - Sequences Used: {len(sequences_used)}")
print(f"  - Campaigns Used: {len(campaigns_used)}")

# Check for any issues
print("\n\n⚠️  POTENTIAL ISSUES:")
print("-"*100)

# Check for VITAC niche with non-VITAC sequences
if lead and lead[3] == 'VITAC':
    non_vitac_sequences = [seq_data['name'] for seq_data in sequences_used.values() if 'VITAC' not in seq_data['name']]
    if non_vitac_sequences:
        print(f"  ❌ VITAC lead enrolled in non-VITAC sequences: {', '.join(non_vitac_sequences)}")
    else:
        print(f"  ✅ All sequences match VITAC niche")
else:
    print(f"  ℹ️  Lead niche is {lead[3] if lead else 'unknown'}")

conn.close()
print("\n" + "="*100)
print("Analysis complete!")
