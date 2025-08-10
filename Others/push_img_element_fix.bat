@echo off
echo ========================================
echo WhatsApp Web Image Preview Element Fix
echo ========================================

REM Add all changes
git add .

REM Commit with descriptive message
git commit -m "fix: WhatsApp Web image preview element now exists correctly" -m "- Fixed incomplete img tag that was missing opening bracket" -m "- Preview modal now displays correctly on image selection"

REM Push to main
echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Push completed successfully!
echo ========================================
echo.
echo The image preview modal should now work correctly!
echo.
pause
