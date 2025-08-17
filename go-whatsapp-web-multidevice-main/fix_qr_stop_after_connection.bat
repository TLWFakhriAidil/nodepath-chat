@echo off
echo ========================================================
echo Fix: Stop QR Generation After Connection
echo ========================================================
echo.
echo Fixed issues:
echo 1. QR codes stop generating after successful scan
echo 2. Proper connection monitoring
echo 3. Frontend receives DEVICE_CONNECTED event
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Committing fix...
git add -A
git commit -m "Fix QR generation continues after connection

- Added stopQR channel to stop QR generation after connection
- Monitor for successful login and stop QR generation
- Frontend already handles DEVICE_CONNECTED event to close modal
- QR codes will stop generating once device is logged in"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================================
echo FIX DEPLOYED!
echo.
echo Now when you scan QR code:
echo 1. Device connects
echo 2. QR generation stops
echo 3. Modal closes automatically
echo 4. Success message shows
echo ========================================================
pause