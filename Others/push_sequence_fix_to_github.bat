@echo off
echo Pushing Sequence Device Report fix to GitHub...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Current directory:
cd

echo.
echo Git status:
git status

echo.
echo Adding modified files...
git add src/ui/rest/app.go

echo.
echo Creating commit...
git commit -m "Fix sequence device report calculation to match summary page logic" -m "- Changed shouldSend calculation to: done + failed + remaining" -m "- Updated query to get remaining_send count separately" -m "- Fixed overall totals to use same logic as summary page" -m "- Added debug logging for troubleshooting" -m "- Ensures consistent numbers between summary and device report"

echo.
echo Pushing to origin/master...
git push origin master

echo.
echo Done! Changes pushed to GitHub.
echo Repository: https://github.com/aqilrvsb/go-whatsapp-web-multidevice
pause