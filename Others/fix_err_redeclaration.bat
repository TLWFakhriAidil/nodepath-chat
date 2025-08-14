@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix: Remove err variable redeclaration in whatsapp_web.go"
git push origin main
echo.
echo Error fix deployed!
pause
