#!/usr/bin/env python3
"""
Fix specific MySQL syntax errors found in logs
"""

import os
import re

def fix_double_backticks(content):
    """Fix double backticks like ``trigger``"""
    content = content.replace('``trigger``', '`trigger`')
    content = content.replace("``", "`")  # Fix any other double backticks
    return content

def fix_empty_where_clauses(content):
    """Fix queries with incomplete WHERE clauses"""
    # Fix patterns like "WHERE user_id = ? AND device_id =" (missing parameter)
    content = re.sub(r'AND device_id =\s*\n', r'AND device_id = ?\n', content)
    content = re.sub(r'AND device_id =\s*$', r'AND device_id = ?', content, flags=re.MULTILINE)
    
    # Fix trailing AND without condition
    content = re.sub(r'WHERE\s+AND\s+', 'WHERE ', content)
    content = re.sub(r'AND\s+AND\s+', 'AND ', content)
    
    return content

def fix_query_syntax(filepath):
    """Fix various SQL syntax issues"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original = content
        
        # Apply fixes
        content = fix_double_backticks(content)
        content = fix_empty_where_clauses(content)
        
        # Fix any remaining PostgreSQL-specific syntax
        
        # Fix COUNT queries with incorrect syntax
        content = re.sub(r'COUNT\s*\(\s*\)\s*WHERE', 'COUNT(*) FROM leads WHERE', content)
        
        # Fix empty IN clauses - should have at least one placeholder
        content = re.sub(r'IN\s*\(\s*\)', 'IN (?)', content)
        
        # Fix queries ending with AND but no condition
        content = re.sub(r'AND\s*\n\s*ORDER BY', '\nORDER BY', content)
        content = re.sub(r'AND\s*\n\s*GROUP BY', '\nGROUP BY', content)
        content = re.sub(r'AND\s*\n\s*LIMIT', '\nLIMIT', content)
        
        # Fix queries with WHERE at end with no condition
        content = re.sub(r'WHERE\s*\n\s*ORDER BY', '\nORDER BY', content)
        content = re.sub(r'WHERE\s*\n\s*GROUP BY', '\nGROUP BY', content)
        content = re.sub(r'WHERE\s*\n\s*LIMIT', '\nLIMIT', content)
        
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
    
    # Priority files based on the errors shown
    priority_files = [
        "repository/lead_repository.go",
        "repository/campaign_repository.go",
        "repository/sequence_repository.go",
        "repository/broadcast_repository.go",
        "ui/rest/app.go",
        "ui/rest/dashboard.go",
        "usecase/broadcast_scheduler.go",
    ]
    
    print("Fixing specific MySQL syntax errors from logs...")
    print("-" * 50)
    
    fixed_count = 0
    for relative_path in priority_files:
        filepath = os.path.join(base_path, relative_path)
        if os.path.exists(filepath):
            if fix_query_syntax(filepath):
                print(f"[FIXED] {relative_path}")
                fixed_count += 1
            else:
                print(f"[OK] {relative_path}")
        else:
            print(f"[SKIP] {relative_path} - not found")
    
    # Also check for dashboard handler files
    dashboard_files = [
        "ui/rest/dashboard_handler.go",
        "ui/rest/campaign_handler.go",
        "ui/rest/sequence_handler.go",
    ]
    
    print("\nChecking dashboard handlers...")
    for relative_path in dashboard_files:
        filepath = os.path.join(base_path, relative_path)
        if os.path.exists(filepath):
            if fix_query_syntax(filepath):
                print(f"[FIXED] {relative_path}")
                fixed_count += 1
    
    print("-" * 50)
    print(f"Fixed {fixed_count} files")
    
    # Now let's specifically look for the dashboard stats queries
    print("\nSearching for dashboard statistics queries...")
    
    # Search for the problematic queries
    search_patterns = [
        "Error counting campaigns",
        "Error getting broadcast stats",
        "Error counting sequences",
    ]
    
    # Let's find where these queries are
    for root, dirs, files in os.walk(base_path):
        for file in files:
            if file.endswith('.go'):
                filepath = os.path.join(root, file)
                try:
                    with open(filepath, 'r', encoding='utf-8') as f:
                        content = f.read()
                    
                    for pattern in search_patterns:
                        if pattern in content:
                            print(f"\nFound '{pattern}' in {os.path.relpath(filepath, base_path)}")
                            # Find the query near this error
                            lines = content.split('\n')
                            for i, line in enumerate(lines):
                                if pattern in line:
                                    # Show context
                                    start = max(0, i - 10)
                                    end = min(len(lines), i + 5)
                                    print(f"Context (lines {start}-{end}):")
                                    for j in range(start, end):
                                        print(f"{j}: {lines[j]}")
                except:
                    pass

if __name__ == "__main__":
    main()
