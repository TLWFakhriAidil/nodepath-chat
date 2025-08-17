@echo off
echo Pushing team login with cache bust...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding version comment to force cache refresh...
powershell -Command "(Get-Content src\views\team_login.html) -replace '<!-- Team Member Login -->', '<!-- Team Member Login v2 -->' | Set-Content src\views\team_login.html"

echo Adding files...
git add -A

echo Committing with cache bust...
git commit -m "Force cache refresh for team login

- Added version comment to force browser cache refresh
- Team login should now load the latest version
- All syntax errors have been fixed
- Autocomplete attributes are properly set"

echo Pushing to main branch...
git push origin main

echo Done! Clear your browser cache and try again.
pause