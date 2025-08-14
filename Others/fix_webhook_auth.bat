@echo off
echo ========================================
echo Fixing Webhook Authentication Issue
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files to git...
git add src/ui/rest/middleware/custom_auth.go

echo Committing changes...
git commit -m "fix: Add webhook endpoint to public routes to bypass authentication

- Added /webhook/lead/create to PublicRoutes list
- Webhook is now accessible without login
- Fixed Postman 401 authentication error"

echo Pushing to GitHub main branch...
git push origin main

echo ========================================
echo Fix deployed!
echo ========================================
echo.
echo The webhook should now work without authentication:
echo https://web-production-b777.up.railway.app/webhook/lead/create
echo.
echo Test in Postman without any authentication headers!
echo.
pause
