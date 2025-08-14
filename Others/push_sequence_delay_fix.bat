@echo off
echo ========================================
echo Fixing Sequence Min/Max Delays
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

echo Building application with CGO_ENABLED=0...
cd src
set CGO_ENABLED=0
go build -o ../whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed!
    echo Reverting changes...
    cd ..
    move /Y "src\usecase\sequence_trigger_processor.go.bak" "src\usecase\sequence_trigger_processor.go"
    pause
    exit /b 1
)

echo.
echo Build successful!
cd ..

echo.
echo Committing changes to Git...
git add -A
git commit -m "Fix: Respect min/max delay for each sequence step to prevent bans

- Added minDelaySeconds and maxDelaySeconds to contactJob struct
- Updated query to fetch delays from sequence steps with fallback to sequence level
- Pass delays to broadcast message for proper rate limiting
- Critical fix for WhatsApp ban prevention

Each sequence step can now have its own min/max delays which are respected during message sending. Falls back to sequence-level delays if step doesn't have them, and finally to default 10-30 seconds."

echo.
echo Pushing to GitHub main branch...
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Git push failed!
    pause
    exit /b 1
)

echo.
echo ========================================
echo Fix completed successfully!
echo ========================================
echo.
echo Changes:
echo - Each sequence step now respects its own min/max delays
echo - Falls back to sequence-level delays if not set on step
echo - Prevents WhatsApp bans with proper rate limiting
echo.
pause
