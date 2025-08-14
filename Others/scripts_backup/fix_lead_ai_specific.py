#!/usr/bin/env python3
"""
Fix lead_ai_repository.go syntax errors
"""

import re

def fix_lead_ai_repository():
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\repository\lead_ai_repository.go"
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Fix specific syntax errors
    # 1. Fix target_`status` to target_status
    content = content.replace('target_`status`', 'target_status')
    content = content.replace('`status`', 'status')
    content = content.replace('`order`', 'order')
    content = content.replace('`from`', 'from')
    
    # 2. Fix malformed INSERT query
    content = re.sub(
        r'INSERT INTO leads_ai\(user_id, name, phone, email, niche, source, target_status, notes\)\s*VALUES \(\?, \?, \?, \?, \?, \?, \?, \?\), created_at, updated_at',
        'INSERT INTO leads_ai(user_id, name, phone, email, niche, source, target_status, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())',
        content
    )
    
    # 3. Fix SQL keywords in queries
    # Replace backticked keywords that are at end of line
    content = re.sub(r'`(status|order|from)`\s*$', r'\1', content, flags=re.MULTILINE)
    
    # 4. Fix SELECT queries - ensure proper syntax
    # Find queries and fix them
    def fix_query(match):
        query = match.group(0)
        # Ensure SQL keywords are uppercase
        query = re.sub(r'\bselect\b', 'SELECT', query, flags=re.IGNORECASE)
        query = re.sub(r'\bfrom\b', 'FROM', query, flags=re.IGNORECASE)
        query = re.sub(r'\bwhere\b', 'WHERE', query, flags=re.IGNORECASE)
        query = re.sub(r'\band\b', 'AND', query, flags=re.IGNORECASE)
        query = re.sub(r'\bor\b', 'OR', query, flags=re.IGNORECASE)
        query = re.sub(r'\border\s+by\b', 'ORDER BY', query, flags=re.IGNORECASE)
        query = re.sub(r'\blimit\b', 'LIMIT', query, flags=re.IGNORECASE)
        return query
    
    # Apply to all query blocks
    content = re.sub(r'query\s*:=\s*`[^`]+`', fix_query, content, flags=re.DOTALL)
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    
    print("[FIXED] lead_ai_repository.go")

if __name__ == "__main__":
    fix_lead_ai_repository()
