#!/usr/bin/env python3
"""
Clear all sequence contacts from the database
"""

import psycopg2
import sys

# Database connection string
DB_URI = "postgres://postgres:CNFPbgfjsIVirTuqLMoObNMvoYobDDTU@yamanote.proxy.rlwy.net:49914/railway?sslmode=require"

def clear_sequence_contacts():
    """Clear all sequence contacts with confirmation"""
    try:
        # Connect to database
        conn = psycopg2.connect(DB_URI)
        cur = conn.cursor()
        
        # First, show current count
        cur.execute("SELECT COUNT(*) FROM sequence_contacts")
        total_count = cur.fetchone()[0]
        
        print(f"\nCurrent sequence_contacts count: {total_count}")
        
        if total_count == 0:
            print("No sequence contacts to delete.")
            return
        
        # Show breakdown by status
        print("\nBreakdown by status:")
        cur.execute("""
            SELECT status, COUNT(*) as count 
            FROM sequence_contacts 
            GROUP BY status 
            ORDER BY count DESC
        """)
        
        for row in cur.fetchall():
            print(f"  {row[0]}: {row[1]}")
        
        # Ask for confirmation
        confirm = input(f"\nAre you sure you want to delete ALL {total_count} sequence contacts? (yes/no): ")
        
        if confirm.lower() != 'yes':
            print("Operation cancelled.")
            return
        
        # Delete all records
        print("\nDeleting all sequence contacts...")
        cur.execute("DELETE FROM sequence_contacts")
        deleted_count = cur.rowcount
        
        # Commit the changes
        conn.commit()
        
        print(f"✅ Successfully deleted {deleted_count} sequence contacts")
        
        # Verify
        cur.execute("SELECT COUNT(*) FROM sequence_contacts")
        remaining = cur.fetchone()[0]
        print(f"Remaining contacts: {remaining}")
        
        # Close connection
        cur.close()
        conn.close()
        
    except Exception as e:
        print(f"❌ Error: {e}")
        if 'conn' in locals():
            conn.rollback()
            conn.close()

def clear_specific_sequence(sequence_id=None):
    """Clear contacts for a specific sequence only"""
    try:
        conn = psycopg2.connect(DB_URI)
        cur = conn.cursor()
        
        if not sequence_id:
            # Show all sequences
            cur.execute("""
                SELECT s.id, s.name, COUNT(sc.id) as contact_count
                FROM sequences s
                LEFT JOIN sequence_contacts sc ON sc.sequence_id = s.id
                GROUP BY s.id, s.name
                HAVING COUNT(sc.id) > 0
                ORDER BY COUNT(sc.id) DESC
            """)
            
            sequences = cur.fetchall()
            if not sequences:
                print("No sequences with contacts found.")
                return
                
            print("\nSequences with contacts:")
            for seq in sequences:
                print(f"  {seq[0]}: {seq[1]} ({seq[2]} contacts)")
            
            sequence_id = input("\nEnter sequence ID to clear (or 'all' for all): ")
            
        if sequence_id.lower() == 'all':
            clear_sequence_contacts()
            return
            
        # Delete for specific sequence
        cur.execute("DELETE FROM sequence_contacts WHERE sequence_id = %s", (sequence_id,))
        deleted = cur.rowcount
        conn.commit()
        
        print(f"✅ Deleted {deleted} contacts from sequence {sequence_id}")
        
        cur.close()
        conn.close()
        
    except Exception as e:
        print(f"❌ Error: {e}")

if __name__ == "__main__":
    print("=== Sequence Contacts Cleanup ===")
    print("1. Clear ALL sequence contacts")
    print("2. Clear contacts for specific sequence")
    print("3. Exit")
    
    choice = input("\nSelect option (1-3): ")
    
    if choice == '1':
        clear_sequence_contacts()
    elif choice == '2':
        clear_specific_sequence()
    else:
        print("Exiting...")
