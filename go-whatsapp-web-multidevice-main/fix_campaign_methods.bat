@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Fixing missing campaign repository methods ===
echo.

git add src/repository/campaign_repository.go
git add src/usecase/campaign_trigger.go
git commit -m "fix: Add missing campaign repository methods and fix field names

- Added GetCampaigns, UpdateCampaign, DeleteCampaign, GetCampaignsByUser methods
- Fixed ScheduledTime to TimeSchedule in campaign_trigger.go
- Fixed UpdateCampaign call to use UpdateCampaignStatus
- All methods now properly implemented"

git push origin main

echo.
echo === Fix pushed! ===
pause
