#!/usr/bin/env python3
"""
Fix MySQL-specific syntax errors in all Go files
"""

import os
import re
import glob

def fix_interval_syntax(content):
    """Fix PostgreSQL INTERVAL syntax to MySQL DATE_SUB"""
    # PostgreSQL: updated_at < (CURRENT_TIMESTAMP - INTERVAL '12 hours')
    # MySQL: updated_at < DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 12 HOUR)
    
    patterns = [
        # Fix INTERVAL '12 hours'
        (r"CURRENT_TIMESTAMP\s*-\s*INTERVAL\s*'(\d+)\s*hours?'", r"DATE_SUB(CURRENT_TIMESTAMP, INTERVAL \1 HOUR)"),
        (r"NOW\(\)\s*-\s*INTERVAL\s*'(\d+)\s*hours?'", r"DATE_SUB(NOW(), INTERVAL \1 HOUR)"),
        
        # Fix INTERVAL '30 days'
        (r"CURRENT_TIMESTAMP\s*-\s*INTERVAL\s*'(\d+)\s*days?'", r"DATE_SUB(CURRENT_TIMESTAMP, INTERVAL \1 DAY)"),
        (r"NOW\(\)\s*-\s*INTERVAL\s*'(\d+)\s*days?'", r"DATE_SUB(NOW(), INTERVAL \1 DAY)"),
        
        # Fix INTERVAL '5 minutes'
        (r"CURRENT_TIMESTAMP\s*-\s*INTERVAL\s*'(\d+)\s*minutes?'", r"DATE_SUB(CURRENT_TIMESTAMP, INTERVAL \1 MINUTE)"),
        (r"NOW\(\)\s*-\s*INTERVAL\s*'(\d+)\s*minutes?'", r"DATE_SUB(NOW(), INTERVAL \1 MINUTE)"),
        
        # Fix INTERVAL with addition
        (r"CURRENT_TIMESTAMP\s*\+\s*INTERVAL\s*'(\d+)\s*hours?'", r"DATE_ADD(CURRENT_TIMESTAMP, INTERVAL \1 HOUR)"),
        (r"NOW\(\)\s*\+\s*INTERVAL\s*'(\d+)\s*hours?'", r"DATE_ADD(NOW(), INTERVAL \1 HOUR)"),
        (r"CURRENT_TIMESTAMP\s*\+\s*INTERVAL\s*'(\d+)\s*days?'", r"DATE_ADD(CURRENT_TIMESTAMP, INTERVAL \1 DAY)"),
        (r"NOW\(\)\s*\+\s*INTERVAL\s*'(\d+)\s*days?'", r"DATE_ADD(NOW(), INTERVAL \1 DAY)"),
    ]
    
    for pattern, replacement in patterns:
        content = re.sub(pattern, replacement, content, flags=re.IGNORECASE)
    
    return content

def fix_empty_in_clause(content):
    """Fix empty IN() clauses which are invalid in MySQL"""
    # Replace IN () with IN (NULL) or FALSE condition
    content = re.sub(r'IN\s*\(\s*\)', 'IN (NULL)', content)
    
    # Better approach - replace WHERE x IN () with WHERE FALSE
    content = re.sub(r'WHERE\s+(\w+)\s+IN\s*\(\s*\)', r'WHERE FALSE -- \1 IN empty list', content)
    content = re.sub(r'AND\s+(\w+)\s+IN\s*\(\s*\)', r'AND FALSE -- \1 IN empty list', content)
    
    return content

def fix_limit_offset_syntax(content):
    """Fix LIMIT ? OFFSET ? to LIMIT ?, ?"""
    content = re.sub(r'LIMIT\s*\?\s*OFFSET\s*\?', 'LIMIT ?, ?', content)
    return content

