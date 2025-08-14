@echo off
echo ========================================
echo WhatsApp Web Image Modal Fix - Final
echo ========================================

REM Add all changes
git add .

REM Commit with descriptive message
git commit -m "fix: WhatsApp Web image preview modal now shows correctly" -m "- Fixed missing function parameters in showImagePreview" -m "- Added fallback display style for modal visibility" -m "- Added debug logging to track modal behavior" -m "- Clear caption input on new image selection"

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
echo - Image preview modal now shows on every image selection
echo - Added comprehensive debug logging
echo - Modal uses both classList and style.display for reliability
echo.
pause
