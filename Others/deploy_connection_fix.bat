@echo off
echo Deploying WhatsApp Web fixes to Railway...
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "Fix check-connection endpoint - ensure route is properly registered"

REM Push to main branch (Railway auto-deploys from main)
git push origin main

echo.
echo Deployment triggered! Check Railway dashboard for build status.
echo.
pause
