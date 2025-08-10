#!/usr/bin/env python3
"""
Fix SQL keyword capitalization in repository files
"""

import os
import re

def fix_sql_keywords(content):
    """Fix lowercase SQL keywords in queries"""
    
    # Find SQL query blocks
    sql_patterns = [
        # Backtick strings
        r'`([^`]+)`',
        # Double quote strings
        r'"([^"]+)"',
    ]
    
    def fix_keywords_in_sql(sql):
        """Fix keywords within SQL string"""
        # List of SQL keywords that should be uppercase
        keywords = [
            'select', 'from', 'where', 'and', 'or', 'not', 'in', 'like', 
            'order by', 'group by', 'having', 'limit', 'offset',
            'insert into', 'values', 'update', 'set', 'delete',
            'join', 'left join', 'right join', 'inner join', 'on',
            'as', 'distinct', 'count', 'sum', 'avg', 'max', 'min',
            'case', 'when', 'then', 'else', 'end', 'exists',
            'between', 'is null', 'is not null', 'cast', 'concat'
        ]
        
        # Replace keywords with uppercase version
        for keyword in keywords:
            # Word boundary to avoid replacing parts of words
            sql = re.sub(r'\b' + keyword + r'\b', keyword.upper(), sql, flags=re.IGNORECASE)
        
        return sql
    
    # Process SQL in backticks
    def replace_backtick_sql(match):
        sql = match.group(1)
        if any(kw in sql.lower() for kw in ['select', 'from', 'where', 'insert', 'update', 'delete']):
            sql = fix_keywords_in_sql(sql)
        return f'`{sql}`'
    
    content = re.sub(r'`([^`]+)`', replace_backtick_sql, content, flags=re.DOTALL)
    
    # Also fix specific lowercase keywords that appear as standalone
    # Fix `from` to FROM when it's clearly a SQL keyword
    content = re.sub(r'`from`\s+leads', 'FROM leads', content)
    content = re.sub(r'"from"\s+leads', 'FROM leads', content)
    
    # Fix `status` when used as SQL keyword
    content = re.sub(r'`status`\s*=', 'status =', content)
    content = re.sub(r'`status`\s+FROM', 'status FROM', content)
    
    return content

def fix_repository_files():
    """Fix repository files with SQL issues"""
    
    base_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository"
    
    files_to_fix = [
        "campaign_repository.go",
        "lead_ai_repository.go",
        "lead_repository.go",
    ]
    
    for filename in files_to_fix:
        filepath = os.path.join(base_path, filename)
        
        if not os.path.exists(filepath):
            print(f"[SKIP] {filename} - not found")
            continue
            
        try:
            with open(filepath, 'r', encoding='utf-8') as f:
                content = f.read()
            
            original = content
            content = fix_sql_keywords(content)
            
            if content != original:
                with open(filepath, 'w', encoding='utf-8') as f:
                    f.write(content)
                print(f"[FIXED] {filename}")
            else:
                print(f"[OK] {filename}")
                
        except Exception as e:
            print(f"[ERROR] {filename}: {e}")

def main():
    print("Fixing SQL keyword capitalization...")
    print("-" * 60)
    
    fix_repository_files()
    
    print("-" * 60)
    print("Done!")

if __name__ == "__main__":
    main()
