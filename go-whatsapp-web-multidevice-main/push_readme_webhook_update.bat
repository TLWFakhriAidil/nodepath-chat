@echo off
echo ========================================
echo Updating README with Webhook Summary
echo ========================================

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding README to git...
git add README.md

echo Committing changes...
git commit -m "docs: Update README with comprehensive webhook documentation

- Added detailed webhook quick start guide
- Added Postman testing instructions
- Added PHP cURL example code
- Added UUID format requirements and examples
- Added common errors and solutions
- Updated with actual Railway URL
- Added note about getting IDs from admin dashboard"

echo Pushing to GitHub main branch...
git push origin main

echo ========================================
echo README updated successfully!
echo ========================================
echo.
echo The webhook documentation now includes:
echo - Complete setup instructions
echo - Postman configuration
echo - PHP code examples
echo - UUID format requirements
echo - Error troubleshooting
echo.
pause
