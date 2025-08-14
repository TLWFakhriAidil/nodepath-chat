@echo off
echo Fixing team dashboard authentication and error handling...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing fix...
git commit -m "Fix team dashboard authentication errors and improve error handling

- Added proper 401 authentication error handling to all API calls
- Fixed TypeError issues by checking if responses are arrays before using map/forEach
- Added automatic redirect to team login page on authentication failure
- Improved error handling to prevent JavaScript errors on failed API calls
- All functions now properly validate API responses before processing"

echo Pushing to main branch...
git push origin main

echo Done!
pause