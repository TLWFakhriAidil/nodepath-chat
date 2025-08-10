@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Fixing variable redeclaration error ===
echo.

git add src/usecase/campaign_trigger.go
git commit -m "fix: Fix variable redeclaration error in campaign_trigger.go

- Changed err := to inline if statement to avoid redeclaration
- Uses proper Go idiom for error handling"

git push origin main

echo.
echo === Fix pushed! ===
pause
