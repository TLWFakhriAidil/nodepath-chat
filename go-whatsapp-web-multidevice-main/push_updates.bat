@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Add phone linking functionality and update README with complete status"
git push origin main
echo.
echo ============================================
echo UPDATES PUSHED!
echo ============================================
echo.
echo New features:
echo 1. Added /app/link-device endpoint
echo 2. Added UpdateDevicePhone repository method
echo 3. Updated README with current status and fixes
echo.
echo The dashboard should now be fully functional!
echo.
pause