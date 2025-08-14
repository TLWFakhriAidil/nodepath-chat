@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix auto-reconnect: Actually attempt to reconnect devices instead of marking them offline"
git push origin main
echo Done!
pause