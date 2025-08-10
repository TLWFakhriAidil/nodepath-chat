@echo off
echo ========================================
echo Deploying Duplicate Prevention Fix
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add src/ui/rest/webhook_lead.go
git add src/repository/lead_repository.go

echo Committing changes...
git commit -m "fix: Prevent duplicate leads by checking device_id + user_id + phone

- Added GetLeadByDeviceUserPhone method to check for existing leads
- Webhook now checks if lead exists with same device_id, user_id, and phone
- Returns DUPLICATE_SKIPPED status if lead already exists
- Prevents duplicate leads when webhook is called multiple times
- No more duplicate entries for same phone number"

echo Pushing to GitHub main branch...
git push origin main

echo ========================================
echo Fix deployed!
echo ========================================
echo.
echo Duplicate Prevention Logic:
echo - Checks if lead exists with same: device_id + user_id + phone
echo - If exists: Returns existing lead with duplicate=true flag
echo - If not exists: Creates new lead
echo.
echo This prevents duplicate leads even if webhook is triggered multiple times!
echo.
pause
