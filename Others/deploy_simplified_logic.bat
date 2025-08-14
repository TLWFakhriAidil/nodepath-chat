@echo off
echo ========================================
echo Deploying Simplified Device ID Logic
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add src/ui/rest/webhook_lead.go
git add WEBHOOK_DEVICE_ID_LOGIC.md

echo Committing changes...
git commit -m "fix: Simplified device ID handling logic

- Non-UUID device_ids: Use first 6 chars prefix, generate new UUID for ID
- Valid UUID device_ids: Use full UUID for both ID and JID
- JID always stores the full original device_id
- Clear and simple logic as requested"

echo Pushing to GitHub main branch...
git push origin main

echo ========================================
echo Deployment complete!
echo ========================================
echo.
echo Device ID Logic:
echo 1. Non-UUID (like hulN3t1y...): ID = new UUID, JID = full original
echo 2. Valid UUID: ID = JID = same UUID
echo.
pause
