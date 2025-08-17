@echo off
echo Fixing import cycle error...

cd C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo.
echo Git status...
git status

echo.
echo Adding changes...
git add .

echo.
echo Committing fix...
git commit -m "fix: Resolve import cycle by moving BroadcastMessage to domain layer

- Move BroadcastMessage type to domains/broadcast/types.go
- Update all imports to use domainBroadcast.BroadcastMessage
- Fix circular dependency between broadcast and repository packages"

echo.
echo Pushing to GitHub...
git push origin main

echo.
echo ✅ Import cycle fix pushed successfully!
pause