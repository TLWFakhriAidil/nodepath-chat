@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo === Pushing fixes for Campaign Device Report and Worker Status ===
echo.

git add src/ui/rest/app.go
git add src/views/dashboard.html
git commit -m "fix: Campaign Device Report shows real data and Worker Status improvements

- Fixed connected devices count to recognize 'online' status
- Campaign Device Report now shows actual lead data from database
- Added campaign/sequence filtering to Worker Status page
- Worker Status now shows which campaign/sequence is being processed
- Fixed device status badge to show green for both 'connected' and 'online'
- Updated countConnectedDevices function to handle all status variations"

git push origin main

echo.
echo === Fixes pushed successfully! ===
echo.
echo Campaign Device Report will now show:
echo - Real lead names and phone numbers from database
echo - Actual message status (pending/sent/failed)
echo - Real timestamps
echo.
echo Worker Status now includes:
echo - Correct connected devices count
echo - Filter by Campaign or Sequence
echo - Shows current campaign/sequence being processed
echo.
pause
