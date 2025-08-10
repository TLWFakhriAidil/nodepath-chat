@echo off
echo Fixing sequence step min/max delays...
cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM First, let's backup the original file
copy "src\usecase\sequence_trigger_processor.go" "src\usecase\sequence_trigger_processor.go.bak"

echo.
echo Step 1: Updating contactJob struct to include delay fields...
echo Step 2: Updating SQL query to fetch min/max delays from steps...
echo Step 3: Updating scan to read the delays...
echo Step 4: Passing delays to broadcast message...
echo.

REM Navigate to src directory
cd src

REM Build the application
echo Building application...
go build -o ../whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo Build failed!
    pause
    exit /b 1
)

echo Build successful!
cd ..

REM Git operations
echo.
echo Committing changes...
git add -A
git commit -m "Fix: Respect min/max delay for each sequence step to prevent bans

- Add minDelaySeconds and maxDelaySeconds to contactJob struct
- Update query to fetch delays from sequence steps (with fallback to sequence level)
- Pass delays to broadcast message for proper rate limiting
- Critical fix for WhatsApp ban prevention"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo Fix completed and pushed to GitHub!
pause
