@echo off
echo ========================================
echo Fix Webhook Duplicate Key Error
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add src/ui/rest/webhook_lead.go
git add src/repository/user_repository.go

echo Committing changes...
git commit -m "fix: Handle duplicate key constraint in webhook device creation

- Check for existing device by user_id + jid first
- Added GetDeviceByUserAndJID method
- Prevents duplicate key constraint errors
- Uses existing device if user_id + jid combination exists
- Provides helpful hint in error response"

echo Pushing to GitHub main branch...
git push origin main

echo ========================================
echo Fix deployed!
echo ========================================
echo.
echo The webhook now:
echo 1. Checks if device exists by user_id + device_id (as jid)
echo 2. If not found, checks by device_id alone
echo 3. Only creates new device if neither check finds existing device
echo 4. This prevents the duplicate key constraint error
echo.
pause
