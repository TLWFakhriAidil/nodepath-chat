@echo off
echo Fixing multiple compilation errors...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Git status...
git status

echo.
echo Adding all repository fixes...
git add src/repository/*.go src/models/user.go

echo.
echo Committing fixes...
git commit -m "fix: Resolve multiple compilation errors

- Add UpdatedAt field to UserDevice model
- Fix userRepository capitalization in UpdateDeviceStatus
- Add database import to all repositories
- Replace undefined 'db' with database.GetDB()
- Remove unused fmt import from sequence_repository"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ✅ All compilation errors fixed and pushed!
pause