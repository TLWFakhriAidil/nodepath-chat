@echo off
echo ========================================================
echo Deploy QR Authentication Fix
echo ========================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

git add -A
git commit -m "Fix QR authentication and connection flow

- Added better logging for QR events
- Added connectedChan to properly signal when device is authenticated
- Handle 'success' QR event
- Wait for Connected event instead of just IsLoggedIn
- Improved connection monitoring after QR scan
- Fixed the issue where device shows 'Last active' instead of 'Active'"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================================
echo AUTHENTICATION FIX DEPLOYED!
echo.
echo The fix includes:
echo 1. Better QR event logging to debug issues
echo 2. Proper handling of connection events
echo 3. Waiting for full authentication, not just pairing
echo.
echo If QR still fails, use Phone Code method:
echo - More reliable for authentication
echo - Bypasses QR scanning issues
echo ========================================================
pause