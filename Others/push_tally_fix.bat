@echo off
cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Fix: Sequence Summary totals now properly tally with individual sequences" -m "- Removed separate database query for totals that was returning zeros" -m "- Now calculates totals by summing individual sequence statistics" -m "- This ensures the top boxes match the sum of all sequences in the table" -m "- Example: If sequences show 945 done + 598 failed, totals will show exactly that"

echo Pushing to GitHub...
git push origin main

echo Done!
pause
