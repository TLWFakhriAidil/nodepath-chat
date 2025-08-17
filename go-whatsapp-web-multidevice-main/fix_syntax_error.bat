@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Fixing syntax error in campaign_repository.go ===
echo.

git add src/repository/campaign_repository.go
git commit -m "fix: Remove duplicate return statement in campaign_repository.go

- Fixed syntax error caused by duplicate return statement
- Cleaned up file structure"

git push origin main

echo.
echo === Fix pushed! ===
pause
