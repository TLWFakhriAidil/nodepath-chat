@echo off
echo Fixing PostgreSQL SQL Syntax Errors in Analytics
echo ================================================

echo.
echo Creating Python script to fix SQL syntax...

:: Create Python script to fix the issues
echo import re > fix_analytics_sql.py
echo. >> fix_analytics_sql.py
echo print("Fixing PostgreSQL SQL syntax errors...") >> fix_analytics_sql.py
echo. >> fix_analytics_sql.py
echo # Read the analytics_handlers.go file >> fix_analytics_sql.py
echo with open(r'src\ui\rest\analytics_handlers.go', 'r', encoding='utf-8') as f: >> fix_analytics_sql.py
echo     content = f.read() >> fix_analytics_sql.py
echo. >> fix_analytics_sql.py
echo # Fix 1: Replace backticks around SQL keywords with proper syntax >> fix_analytics_sql.py
echo # PostgreSQL doesn't need backticks around keywords >> fix_analytics_sql.py
echo content = content.replace('`from`', 'FROM') >> fix_analytics_sql.py
echo content = content.replace('`order`', 'ORDER') >> fix_analytics_sql.py
echo. >> fix_analytics_sql.py
echo # Fix 2: Remove undefined argCount variable >> fix_analytics_sql.py
echo content = re.sub(r'\n\s*argCount = \d+\n', '\n', content) >> fix_analytics_sql.py
echo. >> fix_analytics_sql.py
echo # Fix 3: Fix the SQL queries - they're using PostgreSQL so we need proper syntax >> fix_analytics_sql.py
echo # The issue is extra spaces and incorrect FROM keyword >> fix_analytics_sql.py
echo content = content.replace('SELECT COUNT(DISTINCT c.id) `from` campaigns', 'SELECT COUNT(DISTINCT c.id) FROM campaigns') >> fix_analytics_sql.py
echo content = content.replace('SELECT COUNT(DISTINCT s.id) `from` sequences', 'SELECT COUNT(DISTINCT s.id) FROM sequences') >> fix_analytics_sql.py
echo content = content.replace('SELECT COUNT(*) `from` sequence_steps', 'SELECT COUNT(*) FROM sequence_steps') >> fix_analytics_sql.py
echo content = content.replace('SELECT COUNT(*) `from` sequence_contacts', 'SELECT COUNT(*) FROM sequence_contacts') >> fix_analytics_sql.py
echo. >> fix_analytics_sql.py
echo # Fix 4: Fix the INNER JOIN queries >> fix_analytics_sql.py
echo content = content.replace('from broadcast_messages bm', 'FROM broadcast_messages bm') >> fix_analytics_sql.py
echo content = content.replace('from sequence_contacts sc', 'FROM sequence_contacts sc') >> fix_analytics_sql.py
echo content = content.replace('from sequences WHERE', 'FROM sequences WHERE') >> fix_analytics_sql.py
echo content = content.replace('from information_schema.columns', 'FROM information_schema.columns') >> fix_analytics_sql.py
echo. >> fix_analytics_sql.py
echo # Write the fixed content back >> fix_analytics_sql.py
echo with open(r'src\ui\rest\analytics_handlers.go', 'w', encoding='utf-8') as f: >> fix_analytics_sql.py
echo     f.write(content) >> fix_analytics_sql.py
echo. >> fix_analytics_sql.py
echo print("Fixed all PostgreSQL SQL syntax errors!") >> fix_analytics_sql.py

:: Run the Python script
python fix_analytics_sql.py

:: Clean up
del fix_analytics_sql.py

echo.
echo ================================================
echo Fix completed!
echo.
echo Fixed issues:
echo 1. Removed MySQL-style backticks from SQL keywords
echo 2. Fixed FROM keyword capitalization
echo 3. Removed undefined argCount variable
echo 4. Corrected SQL query syntax for PostgreSQL
echo.
pause
