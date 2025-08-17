import psycopg2
import csv
from datetime import datetime
import sys

# Set UTF-8 encoding for console output
sys.stdout.reconfigure(encoding='utf-8')

# Database connection
conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    print("Connecting to PostgreSQL database...")
    conn = psycopg2.connect(conn_string)
    cursor = conn.cursor()
    print("Connected successfully!\n")
    
    # Query to get all leads without triggers
    print("Fetching leads without triggers...")
    cursor.execute("""
        SELECT 
            l.id,
            l.name,
            l.phone,
            l.niche,
            l.trigger,
            l.target_status,
            l.platform,
            l.status,
            l.device_id,
            l.user_id,
            l.created_at,
            l.updated_at,
            u.email as user_email,
            u.full_name as user_name,
            d.device_name,
            d.jid as device_jid,
            d.phone as device_phone
        FROM leads l
        LEFT JOIN users u ON l.user_id = u.id
        LEFT JOIN user_devices d ON l.device_id = d.id
        WHERE l.trigger IS NULL OR l.trigger = ''
        ORDER BY l.created_at DESC
    """)
    
    leads = cursor.fetchall()
    print(f"Found {len(leads)} leads without triggers\n")
    
    # Define output file
    output_file = "C:\\Users\\ROGSTRIX\\go-whatsapp-web-multidevice-main\\debug_webhook\\leads_without_triggers.csv"
    
    # Write to CSV
    with open(output_file, 'w', newline='', encoding='utf-8-sig') as csvfile:
        writer = csv.writer(csvfile)
        
        # Write header
        headers = [
            'Lead ID',
            'Name',
            'Phone',
            'Niche',
            'Trigger',
            'Target Status',
            'Platform',
            'Status',
            'Device ID',
            'User ID',
            'Created At',
            'Updated At',
            'User Email',
            'User Name',
            'Device Name',
            'Device JID',
            'Device Phone'
        ]
        writer.writerow(headers)
        
        # Write data
        for lead in leads:
            # Convert datetime objects to string
            row = list(lead)
            row[10] = str(row[10]) if row[10] else ''  # created_at
            row[11] = str(row[11]) if row[11] else ''  # updated_at
            writer.writerow(row)
    
    print(f"[SUCCESS] Successfully exported {len(leads)} leads to:")
    print(f"   {output_file}")
    
    # Also create a summary
    print("\n[SUMMARY BY NICHE]:")
    print("-" * 50)
    cursor.execute("""
        SELECT niche, COUNT(*) as count
        FROM leads
        WHERE trigger IS NULL OR trigger = ''
        GROUP BY niche
        ORDER BY count DESC
    """)
    niche_summary = cursor.fetchall()
    for niche, count in niche_summary:
        print(f"{(niche or 'No Niche'):20} | {count:6} leads")
    
    print("\n[SUMMARY BY PLATFORM]:")
    print("-" * 50)
    cursor.execute("""
        SELECT platform, COUNT(*) as count
        FROM leads
        WHERE trigger IS NULL OR trigger = ''
        GROUP BY platform
        ORDER BY count DESC
    """)
    platform_summary = cursor.fetchall()
    for platform, count in platform_summary:
        print(f"{(platform or 'No Platform'):20} | {count:6} leads")
    
    print("\n[SUMMARY BY TARGET STATUS]:")
    print("-" * 50)
    cursor.execute("""
        SELECT target_status, COUNT(*) as count
        FROM leads
        WHERE trigger IS NULL OR trigger = ''
        GROUP BY target_status
        ORDER BY count DESC
    """)
    status_summary = cursor.fetchall()
    for status, count in status_summary:
        print(f"{(status or 'No Status'):20} | {count:6} leads")
    
    print("\n[SUMMARY BY DATE]:")
    print("-" * 50)
    cursor.execute("""
        SELECT DATE(created_at) as date, COUNT(*) as count
        FROM leads
        WHERE trigger IS NULL OR trigger = ''
        GROUP BY DATE(created_at)
        ORDER BY date DESC
        LIMIT 10
    """)
    date_summary = cursor.fetchall()
    for date, count in date_summary:
        print(f"{str(date):20} | {count:6} leads")
    
    cursor.close()
    conn.close()
    
    print("\n" + "="*60)
    print("CSV file created successfully!")
    print("You can open it in Excel or any spreadsheet application")
    print("="*60)
    
except Exception as e:
    print(f"Error: {str(e)}")
    import traceback
    traceback.print_exc()
    if 'conn' in locals():
        conn.close()
