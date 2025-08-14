@echo off
REM Deploy device status normalization update

echo ========================================
echo Deploying Device Status Fix
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src"

REM Build without CGO
echo Building application...
set CGO_ENABLED=0
go build -o ../whatsapp.exe .

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo.
echo Build successful!
echo.

cd ..

REM Git operations
echo Adding changes to git...
git add -A

echo.
echo Committing changes...
git commit -m "Fix device status - ensure devices always end up as online or offline, never stuck in refreshing state"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Device Status Fix Deployed!
echo ========================================
echo.
echo What this update does:
echo.
echo 1. Auto-refresh waits 10 seconds then checks status
echo 2. If device is not online/offline, sets to offline
echo 3. Status normalizer runs every 5 minutes
echo 4. Ensures NO device is stuck in "refreshing" state
echo.
echo Device statuses will ALWAYS be:
echo - online (connected and working)
echo - offline (not connected)
echo.
pause
