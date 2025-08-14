#!/usr/bin/env python3
"""
Fix MySQL syntax errors in Go source files
Addresses:
1. Reserved keyword 'trigger' not in backticks
2. PostgreSQL syntax (||) instead of MySQL CONCAT
3. ILIKE instead of LIKE
4. Other PostgreSQL-specific syntax
"""

import os
import re
import glob

def fix_mysql_syntax_in_file(filepath):
    """Fix MySQL syntax errors in a single file"""
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    original_content = content
    fixes_made = []
    
    # Fix 1: Replace unquoted 'trigger' in SQL queries with `trigger`
    # Pattern to match trigger in SELECT, INSERT, UPDATE statements
    patterns = [
        # SELECT statements
        (r'(SELECT\s+[^`]*?)\btrigger\b([^`])', r'\1`trigger`\2'),
        (r'(select\s+[^`]*?)\btrigger\b([^`])', r'\1`trigger`\2'),
        
        # FROM/WHERE clauses
        (r'(\s+)trigger(\s*[,\s])', r'\1`trigger`\2'),
        
        # INSERT INTO
        (r'(INSERT\s+INTO\s+\w+\s*\([^)]*?)\btrigger\b', r'\1`trigger`'),
        (r'(insert\s+into\s+\w+\s*\([^)]*?)\btrigger\b', r'\1`trigger`'),
        
        # UPDATE statements
        (r'(UPDATE\s+\w+\s+SET\s+[^`]*?)\btrigger\b\s*=', r'\1`trigger` ='),
        (r'(update\s+\w+\s+set\s+[^`]*?)\btrigger\b\s*=', r'\1`trigger` ='),
        
        # ORDER BY
        (r'(ORDER\s+BY\s+[^`]*?)\btrigger\b', r'\1`trigger`'),
        (r'(order\s+by\s+[^`]*?)\btrigger\b', r'\1`trigger`'),
        
        # WHERE clauses
        (r'(WHERE\s+[^`]*?)\btrigger\b\s*=', r'\1`trigger` ='),
        (r'(where\s+[^`]*?)\btrigger\b\s*=', r'\1`trigger` ='),
        (r'(WHERE\s+[^`]*?)\btrigger\b\s+IS', r'\1`trigger` IS'),
        (r'(where\s+[^`]*?)\btrigger\b\s+is', r'\1`trigger` is'),
    ]
    
    for pattern, replacement in patterns:
        new_content = re.sub(pattern, replacement, content, flags=re.MULTILINE | re.DOTALL)
        if new_content != content:
            fixes_made.append(f"Fixed 'trigger' keyword")
            content = new_content
    
    # Fix 2: Replace PostgreSQL || concatenation with MySQL CONCAT
    # Match patterns like: '%' || ? || '%'
    concat_pattern = r"'%'\s*\|\|\s*\?\s*\|\|\s*'%'"
    if re.search(concat_pattern, content):
        content = re.sub(concat_pattern, "CONCAT('%', ?, '%')", content)
        fixes_made.append("Fixed PostgreSQL || concatenation")
    
    # Fix 3: Replace ILIKE with LIKE (MySQL is case-insensitive by default)
    if ' ILIKE ' in content or ' ilike ' in content:
        content = re.sub(r'\bILIKE\b', 'LIKE', content)
        content = re.sub(r'\bilike\b', 'LIKE', content)
        fixes_made.append("Fixed ILIKE to LIKE")
    
    # Fix 4: Fix RETURNING clause (not supported in MySQL)
    if ' RETURNING ' in content:
        content = re.sub(r'\s+RETURNING\s+\w+', '', content)
        fixes_made.append("Removed RETURNING clause")
    
    # Fix 5: Fix gen_random_uuid() to UUID()
    if 'gen_random_uuid()' in content:
        content = content.replace('gen_random_uuid()', 'UUID()')
        fixes_made.append("Fixed UUID generation")
    
    # Fix 6: Fix boolean values
    if re.search(r'\bTRUE\b', content) or re.search(r'\bFALSE\b', content):
        content = re.sub(r'\bTRUE\b', '1', content)
        content = re.sub(r'\bFALSE\b', '0', content)
        fixes_made.append("Fixed boolean values")
    
    # Fix 7: Fix parameter placeholders $1, $2 to ?, ?
    param_pattern = r'\$(\d+)'
    if re.search(param_pattern, content):
        # Count parameters and replace
        def replace_params(match):
            return '?'
        content = re.sub(param_pattern, replace_params, content)
        fixes_made.append("Fixed parameter placeholders")
    
    # Only write if changes were made
    if content != original_content:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        return filepath, fixes_made
    
    return None, []

def fix_repository_files():
    """Fix all repository files"""
    repo_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository"
    
    files_to_fix = [
        "lead_repository.go",
        "campaign_repository.go",
        "sequence_repository.go",
        "user_repository.go",
        "lead_ai_repository.go",
        "broadcast_repository.go"
    ]
    
    print("Fixing MySQL syntax errors in repository files...")
    print("=" * 70)
    
    for filename in files_to_fix:
        filepath = os.path.join(repo_path, filename)
        if os.path.exists(filepath):
            result, fixes = fix_mysql_syntax_in_file(filepath)
            if result:
                print(f"\n[OK] Fixed {filename}:")
                for fix in fixes:
                    print(f"   - {fix}")
            else:
                print(f"[OK] {filename} - No fixes needed")
        else:
            print(f"[ERROR] {filename} - File not found")
    
    # Also fix usecase files
    usecase_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\usecase"
    usecase_files = [
        "direct_broadcast_processor.go",
        "broadcast_scheduler.go",
        "sequence.go"
    ]
    
    print("\n\nFixing MySQL syntax errors in usecase files...")
    print("=" * 70)
    
    for filename in usecase_files:
        filepath = os.path.join(usecase_path, filename)
        if os.path.exists(filepath):
            result, fixes = fix_mysql_syntax_in_file(filepath)
            if result:
                print(f"\n✅ Fixed {filename}:")
                for fix in fixes:
                    print(f"   - {fix}")
            else:
                print(f"✅ {filename} - No fixes needed")

    # Fix UI/REST files
    ui_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest"
    ui_files = [
        "app.go",
        "team_member_handlers.go",
        "webhook_lead.go",
        "sequence.go"
    ]
    
    print("\n\nFixing MySQL syntax errors in UI/REST files...")
    print("=" * 70)
    
    for filename in ui_files:
        filepath = os.path.join(ui_path, filename)
        if os.path.exists(filepath):
            result, fixes = fix_mysql_syntax_in_file(filepath)
            if result:
                print(f"\n✅ Fixed {filename}:")
                for fix in fixes:
                    print(f"   - {fix}")
            else:
                print(f"✅ {filename} - No fixes needed")

if __name__ == "__main__":
    fix_repository_files()
    print("\n\n✅ MySQL syntax fixes completed!")
