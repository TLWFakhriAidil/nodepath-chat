@echo off
echo Pushing campaign display fixes to GitHub...
echo.

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo.
echo Committing changes...
git commit -m "Fix campaign display and scheduled_time issues

- Fixed campaign labels not showing on calendar days
- Added proper null check for campaigns[dateStr]
- Fixed debug div HTML syntax
- Added more detailed logging for troubleshooting
- Added campaign_date to saveCampaign function
- Created fix_campaign_display.sql for database cleanup
- Created test_calendar.html for testing calendar functionality
- Improved error handling in campaign save/update"

echo.
echo Pushing to main branch...
git push origin main

echo.
echo Done! Campaign display fixes have been pushed to GitHub.
pause
