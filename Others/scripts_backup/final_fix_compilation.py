#!/usr/bin/env python3
"""
Final comprehensive fix for all compilation errors
"""

import os
import re

def fix_all_backtick_issues():
    """Remove problematic backticks from Go code"""
    
    base_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"
    
    # List of files with known issues
    files_to_fix = {
        "database/emergency_fix.go": [
            # Remove backticks in the middle of SQL strings
            ("`trigger`", "trigger"),
            ("`from`", "from"),
            ("`to`", "to"),
        ],
        "database/migrate_sequence_steps.go": [
            # Fix standalone backticked words
            ("rows.Scan(&step.ID, &`trigger`)", "rows.Scan(&step.ID, &trigger)"),
            ("`trigger`\n", "trigger\n"),
        ],
        "database/migrations.go": [
            # Fix struct fields
            ("MigrationName: `to`", 'MigrationName: "to"'),
            ("if `trigger`", "if trigger"),
            ("step.`trigger`", "step.trigger"),
        ],
        "infrastructure/whatsapp/stability/ultra_stable_connection.go": [
            # Fix Go keywords
            ("`case`", "case"),
            ("`if`", "if"),
            ("`else`", "else"),
            ("`for`", "for"),
            ("`break`", "break"),
            ("`continue`", "continue"),
            ("`return`", "return"),
        ],
    }
    
    for relative_path, replacements in files_to_fix.items():
        filepath = os.path.join(base_path, relative_path)
        
        if not os.path.exists(filepath):
            print(f"[SKIP] {relative_path} - not found")
            continue
            
        try:
            with open(filepath, 'r', encoding='utf-8') as f:
                content = f.read()
            
            original = content
            
            # Apply all replacements
            for old, new in replacements:
                content = content.replace(old, new)
            
            # Additional generic fixes
            # Remove backticks from variable names
            content = re.sub(r'&`(\w+)`', r'&\1', content)
            content = re.sub(r'\.`(\w+)`', r'.\1', content)
            
            # Fix SQL queries - add backticks only where needed
            # In UPDATE SET clauses
            content = re.sub(r'UPDATE\s+(\w+)\s+SET\s+trigger\s*=', r'UPDATE \1 SET `trigger` =', content)
            content = re.sub(r'SET\s+trigger\s*=', 'SET `trigger` =', content)
            
            # In WHERE clauses
            content = re.sub(r'WHERE\s+trigger\s*=', 'WHERE `trigger` =', content)
            content = re.sub(r'WHERE\s+trigger\s+IS', 'WHERE `trigger` IS', content)
            
            # Write back if changed
            if content != original:
                with open(filepath, 'w', encoding='utf-8') as f:
                    f.write(content)
                print(f"[FIXED] {relative_path}")
            else:
                print(f"[OK] {relative_path}")
                
        except Exception as e:
            print(f"[ERROR] {relative_path}: {e}")

def main():
    print("Final comprehensive fix for compilation errors...")
    print("-" * 60)
    
    fix_all_backtick_issues()
    
    print("-" * 60)
    print("All fixes applied!")
    print("\nNext: run build_local.bat")

if __name__ == "__main__":
    main()
