#!/usr/bin/env python3
"""
Fix repository files with broken backticks in SQL column names
"""

import os
import re

def fix_broken_backticks(content):
    """Fix backticks that are breaking up words"""
    
    # Fix common patterns where backticks are breaking words
    patterns = [
        # Fix message_`type` -> message_type
        (r'message_`type`', 'message_type'),
        # Fix group_`order` -> group_order
        (r'group_`order`', 'group_order'),
        # Fix `status` when it's a column name
        (r'`status`', 'status'),
        # Fix `limit` when it's a parameter
        (r'`limit`', 'limit'),
        # Fix `order` when it's part of group_order
        (r'`order`', 'order'),
        # Fix `type` when it's part of message_type
        (r'`type`', 'type'),
        # Fix `key` if it appears
        (r'`key`', 'key'),
    ]
    
    for old, new in patterns:
        content = content.replace(old, new)
    
    # Now add backticks properly for MySQL reserved words in SQL contexts
    # In SQL strings (between backticks or quotes)
    sql_blocks = re.findall(r'`[^`]+`', content, re.DOTALL)
    
    for block in sql_blocks:
        if any(keyword in block.upper() for keyword in ['SELECT', 'INSERT', 'UPDATE', 'DELETE', 'FROM', 'WHERE']):
            # This is an SQL block
            fixed_block = block
            
            # Add backticks to reserved words in SQL contexts
            # For trigger column
            fixed_block = re.sub(r'\btrigger\b(?!`)', '`trigger`', fixed_block)
            
            # For order in ORDER BY
            fixed_block = re.sub(r'\bORDER\s+BY\s+order\b', 'ORDER BY `order`', fixed_block)
            
            # For key column
            fixed_block = re.sub(r'\bkey\s*=', '`key` =', fixed_block)
            
            content = content.replace(block, fixed_block)
    
    return content

def fix_repository_files():
    """Fix all repository files"""
    base_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository"
    
    files_to_fix = [
        "broadcast_repository.go",
        "campaign_repository.go",
        "lead_repository.go",
        "sequence_repository.go",
        "user_repository.go"
    ]
    
    print("Fixing repository files...")
    print("-" * 60)
    
    for filename in files_to_fix:
        filepath = os.path.join(base_path, filename)
        
        if not os.path.exists(filepath):
            print(f"[SKIP] {filename} - not found")
            continue
            
        try:
            with open(filepath, 'r', encoding='utf-8') as f:
                content = f.read()
            
            original = content
            content = fix_broken_backticks(content)
            
            if content != original:
                with open(filepath, 'w', encoding='utf-8') as f:
                    f.write(content)
                print(f"[FIXED] {filename}")
            else:
                print(f"[OK] {filename}")
                
        except Exception as e:
            print(f"[ERROR] {filename}: {e}")
    
    print("-" * 60)
    print("Repository fixes completed!")

if __name__ == "__main__":
    fix_repository_files()
