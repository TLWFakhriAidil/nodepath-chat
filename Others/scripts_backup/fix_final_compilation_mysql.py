#!/usr/bin/env python3
"""
Fix all compilation errors in Go source files for MySQL
"""

import os
import re

def fix_migrations_go():
    """Fix database/migrations.go"""
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\database\migrations.go"
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Fix comments with backticks
    content = content.replace("-- Create `trigger` to", "-- Create trigger to")
    content = content.replace("-- Add `trigger` column", "-- Add trigger column")
    content = content.replace("-- Create index for `trigger`", "-- Create index for trigger")
    
    # Fix SQL with wrong backtick usage
    # In ALTER TABLE statements
    content = re.sub(r'ALTER TABLE (\w+) ADD COLUMN IF NOT EXISTS `trigger`', r'ALTER TABLE \1 ADD COLUMN IF NOT EXISTS trigger', content)
    
    # Fix WHERE clauses
    content = re.sub(r'WHERE `trigger` IS NOT NULL', r'WHERE trigger IS NOT NULL', content)
    
    # Fix the specific lines causing issues
    # Line 411 issue - removing backticks from column definition
    content = content.replace(
        "ALTER TABLE leads ADD COLUMN IF NOT EXISTS `trigger` VARCHAR(1000);",
        "ALTER TABLE leads ADD COLUMN IF NOT EXISTS trigger VARCHAR(1000);"
    )
    
    # Line 415 issue
    content = content.replace(
        "CREATE INDEX IF NOT EXISTS idx_leads_trigger ON leads(trigger) WHERE `trigger` IS NOT NULL;",
        "CREATE INDEX IF NOT EXISTS idx_leads_trigger ON leads(trigger) WHERE trigger IS NOT NULL;"
    )
    
    # Fix lines 469-476
    content = content.replace(
        "-- 1. Find all leads with the target `trigger`",
        "-- 1. Find all leads with the target trigger"
    )
    content = content.replace(
        "WHERE `trigger` LIKE '%' || $2 || '%'",
        "WHERE trigger LIKE CONCAT('%', ?, '%')"
    )
    content = content.replace(
        "AND `trigger` IS NOT NULL",
        "AND trigger IS NOT NULL"
    )
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print("[FIXED] database/migrations.go")

def fix_ultra_stable_connection():
    """Fix infrastructure/whatsapp/stability/ultra_stable_connection.go"""
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\infrastructure\whatsapp\stability\ultra_stable_connection.go"
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Check around line 187 for the case statement issue
    # Replace any backticked Go keywords
    content = re.sub(r'`(case|if|else|for|switch|return|break|continue|type|func|var|const)`', r'\1', content)
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print("[FIXED] infrastructure/whatsapp/stability/ultra_stable_connection.go")

def main():
    print("Fixing remaining compilation errors...")
    print("-" * 60)
    
    fix_migrations_go()
    fix_ultra_stable_connection()
    
    print("-" * 60)
    print("Fixes completed!")
    print("\nNow run: build_local.bat")

if __name__ == "__main__":
    main()
