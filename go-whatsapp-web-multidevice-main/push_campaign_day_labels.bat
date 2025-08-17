@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo Adding campaign day labels to calendar...

echo.
echo Adding files...
git add src/views/dashboard.html

echo.
echo Committing changes...
git commit -m "Add day labels to campaign calendar and summary - Added day names (Mon, Tue, Wed) to calendar cells below date numbers - Updated campaign modal titles to show day name with date - Enhanced campaign summary table to display day names - Added formatCampaignDate helper function for consistent date formatting - Improved calendar cell styling for better day label visibility"

echo.
echo Pushing to remote...
git push origin main

echo.
echo Campaign day labels added successfully!
echo.
pause
