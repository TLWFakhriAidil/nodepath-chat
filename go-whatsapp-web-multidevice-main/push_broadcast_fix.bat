@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix unused broadcastManager variable in cmd/rest.go"
git push origin main
echo.
echo Push completed!
pause
