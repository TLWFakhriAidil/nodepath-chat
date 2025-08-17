@echo off
echo ========================================
echo Deploying Lead Creation Webhook
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all webhook files to git...
git add src/ui/rest/webhook_lead.go
git add src/cmd/rest.go
git add SIMPLE_WEBHOOK_DOCS.md
git add clean_webhook_test.md
git add README.md

echo Committing changes...
git commit -m "feat: Add lead creation webhook for external integrations

- Added POST /webhook/lead/create endpoint
- No authentication required for easy integration
- Direct field mapping: name, phone, niche, trigger, target_status, device_id, user_id
- Perfect for WhatsApp bot integration
- Updated README with webhook documentation and examples
- Added PHP cURL examples as requested"

echo Pushing to GitHub main branch...
git push origin main

echo ========================================
echo Deployment complete!
echo ========================================
echo.
echo Your webhook URL:
echo https://web-production-b777.up.railway.app/webhook/lead/create
echo.
echo Test example:
echo curl -X POST https://web-production-b777.up.railway.app/webhook/lead/create -H "Content-Type: application/json" -d "{\"name\":\"Test Lead\",\"phone\":\"60123456789\",\"target_status\":\"prospect\",\"device_id\":\"device-id\",\"user_id\":\"user_id\",\"niche\":\"EXSTART\",\"trigger\":\"NEWNP\"}"
echo.
pause
