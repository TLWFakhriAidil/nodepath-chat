@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo. >> README.md
echo ## Deployment Trigger - %date% %time% >> README.md
git add -A
git commit -m "Trigger Railway deployment - fix foreign key constraint issue"
git push origin main
echo.
echo Push completed!
pause
