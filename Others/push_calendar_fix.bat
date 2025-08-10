@echo off
echo Pushing fix for campaigns not showing on calendar...
echo.

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding changes...
git add src/repository/campaign_repository.go

echo.
echo Committing fix...
git commit -m "Fix campaigns not showing on calendar - remove device_id from query

- Removed device_id from GetCampaigns query and scan
- Campaigns now load correctly based only on user_id
- This fixes the empty results issue in the calendar view"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo Done! The fix has been pushed.
echo Campaigns should now appear on the calendar correctly.
pause
