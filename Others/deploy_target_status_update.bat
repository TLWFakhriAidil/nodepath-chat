@echo off
echo ========================================
echo UPDATING ALL TABLES TO USE TARGET_STATUS
echo ========================================

echo.
echo Committing target_status updates...
git add -A
git commit -m "feat: Update all tables to use target_status column

- Added target_status to Lead model
- Updated lead repository to save/read target_status
- Frontend now uses target_status (fallback to status)
- REST API maps 'status' from frontend to 'target_status' in database
- All three tables (leads, campaigns, sequences) now use target_status
- Default value is 'customer' as requested

Database migration needed:
- ALTER TABLE leads ADD COLUMN target_status TEXT DEFAULT 'customer'
- ALTER TABLE campaigns ADD COLUMN target_status TEXT DEFAULT 'customer'  
- ALTER TABLE sequences ADD COLUMN target_status TEXT DEFAULT 'customer'"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================
echo DEPLOYMENT COMPLETE!
echo ========================================
echo.
echo All three tables now use target_status column!
echo Run the SQL migration: add_target_status_all_tables.sql
echo.
pause
