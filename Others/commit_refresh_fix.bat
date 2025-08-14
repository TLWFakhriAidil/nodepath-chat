@echo off
cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Committing refresh loop fixes...
git add -A
git commit -m "Fix team dashboard infinite refresh loop in all browsers

- Fixed missing function parentheses: loadTeamMemberInfo()
- Fixed incomplete sessionStorage and window.location statements
- Clear redirect flag when successfully on team dashboard
- Prevent multiple redirects with proper auth handling
- Fixed button onclick syntax
- Should work in incognito/private browsing modes"

echo Pushing to GitHub...
git push origin main

echo.
echo Fix complete!
pause
