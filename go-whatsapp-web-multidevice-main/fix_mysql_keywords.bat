@echo off
echo Fixing MySQL reserved keyword issues...

echo.
echo [1/2] Backing up files...
copy /Y "src\repository\lead_repository.go" "src\repository\lead_repository.go.bak" >nul 2>&1
copy /Y "src\usecase\queued_message_cleaner.go" "src\usecase\queued_message_cleaner.go.bak" >nul 2>&1

echo.
echo [2/2] Applying fixes...

echo.
echo Fixing lead_repository.go - escaping 'trigger' column...

:: Create a Python script to handle the complex replacements
echo import re > fix_triggers.py
echo with open(r'src\repository\lead_repository.go', 'r', encoding='utf-8') as f: >> fix_triggers.py
echo     content = f.read() >> fix_triggers.py
echo. >> fix_triggers.py
echo # The trigger column is already escaped in most places, just ensure all are escaped >> fix_triggers.py
echo # Pattern to find unescaped trigger column references >> fix_triggers.py
echo pattern = r'(?<!`)trigger(?!`)' >> fix_triggers.py
echo # Replace with escaped version >> fix_triggers.py
echo content = re.sub(pattern, '`trigger`', content) >> fix_triggers.py
echo. >> fix_triggers.py
echo with open(r'src\repository\lead_repository.go', 'w', encoding='utf-8') as f: >> fix_triggers.py
echo     f.write(content) >> fix_triggers.py

python fix_triggers.py
del fix_triggers.py

echo.
echo Fixing queued_message_cleaner.go - removing extra parenthesis...

:: Fix the SQL syntax error in queued_message_cleaner.go
powershell -Command "(Get-Content 'src\usecase\queued_message_cleaner.go') -replace 'AND updated_at < \(DATE_SUB\(CURRENT_TIMESTAMP, INTERVAL 12 HOUR\)\)', 'AND updated_at < DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 12 HOUR)' | Set-Content 'src\usecase\queued_message_cleaner.go'"

echo.
echo Done! All MySQL reserved keyword issues have been fixed.
echo.
echo Changes made:
echo 1. Escaped all 'trigger' column references with backticks in lead_repository.go
echo 2. Fixed SQL syntax error in queued_message_cleaner.go (removed extra parenthesis)
echo.
echo Please rebuild and restart the application:
echo   go build -o whatsapp.exe
echo   whatsapp.exe rest
echo.
pause
