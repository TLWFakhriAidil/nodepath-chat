@echo off
echo ========================================
echo Deploying Enhanced Webhook with Auto Device Creation
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add src/ui/rest/webhook_lead.go
git add src/models/lead.go
git add src/models/user.go
git add src/repository/lead_repository.go
git add src/repository/user_repository.go
git add add_platform_columns.sql
git add ENHANCED_WEBHOOK_DOCS.md

echo Committing changes...
git commit -m "feat: Enhanced webhook with auto device creation and platform tracking

- Webhook now checks if device exists before creating lead
- If device doesn't exist, creates it automatically with provided details
- Added platform field to both leads and user_devices tables
- Added device_name field support in webhook request
- CreateDevice method added to UserRepository
- Returns device_created flag in response
- Perfect for external integrations like Whacenter"

echo Pushing to GitHub main branch...
git push origin main

echo ========================================
echo Deployment complete!
echo ========================================
echo.
echo IMPORTANT: Run this SQL migration on Railway PostgreSQL:
echo.
echo ALTER TABLE user_devices ADD COLUMN IF NOT EXISTS platform VARCHAR(255);
echo ALTER TABLE leads ADD COLUMN IF NOT EXISTS platform VARCHAR(255);
echo.
echo Test with enhanced request format:
echo {
echo   "name": "Test Customer",
echo   "phone": "60123456789",
echo   "target_status": "prospect",
echo   "device_id": "test-device-123",
echo   "user_id": "your-user-id",
echo   "device_name": "Test-Device",
echo   "platform": "Whacenter",
echo   "niche": "EXSTART",
echo   "trigger": "NEWNP"
echo }
echo.
pause
