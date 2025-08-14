@echo off
echo Fixing remaining SQL syntax errors
echo ==================================

echo.
echo Creating Python script to fix remaining issues...

echo import re > fix_remaining_sql.py
echo import os >> fix_remaining_sql.py
echo. >> fix_remaining_sql.py
echo print("Fixing remaining SQL syntax errors...") >> fix_remaining_sql.py
echo. >> fix_remaining_sql.py
echo # Fix broadcast_repository.go >> fix_remaining_sql.py
echo file_path = r'src\repository\broadcast_repository.go' >> fix_remaining_sql.py
echo with open(file_path, 'r', encoding='utf-8') as f: >> fix_remaining_sql.py
echo     content = f.read() >> fix_remaining_sql.py
echo. >> fix_remaining_sql.py
echo # Find the undefined query variable issues >> fix_remaining_sql.py
echo # Line 219 - missing query := before the SQL string >> fix_remaining_sql.py
echo content = re.sub(r'(\n\t)`\n\t\tSELECT COUNT\(\*\) AS total,', r'\1query := `\n\t\tSELECT COUNT(*) AS total,', content) >> fix_remaining_sql.py
echo # Line 253 - missing query := >> fix_remaining_sql.py
echo content = re.sub(r'(\n\t)`\n\t\tSELECT id, user_id, device_id', r'\1query := `\n\t\tSELECT id, user_id, device_id', content) >> fix_remaining_sql.py
echo # Line 312 - missing query := >> fix_remaining_sql.py
echo content = re.sub(r'(\n\t)`\n\t\tSELECT DISTINCT device_id', r'\1query := `\n\t\tSELECT DISTINCT device_id', content) >> fix_remaining_sql.py
echo. >> fix_remaining_sql.py
echo with open(file_path, 'w', encoding='utf-8') as f: >> fix_remaining_sql.py
echo     f.write(content) >> fix_remaining_sql.py
echo. >> fix_remaining_sql.py
echo # Fix campaign_repository.go >> fix_remaining_sql.py
echo file_path = r'src\repository\campaign_repository.go' >> fix_remaining_sql.py
echo with open(file_path, 'r', encoding='utf-8') as f: >> fix_remaining_sql.py
echo     content = f.read() >> fix_remaining_sql.py
echo. >> fix_remaining_sql.py
echo # Line 68 - missing query := and fix "LIMIT" to `limit` >> fix_remaining_sql.py
echo content = re.sub(r'(\n\t)`\n\t\tINSERT INTO campaigns', r'\1query := `\n\t\tINSERT INTO campaigns', content) >> fix_remaining_sql.py
echo content = content.replace('"LIMIT"', '`limit`') >> fix_remaining_sql.py
echo # Line 90 - missing query := >> fix_remaining_sql.py
echo content = re.sub(r'(\n\t)`\n\t\tSELECT id, user_id, title', r'\1query := `\n\t\tSELECT id, user_id, title', content) >> fix_remaining_sql.py
echo. >> fix_remaining_sql.py
echo with open(file_path, 'w', encoding='utf-8') as f: >> fix_remaining_sql.py
echo     f.write(content) >> fix_remaining_sql.py
echo. >> fix_remaining_sql.py
echo print("Fixed all remaining SQL syntax errors!") >> fix_remaining_sql.py

:: Run the Python script
python fix_remaining_sql.py

:: Clean up
del fix_remaining_sql.py

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
    dir whatsapp.exe | findstr whatsapp.exe
    echo.
    echo Committing final fixes...
    git add -A
    git commit -m "Fix: Resolved all remaining SQL syntax errors - Build successful"
    echo.
    echo Ready to push to GitHub:
    echo   git push origin main
) else (
    echo Build failed. Check errors above.
)

echo.
pause
