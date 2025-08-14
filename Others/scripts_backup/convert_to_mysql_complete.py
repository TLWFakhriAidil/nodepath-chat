import os
import re

BASE_DIR = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

def convert_file(filepath):
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    original = content
    
    # 1. Replace NULLS LAST (MySQL doesn't support this)
    content = re.sub(r'\s+NULLS\s+LAST', '', content)
    
    # 2. Replace $1, $2, etc with ?
    content = re.sub(r'\$(\d+)', '?', content)
    
    # 3. Replace ILIKE with LIKE (MySQL is case-insensitive by default)
    content = content.replace(' ILIKE ', ' LIKE ')
    
    # 4. Replace gen_random_uuid() with UUID()
    content = content.replace('gen_random_uuid()', 'UUID()')
    
    # 5. Replace RETURNING clauses
    content = re.sub(r'\s+RETURNING\s+\w+', '', content)
    
    # 6. Replace TRUE/FALSE with 1/0
    content = re.sub(r'\bTRUE\b', '1', content)
    content = re.sub(r'\bFALSE\b', '0', content)
    
    if content != original:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        return True
    return False

# Get all Go files
go_files = []
for root, dirs, files in os.walk(BASE_DIR):
    for file in files:
        if file.endswith('.go'):
            go_files.append(os.path.join(root, file))

print(f"Found {len(go_files)} Go files")

# Convert all files
converted = 0
for filepath in go_files:
    if convert_file(filepath):
        converted += 1
        print(f"Converted: {os.path.relpath(filepath, BASE_DIR)}")

print(f"\nTotal files converted: {converted}")
