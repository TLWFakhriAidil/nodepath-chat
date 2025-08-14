#!/usr/bin/env python3
"""
Final manual fixes for compilation errors
"""

import os

def fix_connection_go():
    """Fix database/connection.go"""
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\database\connection.go"
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Remove the trigger comment that's causing issues
    content = content.replace("-- Add `trigger` to update updated_at timestamp", "-- Add database trigger to update updated_at timestamp")
    
    # Also check if this is a PostgreSQL-specific migration that should be skipped for MySQL
    # This CREATE TRIGGER syntax is PostgreSQL-specific
    content = content.replace("""CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';

		CREATE TRIGGER update_team_members_updated_at BEFORE UPDATE
			ON team_members FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();""", 
    """-- PostgreSQL trigger removed for MySQL compatibility
		-- MySQL handles updated_at with ON UPDATE CURRENT_TIMESTAMP in column definition""")
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print("[FIXED] database/connection.go")

def fix_emergency_fix_go():
    """Fix database/emergency_fix.go"""
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\database\emergency_fix.go"
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Fix struct field names that were incorrectly backticked
    content = content.replace("`trigger`:", "trigger:")
    content = content.replace("`from`:", "from:")
    content = content.replace("`to`:", "to:")
    content = content.replace("`status`:", "status:")
    content = content.replace("`type`:", "type:")
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print("[FIXED] database/emergency_fix.go")

def fix_migrations_go():
    """Fix database/migrations.go"""
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\database\migrations.go"
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Fix struct field names
    content = content.replace("`trigger`:", "trigger:")
    content = content.replace("`from`:", "from:")
    content = content.replace("`to`:", "to:")
    content = content.replace("`status`:", "status:")
    content = content.replace("`type`:", "type:")
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print("[FIXED] database/migrations.go")

def fix_migrate_sequence_steps():
    """Fix database/migrate_sequence_steps.go"""
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\database\migrate_sequence_steps.go"
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Fix SQL query to have proper backticks only in SQL context
    content = content.replace("SELECT id, `trigger`", "SELECT id, `trigger`")
    content = content.replace("`trigger`\n\tFROM", "`trigger` FROM")
    
    # Fix any standalone trigger that's not in SQL
    lines = content.split('\n')
    fixed_lines = []
    in_sql = False
    
    for line in lines:
        if '`SELECT' in line or '`INSERT' in line or '`UPDATE' in line:
            in_sql = True
        elif '`' in line and in_sql and line.strip().endswith('`'):
            in_sql = False
        
        # If not in SQL and has standalone `trigger`
        if not in_sql and line.strip() == '`trigger`':
            line = line.replace('`trigger`', 'trigger')
        
        fixed_lines.append(line)
    
    content = '\n'.join(fixed_lines)
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print("[FIXED] database/migrate_sequence_steps.go")

def fix_ultra_stable_connection():
    """Fix infrastructure/whatsapp/stability/ultra_stable_connection.go"""
    filepath = r"C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src\infrastructure\whatsapp\stability\ultra_stable_connection.go"
    
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()
    
    # Fix backticked keywords in Go code
    content = content.replace("`case`", "case")
    content = content.replace("`if`", "if")
    content = content.replace("`from`", "from")
    content = content.replace("`to`", "to")
    content = content.replace("`type`", "type")
    
    with open(filepath, 'w', encoding='utf-8') as f:
        f.write(content)
    print("[FIXED] infrastructure/whatsapp/stability/ultra_stable_connection.go")

def main():
    print("Applying final manual fixes for compilation errors...")
    print("-" * 60)
    
    fix_connection_go()
    fix_emergency_fix_go()
    fix_migrations_go()
    fix_migrate_sequence_steps()
    fix_ultra_stable_connection()
    
    print("-" * 60)
    print("Manual fixes completed!")
    print("\nNow run: build_local.bat")

if __name__ == "__main__":
    main()
