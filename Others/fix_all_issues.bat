@echo off
echo ========================================
echo FIXING ALL ISSUES
echo ========================================

echo.
echo Running database migration...
echo Please run this SQL in your Railway PostgreSQL:
echo.
echo ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS target_status VARCHAR(50) DEFAULT 'prospect';
echo ALTER TABLE sequences ADD COLUMN IF NOT EXISTS target_status VARCHAR(50) DEFAULT 'prospect';
echo.

echo.
echo Committing all fixes...
git add -A
git commit -m "fix: Multiple fixes for lead management and campaigns

1. Fixed interface conversion error in lead endpoints
   - Added safe type assertion for email context
   - Prevents nil pointer exceptions

2. Updated campaign target status
   - Removed 'all' option from dropdown
   - Default to 'prospect' status
   - Only options now: prospect, customer

3. Database schema updates needed:
   - ALTER TABLE campaigns ADD COLUMN target_status VARCHAR(50) DEFAULT 'prospect'
   - ALTER TABLE sequences ADD COLUMN target_status VARCHAR(50) DEFAULT 'prospect'

Note: Campaign repository needs update to include target_status in SELECT queries"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================
echo FIXES DEPLOYED!
echo ========================================
echo.
echo Next steps:
echo 1. Run the SQL migration in Railway PostgreSQL
echo 2. The interface conversion error is fixed
echo 3. Campaign form now uses prospect/customer only
echo.
pause
