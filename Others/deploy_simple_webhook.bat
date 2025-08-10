@echo off
echo ========================================
echo Deploying Simple Webhook Lead Feature
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add src/ui/rest/webhook_lead.go
git add src/cmd/rest.go
git add SIMPLE_WEBHOOK_DOCS.md

echo Committing changes...
git commit -m "feat: Add simple webhook endpoint for creating leads

- Added POST /webhook/lead/create endpoint (no auth required)
- Direct field mapping: name, phone, niche, trigger, target_status, device_id, user_id
- Creates leads in PostgreSQL leads table
- Returns lead_id on successful creation
- Simple JSON in/out format as requested"

echo Pushing to GitHub main branch...
git push origin main

echo ========================================
echo Deployment complete!
echo ========================================
echo.
echo Your webhook will be available at:
echo https://your-app.railway.app/webhook/lead/create
echo.
echo Test with:
echo curl -X POST https://your-app.railway.app/webhook/lead/create -H "Content-Type: application/json" -d "{\"name\":\"John Doe\",\"phone\":\"60123456789\",\"target_status\":\"prospect\",\"device_id\":\"device-id\",\"user_id\":\"user_id\",\"niche\":\"EXSTART\",\"trigger\":\"NEWNP\"}"
echo.
pause
