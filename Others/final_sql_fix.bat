@echo off
echo Final comprehensive fix for SQL syntax
echo =====================================

:: Create a Python script to properly fix all SQL issues
echo import re > final_sql_fix.py
echo. >> final_sql_fix.py
echo # Read campaign_repository.go >> final_sql_fix.py
echo with open(r'src\repository\campaign_repository.go', 'r', encoding='utf-8') as f: >> final_sql_fix.py
echo     content = f.read() >> final_sql_fix.py
echo. >> final_sql_fix.py
echo # The issue is that backticks inside SQL strings are being interpreted by Go >> final_sql_fix.py
echo # We need to ensure the SQL strings are properly formatted >> final_sql_fix.py
echo. >> final_sql_fix.py
echo # Replace `limit` with a different approach - use double quotes in SQL >> final_sql_fix.py
echo content = content.replace('`limit`', '"limit"') >> final_sql_fix.py
echo. >> final_sql_fix.py
echo # Save >> final_sql_fix.py
echo with open(r'src\repository\campaign_repository.go', 'w', encoding='utf-8') as f: >> final_sql_fix.py
echo     f.write(content) >> final_sql_fix.py
echo. >> final_sql_fix.py
echo print("Fixed limit syntax") >> final_sql_fix.py

python final_sql_fix.py
del final_sql_fix.py

echo.
echo Building application...
cd src
set CGO_ENABLED=0
go build -o ..\whatsapp.exe
cd ..

echo.
if exist whatsapp.exe (
    echo =============================
    echo BUILD SUCCESSFUL!
    echo =============================
    echo.
    echo Executable created: whatsapp.exe
    echo.
    echo To push to GitHub:
    echo   git push origin main
) else (
    echo Build failed.
)

pause
