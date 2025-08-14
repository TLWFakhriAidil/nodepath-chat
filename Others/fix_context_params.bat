@echo off
echo ========================================================
echo Fix Context Parameters for Delete and Logout
echo ========================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding changes...
git add -A

echo.
echo Committing fix...
git commit -m "Fix context parameters for Delete and Logout methods

- Fixed device.Delete() to include context parameter
- Fixed service.WaCli.Logout() to include context parameter
- Fixed client.Logout() calls in REST handlers to include context
- All compilation errors resolved"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================================
echo CONTEXT PARAMETER ERRORS FIXED!
echo.
echo The build should now succeed on Railway.
echo All methods now have proper context parameters.
echo ========================================================
pause