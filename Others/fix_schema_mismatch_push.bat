@echo off
echo Fixing schema column name mismatch...

REM Build the application
echo Building application...
cd src
go build -o ../whatsapp.exe

if %errorlevel% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo Build successful!
cd ..

REM Commit and push
git add -A
git commit -m "Fix schema column name mismatch: next_send_at -> next_trigger_time

- Updated sequence model to use correct database column name
- Changed db tag from 'next_send_at' to 'next_trigger_time'
- Fixed schema initialization to use next_trigger_time
- Database already has correct structure, just needed code to match"

git push origin main

echo Fix pushed successfully!
pause