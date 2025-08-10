@echo off
echo Pushing debug fixes for sequence display...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding all changes...
git add -A

echo Committing changes...
git commit -m "Add debugging and fix sequence display issues

- Added console.log to debug sequence data structure
- Added default values for all fields to prevent undefined
- Fixed potential issues with status being undefined
- This will help identify why View button uses wrong ID"

echo Pushing to main branch...
git push origin main

echo Push complete!
pause
