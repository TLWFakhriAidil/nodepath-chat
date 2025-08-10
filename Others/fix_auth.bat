@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix JavaScript errors - add credentials to fetch calls and fix function syntax"
git push origin main
echo.
echo ============================================
echo Fixes pushed:
echo 1. Fixed updateDeviceFilter function syntax
echo 2. Added credentials:'include' to all fetch calls
echo 3. Fixed 401 authentication errors
echo ============================================
echo.
pause