@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo Adding campaign debugging features...

echo.
echo Adding files...
git add src/views/dashboard.html

echo.
echo Committing changes...
git commit -m "Add debugging for campaign calendar display issue

- Added Refresh Campaigns button to manually reload campaigns
- Added debug info display showing campaign count and dates
- Added console logging for campaign date processing
- Fixed potential duplicate calendar controls issue
- Enhanced error handling with visible error messages"

echo.
echo Pushing to remote...
git push origin main

echo.
echo Campaign debugging features added!
echo.
pause
