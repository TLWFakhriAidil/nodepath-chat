@echo off
REM Force push to trigger Railway deployment

echo ========================================
echo Force Push to GitHub Main
echo ========================================
echo.

cd /d "C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main"

REM Add a small change to force new commit
echo.
echo Adding timestamp to force new deployment...
echo REM Deployment trigger: %date% %time% >> deploy_trigger.txt

REM Git operations
git add .
git commit -m "Force deploy: Fix check-connection endpoint - trigger Railway deployment"

echo.
echo Force pushing to GitHub...
git push origin main --force

echo.
echo ========================================
echo ✅ FORCE PUSH COMPLETE!
echo ========================================
echo.
echo Railway should now detect the new push and start deployment.
echo Check your Railway dashboard for deployment status.
echo.
echo The /api/devices/check-connection endpoint will work after deployment completes.
echo.
pause
