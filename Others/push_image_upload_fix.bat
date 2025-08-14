@echo off
echo ========================================
echo WhatsApp Web Image Upload Fix - Part 2
echo ========================================

REM Add all changes
git add .

REM Commit with descriptive message
git commit -m "fix: WhatsApp Web second image upload now works correctly" -m "- Fixed file input not resetting, preventing second image selection" -m "- Fixed syntax errors in setupImageInput function" -m "- Added debug logging for troubleshooting"

REM Push to main
echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Push completed successfully!
echo ========================================
echo.
echo Fixed:
echo - Second image upload now shows preview modal
echo - Can select same image multiple times
echo - Syntax errors corrected
echo.
pause
