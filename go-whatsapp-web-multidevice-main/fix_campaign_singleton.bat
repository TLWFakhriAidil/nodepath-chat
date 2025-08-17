@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Fixing GetCampaignRepository undefined error ===
echo.

git add src/repository/campaign_repository.go
git commit -m "fix: Add GetCampaignRepository singleton function

- Added singleton pattern for campaign repository
- Added sync import for thread safety
- Added database import for DB access
- Fixes undefined GetCampaignRepository errors"

git push origin main

echo.
echo === Fix pushed! ===
pause
