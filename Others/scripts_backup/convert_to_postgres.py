#!/usr/bin/env python3
import os
import re
import glob

def convert_sqlite_to_postgres(file_path):
    """Convert SQLite placeholders to PostgreSQL and fix SQL syntax"""
    
    # Read the file
    with open(file_path, 'r', encoding='utf-8') as f:
        content = f.read()
    
    original_content = content
    
    # Find all SQL queries with ? placeholders
    # Look for patterns like "query := `...`" or similar
    query_pattern = r'(query\s*:?=\s*`[^`]+`|\.Query\(["`][^"`]+["`]|\.QueryRow\(["`][^"`]+["`]|\.Exec\(["`][^"`]+["`])'
    
    matches = list(re.finditer(query_pattern, content, re.DOTALL))
    
    # Process from end to start to maintain positions
    for match in reversed(matches):
        query_text = match.group(0)
        
        # Count the number of ? in this query
        question_marks = query_text.count('?')
        
        if question_marks > 0:
            # Replace ? with $1, $2, etc.
            modified_query = query_text
            for i in range(1, question_marks + 1):
                # Replace the first ? with $i
                modified_query = modified_query.replace('?', f'${i}', 1)
            
            # Replace in the content
            content = content[:match.start()] + modified_query + content[match.end():]
    
    # Fix ON CONFLICT syntax for PostgreSQL
    # SQLite: ON CONFLICT IGNORE
    # PostgreSQL: ON CONFLICT DO NOTHING
    content = re.sub(r'ON CONFLICT IGNORE', 'ON CONFLICT DO NOTHING', content)
    
    # SQLite: ON CONFLICT REPLACE
    # PostgreSQL: ON CONFLICT ... DO UPDATE SET
    if 'ON CONFLICT REPLACE' in content:
        # This needs manual review as it depends on the specific columns
        print(f"WARNING: {file_path} contains ON CONFLICT REPLACE which needs manual review")
    
    # SQLite: AUTOINCREMENT
    # PostgreSQL: SERIAL or IDENTITY
    content = re.sub(r'INTEGER\s+AUTOINCREMENT', 'SERIAL', content, flags=re.IGNORECASE)
    
    # Save the file if there were changes
    if content != original_content:
        with open(file_path, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"[CONVERTED] {file_path}")
        return True
    return False

def fix_duplicate_methods(file_path):
    """Fix duplicate method declarations"""
    
    with open(file_path, 'r', encoding='utf-8') as f:
        lines = f.readlines()
    
    # Track method signatures
    method_signatures = {}
    lines_to_remove = set()
    
    i = 0
    while i < len(lines):
        line = lines[i]
        
        # Check for function declaration
        func_match = re.match(r'^func\s+\([^)]+\)\s+(\w+)\s*\([^)]*\)', line)
        if func_match:
            method_name = func_match.group(1)
            
            if method_name in method_signatures:
                # Found duplicate, mark lines to remove
                print(f"Found duplicate method: {method_name} at line {i+1}")
                
                # Find the end of this method
                brace_count = 0
                start_line = i
                
                # Look for opening brace
                j = i
                while j < len(lines) and '{' not in lines[j]:
                    j += 1
                
                if j < len(lines):
                    # Count braces to find method end
                    while j < len(lines):
                        brace_count += lines[j].count('{')
                        brace_count -= lines[j].count('}')
                        
                        if brace_count == 0 and '}' in lines[j]:
                            # Found end of method
                            for k in range(start_line, j + 1):
                                lines_to_remove.add(k)
                            break
                        j += 1
            else:
                method_signatures[method_name] = i
        
        i += 1
    
    # Remove duplicate lines
    if lines_to_remove:
        new_lines = [lines[i] for i in range(len(lines)) if i not in lines_to_remove]
        
        with open(file_path, 'w', encoding='utf-8') as f:
            f.writelines(new_lines)
        
        print(f"[FIXED] Removed {len(lines_to_remove)} lines from {file_path}")
        return True
    
    return False

def main():
    # Get all Go files in repository
    repo_path = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository"
    
    if not os.path.exists(repo_path):
        print(f"Error: Repository path not found: {repo_path}")
        return
    
    go_files = glob.glob(os.path.join(repo_path, "*.go"))
    
    print(f"Found {len(go_files)} Go files in repository")
    
    # First fix duplicate methods in user_repository.go
    user_repo_path = os.path.join(repo_path, "user_repository.go")
    if os.path.exists(user_repo_path):
        print("\nFixing duplicate methods in user_repository.go...")
        fix_duplicate_methods(user_repo_path)
    
    # Convert all files
    print("\nConverting SQLite syntax to PostgreSQL...")
    converted_count = 0
    
    for file_path in go_files:
        if convert_sqlite_to_postgres(file_path):
            converted_count += 1
    
    print(f"\n[DONE] Conversion complete! Modified {converted_count} files")

if __name__ == "__main__":
    main()
