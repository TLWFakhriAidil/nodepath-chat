#!/usr/bin/env python3
"""
Targeted MySQL fixes - only modify actual SQL query strings
"""

import os
import re

def fix_sql_in_go_files(filepath):
    """Fix MySQL syntax only in SQL query strings"""
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    original = content
    
    # Find SQL query blocks (between backticks or quotes containing SQL keywords)
    # Pattern to match SQL queries in Go
    sql_patterns = [
        # Backtick strings
        (r'`([^`]*\b(?:SELECT|INSERT|UPDATE|DELETE|FROM|WHERE|SET|ORDER BY|GROUP BY|HAVING|JOIN|LEFT JOIN|RIGHT JOIN|INNER JOIN)\b[^`]*)`', '`'),
        # Double quote strings
        (r'"([^"]*\b(?:SELECT|INSERT|UPDATE|DELETE|FROM|WHERE|SET|ORDER BY|GROUP BY|HAVING|JOIN|LEFT JOIN|RIGHT JOIN|INNER JOIN)\b[^"]*)"', '"'),
    ]
    
    for pattern, quote in sql_patterns:
        def fix_sql(match):
            sql = match.group(1)
            original_sql = sql
            
            # Fix trigger keyword in SQL contexts only
            # Look for patterns where trigger is used as a column
            sql = re.sub(r'\b(SELECT\s+.*?)(\s+)trigger(\s*[,\s])', r'\1\2`trigger`\3', sql, flags=re.IGNORECASE)
            sql = re.sub(r'(\s+)trigger(\s+FROM\b)', r'\1`trigger`\2', sql, flags=re.IGNORECASE)
            sql = re.sub(r'(,\s*)trigger(\s*[,)])', r'\1`trigger`\2', sql, flags=re.IGNORECASE)
            sql = re.sub(r'\b(WHERE\s+)trigger(\s*[=<>])', r'\1`trigger`\2', sql, flags=re.IGNORECASE)
            sql = re.sub(r'\b(AND\s+)trigger(\s*[=<>])', r'\1`trigger`\2', sql, flags=re.IGNORECASE)
            sql = re.sub(r'\b(SET\s+)trigger(\s*=)', r'\1`trigger`\2', sql, flags=re.IGNORECASE)
            sql = re.sub(r'\b(ORDER\s+BY\s+)trigger\b', r'\1`trigger`', sql, flags=re.IGNORECASE)
            sql = re.sub(r'\btrigger(\s+IS\s+(?:NOT\s+)?NULL)', r'`trigger`\1', sql, flags=re.IGNORECASE)
            
            # Fix INTERVAL syntax
            sql = re.sub(r"CURRENT_TIMESTAMP\s*-\s*INTERVAL\s*'(\d+)\s*hour[s]?'", r"DATE_SUB(CURRENT_TIMESTAMP, INTERVAL \1 HOUR)", sql, flags=re.IGNORECASE)
            sql = re.sub(r"CURRENT_TIMESTAMP\s*-\s*INTERVAL\s*'(\d+)\s*day[s]?'", r"DATE_SUB(CURRENT_TIMESTAMP, INTERVAL \1 DAY)", sql, flags=re.IGNORECASE)
            sql = re.sub(r"CURRENT_TIMESTAMP\s*-\s*INTERVAL\s*'(\d+)\s*minute[s]?'", r"DATE_SUB(CURRENT_TIMESTAMP, INTERVAL \1 MINUTE)", sql, flags=re.IGNORECASE)
            sql = re.sub(r"NOW\(\)\s*-\s*INTERVAL\s*'(\d+)\s*hour[s]?'", r"DATE_SUB(NOW(), INTERVAL \1 HOUR)", sql, flags=re.IGNORECASE)
            sql = re.sub(r"NOW\(\)\s*-\s*INTERVAL\s*'(\d+)\s*day[s]?'", r"DATE_SUB(NOW(), INTERVAL \1 DAY)", sql, flags=re.IGNORECASE)
            
            # Fix || concatenation
            sql = re.sub(r"'%'\s*\|\|\s*\?\s*\|\|\s*'%'", r"CONCAT('%', ?, '%')", sql)
            
            # Fix parameter placeholders
            sql = re.sub(r'\$(\d+)', r'?', sql)
            
            # Fix ILIKE
            sql = re.sub(r'\bILIKE\b', 'LIKE', sql, flags=re.IGNORECASE)
            
            # Fix boolean values
            sql = re.sub(r'\bTRUE\b', '1', sql)
            sql = re.sub(r'\bFALSE\b', '0', sql)
            
            return quote + sql + quote
        
        content = re.sub(pattern, fix_sql, content, flags=re.DOTALL)
    
    if content != original:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        return True
    return False

def main():
    base_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"
    
    # Priority files that need SQL fixes
    priority_files = [
        "repository/lead_repository.go",
        "repository/campaign_repository.go", 
        "repository/sequence_repository.go",
        "repository/broadcast_repository.go",
        "repository/user_repository.go",
        "repository/lead_ai_repository.go",
        "ui/rest/analytics_handlers.go",
        "ui/rest/app.go",
        "ui/rest/team_member_handlers.go",
        "usecase/queued_message_cleaner.go",
        "usecase/direct_broadcast_processor.go",
        "usecase/broadcast_scheduler.go",
    ]
    
    print("Applying targeted MySQL fixes to SQL queries only...")
    print("-" * 60)
    
    fixed_count = 0
    for relative_path in priority_files:
        filepath = os.path.join(base_path, relative_path)
        if os.path.exists(filepath):
            if fix_sql_in_go_files(filepath):
                print(f"[FIXED] {relative_path}")
                fixed_count += 1
            else:
                print(f"[OK] {relative_path}")
    
    print("-" * 60)
    print(f"Fixed {fixed_count} files")
    
    # Check for any remaining issues
    print("\nChecking for remaining PostgreSQL syntax...")
    
    for relative_path in priority_files:
        filepath = os.path.join(base_path, relative_path)
        if os.path.exists(filepath):
            with open(filepath, 'r') as f:
                content = f.read()
            
            issues = []
            # Check in SQL strings only
            sql_content = ' '.join(re.findall(r'`([^`]+)`', content))
            
            if "INTERVAL '" in sql_content:
                issues.append("INTERVAL syntax")
            if re.search(r'\$\d+', sql_content):
                issues.append("$ parameters")
            if " || " in sql_content:
                issues.append("|| concatenation")
            if " ILIKE " in sql_content.upper():
                issues.append("ILIKE")
                
            if issues:
                print(f"[WARN] {relative_path}: {', '.join(issues)}")

if __name__ == "__main__":
    main()
