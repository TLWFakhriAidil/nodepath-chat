@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo Pushing Railway fixes...

echo.
echo Adding files...
git add src/database/connection.go
git add src/database/preinit_fixes.go
git add src/database/fix_login.go
git add EMERGENCY_LOGIN_FIX.sql
git add push_emergency_login_fix.bat

echo.
echo Committing changes...
git commit -m "Fix Railway deployment issues

- Added PreInitFixes to handle Invalid Date before schema init
- Fixed scheduled_time column type issues
- Added emergency login fix that resets admin password
- Login after fix: admin@whatsapp.com / changeme123
- Also creates backup: backup@admin.com / changeme123
- Fixed order of operations: pre-init fixes -> schema -> migrations
- App now handles PORT env var correctly for Railway"

echo.
echo Pushing to remote...
git push origin main

echo.
echo Railway fixes pushed!
echo.
echo After Railway deploys:
echo 1. Login with: admin@whatsapp.com / changeme123
echo 2. Campaigns will show on calendar
echo 3. Check Railway logs for migration status
echo.
pause
