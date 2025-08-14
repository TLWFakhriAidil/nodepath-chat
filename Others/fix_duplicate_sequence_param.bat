@echo off
echo Fixing duplicate sequence parameter in team dashboard...

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Adding files...
git add -A

echo Committing fix...
git commit -m "Fix duplicate sequence parameter causing syntax error

- Removed duplicate (sequence => from sequences.map() call
- Line 1767 now has correct syntax
- Team dashboard JavaScript loads without errors"

echo Pushing to main branch...
git push origin main

echo Done!
pause