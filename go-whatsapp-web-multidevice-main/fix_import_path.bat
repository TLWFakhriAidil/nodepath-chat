@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Fixing import path error ===
echo.

git add src/repository/campaign_repository.go
git commit -m "fix: Correct import path in campaign_repository.go

- Changed from 'whatsapp-go/src/models' to 'github.com/aldinokemal/go-whatsapp-web-multidevice/models'
- Fixes Railway build error"

git push origin main

echo.
echo === Fix pushed! ===
pause
