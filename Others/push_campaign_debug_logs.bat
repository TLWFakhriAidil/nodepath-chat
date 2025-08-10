@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo Pushing campaign debugging logs...

echo.
echo Adding files...
git add src/repository/campaign_repository.go
git add CAMPAIGN_DEBUG_GUIDE.md
git add check_campaigns_table.sql

echo.
echo Committing changes...
git commit -m "Add logging to debug campaign issues

- Added detailed logging to GetCampaigns to see what's returned
- Added logging to CreateCampaign to track saves
- Created debug guide with SQL queries
- Issues to investigate:
  1. scheduled_time not saving (might be frontend issue)
  2. Campaigns not showing on calendar (might be user_id mismatch)
  3. Calendar might be showing wrong month"

echo.
echo Pushing to remote...
git push origin main

echo.
echo Debug logging added!
echo.
echo Check browser console (F12) when clicking Refresh Campaigns
echo to see what campaigns are loaded.
echo.
pause
