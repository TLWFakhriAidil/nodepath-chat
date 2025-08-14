@echo off
echo ========================================
echo WhatsApp Web - Direct Image Send Update
echo ========================================

REM Add all changes
git add .

REM Commit with descriptive message
git commit -m "feat: Direct image send without preview modal" -m "- Images now send immediately upon selection" -m "- Removed preview modal entirely to eliminate errors" -m "- Faster, simpler workflow: select -> send -> done" -m "- No more modal issues or refresh needed"

REM Push to main
echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo Push completed successfully!
echo ========================================
echo.
echo New workflow:
echo 1. Click paperclip
echo 2. Select image
echo 3. Image sends automatically
echo 4. Ready for next image immediately!
echo.
pause
