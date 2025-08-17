@echo off
echo Fixing constant refresh issue in team dashboard...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing fix...
git commit -m "Fix constant refresh issue in team dashboard

- Added missing parentheses to loadDashboardData() function call
- Added proper error handling for dashboard API to prevent auth loops
- Dashboard data endpoint now handles 401 errors without redirecting
- Fixed missing function call parentheses that could cause evaluation loops
- Team dashboard should no longer refresh constantly"

echo Pushing to main branch...
git push origin main

echo Done! The constant refresh should be fixed.
pause