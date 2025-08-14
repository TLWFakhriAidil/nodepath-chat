@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix: JavaScript syntax error in dashboard.html" -m "- Removed orphaned code after clearSequenceFilter function" -m "- Fixed duplicate clearSequenceFilter function definition" -m "- Cleaned up leftover code from old filter implementation"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
