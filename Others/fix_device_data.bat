@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix: Use real device data and add chat sync functionality"
git push origin main
echo.
echo Device data fix deployed!
echo.
echo Changes:
echo - GetDevice now returns real device data from database
echo - Added debug logging to chat sync
echo - Added manual sync endpoint /api/devices/:id/sync
echo - Auto-sync chats after device connection
pause
