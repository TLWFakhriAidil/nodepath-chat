@echo off
echo ========================================
echo Simplified WhatsApp Web Image Upload
echo ========================================

REM Add all changes
git add .

REM Commit with descriptive message
git commit -m "refactor: Simplified image upload - removed unnecessary complexity" -m "- Direct DOM manipulation without checks" -m "- Simple show/hide modal with style.display" -m "- Removed all debug logging and error handling" -m "- Clean, minimal code that just works"

REM Push to main
echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Push completed successfully!
echo ========================================
echo.
echo Image upload is now simple and working!
echo.
pause
