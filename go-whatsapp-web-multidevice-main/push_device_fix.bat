@echo off
echo ========================================
echo Committing Device Registration Fix
echo ========================================
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "fix: Fix device registration for multidevice support

- Added RegisterDevice method to properly register reconnected devices
- Fixed 'device not connected' error by registering with both managers
- DeviceManager and ClientManager now work together properly
- Resolved rapid connect/disconnect cycle issue

This ensures devices stay properly registered after reconnection."

REM Push to main branch
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo.
    echo Push failed! You may need to pull first.
    pause
    exit /b 1
)

echo.
echo ========================================
echo Successfully pushed device registration fix!
echo ========================================
echo.
pause
