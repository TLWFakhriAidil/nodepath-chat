@echo off
echo Fixing all JavaScript syntax errors in team dashboard...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing comprehensive fix...
git commit -m "Fix all JavaScript syntax errors in team_dashboard.html

- Fixed broken HTML table rows in displayCampaignSummary function
- Added missing closing tags for campaign table
- Added proper spacing between functions
- Fixed template literal syntax issues
- Ensured all JavaScript functions are properly closed
- Team dashboard now loads without any syntax errors"

echo Pushing to main branch...
git push origin main

echo Done! All syntax errors should be fixed now.
pause