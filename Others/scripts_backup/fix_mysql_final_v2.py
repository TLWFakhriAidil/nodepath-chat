#!/usr/bin/env python3
"""
Fix remaining MySQL compatibility issues in Go files
"""

import os
import re

def fix_postgresql_specific_code(content):
    """Remove or comment out PostgreSQL-specific code blocks"""
    
    # Remove PostgreSQL DO blocks
    content = re.sub(r'DO\s*\$\$\s*BEGIN.*?\$\$;', '-- PostgreSQL DO block removed for MySQL', content, flags=re.DOTALL)
    
    # Remove RAISE NOTICE statements
    content = re.sub(r"RAISE NOTICE.*?;", "-- RAISE NOTICE removed", content)
    
    # Fix BOOLEAN to TINYINT(1) for MySQL
    content = re.sub(r'\bBOOLEAN\b', 'TINYINT(1)', content)
    
    # Fix false/true to 0/1
    content = re.sub(r'\bfalse\b', '0', content)
    content = re.sub(r'\btrue\b', '1', content)
    
    # Remove PostgreSQL functions
    content = re.sub(r'gen_random_uuid\(\)', 'UUID()', content)
    
    return content

def fix_emergency_fix_go():
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\database\emergency_fix.go"
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Ensure proper MySQL syntax in SQL strings
    content = re.sub(r"CONCAT\('step_', id\)", "CONCAT('step_', CAST(id AS CHAR))", content)
    
    # Fix any remaining issues
    content = fix_postgresql_specific_code(content)
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print("[FIXED] database/emergency_fix.go")

def fix_migrations_go():
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\database\migrations.go"
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Comment out PostgreSQL-specific migrations
    # Find migrations with DO blocks
    content = re.sub(
        r'{\s*MigrationName:\s*"[^"]*",\s*SQL:\s*`[^`]*DO\s*\$\$[^`]*\$\$[^`]*`\s*},',
        '// PostgreSQL-specific migration commented out for MySQL',
        content,
        flags=re.DOTALL
    )
    
    # Fix specific known issues
    content = fix_postgresql_specific_code(content)
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print("[FIXED] database/migrations.go")

def fix_ultra_stable_connection():
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\infrastructure\whatsapp\stability\ultra_stable_connection.go"
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # This file shouldn't have SQL, just fix any Go syntax issues
    # Remove any backticks around Go keywords
    go_keywords = ['case', 'if', 'else', 'for', 'break', 'continue', 'return', 'switch', 'default', 'func', 'type', 'struct']
    
    for keyword in go_keywords:
        content = content.replace(f'`{keyword}`', keyword)
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print("[FIXED] infrastructure/whatsapp/stability/ultra_stable_connection.go")

def main():
    print("Fixing remaining MySQL compatibility issues...")
    print("-" * 60)
    
    fix_emergency_fix_go()
    fix_migrations_go()
    fix_ultra_stable_connection()
    
    print("-" * 60)
    print("Fixes completed!")

if __name__ == "__main__":
    main()
