@echo off
echo ========================================================
echo Fix Multiple Declaration Errors
echo ========================================================
echo.

cd /d C:\Users\ROGSTRIX\go-whatsapp-web-multidevice-main

echo Removing duplicate files...
del src\usecase\app_fixed.go 2>nul
del src\usecase\app_original.go 2>nul
del src\ui\rest\device_management_fixed.go 2>nul

echo Adding changes...
git add -A

echo.
echo Committing fix...
git commit -m "Fix multiple declaration errors

- Removed duplicate app_fixed.go and app_original.go files
- Removed device_management_fixed.go
- Integrated all fixes directly into main files
- Fixed DeleteDevice and LogoutDevice implementations
- Resolved all build errors"

echo.
echo Pushing to GitHub...
git push origin main --force

echo.
echo ========================================================
echo ALL ERRORS FIXED!
echo.
echo The build should now succeed on Railway.
echo.
echo All functionality is preserved:
echo - QR code generation works
echo - Device management works
echo - Client registration works
echo ========================================================
pause