@echo off
echo Fixing team dashboard syntax error on line 929...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing fix...
git commit -m "Fix syntax error in team_dashboard.html line 929

- Removed duplicate parentheses from loadTeamMemberInfo() call
- Fixed: loadTeamMemberInfo();(); -> loadTeamMemberInfo();
- Team dashboard should now load without JavaScript errors"

echo Pushing to main branch...
git push origin main

echo Done!
pause