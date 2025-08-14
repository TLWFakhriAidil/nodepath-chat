#!/usr/bin/env python3
import os
import re

# Find all Go files
go_files = []
for root, dirs, files in os.walk('src'):
    for file in files:
        if file.endswith('.go'):
            go_files.append(os.path.join(root, file))

print(f"Found {len(go_files)} Go files to check")

# Pattern to find the broken platform checks
pattern = r'(\s+)Platform != ""'
replacement = r'\1if device.Platform != "" {'

fixed_count = 0

for file_path in go_files:
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
        
        original = content
        
        # Fix the broken platform checks
        content = re.sub(pattern, replacement, content)
        
        # Also fix any standalone platform checks
        content = re.sub(r'(\s+)platform != ""', r'\1if platform != "" {', content)
        
        if content != original:
            with open(file_path, 'w', encoding='utf-8') as f:
                f.write(content)
            fixed_count += 1
            print(f"Fixed: {file_path}")
    except Exception as e:
        print(f"Error processing {file_path}: {e}")

print(f"\nFixed {fixed_count} files with platform check issues")