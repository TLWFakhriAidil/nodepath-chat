@echo off
echo Committing changes...
git add -A
git commit -m "Fix sequence device report - Fixed step_order column to use COALESCE(day_number, day, 1) - Added error logging for debugging - Dynamic step statistics display in frontend - Hide step statistics for regular campaigns"
echo.
echo Pushing to GitHub...
git push origin main
echo.
echo Done!
pause
