@echo off
echo Fixing syntax error on line 935...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing fix...
git commit -m "Fix syntax error on line 935 - duplicate parentheses

- Fixed: loadDashboardData();(); -> loadDashboardData();
- Removed duplicate empty parentheses that caused syntax error
- Team dashboard should now load without JavaScript errors"

echo Pushing to main branch...
git push origin main

echo All syntax errors should be fixed now!
pause