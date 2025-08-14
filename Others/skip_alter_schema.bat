@echo off
echo Disabling alter schema operations since database is already configured...

REM Build without CGO
echo Building application...
cd src
set CGO_ENABLED=0
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
git commit -m "Skip alter schema operations - prevents conflicting column operations

- Commented out alterSchema section that was trying to ADD and DROP same columns
- Fixes 'column next_trigger_time does not exist' warning
- Database already has correct structure, no alterations needed
- Cleaner startup with no conflicting SQL operations"

git push origin main

echo Schema alteration skip pushed successfully!
pause