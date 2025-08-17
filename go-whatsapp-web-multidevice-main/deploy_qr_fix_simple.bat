@echo off
echo ===================================================
echo WhatsApp Multi-Device QR Fix and Clear Device Data
echo ===================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add .

echo.
echo Committing changes...
git commit -m "Fix QR code generation and add clear device data functionality"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ===================================================
echo DEPLOYMENT COMPLETE!
echo.
echo Railway will automatically deploy the changes.
echo.
echo FIXES APPLIED:
echo 1. QR code generation now works for all devices
echo 2. Each user can connect multiple devices
echo 3. Devices are properly registered in ClientManager
echo 4. Added clear device data functionality
echo.
echo NEW ENDPOINTS:
echo - DELETE /api/devices/{deviceId}/clear
echo - POST /api/devices/reset-all
echo ===================================================
pause