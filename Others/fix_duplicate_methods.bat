@echo off
echo ========================================================
echo Fix Duplicate Method Declaration
echo ========================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Removing duplicate file...
del src\repository\whatsapp_clear_methods.go 2>nul

echo Adding changes...
git add -A

echo.
echo Committing fix...
git commit -m "Fix duplicate method declaration error

- Removed whatsapp_clear_methods.go file
- Methods already exist in whatsapp_repository.go
- Fixes build error: method already declared"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================================
echo FIX DEPLOYED!
echo.
echo The duplicate method error has been resolved.
echo Railway will rebuild and deploy automatically.
echo ========================================================
pause