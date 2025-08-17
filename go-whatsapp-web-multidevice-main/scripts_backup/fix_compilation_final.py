#!/usr/bin/env python3
"""
Fix compilation errors by properly handling SQL queries
"""

import os
import re

def fix_go_file_sql_properly(filepath):
    """Fix SQL queries in Go files properly"""
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # First, remove all backticks from field names in struct literals
    # Pattern: word: `word`
    content = re.sub(r'(\w+):\s*`(\w+)`', r'\1: \2', content)
    
    # Fix SQL queries - ensure trigger is backticked only in SQL strings
    # Find all SQL string blocks
    sql_blocks = []
    
    # Pattern 1: sql: `...`,
    pattern1 = re.compile(r'sql:\s*`([^`]+)`', re.DOTALL)
    
    # Pattern 2: db.Exec(`...`)
    pattern2 = re.compile(r'db\.(?:Exec|Query|QueryRow)\s*\(`([^`]+)`', re.DOTALL)
    
    # Pattern 3: query := `...`
    pattern3 = re.compile(r'(?:query|sql|stmt)\s*:=\s*`([^`]+)`', re.DOTALL)
    
    def fix_sql_content(sql):
        """Fix SQL content"""
        # Remove existing backticks around trigger
        sql = sql.replace('`trigger`', 'trigger')
        
        # Add backticks properly in SQL contexts
        # In SET clauses
        sql = re.sub(r'SET\s+trigger\s*=', 'SET `trigger` =', sql)
        # In WHERE clauses
        sql = re.sub(r'WHERE\s+trigger\s*=', 'WHERE `trigger` =', sql)
        sql = re.sub(r'WHERE\s+trigger\s+IS', 'WHERE `trigger` IS', sql)
        # In SELECT lists
        sql = re.sub(r'SELECT\s+(.*?)trigger([,\s])', r'SELECT \1`trigger`\2', sql)
        # After commas
        sql = re.sub(r',\s*trigger([,\s])', r', `trigger`\1', sql)
        
        return sql
    
    # Fix pattern 1
    def replace_sql1(match):
        sql = match.group(1)
        sql = fix_sql_content(sql)
        return f'sql: `{sql}`'
    
    content = pattern1.sub(replace_sql1, content)
    
    # Fix pattern 2
    def replace_sql2(match):
        sql = match.group(1)
        sql = fix_sql_content(sql)
        return f'db.{match.group(0).split(".")[1].split("(")[0]}(`{sql}`'
    
    content = pattern2.sub(replace_sql2, content)
    
    # Fix pattern 3
    def replace_sql3(match):
        var_name = match.group(0).split(':=')[0].strip()
        sql = match.group(1)
        sql = fix_sql_content(sql)
        return f'{var_name} := `{sql}`'
    
    content = pattern3.sub(replace_sql3, content)
    
    # Fix specific issues
    # Remove backticks from Go keywords
    content = content.replace('`case`', 'case')
    content = content.replace('`if`', 'if')
    content = content.replace('`else`', 'else')
    content = content.replace('`for`', 'for')
    content = content.replace('`func`', 'func')
    content = content.replace('`type`', 'type')
    content = content.replace('`struct`', 'struct')
    content = content.replace('`return`', 'return')
    
    # Fix PostgreSQL concatenation in SQL
    content = re.sub(r"'step_'\s*\|\|\s*id", "'step_' + CAST(id AS CHAR)", content)
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)

def main():
    base_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"
    
    # Files with compilation errors
    files_to_fix = [
        "database/emergency_fix.go",
        "database/migrate_sequence_steps.go",
        "database/migrations.go",
        "infrastructure/whatsapp/stability/ultra_stable_connection.go",
    ]
    
    print("Fixing compilation errors...")
    print("-" * 60)
    
    for relative_path in files_to_fix:
        filepath = os.path.join(base_path, relative_path)
        if os.path.exists(filepath):
            try:
                fix_go_file_sql_properly(filepath)
                print(f"[FIXED] {relative_path}")
            except Exception as e:
                print(f"[ERROR] {relative_path}: {e}")
    
    print("-" * 60)
    print("Fixes completed!")

if __name__ == "__main__":
    main()
