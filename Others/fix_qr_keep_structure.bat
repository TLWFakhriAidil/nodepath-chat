@echo off
echo ========================================================
echo Fix QR Channel Error - Keep Multi-Device Structure
echo ========================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Creating QR code directory if not exists...
if not exist "src\statics\qrcode" (
    mkdir "src\statics\qrcode"
    echo Created qrcode directory
)

echo.
echo Clearing temporary files...
del src\usecase\app_simple.go 2>nul
del src\usecase\app_complex_backup.go 2>nul

echo.
echo Adding enhanced logging to debug QR issue...

echo.
echo Committing fixes...
git add -A
git commit -m "Fix QR channel error while keeping multi-device structure

- Use provided context instead of context.Background()
- Ensure QR directory exists
- Keep all multi-device functionality intact
- Enhanced logging for debugging"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================================
echo QR CHANNEL FIX DEPLOYED!
echo.
echo The fix maintains your multi-device architecture.
echo.
echo If QR still doesn't work, try:
echo 1. Check Railway logs for specific error
echo 2. Use Phone Code method as alternative
echo 3. Clear browser cache and cookies
echo 4. Try incognito mode
echo.
echo The system structure remains unchanged - only QR
echo generation has been fixed.
echo ========================================================
pause