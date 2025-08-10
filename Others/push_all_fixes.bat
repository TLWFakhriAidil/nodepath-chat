@echo off
echo Committing and pushing all fixes to GitHub...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo [1/4] Adding files to git...
git add -A

echo [2/4] Committing changes...
git commit -m "Fix: WebSocket QR filtering - prevent QR popups showing for other users' devices"

echo [3/4] Pushing to main branch...
git push origin main

if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Push failed!
    pause
    exit /b 1
)

echo [4/4] Push successful!
echo.
echo All fixes applied:
echo 1. Image uploads converted to URL inputs (sequences, campaigns, AI campaigns)
echo 2. QR code popups only show for the device being connected
echo 3. No cross-user interference with QR codes
echo.
echo Railway will auto-deploy these changes.
echo.
pause
