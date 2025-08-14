@echo off
echo Fixing infinite refresh loop in team dashboard...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing refresh loop fix...
git commit -m "Fix infinite refresh loop in team dashboard

- Fixed missing function call parentheses for loadTeamMemberInfo()
- Fixed incomplete window.location redirect in team_login.html
- Added sessionStorage flag to prevent infinite redirect loops
- Added proper error handling for 401 responses to avoid refresh loops
- Clear redirect flags on successful login and when on login page
- Team dashboard now handles authentication errors gracefully"

echo Pushing to main branch...
git push origin main

echo Done! The refresh loop should be fixed now.
pause