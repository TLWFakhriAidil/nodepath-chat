@echo off
echo ========================================
echo Pushing Device Filter Complete Fix
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add changes
git add src/views/dashboard.html

REM Commit with proper message
git commit -m "Fix device filter to persist selection and filter all analytics

- Device filter now properly maintains selected value when switching between device names
- Filter applies to Device Analytics, Campaign Analytics, and Sequence Analytics
- When selecting a device name, all analytics sections show data only for those devices
- Fixed issue where filter would revert to 'All Devices' after selection
- Supports filtering by multiple devices with same name (comma-separated IDs)"

REM Push
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
