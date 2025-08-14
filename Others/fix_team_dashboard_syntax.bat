@echo off
echo Fixing team dashboard syntax error...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing fix...
git commit -m "Fix syntax error in team_dashboard.html

- Fixed malformed line where }d> should have been proper closing tags
- Corrected displayCampaignSummary function structure
- Team dashboard now loads without JavaScript errors"

echo Pushing to main branch...
git push origin main

echo Done!
pause