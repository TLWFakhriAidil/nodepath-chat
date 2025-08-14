@echo off
echo ========================================
echo Pushing Frontend Undefined Fix to GitHub
echo ========================================

echo.
echo Changes made:
echo - Fixed frontend default object structure to use 'inactive' instead of 'paused' and 'draft'
echo - Updated filter function to calculate inactive count correctly
echo - This fixes the "undefined" display issue for inactive sequences
echo.

REM Add all changes
git add -A

REM Commit with descriptive message
git commit -m "fix: Frontend undefined display for inactive sequences

- Updated default summary object to use 'inactive' instead of 'paused' and 'draft'
- Fixed filter function to calculate inactive as: total - active
- Added missing properties to default objects (total_success, total_remaining, total_devices)
- This resolves the 'undefined' display issue"

REM Push to main branch
git push origin main

echo.
echo ========================================
echo Push completed!
echo ========================================
pause