def fix_mysql_reserved_keywords(content):
    """Add backticks to MySQL reserved keywords in queries"""
    # List of MySQL reserved keywords that might be used as column names
    keywords = ['trigger', 'limit', 'order', 'key', 'status', 'type', 'from', 'to']
    
    for keyword in keywords:
        # In SELECT clauses
        content = re.sub(rf'SELECT\s+([^`]*\s+){keyword}(\s*[,\s])', rf'SELECT \1`{keyword}`\2', content, flags=re.IGNORECASE)
        # In WHERE clauses  
        content = re.sub(rf'WHERE\s+{keyword}\s*=', rf'WHERE `{keyword}` =', content, flags=re.IGNORECASE)
        content = re.sub(rf'AND\s+{keyword}\s*=', rf'AND `{keyword}` =', content, flags=re.IGNORECASE)
        # In INSERT
        content = re.sub(rf'INSERT\s+INTO\s+(\w+)\s*\(([^)]*){keyword}', rf'INSERT INTO \1(\2`{keyword}`', content, flags=re.IGNORECASE)
        # In UPDATE
        content = re.sub(rf'UPDATE\s+(\w+)\s+SET\s+{keyword}\s*=', rf'UPDATE \1 SET `{keyword}` =', content, flags=re.IGNORECASE)
    
    return content

def fix_file(filepath):
    """Fix all MySQL issues in a single file"""
    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original = content
        
        # Apply all fixes
        content = fix_interval_syntax(content)
        content = fix_empty_in_clause(content)
        content = fix_limit_offset_syntax(content)
        content = fix_mysql_reserved_keywords(content)
        
        # Additional fixes
        # Fix PostgreSQL || concatenation
        content = re.sub(r"'%'\s*\|\|\s*\?\s*\|\|\s*'%'", "CONCAT('%', ?, '%')", content)
        
        # Fix ILIKE to LIKE
        content = re.sub(r'\bILIKE\b', 'LIKE', content, flags=re.IGNORECASE)
        
        # Fix TRUE/FALSE
        content = re.sub(r'\bTRUE\b', '1', content)
        content = re.sub(r'\bFALSE\b', '0', content)
        
        # Fix $1, $2 parameters
        content = re.sub(r'\$\d+', '?', content)
        
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
    
    print("Comprehensive MySQL syntax fix for all Go files...")
    print("=" * 70)
    
    # Process all .go files
    fixed_count = 0
    checked_count = 0
    
    for root, dirs, files in os.walk(base_path):
        for file in files:
            if file.endswith('.go'):
                filepath = os.path.join(root, file)
                relative_path = os.path.relpath(filepath, base_path)
                checked_count += 1
                
                if fix_file(filepath):
                    print(f"[FIXED] {relative_path}")
                    fixed_count += 1
    
    print("=" * 70)
    print(f"Checked {checked_count} files, fixed {fixed_count} files")
    
    # List of critical files to double-check
    critical_files = [
        "usecase/queued_message_cleaner.go",
        "ui/rest/analytics_handlers.go",
        "repository/lead_repository.go",
        "repository/campaign_repository.go",
        "repository/sequence_repository.go",
        "repository/broadcast_repository.go",
    ]
    
    print("\nDouble-checking critical files...")
    print("-" * 50)
    
    for relative_path in critical_files:
        filepath = os.path.join(base_path, relative_path)
        if os.path.exists(filepath):
            with open(filepath, 'r', encoding='utf-8') as f:
                content = f.read()
            
            # Check for common issues
            issues = []
            if "INTERVAL '" in content:
                issues.append("PostgreSQL INTERVAL syntax")
            if re.search(r'\$\d+', content):
                issues.append("PostgreSQL parameters")
            if " IN ()" in content:
                issues.append("Empty IN clause")
            if " ILIKE " in content:
                issues.append("ILIKE operator")
            
            if issues:
                print(f"[WARN] {relative_path} still has issues: {', '.join(issues)}")
            else:
                print(f"[OK] {relative_path}")

if __name__ == "__main__":
    main()
