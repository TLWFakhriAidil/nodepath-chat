@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Add device diagnostics endpoint and fix client access"
git push origin main
echo.
echo Diagnostics endpoint deployed!
echo.
echo You can now check:
echo https://your-domain.railway.app/api/devices/{deviceId}/diagnose
echo.
echo This will show:
echo - Device connection status
echo - WhatsApp client status
echo - Number of contacts synced
echo - Number of chats in database
pause
