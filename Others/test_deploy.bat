@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "REPLACE dashboard.html completely to test Railway deployment"
git push origin main
echo.
echo Pushed a COMPLETELY DIFFERENT dashboard.html
echo If Railway still shows the old dashboard, it's using a cached build!