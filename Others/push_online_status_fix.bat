@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Pushing ONLINE status fix for device detection ===
echo.

git add src/usecase/campaign_trigger.go
git add src/usecase/optimized_campaign_trigger.go
git add src/usecase/sequence.go
git add README.md
git commit -m "fix: Device status check now includes 'online' status

- Database actually stores 'online'/'offline', not 'connected'
- Updated all device status checks to include online/Online
- Maintains backward compatibility with connected/Connected
- Campaigns and sequences now properly detect online devices
- Updated README with correct status values"

git push origin main

echo.
echo === Fix pushed successfully! ===
echo.
echo Your campaigns should now work properly with devices showing 'online' status.
pause
