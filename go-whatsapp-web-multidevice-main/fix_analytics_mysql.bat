@echo off
echo Fixing Analytics to Use MySQL Instead of PostgreSQL
echo ==================================================

echo.
echo Creating comprehensive fix for analytics handlers...

:: Create Python script to fix the database connection
echo import re > fix_analytics_mysql.py
echo. >> fix_analytics_mysql.py
echo print("Fixing analytics handlers to use MySQL...") >> fix_analytics_mysql.py
echo. >> fix_analytics_mysql.py
echo # Read the analytics_handlers.go file >> fix_analytics_mysql.py
echo with open(r'src\ui\rest\analytics_handlers.go', 'r', encoding='utf-8') as f: >> fix_analytics_mysql.py
echo     content = f.read() >> fix_analytics_mysql.py
echo. >> fix_analytics_mysql.py
echo # Fix 1: Change import from pq to mysql >> fix_analytics_mysql.py
echo content = content.replace('_ "github.com/lib/pq"', '_ "github.com/go-sql-driver/mysql"') >> fix_analytics_mysql.py
echo. >> fix_analytics_mysql.py
echo # Fix 2: Change all sql.Open from postgres to mysql >> fix_analytics_mysql.py
echo content = content.replace('sql.Open("postgres", config.DBURI)', 'sql.Open("mysql", config.MysqlDSN)') >> fix_analytics_mysql.py
echo. >> fix_analytics_mysql.py
echo # Fix 3: Fix SQL syntax for MySQL >> fix_analytics_mysql.py
echo # Remove backticks that were incorrectly added >> fix_analytics_mysql.py
echo content = content.replace('`from`', 'FROM') >> fix_analytics_mysql.py
echo content = content.replace('`order`', 'ORDER') >> fix_analytics_mysql.py
echo. >> fix_analytics_mysql.py
echo # Fix 4: Change PostgreSQL parameter placeholders to MySQL style >> fix_analytics_mysql.py
echo # This is more complex - need to replace $1, $2 with ? >> fix_analytics_mysql.py
echo # But in this file they're already using ? so we're good >> fix_analytics_mysql.py
echo. >> fix_analytics_mysql.py
echo # Fix 5: Remove undefined argCount variables >> fix_analytics_mysql.py
echo content = re.sub(r'\n\s*argCount = \d+\n', '\n', content) >> fix_analytics_mysql.py
echo. >> fix_analytics_mysql.py
echo # Fix 6: Ensure proper FROM capitalization >> fix_analytics_mysql.py
echo content = re.sub(r'(\s+)from(\s+)', r'\1FROM\2', content) >> fix_analytics_mysql.py
echo content = re.sub(r'(\s+)order(\s+)', r'\1ORDER\2', content) >> fix_analytics_mysql.py
echo. >> fix_analytics_mysql.py
echo # Write the fixed content back >> fix_analytics_mysql.py
echo with open(r'src\ui\rest\analytics_handlers.go', 'w', encoding='utf-8') as f: >> fix_analytics_mysql.py
echo     f.write(content) >> fix_analytics_mysql.py
echo. >> fix_analytics_mysql.py
echo print("Analytics handlers now use MySQL!") >> fix_analytics_mysql.py

:: Run the Python script
python fix_analytics_mysql.py

:: Clean up
del fix_analytics_mysql.py

echo.
echo ==================================================
echo Fix completed! Analytics now uses MySQL.
echo.
echo Now building without CGO and preparing for GitHub push...
echo.

:: Build without CGO
echo Building application without CGO...
set CGO_ENABLED=0
go build -o whatsapp.exe

echo.
echo Build complete!
echo.

:: Git operations
echo Preparing to push to GitHub...
git add -A
git commit -m "Fix: Switch analytics from PostgreSQL to MySQL - All application data should use MySQL, not PostgreSQL"

echo.
echo Ready to push! Run: git push origin main
echo.
pause
