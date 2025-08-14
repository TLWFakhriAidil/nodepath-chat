@echo off
echo ========================================
echo Pushing auto-reconnect fix to GitHub
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add all changed files
echo Adding changes...
git add src/infrastructure/whatsapp/multidevice_auto_reconnect.go
git add src/infrastructure/whatsapp/device_manager_init.go
git add src/cmd/rest.go

REM Commit with descriptive message
echo Committing changes...
git commit -m "Fix: Device auto-reconnect after Railway restart

- Added proper DeviceManager initialization with database connection retry
- Fixed nil pointer dereference in CreateDeviceSession
- Ensure DeviceManager is initialized before auto-reconnect starts
- Added database connection verification before initialization
- Fixed reconnection to use UserDevice model from repository
- Auto-reconnect now properly restores device sessions after restart"

REM Push to main branch
echo Pushing to main branch...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully pushed to GitHub!
    echo.
    echo Changes:
    echo - Fixed auto-reconnect nil pointer issue
    echo - Added proper initialization sequence
    echo - Devices will now reconnect after Railway restart
) else (
    echo.
    echo ❌ Push failed! Please check your GitHub credentials.
)

pause
