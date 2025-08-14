@echo off
echo ========================================
echo FIXING CAMPAIGN TIME ERROR
echo ========================================

echo.
echo Committing campaign time fix...
git add -A
git commit -m "fix: Handle empty scheduled_time in campaigns

- Changed COALESCE to return '09:00:00' instead of empty string
- Prevents 'invalid input syntax for type time' error
- Default time is now 9:00 AM for campaigns without time"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================
echo FIX DEPLOYED!
echo ========================================
echo.
echo The campaign calendar should now work properly!
echo Empty scheduled_time values will default to 09:00:00
echo.
pause
