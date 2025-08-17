@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo Committing comprehensive campaign fixes...

echo.
echo Adding files...
git add src/repository/campaign_repository.go
git add comprehensive_campaign_fix.sql
git add fix_scheduled_time.sql

echo.
echo Committing changes...
git commit -m "Fix campaign display and sync issues

- Added COALESCE for all nullable fields to prevent NULL errors
- Added logging to debug campaign queries
- Fixed device_id handling to support NULL values
- Created SQL scripts to fix Invalid Date in scheduled_time
- Enhanced error handling in GetCampaigns and GetCampaignsByDate
- Campaign triggers should now work properly with fixed data"

echo.
echo Pushing to remote...
git push origin main

echo.
pause
