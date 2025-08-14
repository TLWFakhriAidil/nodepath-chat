@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main\src

echo Switching to main branch...
git branch -m master main

echo Setting upstream...
git push -u origin main --force

echo.
echo Done! Check GitHub Actions for deployment status.
pause