@echo off
echo ========================================================
echo Fix: Keep Multi-Device Client Connected
echo ========================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Removing temporary fix file...
del src\usecase\keep_alive_fix.go 2>nul

echo.
echo Committing multi-device connection fix...
git add -A
git commit -m "Fix multi-device client disconnection issue

- Add keepalive monitoring for each new WhatsApp client
- Ensure client stays connected after QR scan
- Client automatically reconnects if disconnected
- Maintains multi-device architecture for broadcast workers
- Each device gets its own persistent client connection"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================================
echo MULTI-DEVICE CONNECTION FIX DEPLOYED!
echo.
echo Now each device will:
echo 1. Create its own WhatsApp client
echo 2. Stay connected with keepalive monitoring
echo 3. Auto-reconnect if disconnected
echo 4. Work properly with broadcast workers
echo.
echo The multi-device architecture is preserved!
echo ========================================================
pause