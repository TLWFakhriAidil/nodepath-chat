@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Simplify sequence system: remove device requirement, fix nil pointer, update UI and README"
git push origin main
echo.
echo Push completed!
pause
