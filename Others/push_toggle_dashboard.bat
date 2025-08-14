@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add src/views/dashboard.html
git commit -m "Fix dashboard toggleSequence to use new toggle endpoint"
git push origin main
pause
