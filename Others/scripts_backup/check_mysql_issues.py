#!/usr/bin/env python3
"""
Check MySQL database for issues and test queries
"""

import pymysql
import sys

def check_mysql_issues():
    """Check MySQL for the reported issues"""
    
    try:
        # Connect to MySQL
        conn = pymysql.connect(
            host='159.89.198.71',
            port=3306,
            user='admin_aqil',
            password='admin_aqil',
            database='admin_railway',
            charset='utf8mb4'
        )
        print("[OK] Connected to MySQL")
        
        cur = conn.cursor()
        
        # Test 1: Check leads table structure
        print("\n1. Checking leads table structure...")
        cur.execute("DESCRIBE leads")
        columns = cur.fetchall()
        lead_columns = [col[0] for col in columns]
        print(f"Leads columns: {', '.join(lead_columns)}")
        
        if 'trigger' in lead_columns:
            print("[OK] 'trigger' column exists in leads")
        else:
            print("[ERROR] 'trigger' column missing in leads")
        
        # Test 2: Test problematic query from error log
        print("\n2. Testing lead selection query...")
        test_query = """
            SELECT id, device_id, user_id, name, phone, niche, journey, status, 
                   target_status, `trigger`, created_at, updated_at
            FROM leads
            WHERE user_id = %s AND device_id = %s
            LIMIT 5
        """
        
        try:
            # Use dummy values for testing
            cur.execute(test_query, ('test-user-id', 'test-device-id'))
            results = cur.fetchall()
            print(f"[OK] Query executed successfully. Found {len(results)} results")
        except Exception as e:
            print(f"[ERROR] Query failed: {e}")
        
        # Test 3: Check for campaigns table issues
        print("\n3. Checking campaigns table...")
        cur.execute("DESCRIBE campaigns")
        campaign_cols = [col[0] for col in cur.fetchall()]
        print(f"Campaign columns: {', '.join(campaign_cols[:5])}...")
        
        # Test 4: Check sequences table
        print("\n4. Checking sequences table...")
        cur.execute("DESCRIBE sequences")
        sequence_cols = [col[0] for col in cur.fetchall()]
        print(f"Sequence columns: {', '.join(sequence_cols[:5])}...")
        
        # Test 5: Check broadcast_messages
        print("\n5. Checking broadcast_messages table...")
        cur.execute("SELECT COUNT(*) FROM broadcast_messages WHERE status = 'pending'")
        pending = cur.fetchone()[0]
        print(f"Pending messages: {pending}")
        
        # Test 6: Test analytics query with date range
        print("\n6. Testing analytics query...")
        analytics_query = """
            SELECT COUNT(DISTINCT c.id) 
            FROM campaigns c 
            WHERE c.created_at BETWEEN %s AND %s
        """
        
        try:
            cur.execute(analytics_query, ('2025-01-01', '2025-12-31'))
            count = cur.fetchone()[0]
            print(f"[OK] Analytics query works. Campaigns: {count}")
        except Exception as e:
            print(f"[ERROR] Analytics query failed: {e}")
        
        # Test 7: Check for sequence trigger issues
        print("\n7. Testing sequence trigger query...")
        trigger_query = """
            SELECT COUNT(*) 
            FROM sequence_steps 
            WHERE `trigger` IS NOT NULL
        """
        
        try:
            cur.execute(trigger_query)
            count = cur.fetchone()[0]
            print(f"[OK] Sequence steps with triggers: {count}")
        except Exception as e:
            print(f"[ERROR] Trigger query failed: {e}")
            
        # Test 8: Test time interval queries
        print("\n8. Testing time interval queries...")
        interval_query = """
            SELECT COUNT(*) 
            FROM broadcast_messages 
            WHERE created_at < DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 12 HOUR)
        """
        
        try:
            cur.execute(interval_query)
            count = cur.fetchone()[0]
            print(f"[OK] Messages older than 12 hours: {count}")
        except Exception as e:
            print(f"[ERROR] Interval query failed: {e}")
        
        conn.close()
        print("\n[OK] All checks completed")
        
    except Exception as e:
        print(f"\n[ERROR] Connection failed: {e}")
        sys.exit(1)

if __name__ == "__main__":
    check_mysql_issues()
