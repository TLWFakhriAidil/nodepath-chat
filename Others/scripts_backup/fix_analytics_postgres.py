#!/usr/bin/env python3
"""
Fix analytics_handlers.go PostgreSQL syntax
"""

import os
import re

def fix_analytics_handlers():
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\ui\rest\analytics_handlers.go"
    
    if not os.path.exists(filepath):
        print(f"File not found: {filepath}")
        return False
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    original = content
    
    # Fix parameter placeholders $1, $2, etc to ?
    content = re.sub(r'\$\d+', '?', content)
    
    # Fix fmt.Sprintf with $%d to use ?
    content = re.sub(r'fmt\.Sprintf\(["\']([^"\']*)\$%d([^"\']*)["\'],\s*\w+\)', r'"\1?\2"', content)
    
    # Fix specific patterns in analytics queries
    # Fix: broadcastQuery += fmt.Sprintf(" AND u.id = $%d", argCount)
    content = re.sub(r'(\w+Query)\s*\+=\s*fmt\.Sprintf\(["\'][^"\']*\$%d[^"\']*["\'],\s*\w+\)', r'\1 += " AND u.id = ?"', content)
    content = re.sub(r'(\w+Query)\s*\+=\s*fmt\.Sprintf\(["\'][^"\']*\$%d[^"\']*["\'],\s*\w+\)', r'\1 += " AND c.niche = ?"', content)
    
    # Fix argCount usage - remove it as MySQL doesn't need it
    content = re.sub(r'argCount\s*:=\s*\d+\s*\n', '', content)
    content = re.sub(r'argCount\+\+\s*\n', '', content)
    
    # Fix specific query patterns that might have issues
    # Ensure all AND clauses have proper conditions
    content = re.sub(r'WHERE\s+AND\s+', 'WHERE ', content)
    content = re.sub(r'AND\s+\n\s*\)', ')', content)
    
    if content != original:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        print("[FIXED] analytics_handlers.go")
        return True
    else:
        print("[OK] analytics_handlers.go - no changes needed")
        return False

def fix_all_go_files_with_dollar_params():
    """Find and fix all Go files using $1, $2 parameter syntax"""
    base_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"
    fixed_files = []
    
    for root, dirs, files in os.walk(base_path):
        for file in files:
            if file.endswith('.go'):
                filepath = os.path.join(root, file)
                try:
                    with open(filepath, 'r', encoding='utf-8') as f:
                        content = f.read()
                    
                    # Check if file contains PostgreSQL parameter syntax
                    if re.search(r'\$\d+', content):
                        original = content
                        
                        # Fix parameter placeholders
                        content = re.sub(r'\$\d+', '?', content)
                        
                        # Fix dynamic query building with fmt.Sprintf
                        content = re.sub(r'fmt\.Sprintf\("([^"]*)\$%d([^"]*)"\s*,\s*\w+\)', r'"\1?\2"', content)
                        content = re.sub(r"fmt\.Sprintf\('([^']*)\$%d([^']*)'\s*,\s*\w+\)", r"'\1?\2'", content)
                        
                        # Remove argCount patterns
                        content = re.sub(r'argCount\s*:=\s*len\([^)]+\)\s*\+\s*\d+\s*\n', '', content)
                        content = re.sub(r'argCount\+\+\s*\n', '', content)
                        
                        if content != original:
                            with open(filepath, 'w', encoding='utf-8') as f:
                                f.write(content)
                            relative_path = os.path.relpath(filepath, base_path)
                            fixed_files.append(relative_path)
                            print(f"[FIXED] {relative_path}")
                except Exception as e:
                    print(f"[ERROR] {filepath}: {e}")
    
    return fixed_files

def main():
    print("Fixing PostgreSQL parameter syntax in Go files...")
    print("-" * 50)
    
    # First fix analytics_handlers.go specifically
    fix_analytics_handlers()
    
    print("\nSearching for other files with PostgreSQL syntax...")
    print("-" * 50)
    
    # Fix all other files
    fixed_files = fix_all_go_files_with_dollar_params()
    
    print("-" * 50)
    print(f"Total files fixed: {len(fixed_files)}")
    
    if fixed_files:
        print("\nFixed files:")
        for f in fixed_files:
            print(f"  - {f}")

if __name__ == "__main__":
    main()
