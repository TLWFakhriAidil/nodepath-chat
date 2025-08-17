import psycopg2
import sys

def fix_sequence_trigger_mapping():
    """Fix sequence trigger mapping issues"""
    
    conn = psycopg2.connect("postgresql://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway")
    cursor = conn.cursor()
    
    try:
        # First, let's check current sequence triggers
        print("=== CHECKING CURRENT SEQUENCE TRIGGER SETUP ===\n")
        
        cursor.execute("""
            SELECT s.id, s.name, s.trigger, ss.id as step_id, ss.trigger as step_trigger, ss.is_entry_point
            FROM sequences s
            JOIN sequence_steps ss ON s.id = ss.sequence_id
            WHERE ss.is_entry_point = true
            ORDER BY s.name
        """)
        
        sequences = cursor.fetchall()
        print("Current Entry Points:")
        for seq in sequences:
            print(f"  {seq[1]}: Seq Trigger='{seq[2]}', Entry Step Trigger='{seq[4]}'")
        
        # Check for mismatched broadcast messages
        print("\n=== CHECKING MISMATCHED BROADCAST MESSAGES ===\n")
        
        cursor.execute("""
            SELECT COUNT(*), l.trigger, s.name, ss.trigger as step_trigger
            FROM leads l
            JOIN broadcast_messages bm ON l.phone = bm.recipient_phone
            JOIN sequences s ON bm.sequence_id = s.id
            LEFT JOIN sequence_steps ss ON bm.sequence_stepid = ss.id
            WHERE l.trigger LIKE '%VITAC%' 
            AND s.name NOT LIKE '%VITAC%'
            GROUP BY l.trigger, s.name, ss.trigger
        """)
        
        mismatches = cursor.fetchall()
        if mismatches:
            print("Found mismatched assignments:")
            for m in mismatches:
                print(f"  {m[0]} leads with trigger '{m[1]}' assigned to '{m[2]}'")
        else:
            print("No mismatched assignments found")
        
        # Fix sequence_stepid NULL values for VITAC sequences
        print("\n=== FIXING NULL sequence_stepid FOR VITAC SEQUENCES ===\n")
        
        # Get VITAC sequence IDs
        cursor.execute("""
            SELECT id, name FROM sequences WHERE name LIKE '%VITAC%'
        """)
        vitac_sequences = cursor.fetchall()
        
        for seq_id, seq_name in vitac_sequences:
            # Get sequence steps for this sequence
            cursor.execute("""
                SELECT id, day_number, trigger FROM sequence_steps 
                WHERE sequence_id = %s 
                ORDER BY day_number
            """, (seq_id,))
            steps = cursor.fetchall()
            
            # Update broadcast messages to have correct sequence_stepid
            for step_id, day_num, step_trigger in steps:
                cursor.execute("""
                    UPDATE broadcast_messages 
                    SET sequence_stepid = %s
                    WHERE sequence_id = %s 
                    AND sequence_stepid IS NULL
                    AND recipient_phone IN (
                        SELECT phone FROM leads WHERE niche = 'VITAC'
                    )
                    AND content LIKE (
                        SELECT CONCAT('%%', SUBSTRING(content, 1, 50), '%%')
                        FROM sequence_steps WHERE id = %s
                    )
                """, (step_id, seq_id, step_id))
                
                updated = cursor.rowcount
                if updated > 0:
                    print(f"  Updated {updated} messages for {seq_name} day {day_num}")
        
        # Commit changes
        conn.commit()
        print("\n✅ Sequence trigger mapping fixes applied successfully!")
        
    except Exception as e:
        conn.rollback()
        print(f"❌ Error: {e}")
        sys.exit(1)
    finally:
        cursor.close()
        conn.close()

if __name__ == "__main__":
    fix_sequence_trigger_mapping()
