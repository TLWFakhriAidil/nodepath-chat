#!/usr/bin/env python3
"""
Revert overly aggressive keyword replacements in Go code
Only fix keywords inside SQL query strings
"""

import os
import re

def revert_go_code_keywords(content):
    """Fix Go code that was incorrectly modified"""
    
    # Revert Go code keywords that were incorrectly modified
    # These patterns indicate we're NOT in a SQL query
    
    # Fix struct field declarations
    content = re.sub(r'(\w+\s+)`trigger`(\s+\w+)', r'\1trigger\2', content)
    content = re.sub(r'(\w+\s+)`from`(\s+\w+)', r'\1from\2', content)
    content = re.sub(r'(\w+\s+)`to`(\s+\w+)', r'\1to\2', content)
    content = re.sub(r'(\w+\s+)`status`(\s+\w+)', r'\1status\2', content)
    content = re.sub(r'(\w+\s+)`type`(\s+\w+)', r'\1type\2', content)
    content = re.sub(r'(\w+\s+)`key`(\s+\w+)', r'\1key\2', content)
    content = re.sub(r'(\w+\s+)`order`(\s+\w+)', r'\1order\2', content)
    content = re.sub(r'(\w+\s+)`limit`(\s+\w+)', r'\1limit\2', content)
    
    # Fix case statements
    content = re.sub(r'`case`\s+', 'case ', content)
    content = re.sub(r'`if`\s+', 'if ', content)
    
    # Fix struct fields in composite literals
    content = re.sub(r'(\w+:\s*)`trigger`', r'\1trigger', content)
    content = re.sub(r'(\w+:\s*)`from`', r'\1from', content)
    content = re.sub(r'(\w+:\s*)`to`', r'\1to', content)
    content = re.sub(r'(\w+:\s*)`status`', r'\1status', content)
    content = re.sub(r'(\w+:\s*)`type`', r'\1type', content)
    
    # Fix function parameters
    content = re.sub(r'(func\s+\w+\([^)]*)`trigger`', r'\1trigger', content)
    content = re.sub(r'(func\s+\w+\([^)]*)`from`', r'\1from', content)
    content = re.sub(r'(func\s+\w+\([^)]*)`to`', r'\1to', content)
    
    return content

def fix_sql_queries_only(content):
    """Fix keywords only inside SQL query strings"""
    
    # Find all SQL queries (strings that look like SQL)
    sql_pattern = r'`([^`]+(?:SELECT|INSERT|UPDATE|DELETE|FROM|WHERE|SET)[^`]+)`'
    
    def fix_sql_string(match):
        sql = match.group(1)
        
        # Only fix keywords in SQL context
        # Add backticks to trigger in SELECT/WHERE/INSERT contexts
        sql = re.sub(r'(\s+)trigger(\s*[,\s\)])', r'\1`trigger`\2', sql)
        sql = re.sub(r'(\s+)trigger(\s+FROM)', r'\1`trigger`\2', sql)
        sql = re.sub(r'(,\s*)trigger(\s*[,\)])', r'\1`trigger`\2', sql)
        sql = re.sub(r'(SET\s+)trigger(\s*=)', r'\1`trigger`\2', sql)
        sql = re.sub(r'(WHERE\s+)trigger(\s*=)', r'\1`trigger`\2', sql)
        
        return f'`{sql}`'
    
    content = re.sub(sql_pattern, fix_sql_string, content, flags=re.DOTALL)
    
    return content

def fix_file(filepath):
    """Fix a single file"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original = content
        
        # First revert incorrect replacements
        content = revert_go_code_keywords(content)
        
        # Then fix only SQL queries
        content = fix_sql_queries_only(content)
        
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
    
    # Files that had compilation errors
    problem_files = [
        "database/connection.go",
        "database/emergency_fix.go",
        "database/migrate_sequence_steps.go",
        "database/migrations.go",
        "infrastructure/whatsapp/stability/ultra_stable_connection.go",
    ]
    
    print("Fixing overly aggressive keyword replacements...")
    print("-" * 50)
    
    for relative_path in problem_files:
        filepath = os.path.join(base_path, relative_path)
        if os.path.exists(filepath):
            if fix_file(filepath):
                print(f"[FIXED] {relative_path}")
            else:
                print(f"[OK] {relative_path}")
        else:
            print(f"[SKIP] {relative_path} - not found")
    
    # Also check all files that were modified
    print("\nChecking all previously modified files...")
    
    for root, dirs, files in os.walk(base_path):
        for file in files:
            if file.endswith('.go'):
                filepath = os.path.join(root, file)
                
                # Read file to check for issues
                try:
                    with open(filepath, 'r', encoding='utf-8') as f:
                        content = f.read()
                    
                    # Check for incorrectly modified Go code
                    if re.search(r'`(case|if|from|to|type)`\s+[^`]', content):
                        relative_path = os.path.relpath(filepath, base_path)
                        if fix_file(filepath):
                            print(f"[FIXED] {relative_path}")
                except:
                    pass

if __name__ == "__main__":
    main()
