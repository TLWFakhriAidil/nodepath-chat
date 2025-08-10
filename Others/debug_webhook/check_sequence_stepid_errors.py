import psycopg2
import json
from datetime import datetime
import sys

# Set UTF-8 encoding
sys.stdout.reconfigure(encoding='utf-8')

# Database connection
conn_string = "postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway"

try:
    print("Connecting to PostgreSQL database...")
    conn = psycopg2.connect(conn_string)
    cursor = conn.cursor()
    print("[SUCCESS] Connected successfully!\n")
    
    # First check if sequence_stepid column exists
    cursor.execute("""
        SELECT column_name 
        FROM information_schema.columns 
        WHERE table_name = 'broadcast_messages' 
        AND column_name = 'sequence_stepid'
    """)
    
    if not cursor.fetchone():
        print("[WARNING] Column 'sequence_stepid' does not exist in broadcast_messages table")
        # Check for similar columns
        cursor.execute("""
            SELECT column_name 
            FROM information_schema.columns 
            WHERE table_name = 'broadcast_messages' 
            AND column_name LIKE '%step%'
            ORDER BY column_name
        """)
        similar_cols = cursor.fetchall()
        if similar_cols:
            print("\nSimilar columns found:")
            for col in similar_cols:
                print(f"  - {col[0]}")
    else:
        # 1. Failed messages with sequence_stepid
        print("[1. FAILED MESSAGES WITH SEQUENCE_STEPID]:")
        print("-" * 100)
        cursor.execute("""
            SELECT 
                bm.id,
                bm.recipient_phone,
                bm.status,
                bm.error_message,
                bm.created_at,
                bm.scheduled_at,
                bm.sequence_stepid,
                bm.sequence_id,
                s.name as sequence_name
            FROM broadcast_messages bm
            LEFT JOIN sequences s ON bm.sequence_id = s.id
            WHERE bm.status = 'failed' 
            AND bm.sequence_stepid IS NOT NULL
            AND bm.error_message IS NOT NULL
            ORDER BY bm.created_at DESC
            LIMIT 30
        """)
        failed_with_stepid = cursor.fetchall()
        
        if failed_with_stepid:
            for msg in failed_with_stepid:
                print(f"ID: {msg[0]}")
                print(f"  Phone: {msg[1]} | Status: {msg[2]}")
                print(f"  Error: {msg[3]}")
                print(f"  Step ID: {msg[6]} | Sequence: {msg[8] or 'Unknown'}")
                print(f"  Created: {msg[4]} | Scheduled: {msg[5]}")
                print()
        else:
            print("No failed messages with sequence_stepid found")
        
        # 2. Count by error type for messages with sequence_stepid
        print("\n[2. ERROR TYPES FOR MESSAGES WITH SEQUENCE_STEPID]:")
        print("-" * 100)
        cursor.execute("""
            SELECT 
                error_message,
                COUNT(*) as count,
                COUNT(DISTINCT sequence_id) as sequences_affected,
                COUNT(DISTINCT sequence_stepid) as steps_affected
            FROM broadcast_messages
            WHERE status = 'failed'
            AND sequence_stepid IS NOT NULL
            AND error_message IS NOT NULL
            GROUP BY error_message
            ORDER BY count DESC
        """)
        error_summary = cursor.fetchall()
        
        if error_summary:
            for error, count, seq_count, step_count in error_summary:
                print(f"Error: {error[:80]}...")
                print(f"  Count: {count} | Sequences: {seq_count} | Steps: {step_count}")
                print()
        else:
            print("No errors found for messages with sequence_stepid")
        
        # 3. Failed messages by sequence
        print("\n[3. FAILURES BY SEQUENCE (with sequence_stepid)]:")
        print("-" * 100)
        cursor.execute("""
            SELECT 
                s.name as sequence_name,
                s.id as sequence_id,
                COUNT(bm.id) as total_failed,
                COUNT(DISTINCT bm.sequence_stepid) as unique_steps,
                MIN(bm.created_at) as first_failure,
                MAX(bm.created_at) as last_failure
            FROM broadcast_messages bm
            JOIN sequences s ON bm.sequence_id = s.id
            WHERE bm.status = 'failed'
            AND bm.sequence_stepid IS NOT NULL
            GROUP BY s.id, s.name
            ORDER BY total_failed DESC
        """)
        seq_failures = cursor.fetchall()
        
        if seq_failures:
            for seq_name, seq_id, failed, steps, first, last in seq_failures:
                print(f"Sequence: {seq_name}")
                print(f"  ID: {seq_id}")
                print(f"  Failed messages: {failed} | Unique steps: {steps}")
                print(f"  First failure: {first} | Last failure: {last}")
                print()
        else:
            print("No sequence failures with sequence_stepid found")
        
        # 4. Check specific step patterns
        print("\n[4. FAILED MESSAGES BY STEP PATTERN]:")
        print("-" * 100)
        cursor.execute("""
            SELECT 
                ss.day_number,
                ss.trigger,
                COUNT(bm.id) as failures,
                bm.error_message
            FROM broadcast_messages bm
            JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
            WHERE bm.status = 'failed'
            AND bm.error_message IS NOT NULL
            GROUP BY ss.day_number, ss.trigger, bm.error_message
            ORDER BY failures DESC
            LIMIT 20
        """)
        step_patterns = cursor.fetchall()
        
        if step_patterns:
            for day, trigger, failures, error in step_patterns:
                print(f"Day {day} | Trigger: {trigger or 'None'}")
                print(f"  Failures: {failures}")
                print(f"  Error: {error[:60]}...")
                print()
        else:
            print("No step pattern data found")
        
        # 5. Summary statistics
        print("\n[5. SUMMARY STATISTICS]:")
        print("-" * 100)
        cursor.execute("""
            SELECT 
                COUNT(*) as total_with_stepid,
                COUNT(CASE WHEN status = 'failed' THEN 1 END) as failed_with_stepid,
                COUNT(CASE WHEN status = 'sent' THEN 1 END) as sent_with_stepid,
                COUNT(CASE WHEN status = 'pending' THEN 1 END) as pending_with_stepid
            FROM broadcast_messages
            WHERE sequence_stepid IS NOT NULL
        """)
        stats = cursor.fetchone()
        
        print(f"Total messages with sequence_stepid: {stats[0]}")
        print(f"  Failed: {stats[1]} ({stats[1]/stats[0]*100:.1f}%)" if stats[0] > 0 else "  Failed: 0")
        print(f"  Sent: {stats[2]} ({stats[2]/stats[0]*100:.1f}%)" if stats[0] > 0 else "  Sent: 0")
        print(f"  Pending: {stats[3]} ({stats[3]/stats[0]*100:.1f}%)" if stats[0] > 0 else "  Pending: 0")
    
    cursor.close()
    conn.close()
    print("\n[SUCCESS] Analysis complete!")
    
except Exception as e:
    print(f"[ERROR] {str(e)}")
    import traceback
    traceback.print_exc()
