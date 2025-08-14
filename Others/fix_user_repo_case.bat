@echo off
echo Fixing userRepository capitalization errors...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Adding changes...
git add src/repository/user_repository.go

echo.
echo Committing fix...
git commit -m "fix: Correct UserRepository receiver type capitalization"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ✅ UserRepository capitalization fixed!
pause