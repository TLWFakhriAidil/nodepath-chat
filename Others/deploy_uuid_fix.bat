@echo off
echo ========================================
echo Deploying UUID Handling Fix for Webhook
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add src/ui/rest/webhook_lead.go
git add WEBHOOK_UUID_HANDLING.md

echo Committing changes...
git commit -m "fix: Handle non-UUID device_ids in webhook

- Automatically generates UUID for non-UUID device_ids
- Stores full original device_id in JID column
- Uses first 6 chars as prefix for device name
- Prevents 'invalid input syntax for type uuid' errors
- Returns both device_id (UUID) and device_jid (original) in response"

echo Pushing to GitHub main branch...
git push origin main

echo ========================================
echo Fix deployed!
echo ========================================
echo.
echo Now the webhook accepts any device_id format:
echo - Valid UUIDs: Used as-is
echo - Non-UUIDs (like hulN3t1y...): Generate new UUID, store original in JID
echo.
echo Example non-UUID device_id:
echo hulN3t1yRMe2J48NYABNss3BOcYRMY1L9UsxdBLDGvkHlR53pBMYCYW.etRzghNw
echo.
echo Will create:
echo - id: generated-uuid
echo - jid: hulN3t1yRMe2J48NYABNss3BOcYRMY1L9UsxdBLDGvkHlR53pBMYCYW.etRzghNw
echo.
pause
