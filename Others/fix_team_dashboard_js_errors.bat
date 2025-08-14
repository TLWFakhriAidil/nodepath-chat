@echo off
echo Fixing team dashboard JavaScript syntax errors...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing fix...
git commit -m "Fix JavaScript syntax errors in team_dashboard.html

- Fixed missing arrow function syntax in sequences.map()
- Added proper template literal backticks
- Fixed formatting and spacing issues
- Team dashboard now loads without any JavaScript errors"

echo Pushing to main branch...
git push origin main

echo Done!
pause