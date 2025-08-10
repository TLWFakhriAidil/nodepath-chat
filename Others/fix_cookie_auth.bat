@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix: Use cookie-based auth directly instead of locals for WhatsApp Web handlers"
git push origin main
echo.
echo Cookie auth fix deployed!
pause
