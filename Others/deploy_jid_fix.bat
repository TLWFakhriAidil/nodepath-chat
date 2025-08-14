@echo off
echo ========================================
echo Updating Webhook - JID = device_id
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add src/ui/rest/webhook_lead.go

echo Committing changes...
git commit -m "fix: Set JID field to device_id when creating device via webhook

- JID column now stores the device_id value
- This ensures device_id is saved in both id and jid columns
- Maintains consistency for device identification"

echo Pushing to GitHub main branch...
git push origin main

echo ========================================
echo Update deployed!
echo ========================================
echo.
echo Device creation now sets:
echo - id = device_id
echo - jid = device_id (same value)
echo.
pause
