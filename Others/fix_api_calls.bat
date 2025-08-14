@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix: Update WhatsApp API calls to match current whatsmeow library version"
git push origin main
echo.
echo API fixes deployed!
pause
