@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
git add -A
git commit -m "Fix: Added missing GetDevice handler and updated README with read-only WhatsApp Web implementation"
git push origin main
echo.
echo Deployment complete!
pause
