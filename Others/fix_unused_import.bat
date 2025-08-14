@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Fixing unused import ===
echo.

git add src/repository/campaign_repository.go
git commit -m "fix: Remove unused fmt import in campaign_repository.go"

git push origin main

echo.
echo === Fix pushed! ===
pause
