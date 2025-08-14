@echo off
echo Comprehensive MySQL Reserved Keywords Fix
echo ========================================

:: Check if Python is available
python --version >nul 2>&1
if %errorlevel% neq 0 (
    echo Error: Python is not installed or not in PATH
    echo Please install Python or use manual fix
    pause
    exit /b 1
)

echo.
echo Creating comprehensive fix script...

:: Create a comprehensive Python script to fix all issues
echo import os > comprehensive_fix.py
echo import re >> comprehensive_fix.py
echo. >> comprehensive_fix.py
echo print("Fixing MySQL reserved keyword issues...") >> comprehensive_fix.py
echo. >> comprehensive_fix.py
echo # Fix 1: Check and fix queued_message_cleaner.go >> comprehensive_fix.py
echo print("\n1. Checking queued_message_cleaner.go...") >> comprehensive_fix.py
echo cleaner_path = r'src\usecase\queued_message_cleaner.go' >> comprehensive_fix.py
echo if os.path.exists(cleaner_path): >> comprehensive_fix.py
echo     with open(cleaner_path, 'r', encoding='utf-8') as f: >> comprehensive_fix.py
echo         content = f.read() >> comprehensive_fix.py
echo     # Check if already fixed >> comprehensive_fix.py
echo     if 'AND updated_at ^< (DATE_SUB' in content: >> comprehensive_fix.py
echo         print("  - Fixing extra parenthesis in SQL query...") >> comprehensive_fix.py
echo         content = content.replace('AND updated_at ^< (DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 12 HOUR))', >> comprehensive_fix.py
echo                                 'AND updated_at ^< DATE_SUB(CURRENT_TIMESTAMP, INTERVAL 12 HOUR)') >> comprehensive_fix.py
echo         with open(cleaner_path, 'w', encoding='utf-8') as f: >> comprehensive_fix.py
echo             f.write(content) >> comprehensive_fix.py
echo         print("  - Fixed!") >> comprehensive_fix.py
echo     else: >> comprehensive_fix.py
echo         print("  - Already fixed or different format") >> comprehensive_fix.py
echo. >> comprehensive_fix.py
echo # Fix 2: Ensure all trigger columns are escaped in lead_repository.go >> comprehensive_fix.py
echo print("\n2. Checking lead_repository.go...") >> comprehensive_fix.py
echo lead_repo_path = r'src\repository\lead_repository.go' >> comprehensive_fix.py
echo if os.path.exists(lead_repo_path): >> comprehensive_fix.py
echo     with open(lead_repo_path, 'r', encoding='utf-8') as f: >> comprehensive_fix.py
echo         content = f.read() >> comprehensive_fix.py
echo     # Count current backtick escapes >> comprehensive_fix.py
echo     escaped_count = content.count('`trigger`') >> comprehensive_fix.py
echo     print(f"  - Found {escaped_count} escaped 'trigger' references") >> comprehensive_fix.py
echo     # Check if we need to escape any unescaped ones >> comprehensive_fix.py
echo     # Look for patterns like: trigger, trigger = , etc. >> comprehensive_fix.py
echo     patterns_to_check = [ >> comprehensive_fix.py
echo         (r'(\s+)trigger,', r'\1`trigger`,'),  # column in SELECT >> comprehensive_fix.py
echo         (r'(\s+)trigger\s+=', r'\1`trigger` ='),  # column in WHERE >> comprehensive_fix.py
echo         (r',\s*trigger\s*,', ', `trigger`,'),  # column in middle of list >> comprehensive_fix.py
echo         (r',\s*trigger\s*FROM', ', `trigger` FROM'),  # column before FROM >> comprehensive_fix.py
echo     ] >> comprehensive_fix.py
echo     changes_made = False >> comprehensive_fix.py
echo     for pattern, replacement in patterns_to_check: >> comprehensive_fix.py
echo         if re.search(pattern, content): >> comprehensive_fix.py
echo             content = re.sub(pattern, replacement, content) >> comprehensive_fix.py
echo             changes_made = True >> comprehensive_fix.py
echo     if changes_made: >> comprehensive_fix.py
echo         with open(lead_repo_path, 'w', encoding='utf-8') as f: >> comprehensive_fix.py
echo             f.write(content) >> comprehensive_fix.py
echo         print("  - Fixed unescaped trigger references!") >> comprehensive_fix.py
echo     else: >> comprehensive_fix.py
echo         print("  - All trigger references are properly escaped") >> comprehensive_fix.py
echo. >> comprehensive_fix.py
echo print("\nDone! All fixes have been applied.") >> comprehensive_fix.py

:: Run the Python script
python comprehensive_fix.py

:: Clean up
del comprehensive_fix.py

echo.
echo ========================================
echo Fix completed!
echo.
echo Next steps:
echo 1. Rebuild the application:
echo    go build -o whatsapp.exe
echo.
echo 2. Restart the application:
echo    whatsapp.exe rest
echo.
pause
