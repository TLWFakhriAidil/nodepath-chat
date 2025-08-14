@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main
echo Improving campaign display on calendar...

echo.
echo Adding files...
git add src/views/dashboard.html

echo.
echo Committing changes...
git commit -m "Improve campaign display on calendar cells

- Show all campaigns on each day (removed 3 campaign limit)
- Made campaign names clickable with hover effect for editing
- Changed delete icon to X circle for better visibility
- Added icons for niche, time, and image indicators
- Improved campaign card styling with background and borders
- Increased calendar day height to 120px for more campaigns
- Enhanced scrollbar styling for campaign container
- Fixed editCampaign function to handle event parameter properly"

echo.
echo Pushing to remote...
git push origin main

echo.
echo Campaign display improvements completed!
echo.
pause
