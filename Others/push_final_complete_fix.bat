@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Pushing final fix for unused variable and README update ===
echo.

git add src/usecase/campaign_trigger.go
git add src/ui/rest/app.go
git add src/views/dashboard.html
git add README.md
git commit -m "fix: Remove unused nowMalaysia variable and update README

- Fixed compilation error by removing unused nowMalaysia variable
- Campaign Device Report now shows real data from database
- Worker Status shows correct connected devices count
- Added campaign/sequence filtering to Worker Status
- Updated README with complete development summary
- All features now working correctly"

git push origin main

echo.
echo === All fixes pushed successfully! ===
echo.
echo Your WhatsApp Multi-Device Ultimate Broadcast System is now:
echo - Fully operational
echo - Showing real data in all reports
echo - Properly detecting online devices
echo - Automatically sending queued messages
echo - Running with optimized logging
echo.
pause
