#!/usr/bin/env python3
"""
Fix MySQL syntax errors - Simple version
"""

import os
import re

def fix_trigger_keyword(content):
    """Fix unquoted trigger keyword in SQL queries"""
    # List of patterns to fix trigger keyword
    patterns = [
        # In SELECT lists
        (r'SELECT ([^;]+?)\btrigger\b', r'SELECT \1`trigger`'),
        (r'select ([^;]+?)\btrigger\b', r'select \1`trigger`'),
        
        # After commas in SELECT
        (r',\s*trigger\s*,', r', `trigger`,'),
        (r',\s*trigger\s+FROM', r', `trigger` FROM'),
        (r',\s*trigger\s+from', r', `trigger` from'),
        
        # At end of SELECT before FROM
        (r'\s+trigger\s+FROM\s+', r' `trigger` FROM '),
        (r'\s+trigger\s+from\s+', r' `trigger` from '),
        
        # In WHERE clauses
        (r'WHERE\s+trigger\s*=', r'WHERE `trigger` ='),
        (r'where\s+trigger\s*=', r'where `trigger` ='),
        (r'AND\s+trigger\s*=', r'AND `trigger` ='),
        (r'and\s+trigger\s*=', r'and `trigger` ='),
        
        # With IS NULL/IS NOT NULL
        (r'\btrigger\s+IS\s+NULL', r'`trigger` IS NULL'),
        (r'\btrigger\s+is\s+null', r'`trigger` is null'),
        (r'\btrigger\s+IS\s+NOT\s+NULL', r'`trigger` IS NOT NULL'),
        (r'\btrigger\s+is\s+not\s+null', r'`trigger` is not null'),
        
        # In INSERT statements
        (r'INSERT INTO ([^(]+)\(([^)]*)\btrigger\b', r'INSERT INTO \1(\2`trigger`'),
        (r'insert into ([^(]+)\(([^)]*)\btrigger\b', r'insert into \1(\2`trigger`'),
        
        # In UPDATE statements
        (r'SET\s+trigger\s*=', r'SET `trigger` ='),
        (r'set\s+trigger\s*=', r'set `trigger` ='),
        
        # In ORDER BY
        (r'ORDER\s+BY\s+trigger', r'ORDER BY `trigger`'),
        (r'order\s+by\s+trigger', r'order by `trigger`'),
    ]
    
    for pattern, replacement in patterns:
        content = re.sub(pattern, replacement, content)
    
    return content

def fix_postgresql_concat(content):
    """Fix PostgreSQL || concatenation"""
    # Fix patterns like '%' || ? || '%'
    content = re.sub(r"'%'\s*\|\|\s*\?\s*\|\|\s*'%'", "CONCAT('%', ?, '%')", content)
    
    # Fix patterns like ? || '%'
    content = re.sub(r"\?\s*\|\|\s*'%'", "CONCAT(?, '%')", content)
    
    # Fix patterns like '%' || ?
    content = re.sub(r"'%'\s*\|\|\s*\?", "CONCAT('%', ?)", content)
    
    return content

def fix_sql_syntax(filepath):
    """Fix SQL syntax in a file"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original = content
        
        # Apply fixes
        content = fix_trigger_keyword(content)
        content = fix_postgresql_concat(content)
        
        # Fix ILIKE to LIKE
        content = re.sub(r'\bILIKE\b', 'LIKE', content)
        content = re.sub(r'\bilike\b', 'LIKE', content)
        
        # Fix parameter placeholders $1, $2, etc to ?
        content = re.sub(r'\$\d+', '?', content)
        
        # Fix gen_random_uuid() to UUID()
        content = content.replace('gen_random_uuid()', 'UUID()')
        
        # Fix TRUE/FALSE to 1/0
        content = re.sub(r'\bTRUE\b', '1', content)
        content = re.sub(r'\bFALSE\b', '0', content)
        
        if content != original:
            with open(filepath, 'w', encoding='utf-8') as f:
                f.write(content)
            return True
        return False
    except Exception as e:
        print(f"Error processing {filepath}: {e}")
        return False

def main():
    base_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"
    
    # Files to fix
    files_to_fix = [
        # Repository files
        "repository/lead_repository.go",
        "repository/campaign_repository.go",
        "repository/sequence_repository.go",
        "repository/user_repository.go",
        "repository/lead_ai_repository.go",
        "repository/broadcast_repository.go",
        
        # Usecase files
        "usecase/direct_broadcast_processor.go",
        "usecase/broadcast_scheduler.go",
        "usecase/sequence.go",
        
        # UI/REST files
        "ui/rest/app.go",
        "ui/rest/team_member_handlers.go",
        "ui/rest/webhook_lead.go",
        "ui/rest/sequence.go",
        
        # Infrastructure files
        "infrastructure/whatsapp/chat_to_leads.go",
    ]
    
    print("Fixing MySQL syntax errors...")
    print("-" * 50)
    
    fixed_count = 0
    for relative_path in files_to_fix:
        filepath = os.path.join(base_path, relative_path)
        if os.path.exists(filepath):
            if fix_sql_syntax(filepath):
                print(f"[FIXED] {relative_path}")
                fixed_count += 1
            else:
                print(f"[OK] {relative_path}")
        else:
            print(f"[SKIP] {relative_path} - not found")
    
    print("-" * 50)
    print(f"Fixed {fixed_count} files")
    print("\nNext steps:")
    print("1. Rebuild the application: build_local.bat")
    print("2. Test all CRUD operations")
    print("3. Check logs for any remaining SQL errors")

if __name__ == "__main__":
    main()
