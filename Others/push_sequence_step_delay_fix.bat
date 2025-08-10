@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix: Sequence steps now use their own delay settings" -m "- Updated GetPendingMessages to join with sequence_steps table" -m "- Now uses delay from specific sequence step, not just main sequence" -m "- Priority: Campaign delays > Sequence step delays > Sequence delays > Default (10-30s)" -m "- Each sequence step can have different delay settings"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
