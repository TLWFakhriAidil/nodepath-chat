@echo off
echo ========================================
echo Deploying Webhook Lead Feature
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add src/ui/rest/webhook_lead.go
git add src/cmd/rest.go
git add src/.env.example
git add WEBHOOK_LEAD_DOCUMENTATION.md
git add test_webhook_examples.md
git add build_webhook_test.bat

echo Committing changes...
git commit -m "feat: Add webhook endpoint for creating leads

- Added POST /webhook/lead/create endpoint
- Simple webhook key authentication via WEBHOOK_LEAD_KEY env var
- Supports creating prospects and customers
- Validates user_id and optional device_id
- Returns lead_id on successful creation
- Comprehensive documentation and examples included"

echo Pushing to GitHub main branch...
git push origin main

echo ========================================
echo Deployment complete!
echo ========================================
echo.
echo Next steps:
echo 1. Set WEBHOOK_LEAD_KEY in Railway environment variables
echo 2. Your webhook will be available at: https://your-app.railway.app/webhook/lead/create
echo 3. Use the examples in test_webhook_examples.md to test
echo.
pause
