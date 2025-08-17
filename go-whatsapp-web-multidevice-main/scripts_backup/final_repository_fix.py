#!/usr/bin/env python3
"""
Final comprehensive fix for all repository syntax errors
"""

import os
import re

def fix_repository_file(filepath):
    """Fix all syntax issues in a repository file"""
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    original = content
    
    # 1. Remove double backticks
    content = content.replace('``', '`')
    
    # 2. Fix backticked words at end of statements
    # Remove backticks from words that appear at end of line
    content = re.sub(r'`(\w+)`\s*$', r'\1', content, flags=re.MULTILINE)
    content = re.sub(r'`(\w+)`\s*\n', r'\1\n', content)
    
    # 3. Fix SQL queries - add backticks only in SQL context
    def fix_sql_query(match):
        query = match.group(1)
        
        # Fix trigger keyword in SQL
        query = re.sub(r'\btrigger\b(?!\`)', '`trigger`', query)
        
        # Fix SQL keywords to uppercase
        keywords = ['SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'ORDER BY', 
                   'GROUP BY', 'INSERT INTO', 'UPDATE', 'SET', 'VALUES',
                   'DELETE', 'LIMIT', 'OFFSET', 'JOIN', 'LEFT JOIN', 
                   'AS', 'ON', 'IN', 'LIKE', 'IS NULL', 'IS NOT NULL']
        
        for kw in keywords:
            query = re.sub(r'\b' + kw + r'\b', kw, query, flags=re.IGNORECASE)
        
        return f'`{query}`'
    
    # Apply to query assignments
    content = re.sub(r'query\s*:=\s*`([^`]+)`', fix_sql_query, content, flags=re.DOTALL)
    
    # 4. Fix specific patterns
    # Fix SELECT statements with trigger
    content = re.sub(r'SELECT\s+(.*?)\s+trigger\s*\n', r'SELECT \1 `trigger`\n', content)
    content = re.sub(r'SELECT\s+(.*?)\s+trigger\s+FROM', r'SELECT \1 `trigger` FROM', content)
    
    # 5. Fix status/from at end of statement
    content = re.sub(r'`status`\s*$', 'status', content, flags=re.MULTILINE)
    content = re.sub(r'`from`\s*$', 'from', content, flags=re.MULTILINE)
    content = re.sub(r'`order`\s*$', 'order', content, flags=re.MULTILINE)
    
    # 6. Fix standalone backticked words
    content = re.sub(r'^`(\w+)`$', r'\1', content, flags=re.MULTILINE)
    
    if content != original:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        return True
    return False

def main():
    base_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository"
    
    # Files that need fixing based on errors
    files_to_fix = [
        "lead_repository.go",
        "lead_ai_repository.go",
        "message_analytics_repository.go",
        "campaign_repository.go",
        "sequence_repository.go",
        "broadcast_repository.go",
    ]
    
    print("Applying final repository fixes...")
    print("-" * 60)
    
    fixed_count = 0
    for filename in files_to_fix:
        filepath = os.path.join(base_path, filename)
        
        if not os.path.exists(filepath):
            print(f"[SKIP] {filename} - not found")
            continue
            
        try:
            if fix_repository_file(filepath):
                print(f"[FIXED] {filename}")
                fixed_count += 1
            else:
                print(f"[OK] {filename}")
        except Exception as e:
            print(f"[ERROR] {filename}: {e}")
    
    print("-" * 60)
    print(f"Fixed {fixed_count} files")

if __name__ == "__main__":
    main()
