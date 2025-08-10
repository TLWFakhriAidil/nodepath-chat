@echo off
echo ========================================
echo Pushing Device Filter Selection Fix
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add changes
git add src/views/dashboard.html

REM Commit with proper message
git commit -m "Fix device filter selection persistence and refresh issues

- Removed updateDeviceFilter() call from updateDeviceAnalytics to prevent selection reset
- Fixed dropdown to properly maintain selected device name after data refresh
- Enhanced updateDeviceFilter to save and restore previous selection
- When switching back to 'All Devices', data now refreshes properly
- Filter selection now persists during auto-refresh and manual updates"

REM Push
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
