@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo Committing automatic database migration system...

echo.
echo Adding files...
git add src/database/connection.go
git add src/database/migrations.go
git add .env.railway
git add railway_database_fix.sql

echo.
echo Committing changes...
git commit -m "Add automatic database migration system

- Created RunMigrations() function that auto-fixes database issues on startup
- Added AutoFixCampaigns() that runs every 5 minutes to fix Invalid Date
- Migrations track which fixes have been applied to avoid re-running
- System will automatically:
  * Fix Invalid Date in scheduled_time
  * Add missing columns (device_id, etc.)
  * Create proper indexes
  * Fix empty string values in nullable columns
  * Update campaign status
- Added detailed logging to show migration progress
- Created .env.railway for Railway-specific configuration
- Database fixes now run automatically when app starts on Railway"

echo.
echo Pushing to remote...
git push origin main

echo.
echo Automatic database migration system added!
echo.
echo IMPORTANT: When you deploy to Railway, the database will be fixed automatically!
echo.
pause
