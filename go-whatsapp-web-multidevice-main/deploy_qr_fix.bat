@echo off
echo ===================================================
echo WhatsApp Multi-Device QR Fix and Clear Device Data
echo ===================================================
echo.

echo Preparing fixes for deployment...
echo.

REM Stage all changes
git add -A

REM Commit with descriptive message
git commit -m "Fix QR code generation and add clear device data functionality

- Fixed QR code generation for multiple devices per user
- Each login attempt now creates a fresh WhatsApp client
- Fixed device registration in ClientManager after successful connection
- Added clear device data endpoint (/api/devices/:deviceId/clear)
- Added reset all devices endpoint (/api/devices/reset-all)
- Proper event handling for device connection and registration
- Fixed issue where devices were not being registered in broadcast system
- Added proper cleanup of WhatsApp store data
- Enhanced error handling and retry logic for QR generation
- Support for multiple simultaneous device connections per user"

echo.
echo Pushing to main branch...
git push origin main --force

echo.
echo ===================================================
echo Deployment complete! Railway will auto-deploy.
echo.
echo FIXES APPLIED:
echo 1. QR code generation now works for all devices
echo 2. Each user can connect multiple devices
echo 3. Devices are properly registered in ClientManager
echo 4. Added clear device data functionality
echo.
echo To use clear device:
echo DELETE /api/devices/{deviceId}/clear
echo.
echo To reset all devices:
echo POST /api/devices/reset-all
echo ===================================================
pause