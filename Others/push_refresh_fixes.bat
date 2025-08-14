@echo off
echo ========================================
echo Pushing device refresh and connection fixes
echo ========================================

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add all changed files
echo Adding changes...
git add src/ui/rest/check_device_connection.go
git add src/ui/rest/device_multidevice.go
git add src/ui/rest/device_refresh.go
git add src/ui/rest/app.go
git add src/views/dashboard.html

REM Commit
echo Committing changes...
git commit -m "Fix: Device refresh and connection check endpoints

- Fixed CheckDeviceConnectionStatus to be standalone function (was causing 404)
- Added nil check for DeviceManager to prevent panic
- Created new simpler RefreshDevice endpoint that avoids DeviceManager issues
- Updated dashboard to use new refresh endpoint
- Made loadDevices handle missing check-connection endpoint gracefully
- Refresh now checks ClientManager first, returns QR_REQUIRED if not connected"

REM Push to main branch
echo Pushing to main branch...
git push origin main

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✅ Successfully pushed all fixes!
    echo.
    echo Fixed:
    echo - 404 on check-connection endpoint
    echo - 500 nil pointer on device refresh
    echo - Dashboard now handles missing endpoints gracefully
) else (
    echo.
    echo ❌ Push failed!
)

pause
