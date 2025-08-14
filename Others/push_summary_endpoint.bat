@echo off
echo Pushing sequence summary endpoint and debugging...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Add missing sequence summary endpoint and debugging

- Added /api/sequences/summary endpoint that was missing
- Implemented GetSequencesSummary function with proper calculations
- Added logging to debug why sequences aren't showing
- This should fix the sequence summary 404 error"

echo Pushing to main branch...
git push origin main

echo Push complete!
pause
