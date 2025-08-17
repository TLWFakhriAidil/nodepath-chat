@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo Committing campaign calendar fix...

echo.
echo Adding files...
git add src/repository/campaign_repository.go
git add src/database/connection.go
git add src/views/dashboard.html

echo.
echo Committing changes...
git commit -m "Fix campaign calendar not showing campaigns

- Fixed SQL query to use campaign_date instead of scheduled_date
- Added COALESCE for device_id to handle NULL values
- Added device_id column to campaigns table schema
- Added debug info display to help troubleshoot issues
- Fixed column name mismatch between database and repository"

echo.
echo Pushing to remote...
git push origin main

echo.
echo Campaign calendar fix committed!
echo.
pause
